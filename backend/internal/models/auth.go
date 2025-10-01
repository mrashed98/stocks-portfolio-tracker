package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// LoginRequest represents the request to login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents the request to register a new user
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=255"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
	jwt.RegisteredClaims
}

// RefreshTokenRequest represents the request to refresh a token
type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// SessionData represents user session data stored in Redis
type SessionData struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}