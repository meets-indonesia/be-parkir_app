package repository

import (
	"be-parkir/internal/domain/entities"
	"math"

	"gorm.io/gorm"
)

type ParkingAreaRepository interface {
	Create(area *entities.ParkingArea) error
	GetByID(id uint) (*entities.ParkingArea, error)
	Update(area *entities.ParkingArea) error
	Delete(id uint) error
	List(limit, offset int) ([]entities.ParkingArea, int64, error)
	GetNearbyAreas(lat, lng, radius float64) ([]entities.ParkingArea, error)
	GetActiveAreas() ([]entities.ParkingArea, error)
}

type parkingAreaRepository struct {
	db *gorm.DB
}

func NewParkingAreaRepository(db *gorm.DB) ParkingAreaRepository {
	return &parkingAreaRepository{db: db}
}

func (r *parkingAreaRepository) Create(area *entities.ParkingArea) error {
	return r.db.Create(area).Error
}

func (r *parkingAreaRepository) GetByID(id uint) (*entities.ParkingArea, error) {
	var area entities.ParkingArea
	err := r.db.Preload("Jukirs").First(&area, id).Error
	if err != nil {
		return nil, err
	}
	return &area, nil
}

func (r *parkingAreaRepository) Update(area *entities.ParkingArea) error {
	return r.db.Save(area).Error
}

func (r *parkingAreaRepository) Delete(id uint) error {
	return r.db.Delete(&entities.ParkingArea{}, id).Error
}

func (r *parkingAreaRepository) List(limit, offset int) ([]entities.ParkingArea, int64, error) {
	var areas []entities.ParkingArea
	var count int64

	query := r.db.Model(&entities.ParkingArea{})
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Jukirs").Limit(limit).Offset(offset).Find(&areas).Error
	return areas, count, err
}

func (r *parkingAreaRepository) GetNearbyAreas(lat, lng, radius float64) ([]entities.ParkingArea, error) {
	var areas []entities.ParkingArea

	// Using Haversine formula for distance calculation
	// For simplicity, we'll use a bounding box approach
	// In production, consider using PostGIS for better performance
	latRange := radius / 111.0                               // Approximate km per degree latitude
	lngRange := radius / (111.0 * math.Cos(lat*math.Pi/180)) // Adjust for longitude

	err := r.db.Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ? AND status = ?",
		lat-latRange, lat+latRange, lng-lngRange, lng+lngRange, entities.AreaStatusActive).
		Preload("Jukirs").
		Find(&areas).Error

	return areas, err
}

func (r *parkingAreaRepository) GetActiveAreas() ([]entities.ParkingArea, error) {
	var areas []entities.ParkingArea
	err := r.db.Where("status = ?", entities.AreaStatusActive).Preload("Jukirs").Find(&areas).Error
	return areas, err
}
