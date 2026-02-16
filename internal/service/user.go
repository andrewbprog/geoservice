package service

import (
	"context"
	"errors"
	"fmt"

	"geoservice/internal/entity"
	"geoservice/internal/repository"

	"github.com/go-playground/validator/v10"
)

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, req entity.CreateUserRequest) (*entity.UserResponse, error)
	GetUser(ctx context.Context, id string) (*entity.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req entity.UpdateUserRequest) (*entity.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, conditions entity.Conditions) (*entity.UserListResponse, error)
}

// userService implements UserService interface
type userService struct {
	userRepo  repository.UserRepository
	validator *validator.Validate
}

// NewUserService creates a new user service instance
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo:  userRepo,
		validator: validator.New(),
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req entity.CreateUserRequest) (*entity.UserResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create user model
	user := &entity.User{}
	user.FromCreateRequest(req)

	// Save to repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert to response
	response := user.ToResponse()
	return &response, nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id string) (*entity.UserResponse, error) {
	if id == "" {
		return nil, errors.New("user ID is required")
	}

	// Get from repository
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert to response
	response := user.ToResponse()
	return &response, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, id string, req entity.UpdateUserRequest) (*entity.UserResponse, error) {
	if id == "" {
		return nil, errors.New("user ID is required")
	}

	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update user fields
	user.UpdateFromRequest(req)

	// Save to repository
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Convert to response
	response := user.ToResponse()
	return &response, nil
}

// DeleteUser deletes a user (soft delete)
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("user ID is required")
	}

	// Delete from repository
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves a list of users with pagination and filtering
func (s *userService) ListUsers(ctx context.Context, conditions entity.Conditions) (*entity.UserListResponse, error) {
	// Set default values
	if conditions.Limit == 0 {
		conditions.Limit = 10 // Default limit
	}
	if conditions.Limit > 100 {
		conditions.Limit = 100 // Maximum limit
	}

	// Validate conditions
	if err := s.validator.Struct(conditions); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get from repository
	users, total, err := s.userRepo.List(ctx, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to response
	userResponses := make([]entity.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	response := &entity.UserListResponse{
		Users:  userResponses,
		Total:  total,
		Limit:  conditions.Limit,
		Offset: conditions.Offset,
	}

	return response, nil
}
