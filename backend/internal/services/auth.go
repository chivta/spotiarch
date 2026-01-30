package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/chivta/spotiarch/internal/logger"
	"github.com/chivta/spotiarch/internal/models"
)

func NewAuthService(db *gorm.DB, log *logger.Logger, secret string) *AuthService {
	return &AuthService{
		db:     db,
		log:    log,
		secret: secret,
	}
}

type AuthService struct {
	log    *logger.Logger
	db     *gorm.DB
	secret string
}

func (s *AuthService) SignUp(email, password string) (string, error) {
	s.log.Infof("Creating new user: %s", email)

	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return "", ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.log.Errorf("Error checking user existence: %v", err)
		return "", ErrInternal
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return "", ErrInternal
	}

	// Create user
	user := models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPassword),
	}
	if err := s.db.Create(&user).Error; err != nil {
		s.log.Errorf("Failed to create user: %v", err)
		return "", ErrInternal
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID)
	if err != nil {
		s.log.Errorf("Failed to generate token: %v", err)
		return "", err
	}

	s.log.Infof("Successfully created user: %s", email)
	return token, nil
}

func (s *AuthService) ParseToken(token string) (*models.JWTPayloadDTO, error) {
	payload := &models.JWTPayloadDTO{}
	parsedToken, err := jwt.ParseWithClaims(token, payload, func(t *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			s.log.Debugf("Unexpected signing method: %v", t.Header["alg"])
			return nil, ErrTokenInvalid
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		s.log.Errorf("Token validation failed: %v", err)
		// Check if token is expired
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !parsedToken.Valid {
		return nil, ErrTokenInvalid
	}
	return payload, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	s.log.Infof("User login attempt: %s", email)

	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Infof("User not found: %s", email)
			return "", ErrInvalidCredentials
		}
		s.log.Errorf("Error retrieving user: %v", err)
		return "", ErrInternal
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.log.Infof("Invalid password for user: %s", email)
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID)
	if err != nil {
		s.log.Errorf("Failed to generate token: %v", err)
		return "", ErrInternal
	}

	s.log.Infof("User logged in successfully: %s", email)
	return token, nil
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := models.JWTPayloadDTO{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	s.log.Infof("Retrieving user by ID: %s", userID)

	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		s.log.Errorf("Error retrieving user: %v", err)
		return nil, ErrInternal
	}
	return &user, nil
}
