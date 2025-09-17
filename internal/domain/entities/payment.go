package entities

import (
	"time"

	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodQRIS         PaymentMethod = "qris"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

type Payment struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	SessionID     uint           `json:"session_id" gorm:"not null"`
	Amount        float64        `json:"amount" gorm:"not null" validate:"required,min=0"`
	PaymentMethod PaymentMethod  `json:"payment_method" gorm:"type:varchar(20);not null" validate:"required,oneof=cash qris bank_transfer"`
	ConfirmedBy   *uint          `json:"confirmed_by,omitempty"` // Jukir ID
	ConfirmedAt   *time.Time     `json:"confirmed_at,omitempty"`
	Status        PaymentStatus  `json:"status" gorm:"type:varchar(20);not null;default:'pending'" validate:"required,oneof=pending paid failed refunded"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Session ParkingSession `json:"session" gorm:"foreignKey:SessionID"`
	Jukir   *Jukir         `json:"jukir,omitempty" gorm:"foreignKey:ConfirmedBy"`
}

type ConfirmPaymentRequest struct {
	SessionID     uint          `json:"session_id" validate:"required"`
	PaymentMethod PaymentMethod `json:"payment_method" validate:"required,oneof=cash qris bank_transfer"`
}

type ConfirmPaymentResponse struct {
	PaymentID     uint          `json:"payment_id"`
	Amount        float64       `json:"amount"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	ConfirmedAt   time.Time     `json:"confirmed_at"`
	Status        PaymentStatus `json:"status"`
}

type PendingPaymentResponse struct {
	SessionID     uint      `json:"session_id"`
	UserID        uint      `json:"user_id"`
	UserName      string    `json:"user_name"`
	UserPhone     string    `json:"user_phone"`
	CheckinTime   time.Time `json:"checkin_time"`
	CheckoutTime  time.Time `json:"checkout_time"`
	Duration      int       `json:"duration"` // in minutes
	Amount        float64   `json:"amount"`
	AreaName      string    `json:"area_name"`
	PaymentStatus string    `json:"payment_status"`
}

type DailyReportResponse struct {
	Date              string  `json:"date"`
	TotalSessions     int64   `json:"total_sessions"`
	TotalRevenue      float64 `json:"total_revenue"`
	PendingPayments   int64   `json:"pending_payments"`
	CompletedSessions int64   `json:"completed_sessions"`
}
