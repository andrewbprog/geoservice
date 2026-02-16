package repository

import (
	"context"
	"errors"
	"fmt"
	"geoservice/pkg/metrics"
	"time"

	"geoservice/internal/entity"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, conditions entity.Conditions) ([]entity.User, int64, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	start := time.Now()
	defer func() {
		metrics.DBDuration.WithLabelValues("CreateUser").Observe(time.Since(start).Seconds())
	}()

	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Check if user with email already exists
	var existingUser entity.User
	err := r.db.WithContext(ctx).Where("email = ? AND is_deleted = ?", user.Email, false).First(&existingUser).Error
	if err == nil {
		return errors.New("user with this email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	// Create the user
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	start := time.Now()
	defer func() {
		metrics.DBDuration.WithLabelValues("GetUserById").Observe(time.Since(start).Seconds())
	}()

	if id == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = ?", id, false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	start := time.Now()
	defer func() {
		metrics.DBDuration.WithLabelValues("UpdateUser").Observe(time.Since(start).Seconds())
	}()

	if user == nil {
		return errors.New("user cannot be nil")
	}

	if user.ID == "" {
		return errors.New("user ID cannot be empty")
	}

	// Check if user exists and is not deleted
	var existingUser entity.User
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = ?", user.ID, false).First(&existingUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	// Check if email is being changed and if the new email already exists
	if user.Email != existingUser.Email {
		var emailUser entity.User
		err := r.db.WithContext(ctx).Where("email = ? AND id != ? AND is_deleted = ?", user.Email, user.ID, false).First(&emailUser).Error
		if err == nil {
			return errors.New("user with this email already exists")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check email uniqueness: %w", err)
		}
	}

	// Update the user
	result := r.db.WithContext(ctx).Where("id = ? AND is_deleted = ?", user.ID, false).Updates(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no changes made")
	}

	return nil
}

// Delete marks a user as deleted (soft delete)
func (r *userRepository) Delete(ctx context.Context, id string) error {
	start := time.Now()
	defer func() {
		metrics.DBDuration.WithLabelValues("DeleteUser").Observe(time.Since(start).Seconds())
	}()
	if id == "" {
		return errors.New("user ID cannot be empty")
	}

	// Check if user exists and is not already deleted
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = ?", id, false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	// Soft delete: mark as deleted and set deleted_at timestamp
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&user).Updates(entity.User{
		IsDeleted: true,
		DeletedAt: &now,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// List retrieves a list of users with pagination and filtering
func (r *userRepository) List(ctx context.Context, conditions entity.Conditions) ([]entity.User, int64, error) {
	start := time.Now()
	defer func() {
		metrics.DBDuration.WithLabelValues("ListUsers").Observe(time.Since(start).Seconds())
	}()
	var users []entity.User
	var total int64

	// Build the base query
	query := r.db.WithContext(ctx).Model(&entity.User{}).Where("is_deleted = ?", false)

	// Apply search filter if provided
	if conditions.Search != "" {
		searchTerm := "%" + conditions.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR location ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply pagination
	if conditions.Limit > 0 {
		query = query.Limit(conditions.Limit)
	}
	if conditions.Offset > 0 {
		query = query.Offset(conditions.Offset)
	}

	// Execute query with ordering
	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch users: %w", err)
	}

	return users, total, nil
}
