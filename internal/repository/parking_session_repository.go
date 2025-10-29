package repository

import (
	"be-parkir/internal/domain/entities"
	"time"

	"gorm.io/gorm"
)

type ParkingSessionRepository interface {
	Create(session *entities.ParkingSession) error
	GetByID(id uint) (*entities.ParkingSession, error)
	GetActiveByPlatNomor(platNomor string) (*entities.ParkingSession, error)
	GetActiveByQRToken(qrToken string) (*entities.ParkingSession, error)
	Update(session *entities.ParkingSession) error
	Delete(id uint) error
	GetHistoryByPlatNomor(platNomor string, limit, offset int) ([]entities.ParkingSession, int64, error)
	GetJukirActiveSessions(jukirID uint) ([]entities.ParkingSession, error)
	GetPendingPayments(jukirID uint) ([]entities.ParkingSession, error)
	GetSessionsByArea(areaID uint, startDate, endDate time.Time) ([]entities.ParkingSession, error)
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
	err := r.db.Preload("Jukir").Preload("Area").Preload("Payment").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *parkingSessionRepository) GetActiveByPlatNomor(platNomor string) (*entities.ParkingSession, error) {
	var session entities.ParkingSession
	err := r.db.Preload("Jukir").Preload("Area").Preload("Payment").
		Where("plat_nomor = ? AND session_status = ?", platNomor, entities.SessionStatusActive).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *parkingSessionRepository) GetActiveByQRToken(qrToken string) (*entities.ParkingSession, error) {
	var session entities.ParkingSession
	err := r.db.Preload("Jukir").Preload("Area").Preload("Payment").
		Joins("JOIN jukirs ON parking_sessions.jukir_id = jukirs.id").
		Where("jukirs.qr_token = ? AND parking_sessions.session_status = ?", qrToken, entities.SessionStatusActive).
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

func (r *parkingSessionRepository) GetHistoryByPlatNomor(platNomor string, limit, offset int) ([]entities.ParkingSession, int64, error) {
	var sessions []entities.ParkingSession
	var count int64

	query := r.db.Model(&entities.ParkingSession{}).Where("plat_nomor = ?", platNomor)
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Jukir").Preload("Area").Preload("Payment").
		Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, count, err
}

func (r *parkingSessionRepository) GetJukirActiveSessions(jukirID uint) ([]entities.ParkingSession, error) {
	var sessions []entities.ParkingSession
	err := r.db.Preload("Jukir").Preload("Area").Preload("Payment").
		Where("jukir_id = ? AND session_status = ?", jukirID, entities.SessionStatusActive).
		Find(&sessions).Error
	return sessions, err
}

func (r *parkingSessionRepository) GetPendingPayments(jukirID uint) ([]entities.ParkingSession, error) {
	var sessions []entities.ParkingSession
	err := r.db.Preload("Jukir").Preload("Area").Preload("Payment").
		Where("jukir_id = ? AND session_status = ?", jukirID, entities.SessionStatusPendingPayment).
		Find(&sessions).Error
	return sessions, err
}

func (r *parkingSessionRepository) GetSessionsByArea(areaID uint, startDate, endDate time.Time) ([]entities.ParkingSession, error) {
	var sessions []entities.ParkingSession
	err := r.db.Preload("Jukir").Preload("Area").Preload("Payment").
		Where("area_id = ? AND checkin_time >= ? AND checkin_time <= ?", areaID, startDate, endDate).
		Order("checkin_time ASC").
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

	err := query.Preload("Jukir").Preload("Area").Preload("Payment").
		Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, count, err
}
