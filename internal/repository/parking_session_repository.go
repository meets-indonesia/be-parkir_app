package repository

import (
	"be-parkir/internal/domain/entities"
	"time"

	"gorm.io/gorm"
)

type ParkingSessionRepository interface {
	Create(session *entities.ParkingSession) error
	GetByID(id uint) (*entities.ParkingSession, error)
	GetActiveByUserID(userID uint) (*entities.ParkingSession, error)
	Update(session *entities.ParkingSession) error
	Delete(id uint) error
	GetUserHistory(userID uint, limit, offset int) ([]entities.ParkingSession, int64, error)
	GetJukirActiveSessions(jukirID uint) ([]entities.ParkingSession, error)
	GetPendingPayments(jukirID uint) ([]entities.ParkingSession, error)
	GetSessionsByArea(areaID uint, startDate, endDate time.Time) ([]entities.ParkingSession, error)
	GetTimeoutSessions() ([]entities.ParkingSession, error)
	GetAllSessions(limit, offset int, filters map[string]interface{}) ([]entities.ParkingSession, int64, error)
}

type parkingSessionRepository struct {
	db *gorm.DB
}

func NewParkingSessionRepository(db *gorm.DB) ParkingSessionRepository {
	return &parkingSessionRepository{db: db}
}

func (r *parkingSessionRepository) Create(session *entities.ParkingSession) error {
	return r.db.Create(session).Error
}

func (r *parkingSessionRepository) GetByID(id uint) (*entities.ParkingSession, error) {
	var session entities.ParkingSession
	err := r.db.Preload("User").Preload("Jukir").Preload("Area").Preload("Payment").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *parkingSessionRepository) GetActiveByUserID(userID uint) (*entities.ParkingSession, error) {
	var session entities.ParkingSession
	err := r.db.Preload("User").Preload("Jukir").Preload("Area").Preload("Payment").
		Where("user_id = ? AND session_status = ?", userID, entities.SessionStatusActive).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *parkingSessionRepository) Update(session *entities.ParkingSession) error {
	return r.db.Save(session).Error
}

func (r *parkingSessionRepository) Delete(id uint) error {
	return r.db.Delete(&entities.ParkingSession{}, id).Error
}

func (r *parkingSessionRepository) GetUserHistory(userID uint, limit, offset int) ([]entities.ParkingSession, int64, error) {
	var sessions []entities.ParkingSession
	var count int64

	query := r.db.Model(&entities.ParkingSession{}).Where("user_id = ?", userID)
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").Preload("Jukir").Preload("Area").Preload("Payment").
		Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, count, err
}

func (r *parkingSessionRepository) GetJukirActiveSessions(jukirID uint) ([]entities.ParkingSession, error) {
	var sessions []entities.ParkingSession
	err := r.db.Preload("User").Preload("Jukir").Preload("Area").Preload("Payment").
		Where("jukir_id = ? AND session_status = ?", jukirID, entities.SessionStatusActive).
		Find(&sessions).Error
	return sessions, err
}

func (r *parkingSessionRepository) GetPendingPayments(jukirID uint) ([]entities.ParkingSession, error) {
	var sessions []entities.ParkingSession
	err := r.db.Preload("User").Preload("Jukir").Preload("Area").Preload("Payment").
		Where("jukir_id = ? AND session_status = ?", jukirID, entities.SessionStatusPendingPayment).
		Find(&sessions).Error
	return sessions, err
}

func (r *parkingSessionRepository) GetSessionsByArea(areaID uint, startDate, endDate time.Time) ([]entities.ParkingSession, error) {
	var sessions []entities.ParkingSession
	err := r.db.Preload("User").Preload("Jukir").Preload("Area").Preload("Payment").
		Where("area_id = ? AND created_at BETWEEN ? AND ?", areaID, startDate, endDate).
		Find(&sessions).Error
	return sessions, err
}

func (r *parkingSessionRepository) GetTimeoutSessions() ([]entities.ParkingSession, error) {
	var sessions []entities.ParkingSession
	timeoutTime := time.Now().Add(-12 * time.Hour) // 12 hours timeout

	err := r.db.Preload("User").Preload("Jukir").Preload("Area").Preload("Payment").
		Where("session_status = ? AND checkin_time < ?", entities.SessionStatusActive, timeoutTime).
		Find(&sessions).Error
	return sessions, err
}

func (r *parkingSessionRepository) GetAllSessions(limit, offset int, filters map[string]interface{}) ([]entities.ParkingSession, int64, error) {
	var sessions []entities.ParkingSession
	var count int64

	query := r.db.Model(&entities.ParkingSession{})

	// Apply filters
	for key, value := range filters {
		if value != nil {
			query = query.Where(key+" = ?", value)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").Preload("Jukir").Preload("Area").Preload("Payment").
		Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, count, err
}
