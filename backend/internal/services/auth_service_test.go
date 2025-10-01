package services

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"portfolio-app/internal/models"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func setupAuthServiceTest() (*AuthService, *MockUserRepository, *redis.Client, *miniredis.Miniredis) {
	mockRepo := &MockUserRepository{}
	
	// Use miniredis for testing to avoid requiring a real Redis instance
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	
	authService := NewAuthService(mockRepo, redisClient, "test-secret-key")
	return authService, mockRepo, redisClient, mr
}

func TestAuthService_Register(t *testing.T) {
	authService, mockRepo, redisClient, mr := setupAuthServiceTest()
	defer redisClient.Close()
	defer mr.Close()
	
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		req := &models.RegisterRequest{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "password123",
		}

		// Mock that user doesn't exist
		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, nil).Once()

		// Mock successful user creation
		expectedUser := &models.User{
			ID:           uuid.New(),
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: "hashed_password",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(expectedUser, nil).Once()

		result, err := authService.Register(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedUser.Name, result.User.Name)
		assert.Equal(t, expectedUser.Email, result.User.Email)
		assert.NotEmpty(t, result.Token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		req := &models.RegisterRequest{
			Name:     "John Doe",
			Email:    "existing@example.com",
			Password: "password123",
		}

		existingUser := &models.User{
			ID:    uuid.New(),
			Email: req.Email,
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil).Once()

		result, err := authService.Register(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrUserAlreadyExists, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	authService, mockRepo, redisClient, mr := setupAuthServiceTest()
	defer redisClient.Close()
	defer mr.Close()
	
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		req := &models.LoginRequest{
			Email:    "john@example.com",
			Password: password,
		}

		user := &models.User{
			ID:           uuid.New(),
			Name:         "John Doe",
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(user, nil).Once()

		result, err := authService.Login(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.Name, result.User.Name)
		assert.Equal(t, user.Email, result.User.Email)
		assert.NotEmpty(t, result.Token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials - user not found", func(t *testing.T) {
		req := &models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, nil).Once()

		result, err := authService.Login(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials - wrong password", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

		req := &models.LoginRequest{
			Email:    "john@example.com",
			Password: "wrong_password",
		}

		user := &models.User{
			ID:           uuid.New(),
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
		}

		mockRepo.On("GetByEmail", ctx, req.Email).Return(user, nil).Once()

		result, err := authService.Login(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	authService, _, redisClient, mr := setupAuthServiceTest()
	defer redisClient.Close()
	defer mr.Close()

	t.Run("valid token", func(t *testing.T) {
		user := &models.User{
			ID:    uuid.New(),
			Name:  "John Doe",
			Email: "john@example.com",
		}

		token, err := authService.generateToken(user)
		assert.NoError(t, err)

		claims, err := authService.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Name, claims.Name)
	})

	t.Run("invalid token", func(t *testing.T) {
		claims, err := authService.ValidateToken("invalid_token")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
		assert.Nil(t, claims)
	})
}

func TestAuthService_PasswordHashing(t *testing.T) {
	authService, _, redisClient, mr := setupAuthServiceTest()
	defer redisClient.Close()
	defer mr.Close()

	password := "test_password_123"

	t.Run("hash and verify password", func(t *testing.T) {
		hash, err := authService.hashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)

		isValid := authService.verifyPassword(password, hash)
		assert.True(t, isValid)

		isInvalid := authService.verifyPassword("wrong_password", hash)
		assert.False(t, isInvalid)
	})
}