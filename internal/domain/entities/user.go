package entities

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleJukir    UserRole = "jukir"
	RoleAdmin    UserRole = "admin"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusPending  UserStatus = "pending"
)

type User struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
	Email           string         `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Phone           string         `json:"phone" gorm:"not null" validate:"required,min=10,max=15"`
	Password        string         `json:"-" gorm:"not null"`
	DisplayPassword *string        `json:"display_password,omitempty" gorm:"type:varchar(20)"` // For jukir password display
	Role            UserRole       `json:"role" gorm:"type:varchar(20);not null;default:'customer'" validate:"required,oneof=customer jukir admin"`
	Status          UserStatus     `json:"status" gorm:"type:varchar(20);not null;default:'active'" validate:"required,oneof=active inactive pending"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	JukirProfile *Jukir `json:"jukir_profile,omitempty" gorm:"foreignKey:UserID"`
}

type CreateUserRequest struct {
	Name     string   `json:"name" validate:"required,min=2,max=100"`
	Email    string   `json:"email" validate:"required,email"`
	Phone    string   `json:"phone" validate:"required,min=10,max=15"`
	Password string   `json:"password" validate:"required,min=6"`
	Role     UserRole `json:"role" validate:"required,oneof=customer jukir admin"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone *string `json:"phone,omitempty" validate:"omitempty,min=10,max=15"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
