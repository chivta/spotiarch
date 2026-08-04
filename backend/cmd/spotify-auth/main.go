// Command spotify-auth mints the service account's long-lived refresh token.
//
// This is a one-off developer tool, not part of the running system. Spotify has
// no "app-owned playlist" concept, so creating and mutating archive playlists
// requires a token that acts as a real user — the service account. Run this
// once, authorise as that account, and the token it writes is what the api and
// watcher use forever after.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const (
	authorizeURL = "https://accounts.spotify.com/authorize"
	tokenURL     = "https://accounts.spotify.com/api/token"

	// creating and editing private archive playlists is all the service account
	// ever needs to do
	defaultScope = "playlist-modify-private"
	// loopback literals are the only http redirect URIs Spotify still accepts;
	// "localhost" is rejected on apps registered after the 2025 migration
	defaultRedirect = "http://127.0.0.1:8888/callback"

	refreshTokenKey = "SPOTIFY_REFRESH_TOKEN"
	envFileMode     = 0o600
	callbackTimeout = 5 * time.Minute
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	redirect := flag.String("redirect", defaultRedirect,
		"redirect URI, must match one registered on the Spotify app exactly")
	scope := flag.String("scope", defaultScope, "space separated OAuth scopes")
	envPath := flag.String("env", "./.env", "env file to read credentials from and write the token to")
	flag.Parse()

	_ = godotenv.Load(*envPath)
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET must be set in %s or the environment", *envPath)
	}

	redirectURL, err := url.Parse(*redirect)
	if err != nil {
		return fmt.Errorf("invalid redirect URI: %w", err)
	}
	listenAddr := redirectURL.Host
	if redirectURL.Port() == "" {
		return fmt.Errorf("redirect URI %q needs an explicit port", *redirect)
	}

	state, err := randomHex()
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", listenAddr, err)
	}
	defer listener.Close()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	server := &http.Server{Handler: callbackHandler(redirectURL.Path, state, codeCh, errCh)}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	query := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {*redirect},
		"scope":         {*scope},
		"state":         {state},
		// force the account chooser even when the browser already has a session,
		// so the service account is used rather than whoever is logged in
		"show_dialog": {"true"},
	}

	fmt.Println("Open this URL and approve as the SERVICE ACCOUNT (not your personal account):")
	fmt.Println()
	fmt.Println("  " + authorizeURL + "?" + query.Encode())
	fmt.Println()
	fmt.Printf("Waiting for the callback on %s ...\n", *redirect)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return err
	case <-time.After(callbackTimeout):
		return errors.New("timed out waiting for the callback")
	case <-ctx.Done():
		return errors.New("cancelled")
	}

	refreshToken, err := exchangeCode(ctx, clientID, clientSecret, code, *redirect)
	if err != nil {
		return err
	}

	if err := writeEnvValue(*envPath, refreshTokenKey, refreshToken); err != nil {
		return err
	}

	// Confirm which account was actually authorised — the most common mistake
	// here is approving as a personal account.
	if name, err := whoAmI(ctx, clientID, clientSecret, refreshToken); err != nil {
		fmt.Printf("\nWrote %s to %s (could not confirm the account: %v)\n", refreshTokenKey, *envPath, err)
	} else {
		fmt.Printf("\nWrote %s to %s\nAuthorised as: %s\n", refreshTokenKey, *envPath, name)
	}
	fmt.Println("The token is not printed here on purpose; it is a long-lived credential.")
	return nil
}

func callbackHandler(path, state string, codeCh chan<- string, errCh chan<- error) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if authErr := query.Get("error"); authErr != "" {
			http.Error(w, "authorisation denied: "+authErr, http.StatusBadRequest)
			errCh <- fmt.Errorf("authorisation denied: %s", authErr)
			return
		}
		if query.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- errors.New("state mismatch, aborting")
			return
		}
		code := query.Get("code")
		if code == "" {
			http.Error(w, "no code in callback", http.StatusBadRequest)
			errCh <- errors.New("no code in callback")
			return
		}
		fmt.Fprintln(w, "spotiarch: authorised. You can close this tab.")
		codeCh <- code
	})
	return mux
}

func exchangeCode(ctx context.Context, clientID, clientSecret, code, redirect string) (string, error) {
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirect},
	}
	var payload struct {
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := postToken(ctx, clientID, clientSecret, form, &payload); err != nil {
		return "", err
	}
	if payload.Error != "" {
		return "", fmt.Errorf("token exchange failed: %s: %s", payload.Error, payload.ErrorDesc)
	}
	if payload.RefreshToken == "" {
		return "", errors.New("token exchange returned no refresh_token")
	}
	if payload.Scope != "" {
		fmt.Printf("Granted scopes: %s\n", payload.Scope)
	}
	return payload.RefreshToken, nil
}

func whoAmI(ctx context.Context, clientID, clientSecret, refreshToken string) (string, error) {
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}}
	if err := postToken(ctx, clientID, clientSecret, form, &payload); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+payload.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET /me returned %d", resp.StatusCode)
	}

	var me struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s (id %s)", me.DisplayName, me.ID), nil
}

func postToken(ctx context.Context, clientID, clientSecret string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

// writeEnvValue replaces key in the env file, or appends it when absent, leaving
// every other line untouched.
func writeEnvValue(path, key, value string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	lines := []string{}
	if len(existing) > 0 {
		lines = strings.Split(strings.TrimRight(string(existing), "\n"), "\n")
	}

	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			lines[i] = key + "=" + value
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, key+"="+value)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), envFileMode)
}

func randomHex() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
