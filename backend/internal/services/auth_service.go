package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"portfolio-app/internal/models"
	"portfolio-app/internal/repositories"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionNotFound    = errors.New("session not found")
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo      repositories.UserRepository
	redisClient   *redis.Client
	jwtSecret     []byte
	tokenDuration time.Duration
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repositories.UserRepository, redisClient *redis.Client, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		redisClient:   redisClient,
		jwtSecret:     []byte(jwtSecret),
		tokenDuration: 24 * time.Hour, // 24 hours
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{}
	createReq := &models.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password, // This won't be used in FromCreateRequest
	}
	user.FromCreateRequest(createReq, hashedPassword)

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.generateToken(createdUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store session in Redis
	if err := s.storeSession(ctx, createdUser, token); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &models.AuthResponse{
		User:  createdUser.ToResponse(),
		Token: token,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !s.verifyPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store session in Redis
	if err := s.storeSession(ctx, user, token); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &models.AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// ValidateToken validates a JWT token and returns the user claims
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetSession retrieves session data from Redis
func (s *AuthService) GetSession(ctx context.Context, userID uuid.UUID) (*models.SessionData, error) {
	sessionKey := s.getSessionKey(userID)
	sessionJSON, err := s.redisClient.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session models.SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(ctx, userID)
		return nil, ErrSessionNotFound
	}

	return &session, nil
}

// DeleteSession removes a session from Redis
func (s *AuthService) DeleteSession(ctx context.Context, userID uuid.UUID) error {
	sessionKey := s.getSessionKey(userID)
	return s.redisClient.Del(ctx, sessionKey).Err()
}

// RefreshToken generates a new token for an existing session
func (s *AuthService) RefreshToken(ctx context.Context, tokenString string) (*models.AuthResponse, error) {
	// Validate current token
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if session exists
	session, err := s.GetSession(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Get user data
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate new token
	newToken, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %w", err)
	}

	// Update session with new token
	session.ExpiresAt = time.Now().Add(s.tokenDuration)
	if err := s.storeSessionData(ctx, user.ID, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &models.AuthResponse{
		User:  user.ToResponse(),
		Token: newToken,
	}, nil
}

// hashPassword hashes a password using bcrypt
func (s *AuthService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against its hash
func (s *AuthService) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateToken creates a JWT token for a user
func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := &models.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "portfolio-app",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// storeSession stores session data in Redis
func (s *AuthService) storeSession(ctx context.Context, user *models.User, token string) error {
	session := &models.SessionData{
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.tokenDuration),
	}

	return s.storeSessionData(ctx, user.ID, session)
}

// storeSessionData stores session data in Redis
func (s *AuthService) storeSessionData(ctx context.Context, userID uuid.UUID, session *models.SessionData) error {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionKey := s.getSessionKey(userID)
	return s.redisClient.Set(ctx, sessionKey, sessionJSON, s.tokenDuration).Err()
}

// getSessionKey generates a Redis key for a user session
func (s *AuthService) getSessionKey(userID uuid.UUID) string {
	return fmt.Sprintf("session:%s", userID.String())
}

// generateSecureToken generates a cryptographically secure random token
func (s *AuthService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}