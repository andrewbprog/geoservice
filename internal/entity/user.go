package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// User represents a user in the system
type User struct {
	ID        string     `json:"id" gorm:"type:uuid;primaryKey" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string     `json:"email" gorm:"uniqueIndex;not null" validate:"required,email" example:"user@example.com"`
	Name      string     `json:"name" gorm:"not null" validate:"required,min=2,max=100" example:"John Doe"`
	Age       int        `json:"age" gorm:"not null" validate:"required,min=1,max=150" example:"25"`
	Location  string     `json:"location" gorm:"not null" validate:"required,min=2,max=200" example:"New York, USA"`
	IsDeleted bool       `json:"is_deleted" gorm:"default:false"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	Age      int    `json:"age" validate:"required,min=1,max=150" example:"25"`
	Location string `json:"location" validate:"required,min=2,max=200" example:"USA"`
} // @name CreateUserRequest

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty" validate:"omitempty,email" example:"user@example.com"`
	Name     string `json:"name,omitempty" validate:"omitempty,min=2,max=100" example:"John Doe"`
	Age      int    `json:"age,omitempty" validate:"omitempty,min=1,max=150" example:"25"`
	Location string `json:"location,omitempty" validate:"omitempty,min=2,max=200" example:"USA "`
} // @name UpdateUserRequest

// UserResponse represents the response payload for user operations
type UserResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@example.com"`
	Name      string    `json:"name" example:"John Doe"`
	Age       int       `json:"age" example:"25"`
	Location  string    `json:"location" example:"USA"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} // @name UserResponse

// UserListResponse represents the response payload for listing users
type UserListResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int64          `json:"total" example:"10"`
	Limit  int            `json:"limit" example:"10"`
	Offset int            `json:"offset" example:"0"`
} // @name UserListResponse

// Conditions represents query conditions for listing users
type Conditions struct {
	Limit  int    `json:"limit" form:"limit" validate:"min=10,max=100"`
	Offset int    `json:"offset" form:"offset" validate:"min=0"`
	Search string `json:"search" form:"search"`
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// ToResponse converts a User model to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Age:       u.Age,
		Location:  u.Location,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// FromCreateRequest creates a User from CreateUserRequest
func (u *User) FromCreateRequest(req CreateUserRequest) {
	u.Email = req.Email
	u.Name = req.Name
	u.Age = req.Age
	u.Location = req.Location
}

// UpdateFromRequest updates User fields from UpdateUserRequest
func (u *User) UpdateFromRequest(req UpdateUserRequest) {
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Age > 0 {
		u.Age = req.Age
	}
	if req.Location != "" {
		u.Location = req.Location
	}
}
