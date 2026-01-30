package models

import "github.com/golang-jwt/jwt/v5"

type SignUpDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JWTPayloadDTO struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
