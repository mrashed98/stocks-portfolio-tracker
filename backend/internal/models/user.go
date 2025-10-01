package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	Email        string    `json:"email" db:"email" validate:"required,email,max=255"`
	PasswordHash string    `json:"-" db:"password_hash" validate:"required"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=255"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Email *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// FromCreateRequest creates a User from CreateUserRequest
func (u *User) FromCreateRequest(req *CreateUserRequest, passwordHash string) {
	u.ID = uuid.New()
	u.Name = req.Name
	u.Email = req.Email
	u.PasswordHash = passwordHash
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}