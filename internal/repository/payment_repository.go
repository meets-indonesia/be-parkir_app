package repository

import (
	"be-parkir/internal/domain/entities"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *entities.Payment) error
	GetByID(id uint) (*entities.Payment, error)
	GetBySessionID(sessionID uint) (*entities.Payment, error)
	Update(payment *entities.Payment) error
	Delete(id uint) error
	GetJukirDailyRevenue(jukirID uint, date time.Time) (float64, error)
	GetJukirPendingPayments(jukirID uint) ([]entities.Payment, error)
	GetDailyReport(date time.Time) (*entities.DailyReportResponse, error)
	GetRevenueByDateRange(startDate, endDate time.Time) (float64, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *entities.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByID(id uint) (*entities.Payment, error) {
	var payment entities.Payment
	err := r.db.Preload("Session").Preload("Jukir").First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetBySessionID(sessionID uint) (*entities.Payment, error) {
	var payment entities.Payment
	err := r.db.Preload("Session").Preload("Jukir").Where("session_id = ?", sessionID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) Update(payment *entities.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Payment{}, id).Error
}

func (r *paymentRepository) GetJukirDailyRevenue(jukirID uint, date time.Time) (float64, error) {
	var total float64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Model(&entities.Payment{}).
		Where("confirmed_by = ? AND status = ? AND confirmed_at BETWEEN ? AND ?",
			jukirID, entities.PaymentStatusPaid, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error

	return total, err
}

func (r *paymentRepository) GetJukirPendingPayments(jukirID uint) ([]entities.Payment, error) {
	var payments []entities.Payment
	err := r.db.Preload("Session").Preload("Jukir").
		Joins("JOIN parking_sessions ON payments.session_id = parking_sessions.id").
		Where("parking_sessions.jukir_id = ? AND payments.status = ?", jukirID, entities.PaymentStatusPending).
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) GetDailyReport(date time.Time) (*entities.DailyReportResponse, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var report entities.DailyReportResponse
	report.Date = date.Format("2006-01-02")

	// Total sessions
	err := r.db.Model(&entities.ParkingSession{}).
		Where("created_at BETWEEN ? AND ?", startOfDay, endOfDay).
		Count(&report.TotalSessions).Error
	if err != nil {
		return nil, err
	}

	// Total revenue
	err = r.db.Model(&entities.Payment{}).
		Where("status = ? AND confirmed_at BETWEEN ? AND ?", entities.PaymentStatusPaid, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&report.TotalRevenue).Error
	if err != nil {
		return nil, err
	}

	// Pending payments
	err = r.db.Model(&entities.Payment{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", entities.PaymentStatusPending, startOfDay, endOfDay).
		Count(&report.PendingPayments).Error
	if err != nil {
		return nil, err
	}

	// Completed sessions
	err = r.db.Model(&entities.ParkingSession{}).
		Where("session_status = ? AND created_at BETWEEN ? AND ?", entities.SessionStatusCompleted, startOfDay, endOfDay).
		Count(&report.CompletedSessions).Error
	if err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *paymentRepository) GetRevenueByDateRange(startDate, endDate time.Time) (float64, error) {
	var total float64
	err := r.db.Model(&entities.Payment{}).
		Where("status = ? AND confirmed_at BETWEEN ? AND ?", entities.PaymentStatusPaid, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
