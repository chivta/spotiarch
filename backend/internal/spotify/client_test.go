package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// tokenServer serves canned token responses and records the refresh_token sent
// with each request.
func tokenServer(t *testing.T, bodies ...string) (*httptest.Server, *[]string) {
	t.Helper()
	var sent []string
	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}
		sent = append(sent, r.Form.Get("refresh_token"))

		body := bodies[min(call, len(bodies)-1)]
		call++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server, &sent
}

func newTestClient(tokenURL string) *Client {
	client := NewClient("id", "secret", "original-refresh")
	client.tokenURL = tokenURL
	return client
}

func TestGetValidTokenAdoptsRotatedRefreshToken(t *testing.T) {
	// expires_in 0 forces a refresh on the second call rather than reusing the cache
	server, sent := tokenServer(t,
		`{"access_token":"access-1","expires_in":0,"refresh_token":"rotated-refresh"}`,
		`{"access_token":"access-2","expires_in":3600}`,
	)
	client := newTestClient(server.URL)

	if _, err := client.getValidToken(context.Background()); err != nil {
		t.Fatalf("first getValidToken: %v", err)
	}
	if _, err := client.getValidToken(context.Background()); err != nil {
		t.Fatalf("second getValidToken: %v", err)
	}

	want := []string{"original-refresh", "rotated-refresh"}
	if len(*sent) != len(want) {
		t.Fatalf("got %d token requests, want %d", len(*sent), len(want))
	}
	for i, w := range want {
		if (*sent)[i] != w {
			t.Errorf("request %d used refresh_token %q, want %q", i, (*sent)[i], w)
		}
	}
}

func TestGetValidTokenKeepsRefreshTokenWhenNotRotated(t *testing.T) {
	server, sent := tokenServer(t, `{"access_token":"access","expires_in":0}`)
	client := newTestClient(server.URL)

	for i := range 2 {
		if _, err := client.getValidToken(context.Background()); err != nil {
			t.Fatalf("getValidToken %d: %v", i, err)
		}
	}

	for i, got := range *sent {
		if got != "original-refresh" {
			t.Errorf("request %d used refresh_token %q, want the original", i, got)
		}
	}
}

func TestGetValidTokenCachesUntilExpiry(t *testing.T) {
	server, sent := tokenServer(t, `{"access_token":"access","expires_in":3600}`)
	client := newTestClient(server.URL)

	for range 3 {
		token, err := client.getValidToken(context.Background())
		if err != nil {
			t.Fatalf("getValidToken: %v", err)
		}
		if token != "access" {
			t.Fatalf("got token %q, want %q", token, "access")
		}
	}

	if len(*sent) != 1 {
		t.Errorf("made %d token requests, want 1 (the rest should be cached)", len(*sent))
	}
}

func TestGetValidTokenReturnsAPIErrorOnFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(server.URL).getValidToken(context.Background())
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("got error %T (%v), want *APIError", err, err)
	}
	if apiErr.Status != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", apiErr.Status, http.StatusBadRequest)
	}
}
