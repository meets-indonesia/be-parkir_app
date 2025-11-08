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
	GetActiveSessions(jukirID uint, vehicleType *entities.VehicleType) ([]entities.ActiveSessionResponse, error)
	ConfirmPayment(jukirID uint, req *entities.ConfirmPaymentRequest) (*entities.ConfirmPaymentResponse, error)
	GetQRCode(jukirID uint) (*entities.JukirQRResponse, error)
	GetDailyReport(jukirID uint, date time.Time) (*entities.DailyReportResponse, error)
	GetJukirByUserID(userID uint) (*entities.Jukir, error)
	GetVehicleBreakdown(jukirID uint) (*entities.VehicleBreakdownResponse, error)
}

type jukirUsecase struct {
	jukirRepo    repository.JukirRepository
	areaRepo     repository.ParkingAreaRepository
	sessionRepo  repository.ParkingSessionRepository
	paymentRepo  repository.PaymentRepository
	eventManager *EventManager
}

func NewJukirUsecase(jukirRepo repository.JukirRepository, areaRepo repository.ParkingAreaRepository, sessionRepo repository.ParkingSessionRepository, paymentRepo repository.PaymentRepository, eventManager *EventManager) JukirUsecase {
	return &jukirUsecase{
		jukirRepo:    jukirRepo,
		areaRepo:     areaRepo,
		sessionRepo:  sessionRepo,
		paymentRepo:  paymentRepo,
		eventManager: eventManager,
	}
}

func (u *jukirUsecase) GetDashboard(jukirID uint) (*entities.JukirDashboardResponse, error) {
	// Get pending payments count (already filtered by jukir_id)
	pendingSessions, err := u.sessionRepo.GetPendingPayments(jukirID)
	if err != nil {
		return nil, errors.New("failed to get pending payments")
	}

	// Get daily revenue (already filtered by jukir_id)
	dailyRevenue, err := u.paymentRepo.GetJukirDailyRevenue(jukirID, time.Now())
	if err != nil {
		return nil, errors.New("failed to get daily revenue")
	}

	// Get active sessions count (already filtered by jukir_id)
	activeSessions, err := u.sessionRepo.GetJukirActiveSessions(jukirID)
	if err != nil {
		return nil, errors.New("failed to get active sessions")
	}

	// Get total transactions for today (filter by jukir_id, not area_id)
	startOfDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	todaySessions, err := u.sessionRepo.GetSessionsByJukir(jukirID, startOfDay, endOfDay)
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

func (u *jukirUsecase) GetActiveSessions(jukirID uint, vehicleType *entities.VehicleType) ([]entities.ActiveSessionResponse, error) {
	// Get all active sessions for this jukir (includes both manual and QR input)
	sessions, err := u.sessionRepo.GetJukirActiveSessions(jukirID)
	if err != nil {
		return nil, errors.New("failed to get active sessions")
	}

	// Transform to simplified active session response (only essential data)
	var activeSessions []entities.ActiveSessionResponse
	now := nowGMT7() // Use GMT+7 timezone
	for _, session := range sessions {
		// Filter by vehicle_type if provided
		if vehicleType != nil && session.VehicleType != *vehicleType {
			continue
		}

		// Calculate duration (handle negative if checkin_time is in future)
		durationMinutes := int(now.Sub(session.CheckinTime).Minutes())
		if durationMinutes < 0 {
			durationMinutes = 0 // If checkin_time is in future, set duration to 0
		}

		// Biaya parkir adalah FLAT RATE (tidak per jam)
		// HourlyRate sebenarnya adalah flat rate untuk sekali parkir
		currentCost := session.Area.HourlyRate // Flat rate, tidak dikali jam
		if currentCost < 0 {
			currentCost = 0
		}

		activeSessions = append(activeSessions, entities.ActiveSessionResponse{
			SessionID:   session.ID,
			CheckinTime: session.CheckinTime,
			Area:        session.Area.Name,
			PlatNomor:   session.PlatNomor,       // Include plat_nomor in response
			HourlyRate:  session.Area.HourlyRate, // Ini adalah flat rate
			Duration:    durationMinutes,
			CurrentCost: currentCost, // Flat rate, tidak per jam
		})
	}

	return activeSessions, nil
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

	// Notify jukir about payment confirmation via SSE
	eventData := PaymentConfirmedEvent{
		SessionID:     session.ID,
		PaymentID:     payment.ID,
		PaymentMethod: string(payment.PaymentMethod),
		Amount:        payment.Amount,
		ConfirmedBy:   "Jukir",
		ConfirmedAt:   payment.ConfirmedAt.Format(time.RFC3339),
	}

	u.eventManager.NotifyJukir(jukirID, EventPaymentConfirmed, eventData)

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
	// Get all sessions for the date (includes both manual and QR input)
	// Filter by jukir_id, not area_id - includes all session types
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	sessions, err := u.sessionRepo.GetSessionsByJukir(jukirID, startOfDay, endOfDay)
	if err != nil {
		return nil, errors.New("failed to get sessions for date")
	}

	// Calculate metrics
	var totalRevenue float64
	var pendingPayments int64
	var completedSessions int64
	var vehiclesOut int64

	records := make([]entities.DailyReportSessionItem, 0, len(sessions))

	for _, session := range sessions {
		if session.SessionStatus == entities.SessionStatusCompleted {
			completedSessions++
			if session.TotalCost != nil {
				totalRevenue += *session.TotalCost
			}
		}

		if session.SessionStatus == entities.SessionStatusPendingPayment {
			pendingPayments++
		}

		if session.CheckoutTime != nil {
			vehiclesOut++
		}

		var sessionID *uint
		var platNomor *string

		if session.IsManualRecord {
			platNomor = session.PlatNomor
		} else {
			sessionID = &session.ID
		}

		records = append(records, entities.DailyReportSessionItem{
			SessionID:    sessionID,
			PlatNomor:    platNomor,
			CheckinTime:  session.CheckinTime,
			CheckoutTime: session.CheckoutTime,
			VehicleType:  string(session.VehicleType),
			IsManual:     session.IsManualRecord,
			Status:       string(session.SessionStatus),
		})
	}

	return &entities.DailyReportResponse{
		Date:              date.Format("2006-01-02"),
		TotalSessions:     int64(len(sessions)),
		TotalRevenue:      totalRevenue,
		PendingPayments:   pendingPayments,
		CompletedSessions: completedSessions,
		VehiclesIn:        int64(len(sessions)),
		VehiclesOut:       vehiclesOut,
		Records:           records,
	}, nil
}

func (u *jukirUsecase) GetJukirByUserID(userID uint) (*entities.Jukir, error) {
	jukir, err := u.jukirRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}
	return jukir, nil
}

func (u *jukirUsecase) GetVehicleBreakdown(jukirID uint) (*entities.VehicleBreakdownResponse, error) {
	// Get all sessions for this jukir today (includes both manual and QR input)
	// Filter by jukir_id, not area_id - includes all session types
	// Use current date to get sessions that check-in today (regardless of timezone)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour) // End of today

	sessions, err := u.sessionRepo.GetSessionsByJukir(jukirID, startOfDay, endOfDay)
	if err != nil {
		return nil, errors.New("failed to get sessions")
	}

	// Calculate breakdown
	vehiclesIn := 0
	vehiclesOut := 0
	vehiclesInMobil := 0
	vehiclesOutMobil := 0
	vehiclesInMotor := 0
	vehiclesOutMotor := 0

	for _, session := range sessions {
		// Count check-ins
		vehiclesIn++

		// Count by vehicle type
		if session.VehicleType == entities.VehicleTypeMobil {
			vehiclesInMobil++
			if session.CheckoutTime != nil {
				vehiclesOutMobil++
			}
		} else if session.VehicleType == entities.VehicleTypeMotor {
			vehiclesInMotor++
			if session.CheckoutTime != nil {
				vehiclesOutMotor++
			}
		}

		// Count check-outs
		if session.CheckoutTime != nil {
			vehiclesOut++
		}
	}

	return &entities.VehicleBreakdownResponse{
		VehiclesIn:  vehiclesIn,
		VehiclesOut: vehiclesOut,
		VehiclesByType: map[string]struct {
			In  int `json:"in"`
			Out int `json:"out"`
		}{
			"mobil": {In: vehiclesInMobil, Out: vehiclesOutMobil},
			"motor": {In: vehiclesInMotor, Out: vehiclesOutMotor},
		},
	}, nil
}
