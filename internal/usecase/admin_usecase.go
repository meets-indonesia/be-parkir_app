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
	GetJukirsWithRevenue(limit, offset int) ([]map[string]interface{}, int64, error)
	CreateJukir(req *entities.CreateJukirRequest) (*entities.Jukir, error)
	UpdateJukirStatus(jukirID uint, req *entities.UpdateJukirRequest) (*entities.Jukir, error)
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

	// Get all jukirs with status breakdown
	totalJukirs, _, err := u.jukirRepo.List(0, 0)
	if err != nil {
		return nil, errors.New("failed to get total jukirs")
	}

	// Count jukirs by status
	activeJukirs := 0
	inactiveJukirs := 0
	for _, jukir := range totalJukirs {
		if jukir.Status == entities.JukirStatusActive {
			activeJukirs++
		} else if jukir.Status == entities.JukirStatusInactive {
			inactiveJukirs++
		}
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

	// Calculate vehicles in and out
	vehiclesIn := 0
	vehiclesOut := 0
	for _, session := range todaySessions {
		// Check if session has checkout time (vehicle left)
		if session.CheckoutTime != nil {
			vehiclesOut++
		}
		// Count all check-ins
		vehiclesIn++
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

	// Calculate estimated revenue (assuming all active sessions will be paid)
	allAreasList, _ := u.areaRepo.GetActiveAreas()
	var estimatedRevenue float64
	for _, area := range allAreasList {
		sessions, _ := u.sessionRepo.GetSessionsByArea(area.ID, startOfDay, endOfDay)
		for _, session := range sessions {
			if session.SessionStatus == entities.SessionStatusActive && session.CheckoutTime != nil && session.Duration != nil {
				// Calculate estimated cost
				minutes := float64(*session.Duration)
				hours := minutes / 60.0
				estimatedRevenue += area.HourlyRate * hours
			} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
				estimatedRevenue += *session.TotalCost
			}
		}
	}

	// Get revenue for last 7 days for chart (oldest to newest)
	last7DaysRevenue := make([]map[string]interface{}, 7)
	weekdays := []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"}

	for i := 0; i < 7; i++ {
		daysAgo := 6 - i // 6 (oldest) to 0 (today)
		day := today.AddDate(0, 0, -daysAgo)
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		dayEnd := dayStart.Add(24 * time.Hour)

		dayRevenue, _ := u.paymentRepo.GetRevenueByDateRange(dayStart, dayEnd)

		last7DaysRevenue[i] = map[string]interface{}{
			"day":     weekdays[day.Weekday()],
			"date":    day.Format("2006-01-02"),
			"revenue": dayRevenue,
		}
	}

	return map[string]interface{}{
		"total_users":        len(totalUsers),
		"total_jukirs":       len(totalJukirs),
		"total_areas":        len(totalAreas),
		"today_sessions":     len(todaySessions),
		"vehicles_in":        vehiclesIn,
		"vehicles_out":       vehiclesOut,
		"active_sessions":    activeSessions,
		"pending_payments":   pendingPayments,
		"today_revenue":      totalRevenue,
		"estimated_revenue":  estimatedRevenue,
		"last_7days_revenue": last7DaysRevenue,
		"jukir_status": map[string]interface{}{
			"active":   activeJukirs,
			"inactive": inactiveJukirs,
		},
	}, nil
}

func (u *adminUsecase) GetJukirs(limit, offset int) ([]entities.Jukir, int64, error) {
	jukirs, count, err := u.jukirRepo.List(limit, offset)
	if err != nil {
		return nil, 0, errors.New("failed to get jukirs")
	}
	return jukirs, count, nil
}

// GetJukirsWithRevenue returns jukirs with their today's revenue
func (u *adminUsecase) GetJukirsWithRevenue(limit, offset int) ([]map[string]interface{}, int64, error) {
	jukirs, count, err := u.jukirRepo.List(limit, offset)
	if err != nil {
		return nil, 0, errors.New("failed to get jukirs")
	}

	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	result := make([]map[string]interface{}, 0)

	for _, jukir := range jukirs {
		// Get today's revenue for this jukir
		revenue, err := u.paymentRepo.GetJukirDailyRevenue(jukir.ID, time.Now())
		if err != nil {
			revenue = 0
		}

		// Get today's sessions for this jukir's area to count transactions
		sessions, _ := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startOfDay, endOfDay)

		result = append(result, map[string]interface{}{
			"id":         jukir.ID,
			"jukir_code": jukir.JukirCode,
			"user_id":    jukir.UserID,
			"area_id":    jukir.AreaID,
			"status":     jukir.Status,
			"qr_token":   jukir.QRToken,
			"created_at": jukir.CreatedAt,
			"updated_at": jukir.UpdatedAt,
			"user": map[string]interface{}{
				"id":    jukir.User.ID,
				"name":  jukir.User.Name,
				"email": jukir.User.Email,
				"phone": jukir.User.Phone,
				"role":  jukir.User.Role,
			},
			"area": map[string]interface{}{
				"id":          jukir.Area.ID,
				"name":        jukir.Area.Name,
				"address":     jukir.Area.Address,
				"latitude":    jukir.Area.Latitude,
				"longitude":   jukir.Area.Longitude,
				"hourly_rate": jukir.Area.HourlyRate,
				"status":      jukir.Area.Status,
			},
			"revenue":  revenue,
			"sessions": len(sessions),
		})
	}

	return result, count, nil
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

func (u *adminUsecase) UpdateJukirStatus(jukirID uint, req *entities.UpdateJukirRequest) (*entities.Jukir, error) {
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	// Update fields if provided
	if req.JukirCode != nil {
		jukir.JukirCode = *req.JukirCode
	}
	if req.AreaID != nil {
		jukir.AreaID = *req.AreaID
	}
	if req.Status != nil {
		jukir.Status = *req.Status
	}

	if err := u.jukirRepo.Update(jukir); err != nil {
		return nil, errors.New("failed to update jukir status")
	}

	return jukir, nil
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
