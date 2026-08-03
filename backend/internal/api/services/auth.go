package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/chivta/spotiarch/internal/shared/domain"
)

type (
	sessionTokenRepo interface {
		StoreRefreshTokenHash(ctx context.Context, userID int, refreshTokenHash string, expiresAt time.Time) error
		GetRefreshTokenByUserID(ctx context.Context, userID int) (string, time.Time, error)
		DeleteRefreshTokenHash(ctx context.Context, userID int) error
		IncrementAnonRequestCounter(ctx context.Context, anonID, path string, expiration time.Duration) (int, error)
	}
	userRepo interface {
		GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
		GetUserByID(ctx context.Context, id int) (*domain.User, error)
		CreateUser(ctx context.Context, user *domain.User) (int, error)
	}
	pendingClaimRepo interface {
		Claim(ctx context.Context, anonID string, userID int) error
	}
)

func NewAuthService(jwtSecret []byte, tokenRepo sessionTokenRepo, userRepo userRepo, pendingRepo pendingClaimRepo) *AuthService {
	return &AuthService{jwtSecret: jwtSecret, tokenRepo: tokenRepo, userRepo: userRepo, pendingRepo: pendingRepo}
}

type AuthService struct {
	tokenRepo   sessionTokenRepo
	userRepo    userRepo
	pendingRepo pendingClaimRepo
	jwtSecret   []byte
}

type JWTClaims struct {
	UserID string      `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type Session struct {
	JWT          string
	RefreshToken string
	UserID       string
	Role         domain.Role
}

type AnonymousSession struct {
	JWT    string
	UserID string
	Role   domain.Role
}

func (s *AuthService) Login(ctx context.Context, loginDTO domain.LoginDTO, anonID string) (*Session, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, loginDTO.Email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginDTO.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	session, err := s.createSession(ctx, user)
	if err != nil {
		return nil, err
	}
	s.claimPending(ctx, anonID, user.ID)
	return session, nil
}

func (s *AuthService) Signup(ctx context.Context, signupDTO domain.SignupDTO, anonID string) (*Session, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signupDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := domain.User{Email: signupDTO.Email, PasswordHash: string(hashedPassword), Role: domain.RoleUser}
	user.ID, err = s.userRepo.CreateUser(ctx, &user)
	if err != nil {
		return nil, err
	}

	session, err := s.createSession(ctx, &user)
	if err != nil {
		return nil, err
	}
	s.claimPending(ctx, anonID, user.ID)
	return session, nil
}

func (s *AuthService) claimPending(ctx context.Context, anonID string, userID int) {
	if anonID == "" {
		return
	}
	if err := s.pendingRepo.Claim(ctx, anonID, userID); err != nil {
		log.Error().Err(err).Str("anon_id", anonID).Int("user_id", userID).Msg("failed to claim pending selection")
	}
}

func (s *AuthService) createSession(ctx context.Context, user *domain.User) (*Session, error) {
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshExpiresAt := time.Now().Add(time.Duration(domain.RefreshTokenDuration) * time.Second)
	if err := s.tokenRepo.StoreRefreshTokenHash(ctx, user.ID, hashRefreshToken(refreshToken), refreshExpiresAt); err != nil {
		return nil, err
	}

	claims := JWTClaims{
		UserID: strconv.Itoa(user.ID),
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(domain.JWTDuration) * time.Second)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrUnauthorized, err)
	}
	return &Session{JWT: signed, RefreshToken: refreshToken, UserID: claims.UserID, Role: user.Role}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int) error {
	return s.tokenRepo.DeleteRefreshTokenHash(ctx, userID)
}

func (s *AuthService) ParseJWT(jwtStr string) (JWTClaims, error) {
	var claims JWTClaims
	_, err := jwt.ParseWithClaims(jwtStr, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return claims, fmt.Errorf("%w: %w", domain.ErrUnauthorized, err)
	}
	return claims, nil
}

func (s *AuthService) CreateAnonymousSession(context.Context) (*AnonymousSession, error) {
	anonID, err := generateAnonUserID()
	if err != nil {
		return nil, domain.ErrInternal
	}
	claims := JWTClaims{
		UserID: anonID,
		Role:   domain.RoleAnon,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(domain.AnonSessionDuration) * time.Second)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrUnauthorized, err)
	}
	return &AnonymousSession{JWT: signed, UserID: anonID, Role: domain.RoleAnon}, nil
}

func (s *AuthService) IncrementAnonQuota(ctx context.Context, anonID, path string) (int, error) {
	return s.tokenRepo.IncrementAnonRequestCounter(ctx, anonID, path, time.Duration(domain.AnonSessionDuration)*time.Second)
}

func (s *AuthService) ExchangeRefreshToken(ctx context.Context, expiredJWTStr, refreshToken string) (*Session, error) {
	var claims JWTClaims
	_, err := jwt.NewParser(jwt.WithoutClaimsValidation()).ParseWithClaims(expiredJWTStr, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil || claims.Role != domain.RoleUser {
		return nil, domain.ErrUnauthorized
	}

	userID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	storedHash, expiresAt, err := s.tokenRepo.GetRefreshTokenByUserID(ctx, userID)
	if err != nil || subtle.ConstantTimeCompare([]byte(storedHash), []byte(hashRefreshToken(refreshToken))) != 1 || !time.Now().Before(expiresAt) {
		return nil, domain.ErrUnauthorized
	}
	if err := s.tokenRepo.DeleteRefreshTokenHash(ctx, userID); err != nil {
		return nil, domain.ErrUnauthorized
	}
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	return s.createSession(ctx, user)
}

func generateAnonUserID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func generateRefreshToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
