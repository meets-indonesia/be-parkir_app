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
)

type VehicleType string

const (
	VehicleTypeMobil VehicleType = "mobil"
	VehicleTypeMotor VehicleType = "motor"
)

type ParkingSession struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	JukirID        *uint          `json:"jukir_id,omitempty"`
	AreaID         uint           `json:"area_id" gorm:"not null"`
	VehicleType    VehicleType    `json:"vehicle_type" gorm:"type:varchar(10);not null" validate:"required,oneof=mobil motor"`
	PlatNomor      *string        `json:"plat_nomor,omitempty" gorm:"null" validate:"omitempty,min=1,max=20"`
	IsManualRecord bool           `json:"is_manual_record" gorm:"not null;default:false"`
	CheckinTime    time.Time      `json:"checkin_time" gorm:"not null"`
	CheckoutTime   *time.Time     `json:"checkout_time,omitempty"`
	Duration       *int           `json:"duration,omitempty"` // in minutes
	TotalCost      *float64       `json:"total_cost,omitempty"`
	PaymentStatus  PaymentStatus  `json:"payment_status" gorm:"type:varchar(20);not null;default:'pending'" validate:"required,oneof=pending paid failed"`
	SessionStatus  SessionStatus  `json:"session_status" gorm:"type:varchar(20);not null;default:'active'" validate:"required,oneof=active pending_payment completed cancelled"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Jukir   *Jukir      `json:"jukir,omitempty" gorm:"foreignKey:JukirID"`
	Area    ParkingArea `json:"area" gorm:"foreignKey:AreaID"`
	Payment *Payment    `json:"payment,omitempty" gorm:"foreignKey:SessionID"`
}

type CheckinRequest struct {
	QRToken     string      `json:"qr_token" validate:"required"`
	Latitude    *float64    `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude   *float64    `json:"longitude,omitempty" validate:"omitempty,longitude"`
	VehicleType VehicleType `json:"vehicle_type" validate:"required,oneof=mobil motor"`
	PlatNomor   *string     `json:"plat_nomor,omitempty" validate:"omitempty,min=1,max=20"`
}

type CheckoutRequest struct {
	SessionID *uint    `json:"session_id,omitempty" validate:"omitempty"`
	QRToken   string   `json:"qr_token" validate:"required"`
	PlatNomor *string  `json:"plat_nomor,omitempty" validate:"omitempty,min=1,max=20"`
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
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
	PlatNomor   *string   `json:"plat_nomor,omitempty"` // Optional - bisa null jika tidak diisi saat checkin
	HourlyRate  float64   `json:"hourly_rate"`
	Duration    int       `json:"duration"` // in minutes
	CurrentCost float64   `json:"current_cost"`
}

type SessionHistoryResponse struct {
	Sessions []ParkingSession `json:"sessions"`
	Count    int64            `json:"count"`
}

// Manual Record DTOs
type ManualCheckinRequest struct {
	PlatNomor   string      `json:"plat_nomor" validate:"required,min=1,max=20"`
	VehicleType VehicleType `json:"vehicle_type" validate:"required,oneof=mobil motor"`
	WaktuMasuk  time.Time   `json:"waktu_masuk" validate:"required"`
	Latitude    *float64    `json:"latitude" validate:"required,latitude"`
	Longitude   *float64    `json:"longitude" validate:"required,longitude"`
}

type ManualCheckoutRequest struct {
	SessionID   uint      `json:"session_id" validate:"required"`
	WaktuKeluar time.Time `json:"waktu_keluar" validate:"required"`
}

type ManualCheckinResponse struct {
	SessionID   uint      `json:"session_id"`
	PlatNomor   string    `json:"plat_nomor"`
	VehicleType string    `json:"vehicle_type"`
	WaktuMasuk  time.Time `json:"waktu_masuk"`
	Area        string    `json:"area_name"`
	ParkingCost float64   `json:"parking_cost"`
}

type ManualCheckoutResponse struct {
	SessionID     uint      `json:"session_id"`
	PlatNomor     string    `json:"plat_nomor"`
	VehicleType   string    `json:"vehicle_type"`
	WaktuMasuk    time.Time `json:"waktu_masuk"`
	WaktuKeluar   time.Time `json:"waktu_keluar"`
	Duration      int       `json:"duration"` // in minutes
	TotalCost     float64   `json:"total_cost"`
	PaymentStatus string    `json:"payment_status"`
}
