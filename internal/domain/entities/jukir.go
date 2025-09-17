package entities

import (
	"time"

	"gorm.io/gorm"
)

type JukirStatus string

const (
	JukirStatusActive   JukirStatus = "active"
	JukirStatusInactive JukirStatus = "inactive"
	JukirStatusPending  JukirStatus = "pending"
)

type Jukir struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;uniqueIndex"`
	JukirCode string         `json:"jukir_code" gorm:"uniqueIndex;not null" validate:"required,min=3,max=20"`
	AreaID    uint           `json:"area_id" gorm:"not null"`
	QRToken   string         `json:"qr_token" gorm:"uniqueIndex;not null"`
	Status    JukirStatus    `json:"status" gorm:"type:varchar(20);not null;default:'pending'" validate:"required,oneof=active inactive pending"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User     User             `json:"user" gorm:"foreignKey:UserID"`
	Area     ParkingArea      `json:"area" gorm:"foreignKey:AreaID"`
	Sessions []ParkingSession `json:"sessions,omitempty" gorm:"foreignKey:JukirID"`
}

type CreateJukirRequest struct {
	UserID    uint   `json:"user_id" validate:"required"`
	JukirCode string `json:"jukir_code" validate:"required,min=3,max=20"`
	AreaID    uint   `json:"area_id" validate:"required"`
}

type UpdateJukirRequest struct {
	JukirCode *string      `json:"jukir_code,omitempty" validate:"omitempty,min=3,max=20"`
	AreaID    *uint        `json:"area_id,omitempty" validate:"omitempty"`
	Status    *JukirStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive pending"`
}

type JukirDashboardResponse struct {
	PendingPayments   int64   `json:"pending_payments"`
	DailyRevenue      float64 `json:"daily_revenue"`
	ActiveSessions    int64   `json:"active_sessions"`
	TotalTransactions int64   `json:"total_transactions"`
}

type JukirQRResponse struct {
	QRToken string `json:"qr_token"`
	Area    string `json:"area_name"`
	Code    string `json:"jukir_code"`
}
