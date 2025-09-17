package entities

import (
	"time"

	"gorm.io/gorm"
)

type SessionStatus string

const (
	SessionStatusActive         SessionStatus = "active"
	SessionStatusPendingPayment SessionStatus = "pending_payment"
	SessionStatusCompleted      SessionStatus = "completed"
	SessionStatusCancelled      SessionStatus = "cancelled"
	SessionStatusTimeout        SessionStatus = "timeout"
)

type ParkingSession struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	JukirID       *uint          `json:"jukir_id,omitempty"`
	AreaID        uint           `json:"area_id" gorm:"not null"`
	CheckinTime   time.Time      `json:"checkin_time" gorm:"not null"`
	CheckoutTime  *time.Time     `json:"checkout_time,omitempty"`
	Duration      *int           `json:"duration,omitempty"` // in minutes
	TotalCost     *float64       `json:"total_cost,omitempty"`
	PaymentStatus PaymentStatus  `json:"payment_status" gorm:"type:varchar(20);not null;default:'pending'" validate:"required,oneof=pending paid failed"`
	SessionStatus SessionStatus  `json:"session_status" gorm:"type:varchar(20);not null;default:'active'" validate:"required,oneof=active pending_payment completed cancelled timeout"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User    User        `json:"user" gorm:"foreignKey:UserID"`
	Jukir   *Jukir      `json:"jukir,omitempty" gorm:"foreignKey:JukirID"`
	Area    ParkingArea `json:"area" gorm:"foreignKey:AreaID"`
	Payment *Payment    `json:"payment,omitempty" gorm:"foreignKey:SessionID"`
}

type CheckinRequest struct {
	QRToken   string  `json:"qr_token" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

type CheckoutRequest struct {
	QRToken   string  `json:"qr_token" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

type CheckinResponse struct {
	SessionID   uint      `json:"session_id"`
	CheckinTime time.Time `json:"checkin_time"`
	Area        string    `json:"area_name"`
	HourlyRate  float64   `json:"hourly_rate"`
}

type CheckoutResponse struct {
	SessionID     uint      `json:"session_id"`
	CheckoutTime  time.Time `json:"checkout_time"`
	Duration      int       `json:"duration"` // in minutes
	TotalCost     float64   `json:"total_cost"`
	PaymentStatus string    `json:"payment_status"`
}

type ActiveSessionResponse struct {
	SessionID   uint      `json:"session_id"`
	CheckinTime time.Time `json:"checkin_time"`
	Area        string    `json:"area_name"`
	HourlyRate  float64   `json:"hourly_rate"`
	Duration    int       `json:"duration"` // in minutes
	CurrentCost float64   `json:"current_cost"`
}

type SessionHistoryResponse struct {
	Sessions []ParkingSession `json:"sessions"`
	Count    int64            `json:"count"`
}
