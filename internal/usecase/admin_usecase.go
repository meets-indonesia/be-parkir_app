package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"errors"
	"fmt"
	"time"
)

type AdminUsecase interface {
	GetOverview(vehicleType *string, dateRange *string) (map[string]interface{}, error)
	GetJukirs(limit, offset int) ([]entities.Jukir, int64, error)
	GetJukirsWithRevenue(limit, offset int, vehicleType *string, dateRange *string) ([]map[string]interface{}, int64, error)
	GetParkingAreas() ([]entities.ParkingArea, error)
	GetParkingAreaDetail(areaID uint) (map[string]interface{}, error)
	GetAreaTransactions(areaID uint, limit, offset int) ([]entities.ParkingSession, int64, error)
	GetRevenueTable(limit, offset int, areaID *uint) ([]map[string]interface{}, int64, error)
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

// getDateRange calculates start and end times based on date range string
func getDateRange(dateRange string, now time.Time) (time.Time, time.Time) {
	switch dateRange {
	case "minggu_ini": // This week (Monday to Sunday)
		weekday := int(now.Weekday()) - 1 // Monday = 0
		if weekday < 0 {
			weekday = 6 // Sunday
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-weekday, 0, 0, 0, 0, now.Location())
		return start, now
	case "bulan_ini": // This month
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, now
	case "tahun_ini": // This year
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return start, now
	default: // "hari_ini" or empty - today
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	}
}

// getPeriods returns array of period data based on date range
func getPeriods(dateRange string, now time.Time, paymentRepo repository.PaymentRepository) []map[string]interface{} {
	switch dateRange {
	case "bulan_ini": // Last 7 months
		periods := make([]map[string]interface{}, 7)
		months := []string{"Jul", "Agu", "Sep", "Okt", "Nov", "Des", "Jan"}
		for i := 0; i < 7; i++ {
			monthsAgo := 6 - i
			period := now.AddDate(0, -monthsAgo, 0)
			start := time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, now.Location())
			end := start.AddDate(0, 1, 0).Add(-time.Second)

			revenue, _ := paymentRepo.GetRevenueByDateRange(start, end)

			periods[i] = map[string]interface{}{
				"period":  months[period.Month()-1],
				"date":    period.Format("2006-01"),
				"revenue": revenue,
			}
		}
		return periods
	case "tahun_ini": // Last 7 weeks
		periods := make([]map[string]interface{}, 7)
		for i := 0; i < 7; i++ {
			weeksAgo := 6 - i
			period := now.AddDate(0, 0, -7*weeksAgo)
			weekStart := period.AddDate(0, 0, -int(period.Weekday()))
			start := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, now.Location())
			end := start.AddDate(0, 0, 7)

			revenue, _ := paymentRepo.GetRevenueByDateRange(start, end)

			periods[i] = map[string]interface{}{
				"period":  fmt.Sprintf("Minggu %d", weeksAgo+1),
				"date":    start.Format("2006-01-02"),
				"revenue": revenue,
			}
		}
		return periods
	default: // hari_ini or minggu_ini - last 7 days
		periods := make([]map[string]interface{}, 7)
		weekdays := []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"}
		for i := 0; i < 7; i++ {
			daysAgo := 6 - i
			day := now.AddDate(0, 0, -daysAgo)
			start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
			end := start.Add(24 * time.Hour)

			revenue, _ := paymentRepo.GetRevenueByDateRange(start, end)

			periods[i] = map[string]interface{}{
				"period":  weekdays[day.Weekday()],
				"date":    day.Format("2006-01-02"),
				"revenue": revenue,
			}
		}
		return periods
	}
}

func (u *adminUsecase) GetOverview(vehicleType *string, dateRange *string) (map[string]interface{}, error) {
	// Determine date range
	now := time.Now()
	dateRangeStr := "hari_ini" // default
	if dateRange != nil && *dateRange != "" {
		dateRangeStr = *dateRange
	}

	startTime, endTime := getDateRange(dateRangeStr, now)

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

	// Get all areas for sessions
	allAreas, err := u.areaRepo.GetActiveAreas()
	if err != nil {
		return nil, errors.New("failed to get areas")
	}

	var sessions []entities.ParkingSession
	for _, area := range allAreas {
		areaSessions, err := u.sessionRepo.GetSessionsByArea(area.ID, startTime, endTime)
		if err != nil {
			continue
		}

		// Filter by vehicle type if specified
		if vehicleType != nil && *vehicleType != "" {
			for _, session := range areaSessions {
				if string(session.VehicleType) == *vehicleType {
					sessions = append(sessions, session)
				}
			}
		} else {
			sessions = append(sessions, areaSessions...)
		}
	}

	// Calculate vehicles in and out
	vehiclesIn := 0
	vehiclesOut := 0
	vehiclesInMobil := 0
	vehiclesOutMobil := 0
	vehiclesInMotor := 0
	vehiclesOutMotor := 0

	// If filtered, these will contain only the filtered type
	for _, session := range sessions {
		// Count all check-ins (will be filtered count if vehicleType filter is applied)
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

		// Check if session has checkout time (vehicle left)
		if session.CheckoutTime != nil {
			vehiclesOut++
		}
	}

	// Calculate total revenue from filtered sessions
	var totalRevenue float64
	for _, session := range sessions {
		if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
			totalRevenue += *session.TotalCost
		}
	}

	// Count active sessions
	activeSessions := 0
	for _, session := range sessions {
		if session.SessionStatus == entities.SessionStatusActive {
			activeSessions++
		}
	}

	// Count pending payments
	pendingPayments := 0
	for _, session := range sessions {
		if session.SessionStatus == entities.SessionStatusPendingPayment {
			pendingPayments++
		}
	}

	// Calculate estimated revenue from filtered sessions
	var estimatedRevenue float64
	// Get all areas to find hourly rate
	allAreasList, _ := u.areaRepo.GetActiveAreas()
	areaMap := make(map[uint]entities.ParkingArea)
	for _, area := range allAreasList {
		areaMap[area.ID] = area
	}

	for _, session := range sessions {
		if session.SessionStatus == entities.SessionStatusActive && session.CheckoutTime != nil && session.Duration != nil {
			// Get area hourly rate
			area := areaMap[session.AreaID]
			// Calculate estimated cost
			minutes := float64(*session.Duration)
			hours := minutes / 60.0
			estimatedRevenue += area.HourlyRate * hours
		} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
			estimatedRevenue += *session.TotalCost
		}
	}

	// Get chart data based on date range
	chartData := getPeriods(dateRangeStr, now, u.paymentRepo)

	return map[string]interface{}{
		"total_users":    len(totalUsers),
		"total_jukirs":   len(totalJukirs),
		"total_areas":    len(totalAreas),
		"today_sessions": len(sessions),
		"vehicles_in":    vehiclesIn,
		"vehicles_out":   vehiclesOut,
		"vehicles_by_type": map[string]interface{}{
			"mobil": map[string]interface{}{
				"in":  vehiclesInMobil,
				"out": vehiclesOutMobil,
			},
			"motor": map[string]interface{}{
				"in":  vehiclesInMotor,
				"out": vehiclesOutMotor,
			},
		},
		"active_sessions":   activeSessions,
		"pending_payments":  pendingPayments,
		"today_revenue":     totalRevenue,
		"estimated_revenue": estimatedRevenue,
		"chart_data":        chartData,
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

// GetJukirsWithRevenue returns jukirs with their revenue
func (u *adminUsecase) GetJukirsWithRevenue(limit, offset int, vehicleType *string, dateRange *string) ([]map[string]interface{}, int64, error) {
	jukirs, count, err := u.jukirRepo.List(limit, offset)
	if err != nil {
		return nil, 0, errors.New("failed to get jukirs")
	}

	// Determine date range
	now := time.Now()
	dateRangeStr := "hari_ini" // default
	if dateRange != nil && *dateRange != "" {
		dateRangeStr = *dateRange
	}

	startTime, endTime := getDateRange(dateRangeStr, now)

	result := make([]map[string]interface{}, 0)

	for _, jukir := range jukirs {
		// Get sessions for this jukir's area within date range
		sessions, _ := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startTime, endTime)

		// Filter by vehicle type if specified
		if vehicleType != nil && *vehicleType != "" {
			filteredSessions := []entities.ParkingSession{}
			for _, session := range sessions {
				if string(session.VehicleType) == *vehicleType {
					filteredSessions = append(filteredSessions, session)
				}
			}
			sessions = filteredSessions
		}

		// Calculate revenue for filtered sessions
		var revenue float64
		for _, session := range sessions {
			if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
				revenue += *session.TotalCost
			}
		}

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
		Name:              req.Name,
		Address:           req.Address,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		HourlyRate:        req.HourlyRate,
		Status:            entities.AreaStatusActive,
		MaxMobil:          req.MaxMobil,
		MaxMotor:          req.MaxMotor,
		StatusOperasional: req.StatusOperasional,
		JenisArea:         req.JenisArea,
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
	if req.MaxMobil != nil {
		area.MaxMobil = req.MaxMobil
	}
	if req.MaxMotor != nil {
		area.MaxMotor = req.MaxMotor
	}
	if req.StatusOperasional != nil {
		area.StatusOperasional = *req.StatusOperasional
	}
	if req.JenisArea != nil {
		area.JenisArea = *req.JenisArea
	}

	if err := u.areaRepo.Update(area); err != nil {
		return nil, errors.New("failed to update parking area")
	}

	return area, nil
}

func (u *adminUsecase) GetParkingAreas() ([]entities.ParkingArea, error) {
	// Get all areas using List without limit/offset to get all areas with status
	areas, _, err := u.areaRepo.List(1000, 0) // Large limit to get all
	if err != nil {
		return nil, errors.New("failed to get parking areas")
	}
	return areas, nil
}

func (u *adminUsecase) GetParkingAreaDetail(areaID uint) (map[string]interface{}, error) {
	// Get area
	area, err := u.areaRepo.GetByID(areaID)
	if err != nil {
		return nil, errors.New("parking area not found")
	}

	// Get jukirs for this area
	jukirs, _, err := u.jukirRepo.List(1000, 0)
	if err != nil {
		return nil, errors.New("failed to get jukirs")
	}

	// Filter jukirs by area
	var areaJukirs []entities.Jukir
	for _, jukir := range jukirs {
		if jukir.AreaID == areaID {
			areaJukirs = append(areaJukirs, jukir)
		}
	}

	// Count sessions for today
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	sessions, err := u.sessionRepo.GetSessionsByArea(areaID, startOfDay, now)
	if err != nil {
		return nil, errors.New("failed to get sessions")
	}

	// Calculate metrics
	totalSessions := len(sessions)
	activeSessions := 0
	totalRevenue := 0.0

	for _, session := range sessions {
		if session.SessionStatus == entities.SessionStatusActive {
			activeSessions++
		}
		if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
			totalRevenue += *session.TotalCost
		}
	}

	return map[string]interface{}{
		"area":            area,
		"jukirs":          areaJukirs,
		"total_sessions":  totalSessions,
		"active_sessions": activeSessions,
		"total_revenue":   totalRevenue,
		"jukir_count":     len(areaJukirs),
	}, nil
}

func (u *adminUsecase) GetAreaTransactions(areaID uint, limit, offset int) ([]entities.ParkingSession, int64, error) {
	// Get sessions for this area
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	sessions, err := u.sessionRepo.GetSessionsByArea(areaID, startOfDay, now)
	if err != nil {
		return nil, 0, errors.New("failed to get sessions")
	}

	// Count total
	count := int64(len(sessions))

	// Apply pagination
	start := offset
	end := offset + limit
	if end > len(sessions) {
		end = len(sessions)
	}
	if start > len(sessions) {
		return []entities.ParkingSession{}, count, nil
	}

	return sessions[start:end], count, nil
}

func (u *adminUsecase) GetRevenueTable(limit, offset int, areaID *uint) ([]map[string]interface{}, int64, error) {
	// This will return revenue data for the monitor-pendapatan page
	// Get all active areas or specific area
	var areas []entities.ParkingArea
	if areaID != nil {
		area, err := u.areaRepo.GetByID(*areaID)
		if err != nil {
			return nil, 0, errors.New("parking area not found")
		}
		areas = []entities.ParkingArea{*area}
	} else {
		activeAreas, err := u.areaRepo.GetActiveAreas()
		if err != nil {
			return nil, 0, errors.New("failed to get areas")
		}
		areas = activeAreas
	}

	// Get today's date range
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Build revenue table
	revenueTable := []map[string]interface{}{}

	for _, area := range areas {
		sessions, err := u.sessionRepo.GetSessionsByArea(area.ID, startOfDay, now)
		if err != nil {
			continue
		}

		// Calculate metrics
		totalRevenue := 0.0
		totalSessions := len(sessions)
		completedSessions := 0

		for _, session := range sessions {
			if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
				totalRevenue += *session.TotalCost
			}
			if session.CheckoutTime != nil {
				completedSessions++
			}
		}

		revenueTable = append(revenueTable, map[string]interface{}{
			"area_id":            area.ID,
			"area_name":          area.Name,
			"total_sessions":     totalSessions,
			"completed_sessions": completedSessions,
			"total_revenue":      totalRevenue,
		})
	}

	// Count total
	count := int64(len(revenueTable))

	// Apply pagination
	start := offset
	end := offset + limit
	if end > len(revenueTable) {
		end = len(revenueTable)
	}
	if start > len(revenueTable) {
		return []map[string]interface{}{}, count, nil
	}

	return revenueTable[start:end], count, nil
}
