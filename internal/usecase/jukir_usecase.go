package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"errors"
	"time"
)

type JukirUsecase interface {
	GetDashboard(jukirID uint) (*entities.JukirDashboardResponse, error)
	GetPendingPayments(jukirID uint) ([]entities.PendingPaymentResponse, error)
	GetActiveSessions(jukirID uint) ([]entities.ParkingSession, error)
	ConfirmPayment(jukirID uint, req *entities.ConfirmPaymentRequest) (*entities.ConfirmPaymentResponse, error)
	GetQRCode(jukirID uint) (*entities.JukirQRResponse, error)
	GetDailyReport(jukirID uint, date time.Time) (*entities.DailyReportResponse, error)
	GetJukirByUserID(userID uint) (*entities.Jukir, error)
}

type jukirUsecase struct {
	jukirRepo   repository.JukirRepository
	areaRepo    repository.ParkingAreaRepository
	sessionRepo repository.ParkingSessionRepository
	paymentRepo repository.PaymentRepository
}

func NewJukirUsecase(jukirRepo repository.JukirRepository, areaRepo repository.ParkingAreaRepository, sessionRepo repository.ParkingSessionRepository, paymentRepo repository.PaymentRepository) JukirUsecase {
	return &jukirUsecase{
		jukirRepo:   jukirRepo,
		areaRepo:    areaRepo,
		sessionRepo: sessionRepo,
		paymentRepo: paymentRepo,
	}
}

func (u *jukirUsecase) GetDashboard(jukirID uint) (*entities.JukirDashboardResponse, error) {
	// Get jukir info
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	// Get pending payments count
	pendingSessions, err := u.sessionRepo.GetPendingPayments(jukirID)
	if err != nil {
		return nil, errors.New("failed to get pending payments")
	}

	// Get daily revenue
	dailyRevenue, err := u.paymentRepo.GetJukirDailyRevenue(jukirID, time.Now())
	if err != nil {
		return nil, errors.New("failed to get daily revenue")
	}

	// Get active sessions count
	activeSessions, err := u.sessionRepo.GetJukirActiveSessions(jukirID)
	if err != nil {
		return nil, errors.New("failed to get active sessions")
	}

	// Get total transactions for today
	startOfDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	todaySessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startOfDay, endOfDay)
	if err != nil {
		return nil, errors.New("failed to get today's sessions")
	}

	return &entities.JukirDashboardResponse{
		PendingPayments:   int64(len(pendingSessions)),
		DailyRevenue:      dailyRevenue,
		ActiveSessions:    int64(len(activeSessions)),
		TotalTransactions: int64(len(todaySessions)),
	}, nil
}

func (u *jukirUsecase) GetPendingPayments(jukirID uint) ([]entities.PendingPaymentResponse, error) {
	sessions, err := u.sessionRepo.GetPendingPayments(jukirID)
	if err != nil {
		return nil, errors.New("failed to get pending payments")
	}

	var pendingPayments []entities.PendingPaymentResponse
	for _, session := range sessions {
		platNomor := ""
		if session.PlatNomor != nil {
			platNomor = *session.PlatNomor
		}

		pendingPayments = append(pendingPayments, entities.PendingPaymentResponse{
			SessionID:     session.ID,
			PlatNomor:     platNomor,
			VehicleType:   string(session.VehicleType),
			CheckinTime:   session.CheckinTime,
			CheckoutTime:  *session.CheckoutTime,
			Duration:      *session.Duration,
			Amount:        *session.TotalCost,
			AreaName:      session.Area.Name,
			PaymentStatus: string(session.PaymentStatus),
		})
	}

	return pendingPayments, nil
}

func (u *jukirUsecase) GetActiveSessions(jukirID uint) ([]entities.ParkingSession, error) {
	sessions, err := u.sessionRepo.GetJukirActiveSessions(jukirID)
	if err != nil {
		return nil, errors.New("failed to get active sessions")
	}
	return sessions, nil
}

func (u *jukirUsecase) ConfirmPayment(jukirID uint, req *entities.ConfirmPaymentRequest) (*entities.ConfirmPaymentResponse, error) {
	// Get session
	session, err := u.sessionRepo.GetByID(req.SessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	// Verify session belongs to this jukir
	if session.JukirID == nil || *session.JukirID != jukirID {
		return nil, errors.New("session does not belong to this jukir")
	}

	// Check if session is in pending payment status
	if session.SessionStatus != entities.SessionStatusPendingPayment {
		return nil, errors.New("session is not in pending payment status")
	}

	// Get or create payment record
	payment, err := u.paymentRepo.GetBySessionID(req.SessionID)
	if err != nil {
		// Create new payment record
		payment = &entities.Payment{
			SessionID:     req.SessionID,
			Amount:        *session.TotalCost,
			PaymentMethod: req.PaymentMethod,
			Status:        entities.PaymentStatusPaid,
			ConfirmedBy:   &jukirID,
			ConfirmedAt:   &[]time.Time{time.Now()}[0],
		}
		if err := u.paymentRepo.Create(payment); err != nil {
			return nil, errors.New("failed to create payment record")
		}
	} else {
		// Update existing payment record
		payment.PaymentMethod = req.PaymentMethod
		payment.Status = entities.PaymentStatusPaid
		payment.ConfirmedBy = &jukirID
		payment.ConfirmedAt = &[]time.Time{time.Now()}[0]
		if err := u.paymentRepo.Update(payment); err != nil {
			return nil, errors.New("failed to update payment record")
		}
	}

	// Update session status
	session.SessionStatus = entities.SessionStatusCompleted
	session.PaymentStatus = entities.PaymentStatusPaid
	if err := u.sessionRepo.Update(session); err != nil {
		return nil, errors.New("failed to update session status")
	}

	return &entities.ConfirmPaymentResponse{
		PaymentID:     payment.ID,
		Amount:        payment.Amount,
		PaymentMethod: payment.PaymentMethod,
		ConfirmedAt:   *payment.ConfirmedAt,
		Status:        payment.Status,
	}, nil
}

func (u *jukirUsecase) GetQRCode(jukirID uint) (*entities.JukirQRResponse, error) {
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	return &entities.JukirQRResponse{
		QRToken: jukir.QRToken,
		Area:    jukir.Area.Name,
		Code:    jukir.JukirCode,
	}, nil
}

func (u *jukirUsecase) GetDailyReport(jukirID uint, date time.Time) (*entities.DailyReportResponse, error) {
	// Get jukir info
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	// Get sessions for the date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	sessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startOfDay, endOfDay)
	if err != nil {
		return nil, errors.New("failed to get sessions for date")
	}

	// Calculate metrics
	var totalRevenue float64
	var pendingPayments int64
	var completedSessions int64

	for _, session := range sessions {
		if session.SessionStatus == entities.SessionStatusCompleted {
			completedSessions++
			if session.TotalCost != nil {
				totalRevenue += *session.TotalCost
			}
		} else if session.SessionStatus == entities.SessionStatusPendingPayment {
			pendingPayments++
		}
	}

	return &entities.DailyReportResponse{
		Date:              date.Format("2006-01-02"),
		TotalSessions:     int64(len(sessions)),
		TotalRevenue:      totalRevenue,
		PendingPayments:   pendingPayments,
		CompletedSessions: completedSessions,
	}, nil
}

func (u *jukirUsecase) GetJukirByUserID(userID uint) (*entities.Jukir, error) {
	jukir, err := u.jukirRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}
	return jukir, nil
}
