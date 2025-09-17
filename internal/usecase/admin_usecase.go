package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"errors"
	"time"
)

type AdminUsecase interface {
	GetOverview() (map[string]interface{}, error)
	GetJukirs(limit, offset int) ([]entities.Jukir, int64, error)
	CreateJukir(req *entities.CreateJukirRequest) (*entities.Jukir, error)
	ApproveJukir(jukirID uint) error
	GetReports(startDate, endDate time.Time, areaID *uint) (map[string]interface{}, error)
	GetAllSessions(limit, offset int, filters map[string]interface{}) ([]entities.ParkingSession, int64, error)
	CreateParkingArea(req *entities.CreateParkingAreaRequest) (*entities.ParkingArea, error)
	UpdateParkingArea(areaID uint, req *entities.UpdateParkingAreaRequest) (*entities.ParkingArea, error)
}

type adminUsecase struct {
	userRepo    repository.UserRepository
	jukirRepo   repository.JukirRepository
	areaRepo    repository.ParkingAreaRepository
	sessionRepo repository.ParkingSessionRepository
	paymentRepo repository.PaymentRepository
}

func NewAdminUsecase(userRepo repository.UserRepository, jukirRepo repository.JukirRepository, areaRepo repository.ParkingAreaRepository, sessionRepo repository.ParkingSessionRepository, paymentRepo repository.PaymentRepository) AdminUsecase {
	return &adminUsecase{
		userRepo:    userRepo,
		jukirRepo:   jukirRepo,
		areaRepo:    areaRepo,
		sessionRepo: sessionRepo,
		paymentRepo: paymentRepo,
	}
}

func (u *adminUsecase) GetOverview() (map[string]interface{}, error) {
	// Get total users
	totalUsers, _, err := u.userRepo.List(0, 0)
	if err != nil {
		return nil, errors.New("failed to get total users")
	}

	// Get total jukirs
	totalJukirs, _, err := u.jukirRepo.List(0, 0)
	if err != nil {
		return nil, errors.New("failed to get total jukirs")
	}

	// Get total areas
	totalAreas, _, err := u.areaRepo.List(0, 0)
	if err != nil {
		return nil, errors.New("failed to get total areas")
	}

	// Get today's sessions
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get all areas for today's sessions
	allAreas, err := u.areaRepo.GetActiveAreas()
	if err != nil {
		return nil, errors.New("failed to get areas")
	}

	var todaySessions []entities.ParkingSession
	for _, area := range allAreas {
		sessions, err := u.sessionRepo.GetSessionsByArea(area.ID, startOfDay, endOfDay)
		if err != nil {
			continue
		}
		todaySessions = append(todaySessions, sessions...)
	}

	// Get total revenue
	totalRevenue, err := u.paymentRepo.GetRevenueByDateRange(startOfDay, endOfDay)
	if err != nil {
		return nil, errors.New("failed to get total revenue")
	}

	// Count active sessions
	activeSessions := 0
	for _, session := range todaySessions {
		if session.SessionStatus == entities.SessionStatusActive {
			activeSessions++
		}
	}

	// Count pending payments
	pendingPayments := 0
	for _, session := range todaySessions {
		if session.SessionStatus == entities.SessionStatusPendingPayment {
			pendingPayments++
		}
	}

	return map[string]interface{}{
		"total_users":      len(totalUsers),
		"total_jukirs":     len(totalJukirs),
		"total_areas":      len(totalAreas),
		"today_sessions":   len(todaySessions),
		"active_sessions":  activeSessions,
		"pending_payments": pendingPayments,
		"today_revenue":    totalRevenue,
	}, nil
}

func (u *adminUsecase) GetJukirs(limit, offset int) ([]entities.Jukir, int64, error) {
	jukirs, count, err := u.jukirRepo.List(limit, offset)
	if err != nil {
		return nil, 0, errors.New("failed to get jukirs")
	}
	return jukirs, count, nil
}

func (u *adminUsecase) CreateJukir(req *entities.CreateJukirRequest) (*entities.Jukir, error) {
	// Check if user exists
	user, err := u.userRepo.GetByID(req.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user is already a jukir
	existingJukir, err := u.jukirRepo.GetByUserID(req.UserID)
	if err == nil && existingJukir != nil {
		return nil, errors.New("user is already a jukir")
	}

	// Check if jukir code is already taken
	_, err = u.jukirRepo.GetByJukirCode(req.JukirCode)
	if err == nil {
		return nil, errors.New("jukir code already taken")
	}

	// Check if area exists
	area, err := u.areaRepo.GetByID(req.AreaID)
	if err != nil {
		return nil, errors.New("parking area not found")
	}

	// Generate QR token (in production, use a proper QR generation library)
	qrToken := "QR_" + req.JukirCode + "_" + time.Now().Format("20060102150405")

	// Create jukir
	jukir := &entities.Jukir{
		UserID:    req.UserID,
		JukirCode: req.JukirCode,
		AreaID:    req.AreaID,
		QRToken:   qrToken,
		Status:    entities.JukirStatusPending,
	}

	if err := u.jukirRepo.Create(jukir); err != nil {
		return nil, errors.New("failed to create jukir")
	}

	// Load relations
	jukir.User = *user
	jukir.Area = *area

	return jukir, nil
}

func (u *adminUsecase) ApproveJukir(jukirID uint) error {
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return errors.New("jukir not found")
	}

	jukir.Status = entities.JukirStatusActive
	if err := u.jukirRepo.Update(jukir); err != nil {
		return errors.New("failed to approve jukir")
	}

	return nil
}

func (u *adminUsecase) GetReports(startDate, endDate time.Time, areaID *uint) (map[string]interface{}, error) {
	var sessions []entities.ParkingSession
	var err error

	if areaID != nil {
		sessions, err = u.sessionRepo.GetSessionsByArea(*areaID, startDate, endDate)
	} else {
		// Get sessions from all areas
		allAreas, err := u.areaRepo.GetActiveAreas()
		if err != nil {
			return nil, errors.New("failed to get areas")
		}

		for _, area := range allAreas {
			areaSessions, err := u.sessionRepo.GetSessionsByArea(area.ID, startDate, endDate)
			if err != nil {
				continue
			}
			sessions = append(sessions, areaSessions...)
		}
	}

	if err != nil {
		return nil, errors.New("failed to get sessions")
	}

	// Calculate metrics
	totalSessions := len(sessions)
	var totalRevenue float64
	var completedSessions int
	var activeSessions int
	var pendingPayments int

	for _, session := range sessions {
		switch session.SessionStatus {
		case entities.SessionStatusCompleted:
			completedSessions++
			if session.TotalCost != nil {
				totalRevenue += *session.TotalCost
			}
		case entities.SessionStatusActive:
			activeSessions++
		case entities.SessionStatusPendingPayment:
			pendingPayments++
		}
	}

	return map[string]interface{}{
		"total_sessions":     totalSessions,
		"completed_sessions": completedSessions,
		"active_sessions":    activeSessions,
		"pending_payments":   pendingPayments,
		"total_revenue":      totalRevenue,
		"start_date":         startDate.Format("2006-01-02"),
		"end_date":           endDate.Format("2006-01-02"),
	}, nil
}

func (u *adminUsecase) GetAllSessions(limit, offset int, filters map[string]interface{}) ([]entities.ParkingSession, int64, error) {
	sessions, count, err := u.sessionRepo.GetAllSessions(limit, offset, filters)
	if err != nil {
		return nil, 0, errors.New("failed to get sessions")
	}
	return sessions, count, nil
}

func (u *adminUsecase) CreateParkingArea(req *entities.CreateParkingAreaRequest) (*entities.ParkingArea, error) {
	area := &entities.ParkingArea{
		Name:       req.Name,
		Address:    req.Address,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		HourlyRate: req.HourlyRate,
		Status:     entities.AreaStatusActive,
	}

	if err := u.areaRepo.Create(area); err != nil {
		return nil, errors.New("failed to create parking area")
	}

	return area, nil
}

func (u *adminUsecase) UpdateParkingArea(areaID uint, req *entities.UpdateParkingAreaRequest) (*entities.ParkingArea, error) {
	area, err := u.areaRepo.GetByID(areaID)
	if err != nil {
		return nil, errors.New("parking area not found")
	}

	// Update fields if provided
	if req.Name != nil {
		area.Name = *req.Name
	}
	if req.Address != nil {
		area.Address = *req.Address
	}
	if req.Latitude != nil {
		area.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		area.Longitude = *req.Longitude
	}
	if req.HourlyRate != nil {
		area.HourlyRate = *req.HourlyRate
	}
	if req.Status != nil {
		area.Status = *req.Status
	}

	if err := u.areaRepo.Update(area); err != nil {
		return nil, errors.New("failed to update parking area")
	}

	return area, nil
}
