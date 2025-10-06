package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"errors"
	"math"
	"time"
)

type ParkingUsecase interface {
	GetNearbyAreas(req *entities.NearbyAreasRequest) (*entities.NearbyAreasResponse, error)
	Checkin(req *entities.CheckinRequest) (*entities.CheckinResponse, error)
	Checkout(req *entities.CheckoutRequest) (*entities.CheckoutResponse, error)
	GetActiveSession(qrToken string) (*entities.ActiveSessionResponse, error)
	GetHistoryByPlatNomor(platNomor string, limit, offset int) (*entities.SessionHistoryResponse, error)
	ManualCheckin(jukirID uint, req *entities.ManualCheckinRequest) (*entities.ManualCheckinResponse, error)
	ManualCheckout(jukirID uint, req *entities.ManualCheckoutRequest) (*entities.ManualCheckoutResponse, error)
}

type parkingUsecase struct {
	sessionRepo repository.ParkingSessionRepository
	areaRepo    repository.ParkingAreaRepository
	userRepo    repository.UserRepository
	jukirRepo   repository.JukirRepository
	paymentRepo repository.PaymentRepository
}

func NewParkingUsecase(sessionRepo repository.ParkingSessionRepository, areaRepo repository.ParkingAreaRepository, userRepo repository.UserRepository, jukirRepo repository.JukirRepository, paymentRepo repository.PaymentRepository) ParkingUsecase {
	return &parkingUsecase{
		sessionRepo: sessionRepo,
		areaRepo:    areaRepo,
		userRepo:    userRepo,
		jukirRepo:   jukirRepo,
		paymentRepo: paymentRepo,
	}
}

func (u *parkingUsecase) GetNearbyAreas(req *entities.NearbyAreasRequest) (*entities.NearbyAreasResponse, error) {
	radius := req.Radius
	if radius == 0 {
		radius = 1.0 // Default 1km radius
	}

	areas, err := u.areaRepo.GetNearbyAreas(req.Latitude, req.Longitude, radius)
	if err != nil {
		return nil, errors.New("failed to get nearby areas")
	}

	// Filter areas by actual distance (more accurate than bounding box)
	var filteredAreas []entities.ParkingArea
	for _, area := range areas {
		distance := u.calculateDistance(req.Latitude, req.Longitude, area.Latitude, area.Longitude)
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
	// Check if there's already an active session for this QR token
	activeSession, err := u.sessionRepo.GetActiveByQRToken(req.QRToken)
	if err == nil && activeSession != nil {
		return nil, errors.New("there is already an active parking session for this QR code")
	}

	// Get jukir by QR token
	jukir, err := u.jukirRepo.GetByQRToken(req.QRToken)
	if err != nil {
		return nil, errors.New("invalid QR code")
	}

	// Check if jukir is active
	if jukir.Status != entities.JukirStatusActive {
		return nil, errors.New("jukir is not active")
	}

	// Verify GPS location (within 50m radius)
	distance := u.calculateDistance(req.Latitude, req.Longitude, jukir.Area.Latitude, jukir.Area.Longitude)
	if distance > 0.05 { // 50 meters
		return nil, errors.New("you must be within 50 meters of the parking area")
	}

	// Create parking session
	session := &entities.ParkingSession{
		JukirID:        &jukir.ID,
		AreaID:         jukir.AreaID,
		VehicleType:    req.VehicleType,
		PlatNomor:      req.PlatNomor, // Optional for QR-based sessions
		IsManualRecord: false,
		CheckinTime:    time.Now(),
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

	// Get active session - check by license plate if provided, otherwise by QR token
	if req.PlatNomor != nil && *req.PlatNomor != "" {
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

	// Verify it's the same jukir
	if session.JukirID == nil || *session.JukirID != jukir.ID {
		return nil, errors.New("QR code does not match the check-in location")
	}

	// Verify GPS location
	distance := u.calculateDistance(req.Latitude, req.Longitude, jukir.Area.Latitude, jukir.Area.Longitude)
	if distance > 0.05 { // 50 meters
		return nil, errors.New("you must be within 50 meters of the parking area")
	}

	// Calculate duration and cost based on area's hourly rate
	checkoutTime := time.Now()
	duration := int(checkoutTime.Sub(session.CheckinTime).Minutes())
	totalCost := session.Area.HourlyRate // Use area's hourly rate

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

	// Calculate current cost based on area's hourly rate
	duration := int(time.Since(session.CheckinTime).Minutes())
	currentCost := session.Area.HourlyRate // Use area's hourly rate

	return &entities.ActiveSessionResponse{
		SessionID:   session.ID,
		CheckinTime: session.CheckinTime,
		Area:        session.Area.Name,
		HourlyRate:  session.Area.HourlyRate,
		Duration:    duration,
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

	// Create manual parking session
	session := &entities.ParkingSession{
		JukirID:        &jukir.ID,
		AreaID:         jukir.AreaID,
		VehicleType:    req.VehicleType,
		PlatNomor:      &req.PlatNomor,
		IsManualRecord: true,
		CheckinTime:    req.WaktuMasuk,
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

	// Calculate duration and cost based on area's hourly rate
	duration := int(req.WaktuKeluar.Sub(session.CheckinTime).Minutes())
	totalCost := session.Area.HourlyRate // Use area's hourly rate

	// Update session
	session.CheckoutTime = &req.WaktuKeluar
	session.Duration = &duration
	session.TotalCost = &totalCost
	session.SessionStatus = entities.SessionStatusPendingPayment

	if err := u.sessionRepo.Update(session); err != nil {
		return nil, errors.New("failed to update manual parking session")
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

	platNomor := ""
	if session.PlatNomor != nil {
		platNomor = *session.PlatNomor
	}

	return &entities.ManualCheckoutResponse{
		SessionID:     session.ID,
		PlatNomor:     platNomor,
		VehicleType:   string(session.VehicleType),
		WaktuMasuk:    session.CheckinTime,
		WaktuKeluar:   req.WaktuKeluar,
		Duration:      duration,
		TotalCost:     totalCost,
		PaymentStatus: string(entities.PaymentStatusPending),
	}, nil
}
