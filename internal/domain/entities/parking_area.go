package entities

import (
	"time"

	"gorm.io/gorm"
)

type AreaStatus string

const (
	AreaStatusActive      AreaStatus = "active"
	AreaStatusInactive    AreaStatus = "inactive"
	AreaStatusMaintenance AreaStatus = "maintenance"
)

type JenisArea string

const (
	JenisAreaIndoor  JenisArea = "indoor"
	JenisAreaOutdoor JenisArea = "outdoor"
	JenisAreaMix     JenisArea = "mix"
)

type ParkingArea struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Name              string         `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
	Address           string         `json:"address" gorm:"not null" validate:"required"`
	Latitude          float64        `json:"latitude" gorm:"not null" validate:"required,latitude"`
	Longitude         float64        `json:"longitude" gorm:"not null" validate:"required,longitude"`
	Regional          string         `json:"regional" gorm:"type:varchar(50)" validate:"required,max=50"`
	Image             *string        `json:"image,omitempty" gorm:"type:text"`
	HourlyRate        float64        `json:"hourly_rate" gorm:"not null" validate:"required,min=0"`
	Status            AreaStatus     `json:"status" gorm:"type:varchar(20);not null;default:'active'" validate:"required,oneof=active inactive maintenance"`
	MaxMobil          *int           `json:"max_mobil,omitempty" gorm:"type:int;default:0"`
	MaxMotor          *int           `json:"max_motor,omitempty" gorm:"type:int;default:0"`
	StatusOperasional string         `json:"status_operasional" gorm:"type:varchar(20);not null;default:'buka'" validate:"required,oneof=buka tutup maintenance"`
	JenisArea         JenisArea      `json:"jenis_area" gorm:"type:varchar(10);not null;default:'outdoor'" validate:"required,oneof=indoor outdoor mix"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Jukirs   []Jukir          `json:"jukirs,omitempty" gorm:"foreignKey:AreaID"`
	Sessions []ParkingSession `json:"sessions,omitempty" gorm:"foreignKey:AreaID"`
}

type CreateParkingAreaRequest struct {
	Name      string  `json:"name" validate:"required,min=2,max=100"`
	Address   string  `json:"address" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
	Regional  string  `json:"regional" validate:"required,max=50"`
	// image dikirim melalui multipart form-data, bukan JSON
	HourlyRate        float64   `json:"hourly_rate" validate:"required,min=0"`
	MaxMobil          *int      `json:"max_mobil,omitempty" validate:"omitempty,min=0"`
	MaxMotor          *int      `json:"max_motor,omitempty" validate:"omitempty,min=0"`
	StatusOperasional string    `json:"status_operasional" validate:"required,oneof=buka tutup maintenance"`
	JenisArea         JenisArea `json:"jenis_area" validate:"required,oneof=indoor outdoor mix"`
}

type UpdateParkingAreaRequest struct {
	Name      *string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Address   *string  `json:"address,omitempty" validate:"omitempty"`
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
	Regional  *string  `json:"regional,omitempty" validate:"omitempty,max=50"`
	Image     *string  `json:"image,omitempty"`
	// image update via endpoint yang sama: multipart form-data opsional
	HourlyRate        *float64    `json:"hourly_rate,omitempty" validate:"omitempty,min=0"`
	Status            *AreaStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive maintenance"`
	MaxMobil          *int        `json:"max_mobil,omitempty" validate:"omitempty,min=0"`
	MaxMotor          *int        `json:"max_motor,omitempty" validate:"omitempty,min=0"`
	StatusOperasional *string     `json:"status_operasional,omitempty" validate:"omitempty,oneof=buka tutup maintenance"`
	JenisArea         *JenisArea  `json:"jenis_area,omitempty" validate:"omitempty,oneof=indoor outdoor mix"`
}

type NearbyAreasRequest struct {
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
	Radius    float64  `json:"radius" validate:"omitempty,min=0.1,max=10"`
}

type NearbyAreasResponse struct {
	Areas []ParkingArea `json:"areas"`
	Count int64         `json:"count"`
}
