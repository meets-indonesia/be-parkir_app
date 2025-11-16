package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"errors"
	"fmt"
	"math"
	"time"
)

type ParkingUsecase interface {
	GetNearbyAreas(req *entities.NearbyAreasRequest) (*entities.NearbyAreasResponse, error)
	Checkin(req *entities.CheckinRequest) (*entities.CheckinResponse, error)
	Checkout(req *entities.CheckoutRequest) (*entities.CheckoutResponse, error)
	GetActiveSession(qrToken string) (*entities.ActiveSessionResponse, error)
	GetActiveSessionByID(sessionID uint) (*entities.ActiveSessionResponse, error)
	GetSessionByID(sessionID uint) (*entities.ParkingSession, error)
	GetHistoryByPlatNomor(platNomor string, limit, offset int) (*entities.SessionHistoryResponse, error)
	GetHistoryBySession(sessionID uint) (*entities.ParkingSession, error)
	GetHistoryBySessionIDs(sessionIDs []uint) ([]entities.ParkingSession, error)
	ManualCheckin(jukirID uint, req *entities.ManualCheckinRequest) (*entities.ManualCheckinResponse, error)
	ManualCheckout(jukirID uint, req *entities.ManualCheckoutRequest) (*entities.ManualCheckoutResponse, error)
}

type parkingUsecase struct {
	sessionRepo  repository.ParkingSessionRepository
	areaRepo     repository.ParkingAreaRepository
	userRepo     repository.UserRepository
	jukirRepo    repository.JukirRepository
	paymentRepo  repository.PaymentRepository
	eventManager *EventManager
}

const defaultLocationToleranceKM = 0.3

func NewParkingUsecase(sessionRepo repository.ParkingSessionRepository, areaRepo repository.ParkingAreaRepository, userRepo repository.UserRepository, jukirRepo repository.JukirRepository, paymentRepo repository.PaymentRepository, eventManager *EventManager) ParkingUsecase {
	return &parkingUsecase{
		sessionRepo:  sessionRepo,
		areaRepo:     areaRepo,
		userRepo:     userRepo,
		jukirRepo:    jukirRepo,
		paymentRepo:  paymentRepo,
		eventManager: eventManager,
	}
}

func (u *parkingUsecase) GetNearbyAreas(req *entities.NearbyAreasRequest) (*entities.NearbyAreasResponse, error) {
	// If latitude and longitude are not provided, return all active areas
	if req.Latitude == nil || req.Longitude == nil {
		areas, err := u.areaRepo.GetActiveAreas()
		if err != nil {
			return nil, errors.New("failed to get parking areas")
		}

		return &entities.NearbyAreasResponse{
			Areas: areas,
			Count: int64(len(areas)),
		}, nil
	}

	// If lat/long provided, use nearby areas logic
	radius := req.Radius
	if radius == 0 {
		radius = 1.0 // Default 1km radius
	}

	areas, err := u.areaRepo.GetNearbyAreas(*req.Latitude, *req.Longitude, radius)
	if err != nil {
		return nil, errors.New("failed to get nearby areas")
	}

	// Filter areas by actual distance (more accurate than bounding box)
	var filteredAreas []entities.ParkingArea
	for _, area := range areas {
		distance := u.calculateDistance(*req.Latitude, *req.Longitude, area.Latitude, area.Longitude)
		if distance <= radius {
			filteredAreas = append(filteredAreas, area)
		}
	}

	return &entities.NearbyAreasResponse{
		Areas: filteredAreas,
		Count: int64(len(filteredAreas)),
	}, nil
}

func (u *parkingUsecase) Checkin(req *entities.CheckinRequest) (*entities.CheckinResponse, error) {
	// Get jukir by QR token
	jukir, err := u.jukirRepo.GetByQRToken(req.QRToken)
	if err != nil {
		return nil, errors.New("invalid QR code")
	}

	// Check if jukir is active
	if jukir.Status != entities.JukirStatusActive {
		return nil, errors.New("jukir is not active")
	}

	// Optional GPS verification (ignored if not provided)
	if req.Latitude != nil && req.Longitude != nil {
		// We no longer block QR check-in when coordinates are missing, but if the client still
		// sends them we keep the distance check for sanity.
		if err := u.ensureWithinArea(*req.Latitude, *req.Longitude, jukir.Area); err != nil {
			return nil, err
		}
	}

	// Get current time for check-in
	checkinTime := nowGMT7()

	// Create parking session
	session := &entities.ParkingSession{
		JukirID:        &jukir.ID,
		AreaID:         jukir.AreaID,
		VehicleType:    req.VehicleType,
		PlatNomor:      req.PlatNomor, // Optional for QR-based sessions
		IsManualRecord: false,
		CheckinTime:    checkinTime, // Use GMT+7 timezone
		PaymentStatus:  entities.PaymentStatusPending,
		SessionStatus:  entities.SessionStatusActive,
	}

	if err := u.sessionRepo.Create(session); err != nil {
		return nil, errors.New("failed to create parking session")
	}

	return &entities.CheckinResponse{
		SessionID:   session.ID,
		CheckinTime: session.CheckinTime,
		Area:        jukir.Area.Name,
		HourlyRate:  jukir.Area.HourlyRate,
	}, nil
}

func (u *parkingUsecase) Checkout(req *entities.CheckoutRequest) (*entities.CheckoutResponse, error) {
	var session *entities.ParkingSession
	var err error

	// Priority: session_id > plat_nomor > qr_token
	if req.SessionID != nil && *req.SessionID != 0 {
		session, err = u.sessionRepo.GetByID(*req.SessionID)
		if err != nil {
			return nil, errors.New("session not found")
		}
		if session.SessionStatus == entities.SessionStatusCompleted {
			return nil, errors.New("session already completed")
		}
	} else if req.PlatNomor != nil && *req.PlatNomor != "" {
		session, err = u.sessionRepo.GetActiveByPlatNomor(*req.PlatNomor)
		if err != nil {
			return nil, errors.New("no active parking session found for this license plate")
		}
	} else {
		session, err = u.sessionRepo.GetActiveByQRToken(req.QRToken)
		if err != nil {
			return nil, errors.New("no active parking session found for this QR code")
		}
	}

	// Get jukir by QR token
	jukir, err := u.jukirRepo.GetByQRToken(req.QRToken)
	if err != nil {
		return nil, errors.New("invalid QR code")
	}

	// Ensure the QR belongs to the same parking area
	if session.AreaID != jukir.AreaID {
		return nil, errors.New("QR code belongs to a different parking area")
	}

	// Verify it's the same jukir who performed the check-in
	if session.JukirID == nil || *session.JukirID != jukir.ID {
		return nil, errors.New("QR code does not match the check-in location")
	}

	// Optional GPS verification
	if req.Latitude != nil && req.Longitude != nil {
		if err := u.ensureWithinArea(*req.Latitude, *req.Longitude, jukir.Area); err != nil {
			return nil, err
		}
	}

	// Calculate duration and cost (FLAT RATE, not per hour)
	checkoutTime := nowGMT7()
	duration := int(checkoutTime.Sub(session.CheckinTime).Minutes())
	if duration < 0 {
		duration = 0 // Handle edge case
	}
	// Biaya parkir adalah FLAT RATE (bukan per jam)
	totalCost := session.Area.HourlyRate // Flat rate, bukan hourly rate

	// Update session
	session.CheckoutTime = &checkoutTime
	session.Duration = &duration
	session.TotalCost = &totalCost
	session.SessionStatus = entities.SessionStatusPendingPayment

	if err := u.sessionRepo.Update(session); err != nil {
		return nil, errors.New("failed to update parking session")
	}

	// Create payment record
	payment := &entities.Payment{
		SessionID:     session.ID,
		Amount:        totalCost,
		PaymentMethod: entities.PaymentMethodCash,
		Status:        entities.PaymentStatusPending,
	}

	if err := u.paymentRepo.Create(payment); err != nil {
		return nil, errors.New("failed to create payment record")
	}

	// Notify jukir about checkout via SSE
	if session.JukirID != nil {
		platNomor := ""
		if session.PlatNomor != nil {
			platNomor = *session.PlatNomor
		}

		eventData := SessionUpdateEvent{
			SessionID:    session.ID,
			PlatNomor:    platNomor,
			VehicleType:  string(session.VehicleType),
			OldStatus:    string(entities.SessionStatusActive),
			NewStatus:    string(entities.SessionStatusPendingPayment),
			TotalCost:    totalCost,
			CheckoutTime: checkoutTime.Format(time.RFC3339),
			CheckinTime:  session.CheckinTime.Format(time.RFC3339),
		}

		u.eventManager.NotifyJukir(*session.JukirID, EventSessionUpdate, eventData)
	}

	return &entities.CheckoutResponse{
		SessionID:     session.ID,
		CheckoutTime:  checkoutTime,
		Duration:      duration,
		TotalCost:     totalCost,
		PaymentStatus: string(entities.PaymentStatusPending),
	}, nil
}

func (u *parkingUsecase) GetActiveSession(qrToken string) (*entities.ActiveSessionResponse, error) {
	session, err := u.sessionRepo.GetActiveByQRToken(qrToken)
	if err != nil {
		return nil, errors.New("no active parking session found for this QR code")
	}

	// Calculate duration (handle negative if checkin_time is in future)
	durationMinutes := int(nowGMT7().Sub(session.CheckinTime).Minutes())
	if durationMinutes < 0 {
		durationMinutes = 0
	}
	// Biaya parkir adalah FLAT RATE (bukan per jam)
	currentCost := session.Area.HourlyRate // Flat rate, bukan hourly rate

	return &entities.ActiveSessionResponse{
		SessionID:   session.ID,
		CheckinTime: session.CheckinTime,
		Area:        session.Area.Name,
		HourlyRate:  session.Area.HourlyRate, // Ini sebenarnya flat rate
		Duration:    durationMinutes,
		CurrentCost: currentCost, // Flat rate
	}, nil
}

func (u *parkingUsecase) GetActiveSessionByID(sessionID uint) (*entities.ActiveSessionResponse, error) {
	session, err := u.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	if session.SessionStatus != entities.SessionStatusActive {
		return nil, errors.New("no active parking session found for this session ID")
	}

	durationMinutes := int(nowGMT7().Sub(session.CheckinTime).Minutes())
	if durationMinutes < 0 {
		durationMinutes = 0
	}

	currentCost := session.Area.HourlyRate // Flat rate

	return &entities.ActiveSessionResponse{
		SessionID:   session.ID,
		CheckinTime: session.CheckinTime,
		Area:        session.Area.Name,
		HourlyRate:  session.Area.HourlyRate, // Flat rate label retained for compatibility
		Duration:    durationMinutes,
		CurrentCost: currentCost,
	}, nil
}

func (u *parkingUsecase) GetHistoryByPlatNomor(platNomor string, limit, offset int) (*entities.SessionHistoryResponse, error) {
	sessions, count, err := u.sessionRepo.GetHistoryByPlatNomor(platNomor, limit, offset)
	if err != nil {
		return nil, errors.New("failed to get parking history")
	}

	return &entities.SessionHistoryResponse{
		Sessions: sessions,
		Count:    count,
	}, nil
}

func (u *parkingUsecase) GetHistoryBySession(sessionID uint) (*entities.ParkingSession, error) {
	session, err := u.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (u *parkingUsecase) GetHistoryBySessionIDs(sessionIDs []uint) ([]entities.ParkingSession, error) {
	if len(sessionIDs) == 0 {
		return []entities.ParkingSession{}, nil
	}

	var sessions []entities.ParkingSession
	for _, sessionID := range sessionIDs {
		session, err := u.sessionRepo.GetByID(sessionID)
		if err != nil {
			// Skip invalid session IDs, continue with others
			continue
		}
		sessions = append(sessions, *session)
	}

	return sessions, nil
}

// calculateDistance calculates the distance between two coordinates using Haversine formula
func (u *parkingUsecase) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance
}

func (u *parkingUsecase) ensureWithinArea(lat, lng float64, area entities.ParkingArea) error {
	distance := u.calculateDistance(lat, lng, area.Latitude, area.Longitude)
	if distance > defaultLocationToleranceKM {
		return fmt.Errorf("you must be within %.0f meters of the parking area", defaultLocationToleranceKM*1000)
	}
	return nil
}

func (u *parkingUsecase) GetSessionByID(sessionID uint) (*entities.ParkingSession, error) {
	return u.sessionRepo.GetByID(sessionID)
}

func (u *parkingUsecase) ManualCheckin(jukirID uint, req *entities.ManualCheckinRequest) (*entities.ManualCheckinResponse, error) {
	// Get jukir info
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	// Check if jukir is active
	if jukir.Status != entities.JukirStatusActive {
		return nil, errors.New("jukir is not active")
	}

	// Ensure waktu_masuk is in GMT+7 timezone
	gmt7Loc := getGMT7Location()
	checkinTime := req.WaktuMasuk.In(gmt7Loc)

	// Validate GPS coordinates (manual check-in requires location confirmation)
	if req.Latitude == nil || req.Longitude == nil {
		return nil, errors.New("latitude and longitude are required for manual check-in")
	}

	if err := u.ensureWithinArea(*req.Latitude, *req.Longitude, jukir.Area); err != nil {
		return nil, err
	}

	// Create manual parking session
	session := &entities.ParkingSession{
		JukirID:        &jukir.ID,
		AreaID:         jukir.AreaID,
		VehicleType:    req.VehicleType,
		PlatNomor:      &req.PlatNomor,
		IsManualRecord: true,
		CheckinTime:    checkinTime,
		PaymentStatus:  entities.PaymentStatusPending,
		SessionStatus:  entities.SessionStatusActive,
	}

	if err := u.sessionRepo.Create(session); err != nil {
		return nil, errors.New("failed to create manual parking session")
	}

	platNomor := ""
	if session.PlatNomor != nil {
		platNomor = *session.PlatNomor
	}

	return &entities.ManualCheckinResponse{
		SessionID:   session.ID,
		PlatNomor:   platNomor,
		VehicleType: string(session.VehicleType),
		WaktuMasuk:  session.CheckinTime,
		Area:        jukir.Area.Name,
		ParkingCost: jukir.Area.HourlyRate, // Use area's hourly rate
	}, nil
}

func (u *parkingUsecase) ManualCheckout(jukirID uint, req *entities.ManualCheckoutRequest) (*entities.ManualCheckoutResponse, error) {
	// Get session
	session, err := u.sessionRepo.GetByID(req.SessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	// Verify session belongs to this jukir
	if session.JukirID == nil || *session.JukirID != jukirID {
		return nil, errors.New("session does not belong to this jukir")
	}

	// Check if session is active
	if session.SessionStatus != entities.SessionStatusActive {
		return nil, errors.New("session is not active")
	}

	// Check if it's a manual record
	if !session.IsManualRecord {
		return nil, errors.New("session is not a manual record")
	}

	// Validate GPS coordinates
	if req.Latitude == nil || req.Longitude == nil {
		return nil, errors.New("latitude and longitude are required for manual checkout")
	}

	if err := u.ensureWithinArea(*req.Latitude, *req.Longitude, session.Area); err != nil {
		return nil, err
	}

	gmt7Loc := getGMT7Location()
	checkoutTime := req.WaktuKeluar.In(gmt7Loc)

	// Calculate duration and cost (FLAT RATE, not per hour)
	duration := int(checkoutTime.Sub(session.CheckinTime).Minutes())
	if duration < 0 {
		duration = 0 // Handle edge case
	}
	// Biaya parkir adalah FLAT RATE (bukan per jam)
	totalCost := session.Area.HourlyRate // Flat rate, bukan hourly rate

	// For manual checkout, payment is automatically confirmed (no pending payment step)
	confirmedAt := nowGMT7()

	// Update session - directly mark as completed since manual checkout means payment is already received
	session.CheckoutTime = &checkoutTime
	session.Duration = &duration
	session.TotalCost = &totalCost
	session.SessionStatus = entities.SessionStatusCompleted
	session.PaymentStatus = entities.PaymentStatusPaid

	if err := u.sessionRepo.Update(session); err != nil {
		return nil, errors.New("failed to update manual parking session")
	}

	// Create payment record - directly mark as paid (no pending confirmation needed for manual records)
	payment := &entities.Payment{
		SessionID:     session.ID,
		Amount:        totalCost,
		PaymentMethod: entities.PaymentMethodCash,
		Status:        entities.PaymentStatusPaid,
		ConfirmedBy:   &jukirID,
		ConfirmedAt:   &confirmedAt,
	}

	if err := u.paymentRepo.Create(payment); err != nil {
		return nil, errors.New("failed to create payment record")
	}

	platNomor := ""
	if session.PlatNomor != nil {
		platNomor = *session.PlatNomor
	}

	// Notify jukir about manual checkout via SSE
	eventData := SessionUpdateEvent{
		SessionID:    session.ID,
		PlatNomor:    platNomor,
		VehicleType:  string(session.VehicleType),
		OldStatus:    string(entities.SessionStatusActive),
		NewStatus:    string(entities.SessionStatusCompleted),
		TotalCost:    totalCost,
		CheckoutTime: checkoutTime.Format(time.RFC3339),
		CheckinTime:  session.CheckinTime.Format(time.RFC3339),
	}

	u.eventManager.NotifyJukir(jukirID, EventSessionUpdate, eventData)

	// Also notify payment confirmation
	paymentEventData := PaymentConfirmedEvent{
		SessionID:     session.ID,
		PaymentID:     payment.ID,
		Amount:        totalCost,
		PaymentMethod: string(entities.PaymentMethodCash),
		ConfirmedBy:   fmt.Sprintf("%d", jukirID),
		ConfirmedAt:   confirmedAt.Format(time.RFC3339),
	}
	u.eventManager.NotifyJukir(jukirID, EventPaymentConfirmed, paymentEventData)

	return &entities.ManualCheckoutResponse{
		SessionID:     session.ID,
		PlatNomor:     platNomor,
		VehicleType:   string(session.VehicleType),
		WaktuMasuk:    session.CheckinTime,
		WaktuKeluar:   checkoutTime,
		Duration:      duration,
		TotalCost:     totalCost,
		PaymentStatus: string(entities.PaymentStatusPaid),
	}, nil
}
