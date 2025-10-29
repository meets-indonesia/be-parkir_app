package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AdminUsecase interface {
	GetOverview(vehicleType *string, startTime, endTime *time.Time) (map[string]interface{}, error)
	GetJukirs(limit, offset int) ([]entities.Jukir, int64, error)
	GetJukirsWithRevenue(limit, offset int, vehicleType *string, dateRange *string) ([]map[string]interface{}, int64, error)
	GetVehicleStatistics(startTime, endTime *time.Time, vehicleType *string, regional *string) (map[string]interface{}, error)
	GetTotalRevenue(dateRange *string, vehicleType *string) (map[string]interface{}, error)
	GetJukirsListWithRevenue(dateRange *string, vehicleType *string, includeRevenue *bool, status *string) ([]map[string]interface{}, error)
	GetAllJukirsListWithRevenue(dateRange *string) ([]map[string]interface{}, int64, error)
	GetJukirByID(jukirID uint, dateRange *string) (map[string]interface{}, error)
	GetChartDataDetailed(startTime, endTime *time.Time, vehicleType *string, regional *string) ([]map[string]interface{}, error)
	GetParkingAreaStatistics(regional *string) (map[string]interface{}, error)
	GetJukirStatistics(regional *string) (map[string]interface{}, error)
	GetAllJukirsRevenue(startTime, endTime *time.Time, regional *string) ([]entities.JukirRevenueResponse, error)
	AddManualRevenue(req *entities.JukirRevenueRequest) (*entities.JukirRevenueResponse, error)
	GetParkingAreas() ([]map[string]interface{}, error)
	GetParkingAreaDetail(areaID uint) (map[string]interface{}, error)
	GetParkingAreaStatus(areaID uint) (map[string]interface{}, error)
	GetAreaTransactions(areaID uint, limit, offset int) ([]map[string]interface{}, int64, error)
	GetRevenueTable(limit, offset int, areaID *uint) ([]map[string]interface{}, int64, error)
	CreateJukir(req *entities.CreateJukirRequest) (*entities.CreateJukirResponse, error)
	UpdateJukirStatus(jukirID uint, req *entities.UpdateJukirRequest) (*entities.Jukir, error)
	UpdateJukir(jukirID uint, req *entities.UpdateJukirRequest) (*entities.Jukir, error)
	DeleteJukir(jukirID uint) error
	GetReports(startDate, endDate time.Time, areaID *uint) (map[string]interface{}, error)
	GetAllSessions(limit, offset int, filters map[string]interface{}) ([]entities.ParkingSession, int64, error)
	CreateParkingArea(req *entities.CreateParkingAreaRequest) (*entities.ParkingArea, error)
	UpdateParkingArea(areaID uint, req *entities.UpdateParkingAreaRequest) (*entities.ParkingArea, error)
	DeleteParkingArea(areaID uint) error
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

// getDateRange is a temporary helper for methods that still use old date range format
func getDateRange(dateRange string, now time.Time) (time.Time, time.Time) {
	switch dateRange {
	case "minggu_ini":
		weekday := int(now.Weekday()) - 1
		if weekday < 0 {
			weekday = 6
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-weekday, 0, 0, 0, 0, now.Location())
		return start, now
	case "bulan_ini":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, now
	case "tahun_ini":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return start, now
	default:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	}
}

// parseDateRange parses start_date and end_date strings, returns nil if both are empty
func parseDateRange(startDateStr, endDateStr string) (*time.Time, *time.Time, error) {
	if startDateStr == "" && endDateStr == "" {
		return nil, nil, nil
	}

	if startDateStr == "" || endDateStr == "" {
		return nil, nil, errors.New("both start_date and end_date are required")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, nil, errors.New("invalid start_date format. Use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, nil, errors.New("invalid end_date format. Use YYYY-MM-DD")
	}

	// Set end date to end of day
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	return &startDate, &endDate, nil
}

// getPeriods returns array of period data based on date range with actual and estimated revenue
func getPeriods(dateRange string, now time.Time, paymentRepo repository.PaymentRepository, sessionRepo repository.ParkingSessionRepository, areaRepo repository.ParkingAreaRepository, regional *string) []map[string]interface{} {
	switch dateRange {
	case "bulan_ini": // Last 7 months
		periods := make([]map[string]interface{}, 7)
		months := []string{"Jul", "Agu", "Sep", "Okt", "Nov", "Des", "Jan"}
		for i := 0; i < 7; i++ {
			monthsAgo := 6 - i
			period := now.AddDate(0, -monthsAgo, 0)
			start := time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, now.Location())
			end := start.AddDate(0, 1, 0).Add(-time.Second)

			actualRevenue, _ := paymentRepo.GetRevenueByDateRange(start, end)
			estimatedRevenue := calculateEstimatedRevenue(sessionRepo, areaRepo, start, end, regional)

			periods[i] = map[string]interface{}{
				"period":            months[period.Month()-1],
				"date":              period.Format("2006-01"),
				"actual_revenue":    actualRevenue,
				"estimated_revenue": estimatedRevenue,
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

			actualRevenue, _ := paymentRepo.GetRevenueByDateRange(start, end)
			estimatedRevenue := calculateEstimatedRevenue(sessionRepo, areaRepo, start, end, regional)

			periods[i] = map[string]interface{}{
				"period":            fmt.Sprintf("Minggu %d", weeksAgo+1),
				"date":              start.Format("2006-01-02"),
				"actual_revenue":    actualRevenue,
				"estimated_revenue": estimatedRevenue,
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

			actualRevenue, _ := paymentRepo.GetRevenueByDateRange(start, end)
			estimatedRevenue := calculateEstimatedRevenue(sessionRepo, areaRepo, start, end, regional)

			periods[i] = map[string]interface{}{
				"period":            weekdays[day.Weekday()],
				"date":              day.Format("2006-01-02"),
				"actual_revenue":    actualRevenue,
				"estimated_revenue": estimatedRevenue,
			}
		}
		return periods
	}
}

// calculateEstimatedRevenue calculates estimated revenue from active sessions
func calculateEstimatedRevenue(sessionRepo repository.ParkingSessionRepository, areaRepo repository.ParkingAreaRepository, start, end time.Time, regional *string) float64 {
	areas, _ := areaRepo.GetActiveAreas()
	estimatedRevenue := 0.0

	for _, area := range areas {
		// Filter by regional if specified
		if regional != nil && *regional != "" && area.Regional != *regional {
			continue
		}

		sessions, _ := sessionRepo.GetSessionsByArea(area.ID, start, end)
		for _, session := range sessions {
			if session.SessionStatus == entities.SessionStatusActive && session.CheckoutTime == nil {
				// Active session - estimate based on duration so far
				minutes := int(time.Since(session.CheckinTime).Minutes())
				hours := float64(minutes) / 60.0
				estimatedRevenue += area.HourlyRate * hours
			} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
				estimatedRevenue += *session.TotalCost
			}
		}
	}

	return estimatedRevenue
}

func (u *adminUsecase) GetOverview(vehicleType *string, startTime, endTime *time.Time) (map[string]interface{}, error) {
	// Use provided time range or default to today
	var actualStart, actualEnd time.Time
	if startTime != nil && endTime != nil {
		actualStart = *startTime
		actualEnd = *endTime
	} else {
		now := time.Now()
		actualStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		actualEnd = now
	}

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
		areaSessions, err := u.sessionRepo.GetSessionsByArea(area.ID, actualStart, actualEnd)
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

	// Get chart data based on date range - using default minggu_ini for chart
	now := time.Now()
	chartData := getPeriods("minggu_ini", now, u.paymentRepo, u.sessionRepo, u.areaRepo, nil)

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
				"id":       jukir.User.ID,
				"name":     jukir.User.Name,
				"username": jukir.User.Email, // Use email field as username
				"phone":    jukir.User.Phone,
				"role":     jukir.User.Role,
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

func (u *adminUsecase) CreateJukir(req *entities.CreateJukirRequest) (*entities.CreateJukirResponse, error) {
	// Check if area exists
	area, err := u.areaRepo.GetByID(req.AreaID)
	if err != nil {
		return nil, errors.New("parking area not found")
	}

	// Generate jukir code
	jukirCode := generateJukirCode(area.Name, time.Now())

	// Check if jukir code is already taken (retry if needed)
	for i := 0; i < 10; i++ {
		_, err = u.jukirRepo.GetByJukirCode(jukirCode)
		if err != nil {
			break // Code is available
		}
		jukirCode = generateJukirCode(area.Name, time.Now())
	}

	// Generate simple username (lowercase jukir code)
	username := strings.ToLower(jukirCode) // Use jukir code as username

	// Generate username (username)
	email := username

	// Generate simple password (4 random digits)
	password := generateSimplePassword(4)

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user with role jukir
	user := &entities.User{
		Name:     req.Name,
		Email:    email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Role:     entities.RoleJukir,
		Status:   entities.UserStatusActive,
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Generate QR token
	qrToken := "QR_" + jukirCode + "_" + time.Now().Format("20060102150405")

	// Determine status
	status := entities.JukirStatusPending
	if req.Status != nil {
		status = *req.Status
	}

	// Create jukir
	jukir := &entities.Jukir{
		UserID:    user.ID,
		JukirCode: jukirCode,
		AreaID:    req.AreaID,
		QRToken:   qrToken,
		Status:    status,
	}

	if err := u.jukirRepo.Create(jukir); err != nil {
		// Rollback: delete user if jukir creation fails
		u.userRepo.Delete(user.ID)
		return nil, errors.New("failed to create jukir")
	}

	// Load relations
	jukir.User = *user
	jukir.Area = *area

	return &entities.CreateJukirResponse{
		Jukir:    *jukir,
		Username: email,
		Password: password,
	}, nil
}

// generateJukirCode generates a unique jukir code based on area and timestamp
func generateJukirCode(areaName string, timestamp time.Time) string {
	// Get first 3 letters of area name, uppercase
	areaPrefix := strings.ToUpper(areaName[:min(3, len(areaName))])
	// Get last 4 digits of timestamp
	timeSuffix := timestamp.Format("1504") // HHMM format
	return areaPrefix + timeSuffix
}

// generatePassword generates a random password (not used anymore)
func generatePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// generateSimplePassword generates a simple numeric password
func generateSimplePassword(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (u *adminUsecase) GetAllJukirsRevenue(startTime, endTime *time.Time, regional *string) ([]entities.JukirRevenueResponse, error) {
	// Get all jukirs
	jukirs, _, err := u.jukirRepo.List(1000, 0)
	if err != nil {
		return nil, errors.New("failed to get jukirs")
	}

	// Use provided time range or default to today
	var actualStart, actualEnd time.Time
	if startTime != nil && endTime != nil {
		actualStart = *startTime
		actualEnd = *endTime
	} else {
		now := time.Now()
		actualStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		actualEnd = now
	}

	result := []entities.JukirRevenueResponse{}

	for _, jukir := range jukirs {
		// Filter by regional if specified
		if regional != nil && *regional != "" && jukir.Area.Regional != *regional {
			continue
		}
		// Get sessions for this jukir within date range
		allSessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, actualStart, actualEnd)
		if err != nil {
			continue
		}

		// Filter by jukir ID
		var jukirSessions []entities.ParkingSession
		for _, session := range allSessions {
			if session.JukirID != nil && *session.JukirID == jukir.ID {
				jukirSessions = append(jukirSessions, session)
			}
		}

		// Calculate revenue
		totalRevenue := 0.0
		for _, session := range jukirSessions {
			if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
				totalRevenue += *session.TotalCost
			}
		}

		result = append(result, entities.JukirRevenueResponse{
			ID:               jukir.ID,
			JukirName:        jukir.User.Name,
			ActualRevenue:    totalRevenue,
			EstimatedRevenue: totalRevenue,
			TotalRevenue:     totalRevenue,
			Date:             actualStart.Format("2006-01-02"),
		})
	}

	return result, nil
}

func (u *adminUsecase) GetVehicleStatistics(startTime, endTime *time.Time, vehicleType *string, regional *string) (map[string]interface{}, error) {
	// Use provided time range or default to today
	var actualStart, actualEnd time.Time
	if startTime != nil && endTime != nil {
		actualStart = *startTime
		actualEnd = *endTime
	} else {
		now := time.Now()
		actualStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		actualEnd = now
	}

	allAreas, err := u.areaRepo.GetActiveAreas()
	if err != nil {
		return nil, errors.New("failed to get areas")
	}

	var sessions []entities.ParkingSession
	for _, area := range allAreas {
		// Filter by regional if specified
		if regional != nil && *regional != "" && area.Regional != *regional {
			continue
		}
		areaSessions, err := u.sessionRepo.GetSessionsByArea(area.ID, actualStart, actualEnd)
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

	vehiclesIn := 0
	vehiclesOut := 0
	vehiclesInMobil := 0
	vehiclesOutMobil := 0
	vehiclesInMotor := 0
	vehiclesOutMotor := 0

	for _, session := range sessions {
		vehiclesIn++

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

		if session.CheckoutTime != nil {
			vehiclesOut++
		}
	}

	return map[string]interface{}{
		"total_in":  vehiclesIn,
		"total_out": vehiclesOut,
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
	}, nil
}

func (u *adminUsecase) GetTotalRevenue(dateRange *string, vehicleType *string) (map[string]interface{}, error) {
	now := time.Now()
	dateRangeStr := "hari_ini"
	if dateRange != nil && *dateRange != "" {
		dateRangeStr = *dateRange
	}

	startTime, endTime := getDateRange(dateRangeStr, now)

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

	actualRevenue := 0.0
	estimatedRevenue := 0.0

	allAreasList, _ := u.areaRepo.GetActiveAreas()
	areaMap := make(map[uint]entities.ParkingArea)
	for _, area := range allAreasList {
		areaMap[area.ID] = area
	}

	for _, session := range sessions {
		// Actual revenue - from paid sessions
		if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
			actualRevenue += *session.TotalCost
		}

		// Estimated revenue
		if session.SessionStatus == entities.SessionStatusActive && session.CheckoutTime == nil {
			area := areaMap[session.AreaID]
			minutes := int(time.Since(session.CheckinTime).Minutes())
			hours := float64(minutes) / 60.0
			estimatedRevenue += area.HourlyRate * hours
		} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
			estimatedRevenue += *session.TotalCost
		}
	}

	return map[string]interface{}{
		"actual_revenue":    actualRevenue,
		"estimated_revenue": estimatedRevenue,
		"total_revenue":     actualRevenue + estimatedRevenue,
		"date_range":        dateRangeStr,
	}, nil
}

func (u *adminUsecase) GetJukirsListWithRevenue(dateRange *string, vehicleType *string, includeRevenue *bool, status *string) ([]map[string]interface{}, error) {
	// Get all jukirs
	jukirs, _, err := u.jukirRepo.List(1000, 0)
	if err != nil {
		return nil, errors.New("failed to get jukirs")
	}

	// Filter by status if provided
	if status != nil && *status != "" {
		filteredJukirs := []entities.Jukir{}
		for _, jukir := range jukirs {
			if string(jukir.Status) == *status {
				filteredJukirs = append(filteredJukirs, jukir)
			}
		}
		jukirs = filteredJukirs
	}

	now := time.Now()
	dateRangeStr := "hari_ini"
	if dateRange != nil && *dateRange != "" {
		dateRangeStr = *dateRange
	}

	startTime, endTime := getDateRange(dateRangeStr, now)

	result := []map[string]interface{}{}
	for _, jukir := range jukirs {
		allSessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startTime, endTime)
		if err != nil {
			continue
		}

		// Filter by jukir ID and vehicle type
		var jukirSessions []entities.ParkingSession
		for _, session := range allSessions {
			if session.JukirID != nil && *session.JukirID == jukir.ID {
				if vehicleType == nil || *vehicleType == "" {
					jukirSessions = append(jukirSessions, session)
				} else if string(session.VehicleType) == *vehicleType {
					jukirSessions = append(jukirSessions, session)
				}
			}
		}

		// Calculate revenues
		actualRevenue := 0.0
		estimatedRevenue := 0.0

		areaMap := make(map[uint]entities.ParkingArea)
		area, _ := u.areaRepo.GetByID(jukir.AreaID)
		areaMap[jukir.AreaID] = *area

		for _, session := range jukirSessions {
			if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
				actualRevenue += *session.TotalCost
			}

			if session.SessionStatus == entities.SessionStatusActive && session.CheckoutTime == nil {
				area := areaMap[session.AreaID]
				minutes := int(time.Since(session.CheckinTime).Minutes())
				hours := float64(minutes) / 60.0
				estimatedRevenue += area.HourlyRate * hours
			} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
				estimatedRevenue += *session.TotalCost
			}
		}

		// Build response
		item := map[string]interface{}{
			"id":     jukir.ID,
			"name":   jukir.User.Name,
			"status": string(jukir.Status),
			"area": map[string]interface{}{
				"id":   jukir.Area.ID,
				"name": jukir.Area.Name,
			},
			"jukir_code": jukir.JukirCode,
		}

		// Add revenue only if includeRevenue is true
		if includeRevenue != nil && *includeRevenue {
			item["actual_revenue"] = actualRevenue
			item["estimated_revenue"] = estimatedRevenue
			item["total_revenue"] = actualRevenue + estimatedRevenue
			item["date"] = startTime.Format("2006-01-02")
		}

		result = append(result, item)
	}

	return result, nil
}

func (u *adminUsecase) GetJukirByID(jukirID uint, dateRange *string) (map[string]interface{}, error) {
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	response := map[string]interface{}{
		"id":         jukir.ID,
		"name":       jukir.User.Name,
		"status":     string(jukir.Status),
		"jukir_code": jukir.JukirCode,
		"qr_token":   jukir.QRToken,
		"user": map[string]interface{}{
			"id":       jukir.User.ID,
			"name":     jukir.User.Name,
			"username": jukir.User.Email, // Use email field as username
		},
		"area": map[string]interface{}{
			"id":      jukir.Area.ID,
			"name":    jukir.Area.Name,
			"address": jukir.Area.Address,
		},
		"created_at": jukir.CreatedAt,
		"updated_at": jukir.UpdatedAt,
	}

	// Calculate revenue if dateRange is provided
	if dateRange != nil && *dateRange != "" {
		now := time.Now()
		startTime, endTime := getDateRange(*dateRange, now)

		allSessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startTime, endTime)
		if err == nil {
			// Filter by jukir ID
			var jukirSessions []entities.ParkingSession
			for _, session := range allSessions {
				if session.JukirID != nil && *session.JukirID == jukir.ID {
					jukirSessions = append(jukirSessions, session)
				}
			}

			actualRevenue := 0.0
			estimatedRevenue := 0.0

			areaMap := make(map[uint]entities.ParkingArea)
			area, _ := u.areaRepo.GetByID(jukir.AreaID)
			areaMap[jukir.AreaID] = *area

			for _, session := range jukirSessions {
				if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
					actualRevenue += *session.TotalCost
				}

				if session.SessionStatus == entities.SessionStatusActive && session.CheckoutTime == nil {
					area := areaMap[session.AreaID]
					minutes := int(time.Since(session.CheckinTime).Minutes())
					hours := float64(minutes) / 60.0
					estimatedRevenue += area.HourlyRate * hours
				} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
					estimatedRevenue += *session.TotalCost
				}
			}

			// If minggu_ini, show breakdown per day (7 days)
			if *dateRange == "minggu_ini" {
				revenueDays := []map[string]interface{}{}
				weekdays := []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"}

				for i := 0; i < 7; i++ {
					daysAgo := 6 - i
					day := now.AddDate(0, 0, -daysAgo)
					start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
					end := start.Add(24 * time.Hour)

					daySessions, _ := u.sessionRepo.GetSessionsByArea(jukir.AreaID, start, end)
					var jukirDaySessions []entities.ParkingSession
					for _, session := range daySessions {
						if session.JukirID != nil && *session.JukirID == jukir.ID {
							jukirDaySessions = append(jukirDaySessions, session)
						}
					}

					dayActual := 0.0
					dayEstimated := 0.0

					for _, session := range jukirDaySessions {
						if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
							dayActual += *session.TotalCost
						}

						if session.SessionStatus == entities.SessionStatusActive && session.CheckoutTime == nil {
							area := areaMap[session.AreaID]
							minutes := int(time.Since(session.CheckinTime).Minutes())
							hours := float64(minutes) / 60.0
							dayEstimated += area.HourlyRate * hours
						} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
							dayEstimated += *session.TotalCost
						}
					}

					revenueDays = append(revenueDays, map[string]interface{}{
						"day":               weekdays[day.Weekday()],
						"date":              day.Format("2006-01-02"),
						"actual_revenue":    dayActual,
						"estimated_revenue": dayEstimated,
						"total_revenue":     dayActual + dayEstimated,
					})
				}

				response["revenue"] = map[string]interface{}{
					"actual_revenue":    actualRevenue,
					"estimated_revenue": estimatedRevenue,
					"total_revenue":     actualRevenue + estimatedRevenue,
					"date_range":        *dateRange,
					"breakdown":         revenueDays,
				}
			} else {
				response["revenue"] = map[string]interface{}{
					"actual_revenue":    actualRevenue,
					"estimated_revenue": estimatedRevenue,
					"total_revenue":     actualRevenue + estimatedRevenue,
					"date_range":        *dateRange,
				}
			}
		}
	}

	return response, nil
}

func (u *adminUsecase) GetChartDataDetailed(startTime, endTime *time.Time, vehicleType *string, regional *string) ([]map[string]interface{}, error) {
	// Use provided time range or default to this week
	var actualStart, actualEnd time.Time
	if startTime != nil && endTime != nil {
		actualStart = *startTime
		actualEnd = *endTime
	} else {
		now := time.Now()
		// Default to this week (Monday to Sunday)
		weekday := int(now.Weekday()) - 1
		if weekday < 0 {
			weekday = 6
		}
		actualStart = time.Date(now.Year(), now.Month(), now.Day()-weekday, 0, 0, 0, 0, now.Location())
		actualEnd = now
	}

	// Calculate duration in days to determine chart period
	duration := actualEnd.Sub(actualStart)
	days := int(duration.Hours() / 24)

	// Determine chart period based on date range duration
	var chartPeriod string
	if days <= 7 {
		chartPeriod = "minggu_ini" // Show daily data for week view
	} else if days <= 30 {
		chartPeriod = "minggu_ini" // Show daily data
	} else {
		chartPeriod = "bulan_ini" // Show monthly data
	}

	now := time.Now()
	// Get chart data with regional filter
	chartData := getPeriods(chartPeriod, now, u.paymentRepo, u.sessionRepo, u.areaRepo, regional)

	// Calculate summary for the period
	mingguActual, _ := u.paymentRepo.GetRevenueByDateRange(actualStart, actualEnd)
	mingguEstimated := calculateEstimatedRevenue(u.sessionRepo, u.areaRepo, actualStart, actualEnd, regional)

	// Add summary to first item or create new structure
	result := []map[string]interface{}{
		{
			"summary": map[string]interface{}{
				"period": map[string]interface{}{
					"actual_revenue":    mingguActual,
					"estimated_revenue": mingguEstimated,
					"start_date":        actualStart.Format("2006-01-02"),
					"end_date":          actualEnd.Format("2006-01-02"),
				},
			},
			"chart_data": chartData,
		},
	}

	return result, nil
}

func (u *adminUsecase) GetParkingAreaStatistics(regional *string) (map[string]interface{}, error) {
	// Get all areas
	areas, _, err := u.areaRepo.List(1000, 0)
	if err != nil {
		return nil, errors.New("failed to get parking areas")
	}

	// Count by status with regional filter
	activeCount := 0
	inactiveCount := 0
	maintenanceCount := 0
	totalCount := 0

	for _, area := range areas {
		// Filter by regional if specified
		if regional != nil && *regional != "" && area.Regional != *regional {
			continue
		}

		totalCount++
		switch area.Status {
		case entities.AreaStatusActive:
			activeCount++
		case entities.AreaStatusInactive:
			inactiveCount++
		case entities.AreaStatusMaintenance:
			maintenanceCount++
		}
	}

	return map[string]interface{}{
		"total":       totalCount,
		"active":      activeCount,
		"inactive":    inactiveCount,
		"maintenance": maintenanceCount,
	}, nil
}

func (u *adminUsecase) GetJukirStatistics(regional *string) (map[string]interface{}, error) {
	// Get all jukirs
	jukirs, _, err := u.jukirRepo.List(1000, 0)
	if err != nil {
		return nil, errors.New("failed to get jukirs")
	}

	// Count by status with regional filter
	activeCount := 0
	inactiveCount := 0

	for _, jukir := range jukirs {
		// Filter by regional if specified
		if regional != nil && *regional != "" && jukir.Area.Regional != *regional {
			continue
		}

		switch jukir.Status {
		case entities.JukirStatusActive:
			activeCount++
		case entities.JukirStatusInactive:
			inactiveCount++
		}
	}

	// Calculate total after filter
	total := activeCount + inactiveCount

	return map[string]interface{}{
		"total":    total,
		"active":   activeCount,
		"inactive": inactiveCount,
	}, nil
}

func (u *adminUsecase) AddManualRevenue(req *entities.JukirRevenueRequest) (*entities.JukirRevenueResponse, error) {
	// Get jukir
	jukir, err := u.jukirRepo.GetByID(req.JukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format. Use YYYY-MM-DD")
	}

	startTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endTime := startTime.Add(24 * time.Hour)

	// Get jukir sessions for this date
	allSessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startTime, endTime)
	if err != nil {
		return nil, errors.New("failed to get sessions")
	}

	// Filter by jukir ID
	var jukirSessions []entities.ParkingSession
	for _, session := range allSessions {
		if session.JukirID != nil && *session.JukirID == jukir.ID {
			jukirSessions = append(jukirSessions, session)
		}
	}

	// Create a dummy session for manual revenue entry (if no session exists)
	if len(jukirSessions) == 0 {
		// Create a completed session for manual revenue
		duration := 60 // 1 hour dummy
		session := &entities.ParkingSession{
			JukirID:        &jukir.ID,
			AreaID:         jukir.AreaID,
			VehicleType:    entities.VehicleTypeMobil, // Default to mobil
			IsManualRecord: true,
			CheckinTime:    startTime,
			CheckoutTime:   &endTime,
			Duration:       &duration,
			TotalCost:      &req.Amount,
			PaymentStatus:  entities.PaymentStatusPaid,
			SessionStatus:  entities.SessionStatusCompleted,
		}

		if err := u.sessionRepo.Create(session); err != nil {
			return nil, errors.New("failed to create manual session")
		}

		// Create payment record
		payment := &entities.Payment{
			SessionID:     session.ID,
			Amount:        req.Amount,
			PaymentMethod: entities.PaymentMethodCash,
			Status:        entities.PaymentStatusPaid,
		}

		if err := u.paymentRepo.Create(payment); err != nil {
			return nil, errors.New("failed to create payment record")
		}
	} else {
		// If sessions exist, add manual revenue to existing total
		// Just update the first session's total cost to include manual amount
		if len(jukirSessions) > 0 {
			session := jukirSessions[0]
			currentTotal := 0.0
			if session.TotalCost != nil {
				currentTotal = *session.TotalCost
			}
			newTotal := currentTotal + req.Amount
			session.TotalCost = &newTotal
			if err := u.sessionRepo.Update(&session); err != nil {
				return nil, errors.New("failed to update session")
			}
		}
	}

	return &entities.JukirRevenueResponse{
		ID:               jukir.ID,
		JukirName:        jukir.User.Name,
		ActualRevenue:    req.Amount,
		EstimatedRevenue: 0,
		TotalRevenue:     req.Amount,
		Date:             req.Date,
	}, nil
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

func (u *adminUsecase) UpdateJukir(jukirID uint, req *entities.UpdateJukirRequest) (*entities.Jukir, error) {
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	// Update name if provided
	if req.Name != nil {
		jukir.User.Name = *req.Name
	}

	// Update phone if provided
	if req.Phone != nil {
		jukir.User.Phone = *req.Phone
	}

	// Update user if name or phone is provided
	if req.Name != nil || req.Phone != nil {
		if err := u.userRepo.Update(&jukir.User); err != nil {
			return nil, errors.New("failed to update jukir user data")
		}
	}

	// Update fields if provided
	if req.JukirCode != nil {
		jukir.JukirCode = *req.JukirCode
	}
	if req.AreaID != nil {
		// Verify area exists
		_, err := u.areaRepo.GetByID(*req.AreaID)
		if err != nil {
			return nil, errors.New("parking area not found")
		}
		jukir.AreaID = *req.AreaID
	}
	if req.Status != nil {
		jukir.Status = *req.Status
	}

	if err := u.jukirRepo.Update(jukir); err != nil {
		return nil, errors.New("failed to update jukir")
	}

	// Reload jukir with updated user data
	updatedJukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("failed to reload jukir")
	}

	return updatedJukir, nil
}

func (u *adminUsecase) DeleteJukir(jukirID uint) error {
	// Get jukir first to ensure it exists
	_, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return errors.New("jukir not found")
	}

	// Check if jukir has active sessions
	sessions, err := u.sessionRepo.GetJukirActiveSessions(jukirID)
	if err == nil && len(sessions) > 0 {
		return errors.New("cannot delete jukir with active sessions")
	}

	// Delete jukir
	if err := u.jukirRepo.Delete(jukirID); err != nil {
		return errors.New("failed to delete jukir")
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
		Name:              req.Name,
		Address:           req.Address,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		Regional:          req.Regional,
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
	if req.Regional != nil {
		area.Regional = *req.Regional
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

	// Return clean area data without jukirs
	return &entities.ParkingArea{
		ID:                area.ID,
		Name:              area.Name,
		Address:           area.Address,
		Latitude:          area.Latitude,
		Longitude:         area.Longitude,
		Regional:          area.Regional,
		HourlyRate:        area.HourlyRate,
		Status:            area.Status,
		MaxMobil:          area.MaxMobil,
		MaxMotor:          area.MaxMotor,
		StatusOperasional: area.StatusOperasional,
		JenisArea:         area.JenisArea,
		CreatedAt:         area.CreatedAt,
		UpdatedAt:         area.UpdatedAt,
	}, nil
}

func (u *adminUsecase) DeleteParkingArea(areaID uint) error {
	// Check if area exists
	_, err := u.areaRepo.GetByID(areaID)
	if err != nil {
		return errors.New("parking area not found")
	}

	// Delete the area
	if err := u.areaRepo.Delete(areaID); err != nil {
		return errors.New("failed to delete parking area")
	}

	return nil
}

func (u *adminUsecase) GetParkingAreas() ([]map[string]interface{}, error) {
	// Get all areas using List without limit/offset to get all areas with status
	areas, _, err := u.areaRepo.List(1000, 0) // Large limit to get all
	if err != nil {
		return nil, errors.New("failed to get parking areas")
	}

	// Build response with jukirs data for each area
	result := make([]map[string]interface{}, len(areas))
	for i, area := range areas {
		areaMap := map[string]interface{}{
			"id":                 area.ID,
			"name":               area.Name,
			"address":            area.Address,
			"latitude":           area.Latitude,
			"longitude":          area.Longitude,
			"regional":           area.Regional,
			"hourly_rate":        area.HourlyRate,
			"status":             area.Status,
			"max_mobil":          area.MaxMobil,
			"max_motor":          area.MaxMotor,
			"status_operasional": area.StatusOperasional,
			"jenis_area":         area.JenisArea,
			"created_at":         area.CreatedAt,
			"updated_at":         area.UpdatedAt,
			"jukirs":             []map[string]interface{}{},
		}

		// Get jukirs for this area
		jukirs, err := u.jukirRepo.GetByAreaID(area.ID)
		if err == nil {
			jukirsData := make([]map[string]interface{}, len(jukirs))
			for j, jukir := range jukirs {
				jukirsData[j] = map[string]interface{}{
					"id":         jukir.ID,
					"jukir_code": jukir.JukirCode,
					"qr_token":   jukir.QRToken,
					"status":     jukir.Status,
					"user_id":    jukir.UserID,
					"name":       jukir.User.Name,
					"email":      jukir.User.Email,
					"phone":      jukir.User.Phone,
					"created_at": jukir.CreatedAt,
					"updated_at": jukir.UpdatedAt,
				}
			}
			areaMap["jukirs"] = jukirsData
		}

		result[i] = areaMap
	}

	return result, nil
}

func (u *adminUsecase) GetParkingAreaDetail(areaID uint) (map[string]interface{}, error) {
	// Get area
	area, err := u.areaRepo.GetByID(areaID)
	if err != nil {
		return nil, errors.New("parking area not found")
	}

	// Get jukirs for this area with preloaded user data
	jukirs, err := u.jukirRepo.GetByAreaID(areaID)
	if err != nil {
		jukirs = []entities.Jukir{} // Return empty array if error
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

	// Format area data (without jukirs nested)
	areaMap := map[string]interface{}{
		"id":                 area.ID,
		"name":               area.Name,
		"address":            area.Address,
		"latitude":           area.Latitude,
		"longitude":          area.Longitude,
		"regional":           area.Regional,
		"hourly_rate":        area.HourlyRate,
		"status":             area.Status,
		"max_mobil":          area.MaxMobil,
		"max_motor":          area.MaxMotor,
		"status_operasional": area.StatusOperasional,
		"jenis_area":         area.JenisArea,
		"created_at":         area.CreatedAt,
		"updated_at":         area.UpdatedAt,
	}

	// Format jukirs data (without nested area, only user info)
	jukirsData := make([]map[string]interface{}, len(jukirs))
	for i, jukir := range jukirs {
		jukirsData[i] = map[string]interface{}{
			"id":         jukir.ID,
			"user_id":    jukir.UserID,
			"jukir_code": jukir.JukirCode,
			"qr_token":   jukir.QRToken,
			"status":     jukir.Status,
			"created_at": jukir.CreatedAt,
			"updated_at": jukir.UpdatedAt,
			"user": map[string]interface{}{
				"id":         jukir.User.ID,
				"name":       jukir.User.Name,
				"email":      jukir.User.Email,
				"phone":      jukir.User.Phone,
				"role":       jukir.User.Role,
				"status":     jukir.User.Status,
				"created_at": jukir.User.CreatedAt,
				"updated_at": jukir.User.UpdatedAt,
			},
		}
	}

	return map[string]interface{}{
		"area":            areaMap,
		"jukirs":          jukirsData,
		"total_sessions":  totalSessions,
		"active_sessions": activeSessions,
		"total_revenue":   totalRevenue,
		"jukir_count":     len(jukirs),
	}, nil
}

func (u *adminUsecase) GetParkingAreaStatus(areaID uint) (map[string]interface{}, error) {
	// Get area
	area, err := u.areaRepo.GetByID(areaID)
	if err != nil {
		return nil, errors.New("parking area not found")
	}

	// Get all sessions for filtering
	allSessions, _, err := u.sessionRepo.GetAllSessions(1000, 0, map[string]interface{}{})
	if err != nil {
		return nil, errors.New("failed to get sessions")
	}

	// Filter active sessions for this area
	activeMobilCount := 0
	activeMotorCount := 0

	for _, session := range allSessions {
		if session.AreaID == areaID && session.SessionStatus == entities.SessionStatusActive {
			if session.VehicleType == entities.VehicleTypeMobil {
				activeMobilCount++
			} else if session.VehicleType == entities.VehicleTypeMotor {
				activeMotorCount++
			}
		}
	}

	// Calculate available slots
	var maxMobil int = 0
	var maxMotor int = 0
	if area.MaxMobil != nil {
		maxMobil = *area.MaxMobil
	}
	if area.MaxMotor != nil {
		maxMotor = *area.MaxMotor
	}

	availableMobil := maxMobil - activeMobilCount
	availableMotor := maxMotor - activeMotorCount

	// Make sure available is not negative
	if availableMobil < 0 {
		availableMobil = 0
	}
	if availableMotor < 0 {
		availableMotor = 0
	}

	return map[string]interface{}{
		"area": map[string]interface{}{
			"id":   area.ID,
			"name": area.Name,
		},
		"mobil": map[string]interface{}{
			"total":     maxMobil,
			"occupied":  activeMobilCount,
			"available": availableMobil,
			"is_full":   availableMobil == 0 && maxMobil > 0,
		},
		"motor": map[string]interface{}{
			"total":     maxMotor,
			"occupied":  activeMotorCount,
			"available": availableMotor,
			"is_full":   availableMotor == 0 && maxMotor > 0,
		},
		"total_vehicles": activeMobilCount + activeMotorCount,
		"total_capacity": maxMobil + maxMotor,
	}, nil
}

func (u *adminUsecase) GetAreaTransactions(areaID uint, limit, offset int) ([]map[string]interface{}, int64, error) {
	// Get sessions for this area
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	sessions, err := u.sessionRepo.GetSessionsByArea(areaID, startOfDay, now)
	if err != nil {
		return nil, 0, errors.New("failed to get sessions")
	}

	// Count total
	count := int64(len(sessions))

	// Get area info
	area, _ := u.areaRepo.GetByID(areaID)

	// Format sessions to match requirements
	result := []map[string]interface{}{}
	for _, session := range sessions {
		platNomor := ""
		if session.PlatNomor != nil {
			platNomor = *session.PlatNomor
		}

		duration := 0
		if session.Duration != nil {
			duration = *session.Duration
		}

		biaya := 0.0
		if session.TotalCost != nil {
			biaya = *session.TotalCost
		}

		// Get jukir info
		jukirName := ""
		if session.JukirID != nil {
			jukir, _ := u.jukirRepo.GetByID(*session.JukirID)
			if jukir != nil {
				jukirName = jukir.User.Name
			}
		}

		checkinTime := session.CheckinTime
		checkoutTime := ""
		if session.CheckoutTime != nil {
			checkoutTime = session.CheckoutTime.Format("2006-01-02 15:04:05")
		}

		areaName := ""
		if area != nil {
			areaName = area.Name
		}

		result = append(result, map[string]interface{}{
			"session_id":   session.ID,
			"plat_nomor":   platNomor,
			"area":         areaName,
			"jukir_name":   jukirName,
			"masuk":        checkinTime.Format("2006-01-02 15:04:05"),
			"keluar":       checkoutTime,
			"durasi":       duration,
			"biaya":        biaya,
			"status":       string(session.SessionStatus),
			"vehicle_type": string(session.VehicleType),
		})
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	if start > len(result) {
		return []map[string]interface{}{}, count, nil
	}

	return result[start:end], count, nil
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

// Deprecated: Use GetAllJukirsListWithRevenue instead
func (u *adminUsecase) GetJukirsWithRevenueAndDateFilter_OLD(revenue *bool, date *string) ([]map[string]interface{}, int64, error) {
	// Get all jukirs
	jukirs, _, err := u.jukirRepo.List(1000, 0)
	if err != nil {
		return nil, 0, errors.New("failed to get jukirs")
	}

	var startTime, endTime time.Time

	// If date is provided, parse and set the date range
	if date != nil && *date != "" {
		parsedDate, err := time.Parse("2006-01-02", *date)
		if err != nil {
			return nil, 0, errors.New("invalid date format. Use YYYY-MM-DD")
		}
		startTime = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
		endTime = startTime.Add(24 * time.Hour)
	} else {
		// Default to today if no date provided
		now := time.Now()
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endTime = startTime.Add(24 * time.Hour)
	}

	result := []map[string]interface{}{}

	for _, jukir := range jukirs {
		// Get sessions for this jukir on the specified date
		allSessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startTime, endTime)
		if err != nil {
			continue
		}

		// Filter by jukir ID
		var jukirSessions []entities.ParkingSession
		for _, session := range allSessions {
			if session.JukirID != nil && *session.JukirID == jukir.ID {
				jukirSessions = append(jukirSessions, session)
			}
		}

		// Calculate revenue from paid sessions
		hasRevenue := false
		totalRevenue := 0.0
		for _, session := range jukirSessions {
			if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
				totalRevenue += *session.TotalCost
				hasRevenue = true
			}
		}

		// Apply revenue filter
		if revenue != nil {
			if *revenue && !hasRevenue {
				// Filter: revenue=true but jukir has no revenue
				continue
			}
			if !*revenue && hasRevenue {
				// Filter: revenue=false but jukir has revenue
				continue
			}
		}

		// Build response
		item := map[string]interface{}{
			"id":     jukir.ID,
			"name":   jukir.User.Name,
			"status": string(jukir.Status),
			"area": map[string]interface{}{
				"id":   jukir.Area.ID,
				"name": jukir.Area.Name,
			},
			"jukir_code":     jukir.JukirCode,
			"total_sessions": len(jukirSessions),
			"total_revenue":  totalRevenue,
			"has_revenue":    hasRevenue,
			"date":           startTime.Format("2006-01-02"),
			"created_at":     jukir.CreatedAt,
			"updated_at":     jukir.UpdatedAt,
		}

		result = append(result, item)
	}

	count := int64(len(result))

	return result, count, nil
}

func (u *adminUsecase) GetAllJukirsListWithRevenue(dateRange *string) ([]map[string]interface{}, int64, error) {
	// Get all jukirs (with large limit to include all)
	jukirs, count, err := u.jukirRepo.List(1000, 0)
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

	result := []map[string]interface{}{}

	for _, jukir := range jukirs {
		// Get sessions for this jukir's area within date range
		allSessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, startTime, endTime)
		if err != nil {
			continue
		}

		// Filter by jukir ID
		var jukirSessions []entities.ParkingSession
		for _, session := range allSessions {
			if session.JukirID != nil && *session.JukirID == jukir.ID {
				jukirSessions = append(jukirSessions, session)
			}
		}

		// Calculate revenues
		actualRevenue := 0.0
		for _, session := range jukirSessions {
			if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
				actualRevenue += *session.TotalCost
			}
		}

		// Build response
		item := map[string]interface{}{
			"id":             jukir.ID,
			"name":           jukir.User.Name,
			"status":         string(jukir.Status),
			"area_id":        jukir.Area.ID,
			"area_name":      jukir.Area.Name,
			"jukir_code":     jukir.JukirCode,
			"total_sessions": len(jukirSessions),
			"total_revenue":  actualRevenue,
			"date_range":     dateRangeStr,
		}

		result = append(result, item)
	}

	return result, count, nil
}
