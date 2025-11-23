package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

type AdminUsecase interface {
	GetOverview(vehicleType *string, startTime, endTime *time.Time) (map[string]interface{}, error)
	GetJukirs(limit, offset int, regional *string) ([]entities.Jukir, int64, error)
	GetJukirsWithRevenue(limit, offset int, vehicleType *string, dateRange *string) ([]map[string]interface{}, int64, error)
	GetVehicleStatistics(startTime, endTime *time.Time, vehicleType *string, regional *string) (map[string]interface{}, error)
	GetTotalRevenue(startTime, endTime *time.Time, vehicleType *string, regional *string) (map[string]interface{}, error)
	ExportRevenueReport(startTime, endTime *time.Time, regional *string) (*bytes.Buffer, error)
	GetJukirsListWithRevenue(dateRange *string, vehicleType *string, includeRevenue *bool, status *string) ([]map[string]interface{}, error)
	GetAllJukirsListWithRevenue(limit, offset int, dateRange *string) ([]map[string]interface{}, int64, error)
	GetJukirByID(jukirID uint, dateRange *string) (map[string]interface{}, error)
	GetChartDataDetailed(startTime, endTime *time.Time, vehicleType *string, regional *string) ([]map[string]interface{}, error)
	GetParkingAreaStatistics(regional *string) (map[string]interface{}, error)
	GetJukirStatistics(regional *string) (map[string]interface{}, error)
	GetAllJukirsRevenue(startTime, endTime *time.Time, regional *string) ([]entities.JukirRevenueResponse, error)
	AddManualRevenue(req *entities.JukirRevenueRequest) (*entities.JukirRevenueResponse, error)
	GetParkingAreas(regional *string) ([]map[string]interface{}, error)
	GetParkingAreaDetail(areaID uint) (map[string]interface{}, error)
	GetParkingAreaStatus(areaID uint) (map[string]interface{}, error)
	GetAreaTransactions(areaID uint, limit, offset int, startTime, endTime *time.Time) ([]map[string]interface{}, int64, error)
	GetRevenueTable(limit, offset int, areaID *uint, startTime, endTime *time.Time, regional *string) ([]map[string]interface{}, int64, error)
	CreateJukir(req *entities.CreateJukirRequest) (*entities.CreateJukirResponse, error)
	UpdateJukirStatus(jukirID uint, req *entities.UpdateJukirRequest) (*entities.Jukir, error)
	UpdateJukir(jukirID uint, req *entities.UpdateJukirRequest) (*entities.Jukir, error)
	DeleteJukir(jukirID uint) error
	GetReports(startDate, endDate time.Time, areaID *uint) (map[string]interface{}, error)
	GetAllSessions(limit, offset int, filters map[string]interface{}) ([]entities.ParkingSession, int64, error)
	CreateParkingArea(req *entities.CreateParkingAreaRequest) (*entities.ParkingArea, error)
	UpdateParkingArea(areaID uint, req *entities.UpdateParkingAreaRequest) (*entities.ParkingArea, error)
	DeleteParkingArea(areaID uint) error
	GetAreaActivity(startTime, endTime *time.Time, areaID *uint, regional *string) (map[string]interface{}, error)
	GetAreaActivityDetail(areaID uint, startTime, endTime *time.Time) (map[string]interface{}, error)
	GetJukirActivity(startTime, endTime *time.Time, jukirID *uint, regional *string) (map[string]interface{}, error)
	GetJukirActivityDetail(jukirID uint, startTime, endTime *time.Time) (map[string]interface{}, error)
	ExportAreaActivityCSV(startTime, endTime *time.Time, areaID *uint, regional *string) (*bytes.Buffer, error)
	ExportJukirActivityCSV(startTime, endTime *time.Time, jukirID *uint, regional *string) (*bytes.Buffer, error)
	ExportAreaActivityDetailXLSX(areaID uint, startTime, endTime *time.Time) (*bytes.Buffer, error)
	ExportJukirActivityDetailXLSX(jukirID uint, startTime, endTime *time.Time) (*bytes.Buffer, error)
	GetActivityLogs(jukirID *uint, areaID *uint, startTime, endTime time.Time, limit, offset int) (*entities.ActivityLogResponse, error)
	ImportAreasAndJukirsFromCSV(reader io.Reader, regional string) (map[string]interface{}, error)
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
func getPeriods(dateRange string, now time.Time, sessionRepo repository.ParkingSessionRepository, areaRepo repository.ParkingAreaRepository, regional *string) []map[string]interface{} {
	switch dateRange {
	case "bulan_ini": // Last 7 months
		periods := make([]map[string]interface{}, 7)
		months := []string{"Jul", "Agu", "Sep", "Okt", "Nov", "Des", "Jan"}
		for i := 0; i < 7; i++ {
			monthsAgo := 6 - i
			period := now.AddDate(0, -monthsAgo, 0)
			start := time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, now.Location())
			end := start.AddDate(0, 1, 0)

			actualRevenue := calculateActualRevenue(sessionRepo, areaRepo, start, end, regional)
			estimatedRevenue := calculateEstimatedRevenue(sessionRepo, areaRepo, start, end, regional)

			periods[i] = map[string]interface{}{
				"period":            months[period.Month()-1],
				"date":              period.Format("2006-01"),
				"actual_revenue":    roundCurrency(actualRevenue),
				"estimated_revenue": roundCurrency(estimatedRevenue),
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

			actualRevenue := calculateActualRevenue(sessionRepo, areaRepo, start, end, regional)
			estimatedRevenue := calculateEstimatedRevenue(sessionRepo, areaRepo, start, end, regional)

			periods[i] = map[string]interface{}{
				"period":            fmt.Sprintf("Minggu %d", weeksAgo+1),
				"date":              start.Format("2006-01-02"),
				"actual_revenue":    roundCurrency(actualRevenue),
				"estimated_revenue": roundCurrency(estimatedRevenue),
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

			actualRevenue := calculateActualRevenue(sessionRepo, areaRepo, start, end, regional)
			estimatedRevenue := calculateEstimatedRevenue(sessionRepo, areaRepo, start, end, regional)

			periods[i] = map[string]interface{}{
				"period":            weekdays[day.Weekday()],
				"date":              day.Format("2006-01-02"),
				"actual_revenue":    roundCurrency(actualRevenue),
				"estimated_revenue": roundCurrency(estimatedRevenue),
			}
		}
		return periods
	}
}

// calculateEstimatedRevenue calculates estimated revenue from active sessions
func calculateEstimatedRevenue(sessionRepo repository.ParkingSessionRepository, areaRepo repository.ParkingAreaRepository, start, end time.Time, regional *string) float64 {
	// For simplicity, estimated revenue now mirrors confirmed (actual) revenue.
	// Both metrics only increase once a payment has been confirmed.
	return calculateActualRevenue(sessionRepo, areaRepo, start, end, regional)
}

func roundCurrency(value float64) float64 {
	return math.Round(value)
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
			// Get area hourly rate based on vehicle type
			area := areaMap[session.AreaID]
			// Calculate estimated cost
			minutes := float64(*session.Duration)
			hours := minutes / 60.0
			rate := area.GetRateByVehicleType(session.VehicleType)
			estimatedRevenue += rate * hours
		} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
			estimatedRevenue += *session.TotalCost
		} else if session.IsManualRecord && session.TotalCost != nil {
			estimatedRevenue += *session.TotalCost
		}
	}

	// Get chart data based on date range - using default minggu_ini for chart
	now := time.Now()
	chartData := getPeriods("minggu_ini", now, u.sessionRepo, u.areaRepo, nil)

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
		"today_revenue":     roundCurrency(totalRevenue),
		"estimated_revenue": roundCurrency(estimatedRevenue),
		"chart_data":        chartData,
		"jukir_status": map[string]interface{}{
			"active":   activeJukirs,
			"inactive": inactiveJukirs,
		},
	}, nil
}

func (u *adminUsecase) GetJukirs(limit, offset int, regional *string) ([]entities.Jukir, int64, error) {
	jukirs, count, err := u.jukirRepo.List(1000, 0) // Get all to filter by regional
	if err != nil {
		return nil, 0, errors.New("failed to get jukirs")
	}

	// Filter by regional if specified
	if regional != nil && *regional != "" {
		filteredJukirs := []entities.Jukir{}
		for _, jukir := range jukirs {
			if jukir.Area.Regional == *regional {
				filteredJukirs = append(filteredJukirs, jukir)
			}
		}
		jukirs = filteredJukirs
		count = int64(len(filteredJukirs))
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if end > len(jukirs) {
		end = len(jukirs)
	}
	if start > len(jukirs) {
		return []entities.Jukir{}, count, nil
	}

	return jukirs[start:end], count, nil
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
				"hourly_rate_mobil": jukir.Area.HourlyRateMobil,
				"hourly_rate_motor": jukir.Area.HourlyRateMotor,
				"status":      jukir.Area.Status,
			},
			"revenue":  roundCurrency(revenue),
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
		Name:            req.Name,
		Email:           email,
		Phone:           req.Phone,
		Password:        string(hashedPassword),
		DisplayPassword: &password, // Store password for display
		Role:            entities.RoleJukir,
		Status:          entities.UserStatusActive,
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
// Only uses letters and numbers (A-Z, 0-9)
func generateJukirCode(areaName string, timestamp time.Time) string {
	// Filter area name to only letters and numbers, then get first 3 characters
	var cleanAreaName strings.Builder
	for _, char := range areaName {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			cleanAreaName.WriteRune(char)
		}
	}

	cleaned := cleanAreaName.String()
	if len(cleaned) == 0 {
		cleaned = "PAR" // Default prefix if no valid characters
	}

	// Get first 3 characters, uppercase, pad with numbers if needed
	areaPrefix := strings.ToUpper(cleaned[:min(3, len(cleaned))])
	if len(areaPrefix) < 3 {
		// Pad with numbers if less than 3 characters
		for len(areaPrefix) < 3 {
			areaPrefix += fmt.Sprintf("%d", rand.Intn(10))
		}
	}

	// Get last 4 digits of timestamp (HHMM format)
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
			if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
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

func (u *adminUsecase) GetTotalRevenue(startTime, endTime *time.Time, vehicleType *string, regional *string) (map[string]interface{}, error) {
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
			rate := area.GetRateByVehicleType(session.VehicleType)
			estimatedRevenue += rate * hours
		} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
			estimatedRevenue += *session.TotalCost
		} else if session.IsManualRecord && session.TotalCost != nil {
			estimatedRevenue += *session.TotalCost
		}
	}

	return map[string]interface{}{
		"actual_revenue":    roundCurrency(actualRevenue),
		"estimated_revenue": roundCurrency(estimatedRevenue),
		"total_revenue":     roundCurrency(actualRevenue + estimatedRevenue),
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
				rate := area.GetRateByVehicleType(session.VehicleType)
				estimatedRevenue += rate * hours
			} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
				estimatedRevenue += *session.TotalCost
			} else if session.IsManualRecord && session.TotalCost != nil {
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
			item["actual_revenue"] = roundCurrency(actualRevenue)
			item["estimated_revenue"] = roundCurrency(estimatedRevenue)
			item["total_revenue"] = roundCurrency(actualRevenue + estimatedRevenue)
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

	userData := map[string]interface{}{
		"id":       jukir.User.ID,
		"name":     jukir.User.Name,
		"username": jukir.User.Email, // Use email field as username
	}
	var displayPassword interface{}
	if jukir.User.DisplayPassword != nil {
		displayPassword = *jukir.User.DisplayPassword
		userData["password"] = *jukir.User.DisplayPassword
		userData["display_password"] = *jukir.User.DisplayPassword
	}

	response := map[string]interface{}{
		"id":               jukir.ID,
		"name":             jukir.User.Name,
		"status":           string(jukir.Status),
		"jukir_code":       jukir.JukirCode,
		"qr_token":         jukir.QRToken,
		"password":         displayPassword,
		"display_password": displayPassword,
		"user":             userData,
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
					rate := area.GetRateByVehicleType(session.VehicleType)
			estimatedRevenue += rate * hours
				} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
					estimatedRevenue += *session.TotalCost
				} else if session.IsManualRecord && session.TotalCost != nil {
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
							rate := area.GetRateByVehicleType(session.VehicleType)
							dayEstimated += rate * hours
						} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
							dayEstimated += *session.TotalCost
						} else if session.IsManualRecord && session.TotalCost != nil {
							dayEstimated += *session.TotalCost
						}
					}

					dayActualRounded := roundCurrency(dayActual)
					dayEstimatedRounded := roundCurrency(dayEstimated)

					revenueDays = append(revenueDays, map[string]interface{}{
						"day":               weekdays[day.Weekday()],
						"date":              day.Format("2006-01-02"),
						"actual_revenue":    dayActualRounded,
						"estimated_revenue": dayEstimatedRounded,
						"total_revenue":     roundCurrency(dayActual + dayEstimated),
					})
				}

				response["revenue"] = map[string]interface{}{
					"actual_revenue":    roundCurrency(actualRevenue),
					"estimated_revenue": roundCurrency(estimatedRevenue),
					"total_revenue":     roundCurrency(actualRevenue + estimatedRevenue),
					"date_range":        *dateRange,
					"breakdown":         revenueDays,
				}
			} else {
				response["revenue"] = map[string]interface{}{
					"actual_revenue":    roundCurrency(actualRevenue),
					"estimated_revenue": roundCurrency(estimatedRevenue),
					"total_revenue":     roundCurrency(actualRevenue + estimatedRevenue),
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

	// Generate chart data based on actual date range selected by user
	var chartData []map[string]interface{}

	if days <= 30 {
		// Daily breakdown for up to 30 days
		chartData = u.generateDailyChartData(actualStart, actualEnd, regional)
	} else if days <= 90 {
		// Weekly breakdown for 31-90 days
		chartData = u.generateWeeklyChartData(actualStart, actualEnd, regional)
	} else {
		// Monthly breakdown for more than 90 days
		chartData = u.generateMonthlyChartData(actualStart, actualEnd, regional)
	}

	// Calculate summary for the period
	actualRevenue := calculateActualRevenue(u.sessionRepo, u.areaRepo, actualStart, actualEnd, regional)
	estimatedRevenue := calculateEstimatedRevenue(u.sessionRepo, u.areaRepo, actualStart, actualEnd, regional)

	// Add summary to first item or create new structure
	result := []map[string]interface{}{
		{
			"summary": map[string]interface{}{
				"period": map[string]interface{}{
					"actual_revenue":    roundCurrency(actualRevenue),
					"estimated_revenue": roundCurrency(estimatedRevenue),
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
		"total_revenue":      roundCurrency(totalRevenue),
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
		HourlyRateMobil:   req.HourlyRateMobil,
		HourlyRateMotor:   req.HourlyRateMotor,
		Status:            entities.AreaStatusActive,
		MaxMobil:          req.MaxMobil,
		MaxMotor:          req.MaxMotor,
		StatusOperasional: req.StatusOperasional,
		JenisArea:         req.JenisArea,
	}

	if err := u.areaRepo.Create(area); err != nil {
		return nil, fmt.Errorf("failed to create parking area: %w", err)
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
	if req.HourlyRateMobil != nil {
		area.HourlyRateMobil = *req.HourlyRateMobil
	}
	if req.HourlyRateMotor != nil {
		area.HourlyRateMotor = *req.HourlyRateMotor
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
		HourlyRateMobil:   area.HourlyRateMobil,
		HourlyRateMotor:   area.HourlyRateMotor,
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

func (u *adminUsecase) GetParkingAreas(regional *string) ([]map[string]interface{}, error) {
	// Get all areas using List without limit/offset to get all areas with status
	areas, _, err := u.areaRepo.List(1000, 0) // Large limit to get all
	if err != nil {
		return nil, errors.New("failed to get parking areas")
	}

	// Filter by regional if specified
	if regional != nil && *regional != "" {
		filteredAreas := []entities.ParkingArea{}
		for _, area := range areas {
			if area.Regional == *regional {
				filteredAreas = append(filteredAreas, area)
			}
		}
		areas = filteredAreas
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
			"hourly_rate_mobil":  area.HourlyRateMobil,
			"hourly_rate_motor":   area.HourlyRateMotor,
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
		"image":              area.Image,
		"hourly_rate_mobil":  area.HourlyRateMobil,
		"hourly_rate_motor":  area.HourlyRateMotor,
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
		"total_revenue":   roundCurrency(totalRevenue),
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

func (u *adminUsecase) GetAreaTransactions(areaID uint, limit, offset int, startTime, endTime *time.Time) ([]map[string]interface{}, int64, error) {
	// Determine date range - use provided dates or default to today
	var startDate, endDate time.Time
	if startTime != nil && endTime != nil {
		startDate = *startTime
		// Set end date to end of day (23:59:59.999999999)
		endDate = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, endTime.Location())
	} else {
		// Default to today if no date filter provided
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = now
	}

	sessions, err := u.sessionRepo.GetSessionsByArea(areaID, startDate, endDate)
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

func (u *adminUsecase) GetRevenueTable(limit, offset int, areaID *uint, startTime, endTime *time.Time, regional *string) ([]map[string]interface{}, int64, error) {
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

	// Build revenue table
	revenueTable := []map[string]interface{}{}

	for _, area := range areas {
		// Filter by regional if specified
		if regional != nil && *regional != "" && area.Regional != *regional {
			continue
		}

		sessions, err := u.sessionRepo.GetSessionsByArea(area.ID, actualStart, actualEnd)
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
			"total_revenue":      roundCurrency(totalRevenue),
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
			"total_revenue":  roundCurrency(totalRevenue),
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

func (u *adminUsecase) GetAllJukirsListWithRevenue(limit, offset int, dateRange *string) ([]map[string]interface{}, int64, error) {
	// Normalize pagination values
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

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
			"total_revenue":  roundCurrency(actualRevenue),
			"date_range":     dateRangeStr,
		}

		result = append(result, item)
	}

	return result, count, nil
}

// generateDailyChartData generates daily breakdown chart data
func (u *adminUsecase) generateDailyChartData(start, end time.Time, regional *string) []map[string]interface{} {
	var periods []map[string]interface{}
	weekdays := []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"}

	// Iterate through each day in the range
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayStart := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		dayEnd := dayStart.Add(24 * time.Hour)

		// Don't exceed the end date
		if dayEnd.After(end) {
			dayEnd = end
		}

		actualRevenue := calculateActualRevenue(u.sessionRepo, u.areaRepo, dayStart, dayEnd, regional)
		estimatedRevenue := calculateEstimatedRevenue(u.sessionRepo, u.areaRepo, dayStart, dayEnd, regional)

		periods = append(periods, map[string]interface{}{
			"period":            weekdays[d.Weekday()],
			"date":              d.Format("2006-01-02"),
			"actual_revenue":    roundCurrency(actualRevenue),
			"estimated_revenue": roundCurrency(estimatedRevenue),
		})
	}

	return periods
}

// generateWeeklyChartData generates weekly breakdown chart data
func (u *adminUsecase) generateWeeklyChartData(start, end time.Time, regional *string) []map[string]interface{} {
	var periods []map[string]interface{}

	current := start
	weekNum := 1

	for !current.After(end) {
		// Get start of current week (Monday)
		weekday := int(current.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday becomes 7
		}
		daysToMonday := weekday - 1
		weekStart := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location()).AddDate(0, 0, -daysToMonday)

		// Calculate end of week (Sunday)
		weekEnd := weekStart.AddDate(0, 0, 7)

		// Don't exceed the end date
		if weekEnd.After(end) {
			weekEnd = end
		}

		// Only proceed if weekStart is within or before the range
		if !weekStart.After(end) {
			actualRevenue := calculateActualRevenue(u.sessionRepo, u.areaRepo, weekStart, weekEnd, regional)
			estimatedRevenue := calculateEstimatedRevenue(u.sessionRepo, u.areaRepo, weekStart, weekEnd, regional)

			periods = append(periods, map[string]interface{}{
				"period":            fmt.Sprintf("Minggu %d", weekNum),
				"date":              weekStart.Format("2006-01-02"),
				"actual_revenue":    roundCurrency(actualRevenue),
				"estimated_revenue": roundCurrency(estimatedRevenue),
			})
		}

		// Move to next week
		current = weekEnd.AddDate(0, 0, 1)
		weekNum++
	}

	return periods
}

// generateMonthlyChartData generates monthly breakdown chart data
func (u *adminUsecase) generateMonthlyChartData(start, end time.Time, regional *string) []map[string]interface{} {
	var periods []map[string]interface{}
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

	current := start

	for !current.After(end) {
		// Get the start of the month
		monthStart := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())

		// Get the end of the month
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		// Adjust monthStart if it's before our start range
		if monthStart.Before(start) {
			monthStart = start
		}

		// Don't exceed the end date
		if monthEnd.After(end) {
			monthEnd = end
		}

		actualRevenue := calculateActualRevenue(u.sessionRepo, u.areaRepo, monthStart, monthEnd, regional)
		estimatedRevenue := calculateEstimatedRevenue(u.sessionRepo, u.areaRepo, monthStart, monthEnd, regional)

		periods = append(periods, map[string]interface{}{
			"period":            monthNames[current.Month()-1],
			"date":              current.Format("2006-01"),
			"actual_revenue":    roundCurrency(actualRevenue),
			"estimated_revenue": roundCurrency(estimatedRevenue),
		})

		// Move to next month
		current = monthStart.AddDate(0, 1, 0)
	}

	return periods
}

// ExportRevenueReport exports revenue report to Excel
func (u *adminUsecase) ExportRevenueReport(startTime, endTime *time.Time, regional *string) (*bytes.Buffer, error) {
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

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Sheet 1: Revenue by Jukir
	sheet1 := "Revenue by Jukir"
	index1, err := f.NewSheet(sheet1)
	if err != nil {
		return nil, errors.New("failed to create sheet")
	}
	f.SetActiveSheet(index1)

	// Set headers for Jukir revenue
	headersJukir := []string{"No", "Jukir Name", "Area", "Regional", "Actual Revenue", "Estimated Revenue", "Total Revenue"}
	for i, header := range headersJukir {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheet1, cell, header)
	}

	// Get all jukirs
	jukirs, _, err := u.jukirRepo.List(1000, 0)
	if err != nil {
		return nil, errors.New("failed to get jukirs")
	}

	row := 2
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
				rate := area.GetRateByVehicleType(session.VehicleType)
			estimatedRevenue += rate * hours
			} else if session.SessionStatus == entities.SessionStatusPendingPayment && session.TotalCost != nil {
				estimatedRevenue += *session.TotalCost
			} else if session.IsManualRecord && session.TotalCost != nil {
				estimatedRevenue += *session.TotalCost
			}
		}

		f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), row-1)
		f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), jukir.User.Name)
		f.SetCellValue(sheet1, fmt.Sprintf("C%d", row), jukir.Area.Name)
		f.SetCellValue(sheet1, fmt.Sprintf("D%d", row), jukir.Area.Regional)
		f.SetCellValue(sheet1, fmt.Sprintf("E%d", row), actualRevenue)
		f.SetCellValue(sheet1, fmt.Sprintf("F%d", row), estimatedRevenue)
		f.SetCellValue(sheet1, fmt.Sprintf("G%d", row), actualRevenue+estimatedRevenue)
		row++
	}

	// Sheet 2: Revenue by Area
	sheet2 := "Revenue by Area"
	_, err = f.NewSheet(sheet2)
	if err != nil {
		return nil, errors.New("failed to create sheet")
	}

	// Set headers for Area revenue
	headersArea := []string{"No", "Area Name", "Regional", "Total Sessions", "Completed Sessions", "Actual Revenue", "Total Revenue"}
	for i, header := range headersArea {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheet2, cell, header)
	}

	// Get all areas
	areas, _, err := u.areaRepo.List(1000, 0)
	if err != nil {
		return nil, errors.New("failed to get areas")
	}

	rowArea := 2
	for _, area := range areas {
		// Filter by regional if specified
		if regional != nil && *regional != "" && area.Regional != *regional {
			continue
		}

		// Get sessions for this area within date range
		sessions, err := u.sessionRepo.GetSessionsByArea(area.ID, actualStart, actualEnd)
		if err != nil {
			continue
		}

		// Calculate metrics
		totalSessions := len(sessions)
		completedSessions := 0
		actualRevenue := 0.0

		for _, session := range sessions {
			if session.CheckoutTime != nil {
				completedSessions++
			}
			if session.PaymentStatus == entities.PaymentStatusPaid && session.TotalCost != nil {
				actualRevenue += *session.TotalCost
			}
		}

		f.SetCellValue(sheet2, fmt.Sprintf("A%d", rowArea), rowArea-1)
		f.SetCellValue(sheet2, fmt.Sprintf("B%d", rowArea), area.Name)
		f.SetCellValue(sheet2, fmt.Sprintf("C%d", rowArea), area.Regional)
		f.SetCellValue(sheet2, fmt.Sprintf("D%d", rowArea), totalSessions)
		f.SetCellValue(sheet2, fmt.Sprintf("E%d", rowArea), completedSessions)
		f.SetCellValue(sheet2, fmt.Sprintf("F%d", rowArea), actualRevenue)
		f.SetCellValue(sheet2, fmt.Sprintf("G%d", rowArea), actualRevenue)
		rowArea++
	}

	// Delete default Sheet1
	f.DeleteSheet("Sheet1")

	// Set first sheet as active
	f.SetActiveSheet(index1)

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, errors.New("failed to write Excel file")
	}

	return &buf, nil
}

// GetAreaActivity returns activity monitoring data for parking area
// Simple format: total masuk (checkin) and keluar (checkout) per area
func (u *adminUsecase) GetAreaActivity(startTime, endTime *time.Time, areaID *uint, regional *string) (map[string]interface{}, error) {
	// Use provided time range or default to today
	var actualStart, actualEnd time.Time
	if startTime != nil && endTime != nil {
		actualStart = *startTime
		actualEnd = *endTime
		// Set end time to end of day
		actualEnd = time.Date(actualEnd.Year(), actualEnd.Month(), actualEnd.Day(), 23, 59, 59, 0, actualEnd.Location())
	} else {
		now := time.Now()
		actualStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		actualEnd = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	}

	// Get areas to monitor
	var areas []entities.ParkingArea
	if areaID != nil {
		area, err := u.areaRepo.GetByID(*areaID)
		if err != nil {
			return nil, errors.New("parking area not found")
		}
		areas = []entities.ParkingArea{*area}
	} else {
		allAreas, _, err := u.areaRepo.List(1000, 0)
		if err != nil {
			return nil, errors.New("failed to get areas")
		}
		// Filter by regional if specified
		for _, area := range allAreas {
			if regional != nil && *regional != "" && area.Regional != *regional {
				continue
			}
			areas = append(areas, area)
		}
	}

	if len(areas) == 0 {
		return nil, errors.New("no areas found")
	}

	// Process each area
	result := make([]map[string]interface{}, 0, len(areas))
	for _, area := range areas {
		// Get all sessions for this area in date range
		sessions, err := u.sessionRepo.GetSessionsByArea(area.ID, actualStart, actualEnd)
		if err != nil {
			continue
		}

		// Count masuk (checkin) and keluar (checkout) by vehicle type
		mobilMasuk := 0
		mobilKeluar := 0
		motorMasuk := 0
		motorKeluar := 0

		for _, session := range sessions {
			if session.VehicleType == entities.VehicleTypeMobil {
				mobilMasuk++
				if session.CheckoutTime != nil {
					mobilKeluar++
				}
			} else if session.VehicleType == entities.VehicleTypeMotor {
				motorMasuk++
				if session.CheckoutTime != nil {
					motorKeluar++
				}
			}
		}

		result = append(result, map[string]interface{}{
			"area_id":   area.ID,
			"area_name": area.Name,
			"regional":  area.Regional,
			"mobil": map[string]interface{}{
				"masuk":  mobilMasuk,
				"keluar": mobilKeluar,
			},
			"motor": map[string]interface{}{
				"masuk":  motorMasuk,
				"keluar": motorKeluar,
			},
			"total_masuk":  mobilMasuk + motorMasuk,
			"total_keluar": mobilKeluar + motorKeluar,
		})
	}

	return map[string]interface{}{
		"data": result,
		"summary": map[string]interface{}{
			"start_date":  actualStart.Format("2006-01-02"),
			"end_date":    actualEnd.Format("2006-01-02"),
			"total_areas": len(result),
		},
	}, nil
}

// GetAreaActivityDetail returns detailed activity monitoring data for a specific parking area
// Breakdown per 15 minutes interval like the CSV format
func (u *adminUsecase) GetAreaActivityDetail(areaID uint, startTime, endTime *time.Time) (map[string]interface{}, error) {
	// Get area
	area, err := u.areaRepo.GetByID(areaID)
	if err != nil {
		return nil, errors.New("parking area not found")
	}

	// Use provided time range or default to today (in GMT+7)
	var actualStart, actualEnd time.Time
	if startTime != nil && endTime != nil {
		actualStart = *startTime
		actualEnd = *endTime
		// Ensure dates are in GMT+7 timezone
		gmt7Loc := getGMT7Location()
		actualStart = time.Date(actualStart.Year(), actualStart.Month(), actualStart.Day(), 0, 0, 0, 0, gmt7Loc)
		actualEnd = time.Date(actualEnd.Year(), actualEnd.Month(), actualEnd.Day(), 23, 59, 59, 999999999, gmt7Loc)
	} else {
		// Default to today in GMT+7
		now := nowGMT7()
		actualStart = dateGMT7(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0)
		actualEnd = dateGMT7(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999)
	}

	// Validate date range
	if actualEnd.Before(actualStart) {
		return nil, errors.New("end date must be after start date")
	}

	// Limit to max 7 days to prevent too many intervals
	maxDuration := 7 * 24 * time.Hour
	if actualEnd.Sub(actualStart) > maxDuration {
		actualEnd = actualStart.Add(maxDuration)
	}

	// Get all sessions for this area in date range
	sessions, err := u.sessionRepo.GetSessionsByArea(area.ID, actualStart, actualEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Create 15-minute intervals (only from 6am to 10pm)
	intervals := u.create15MinuteIntervalsWorkHours(actualStart, actualEnd)
	if len(intervals) == 0 {
		return nil, errors.New("no intervals created")
	}

	// Process sessions by vehicle type and interval
	mobilData := make([]map[string]interface{}, len(intervals))
	motorData := make([]map[string]interface{}, len(intervals))

	for j, interval := range intervals {
		intervalStart, ok1 := interval["start"].(time.Time)
		intervalEnd, ok2 := interval["end"].(time.Time)
		if !ok1 || !ok2 {
			// If interval parsing fails, create empty stats
			dateStr := ""
			if date, ok := interval["date"].(string); ok {
				dateStr = date
			}
			mobilData[j] = map[string]interface{}{
				"periode":       interval["label"],
				"date":          dateStr, // Add date field for frontend filtering
				"datang":        0,
				"berangkat":     0,
				"akumulasi":     0,
				"volume":        0,
				"indeks_parkir": "0%",
			}
			motorData[j] = map[string]interface{}{
				"periode":       interval["label"],
				"date":          dateStr, // Add date field for frontend filtering
				"datang":        0,
				"berangkat":     0,
				"akumulasi":     0,
				"volume":        0,
				"indeks_parkir": "0%",
			}
			continue
		}

		intervalMap := map[string]interface{}{
			"start": intervalStart,
			"end":   intervalEnd,
			"label": interval["label"],
		}

		mobilStats := u.calculateIntervalStats(sessions, intervalMap, entities.VehicleTypeMobil, area.MaxMobil)
		motorStats := u.calculateIntervalStats(sessions, intervalMap, entities.VehicleTypeMotor, area.MaxMotor)

		// Get date from interval for adding to stats
		dateStr := ""
		if date, ok := interval["date"].(string); ok {
			dateStr = date
		} else {
			// Fallback: extract date from interval start time
			dateStr = intervalStart.Format("2006-01-02")
		}

		// Ensure indeks_parkir is always present and add date field
		if mobilStats == nil {
			mobilStats = map[string]interface{}{
				"periode":       interval["label"],
				"date":          dateStr,
				"datang":        0,
				"berangkat":     0,
				"akumulasi":     0,
				"volume":        0,
				"indeks_parkir": "0%",
			}
		} else {
			// Add date field to existing stats
			mobilStats["date"] = dateStr
		}
		if motorStats == nil {
			motorStats = map[string]interface{}{
				"periode":       interval["label"],
				"date":          dateStr,
				"datang":        0,
				"berangkat":     0,
				"akumulasi":     0,
				"volume":        0,
				"indeks_parkir": "0%",
			}
		} else {
			// Add date field to existing stats
			motorStats["date"] = dateStr
		}

		mobilData[j] = mobilStats
		motorData[j] = motorStats
	}

	// Calculate totals
	totalMobilDatang := 0
	totalMotorDatang := 0
	for _, mobil := range mobilData {
		if mobil != nil {
			if datang, ok := mobil["datang"].(int); ok {
				totalMobilDatang += datang
			}
		}
	}
	for _, motor := range motorData {
		if motor != nil {
			if datang, ok := motor["datang"].(int); ok {
				totalMotorDatang += datang
			}
		}
	}

	return map[string]interface{}{
		"area_id":            area.ID,
		"area_name":          area.Name,
		"regional":           area.Regional,
		"kapasitas_mobil":    area.MaxMobil,
		"kapasitas_motor":    area.MaxMotor,
		"start_date":         actualStart.Format("2006-01-02"),
		"end_date":           actualEnd.Format("2006-01-02"),
		"mobil":              mobilData,
		"motor":              motorData,
		"total_mobil_datang": totalMobilDatang,
		"total_motor_datang": totalMotorDatang,
	}, nil
}

// create15MinuteIntervals creates time intervals of 15 minutes
func (u *adminUsecase) create15MinuteIntervals(start, end time.Time) []map[string]interface{} {
	var intervals []map[string]interface{}
	current := start

	// Safety limit: max 7 days = 672 intervals (7 * 24 * 4)
	maxIntervals := 672
	intervalCount := 0

	for !current.After(end) && intervalCount < maxIntervals {
		intervalEnd := current.Add(15 * time.Minute)
		if intervalEnd.After(end) {
			intervalEnd = end
		}

		intervals = append(intervals, map[string]interface{}{
			"start": current,
			"end":   intervalEnd,
			"label": fmt.Sprintf("%02d.%02d - %02d.%02d", current.Hour(), current.Minute(), intervalEnd.Hour(), intervalEnd.Minute()),
		})

		nextCurrent := intervalEnd
		intervalCount++

		// Safety check: if interval didn't advance, break to prevent infinite loop
		if nextCurrent.Equal(current) || nextCurrent.Before(current) {
			break
		}

		current = nextCurrent
	}

	return intervals
}

// create15MinuteIntervalsWorkHours creates time intervals of 15 minutes from 6am to 10pm (06:00-22:00)
func (u *adminUsecase) create15MinuteIntervalsWorkHours(start, end time.Time) []map[string]interface{} {
	var intervals []map[string]interface{}

	// Process each day in the range
	currentDate := start
	maxDays := 7
	dayCount := 0

	for dayCount < maxDays {
		// Set work hours: 6am to 10pm for this day
		workStart := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 6, 0, 0, 0, currentDate.Location())
		workEnd := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 22, 0, 0, 0, currentDate.Location())

		// Skip if work hours don't overlap with requested range
		if workEnd.Before(start) || workStart.After(end) {
			currentDate = currentDate.Add(24 * time.Hour)
			dayCount++
			if currentDate.After(end) {
				break
			}
			continue
		}

		// Adjust work hours to fit within requested range (but still within 06:00-22:00)
		if workStart.Before(start) && start.Hour() >= 6 {
			// If start is after 6am but before 10pm, use start time
			if start.Hour() < 22 {
				workStart = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute(), 0, 0, start.Location())
				// Round down to nearest 15 minutes
				workStart = workStart.Add(-time.Duration(workStart.Minute()%15) * time.Minute)
			}
		}

		if workEnd.After(end) && end.Hour() >= 6 && end.Hour() < 22 {
			workEnd = end
		}

		// Only create intervals if work hours are valid (06:00-22:00)
		if workStart.Hour() >= 6 && workStart.Hour() < 22 && workEnd.Hour() >= 6 && (workEnd.Hour() < 22 || (workEnd.Hour() == 22 && workEnd.Minute() == 0)) && !workStart.After(workEnd) {
			// Create intervals for this day's work hours
			current := workStart
			intervalCount := 0
			maxIntervalsPerDay := 64 // Max 16 hours * 4 intervals per hour

			for intervalCount < maxIntervalsPerDay {
				// Stop if we've reached or passed 10pm
				if current.Hour() >= 22 || current.After(workEnd) {
					break
				}

				intervalEnd := current.Add(15 * time.Minute)
				if intervalEnd.After(workEnd) {
					intervalEnd = workEnd
				}

				// Stop if interval end passes 10pm
				if intervalEnd.Hour() > 22 || (intervalEnd.Hour() == 22 && intervalEnd.Minute() > 0) {
					break
				}

				// Only add interval if it's within work hours (06:00-22:00)
				if current.Hour() >= 6 && current.Hour() < 22 {
					// Add date to interval for frontend filtering
					dateStr := current.Format("2006-01-02")
					intervals = append(intervals, map[string]interface{}{
						"start": current,
						"end":   intervalEnd,
						"label": fmt.Sprintf("%02d.%02d - %02d.%02d", current.Hour(), current.Minute(), intervalEnd.Hour(), intervalEnd.Minute()),
						"date":  dateStr, // Add date field for frontend filtering
					})
				}

				nextCurrent := intervalEnd
				intervalCount++

				// Safety check
				if nextCurrent.Equal(current) || nextCurrent.Before(current) {
					break
				}

				current = nextCurrent
			}
		}

		// Move to next day
		currentDate = currentDate.Add(24 * time.Hour)
		dayCount++

		// Stop if we've passed the end date
		if currentDate.After(end) {
			break
		}
	}

	return intervals
}

// calculateIntervalStats calculates statistics for a specific interval
func (u *adminUsecase) calculateIntervalStats(sessions []entities.ParkingSession, interval map[string]interface{}, vehicleType entities.VehicleType, maxCapacity *int) map[string]interface{} {
	start := interval["start"].(time.Time)
	end := interval["end"].(time.Time)

	datang := 0    // checkin in this interval
	berangkat := 0 // checkout in this interval
	akumulasi := 0 // active vehicles at the END of this interval
	volume := 0    // total cumulative checkins up to end of interval

	// Count checkins and checkouts in this interval
	for _, session := range sessions {
		if session.VehicleType != vehicleType {
			continue
		}

		// Count datang (checkin dalam interval ini)
		// Checkin time is within [start, end]
		if (session.CheckinTime.After(start) || session.CheckinTime.Equal(start)) &&
			(session.CheckinTime.Before(end) || session.CheckinTime.Equal(end)) {
			datang++
		}

		// Count berangkat (checkout dalam interval ini)
		if session.CheckoutTime != nil {
			if (session.CheckoutTime.After(start) || session.CheckoutTime.Equal(start)) &&
				(session.CheckoutTime.Before(end) || session.CheckoutTime.Equal(end)) {
				berangkat++
			}
		}

		// Calculate akumulasi: active at the END of interval
		// Session aktif jika: checkin sebelum/sama dengan end, dan (belum checkout atau checkout setelah end)
		if session.CheckinTime.Before(end) || session.CheckinTime.Equal(end) {
			if session.CheckoutTime == nil || session.CheckoutTime.After(end) {
				akumulasi++
			}
		}

		// Calculate volume: total checkins up to end of interval
		if session.CheckinTime.Before(end) || session.CheckinTime.Equal(end) {
			volume++
		}
	}

	// Calculate parking index (% of capacity)
	var indeksParkir float64
	if maxCapacity != nil && *maxCapacity > 0 {
		indeksParkir = (float64(akumulasi) / float64(*maxCapacity)) * 100
		if indeksParkir > 100 {
			indeksParkir = 100
		}
	}

	return map[string]interface{}{
		"periode":       interval["label"],
		"datang":        datang,
		"berangkat":     berangkat,
		"akumulasi":     akumulasi,
		"volume":        volume,
		"indeks_parkir": fmt.Sprintf("%.0f%%", indeksParkir),
	}
}

// GetJukirActivity returns activity monitoring data for jukir
// Simple format: total masuk (checkin) and keluar (checkout) per jukir
func (u *adminUsecase) GetJukirActivity(startTime, endTime *time.Time, jukirID *uint, regional *string) (map[string]interface{}, error) {
	// Use provided time range or default to today (in GMT+7)
	var actualStart, actualEnd time.Time
	if startTime != nil && endTime != nil {
		actualStart = *startTime
		actualEnd = *endTime
		// Ensure dates are in GMT+7 timezone
		gmt7Loc := getGMT7Location()
		actualStart = time.Date(actualStart.Year(), actualStart.Month(), actualStart.Day(), 0, 0, 0, 0, gmt7Loc)
		actualEnd = time.Date(actualEnd.Year(), actualEnd.Month(), actualEnd.Day(), 23, 59, 59, 999999999, gmt7Loc)
	} else {
		// Default to today in GMT+7
		now := nowGMT7()
		actualStart = dateGMT7(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0)
		actualEnd = dateGMT7(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999)
	}

	// Get jukirs to monitor
	var jukirs []entities.Jukir
	if jukirID != nil {
		jukir, err := u.jukirRepo.GetByID(*jukirID)
		if err != nil {
			return nil, errors.New("jukir not found")
		}
		jukirs = []entities.Jukir{*jukir}
	} else {
		allJukirs, _, err := u.jukirRepo.List(1000, 0)
		if err != nil {
			return nil, errors.New("failed to get jukirs")
		}
		// Filter by regional if specified
		for _, jukir := range allJukirs {
			if regional != nil && *regional != "" && jukir.Area.Regional != *regional {
				continue
			}
			jukirs = append(jukirs, jukir)
		}
	}

	if len(jukirs) == 0 {
		return nil, errors.New("no jukirs found")
	}

	// Process each jukir
	result := make([]map[string]interface{}, 0, len(jukirs))
	for _, jukir := range jukirs {
		// Get all sessions for this area in date range
		allSessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, actualStart, actualEnd)
		if err != nil {
			continue
		}

		// Filter sessions by jukir ID
		var sessions []entities.ParkingSession
		for _, session := range allSessions {
			if session.JukirID != nil && *session.JukirID == jukir.ID {
				sessions = append(sessions, session)
			}
		}

		// Count masuk (checkin) and keluar (checkout) by vehicle type
		mobilMasuk := 0
		mobilKeluar := 0
		motorMasuk := 0
		motorKeluar := 0

		for _, session := range sessions {
			if session.VehicleType == entities.VehicleTypeMobil {
				mobilMasuk++
				if session.CheckoutTime != nil {
					mobilKeluar++
				}
			} else if session.VehicleType == entities.VehicleTypeMotor {
				motorMasuk++
				if session.CheckoutTime != nil {
					motorKeluar++
				}
			}
		}

		// Add date field for frontend filtering (use start date as reference)
		dateStr := actualStart.Format("2006-01-02")

		result = append(result, map[string]interface{}{
			"jukir_id":   jukir.ID,
			"jukir_name": jukir.User.Name,
			"area_id":    jukir.Area.ID,
			"area_name":  jukir.Area.Name,
			"regional":   jukir.Area.Regional,
			"date":       dateStr, // Add date field for frontend filtering
			"mobil": map[string]interface{}{
				"masuk":  mobilMasuk,
				"keluar": mobilKeluar,
			},
			"motor": map[string]interface{}{
				"masuk":  motorMasuk,
				"keluar": motorKeluar,
			},
			"total_masuk":  mobilMasuk + motorMasuk,
			"total_keluar": mobilKeluar + motorKeluar,
		})
	}

	return map[string]interface{}{
		"data": result,
		"summary": map[string]interface{}{
			"start_date":   actualStart.Format("2006-01-02"),
			"end_date":     actualEnd.Format("2006-01-02"),
			"total_jukirs": len(result),
		},
	}, nil
}

// GetJukirActivityDetail returns detailed activity monitoring data for a specific jukir
// Breakdown per 15 minutes interval from 6am to 10pm like the CSV format
func (u *adminUsecase) GetJukirActivityDetail(jukirID uint, startTime, endTime *time.Time) (map[string]interface{}, error) {
	// Get jukir
	jukir, err := u.jukirRepo.GetByID(jukirID)
	if err != nil {
		return nil, errors.New("jukir not found")
	}

	// Use provided time range or default to today (in GMT+7)
	var actualStart, actualEnd time.Time
	if startTime != nil && endTime != nil {
		actualStart = *startTime
		actualEnd = *endTime
		// Ensure dates are in GMT+7 timezone
		gmt7Loc := getGMT7Location()
		actualStart = time.Date(actualStart.Year(), actualStart.Month(), actualStart.Day(), 0, 0, 0, 0, gmt7Loc)
		actualEnd = time.Date(actualEnd.Year(), actualEnd.Month(), actualEnd.Day(), 23, 59, 59, 999999999, gmt7Loc)
	} else {
		// Default to today in GMT+7
		now := nowGMT7()
		actualStart = dateGMT7(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0)
		actualEnd = dateGMT7(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999)
	}

	// Validate date range
	if actualEnd.Before(actualStart) {
		return nil, errors.New("end date must be after start date")
	}

	// Limit to max 7 days to prevent too many intervals
	maxDuration := 7 * 24 * time.Hour
	if actualEnd.Sub(actualStart) > maxDuration {
		actualEnd = actualStart.Add(maxDuration)
	}

	// Get all sessions for this area in date range
	allSessions, err := u.sessionRepo.GetSessionsByArea(jukir.AreaID, actualStart, actualEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Filter sessions by jukir ID
	var sessions []entities.ParkingSession
	for _, session := range allSessions {
		if session.JukirID != nil && *session.JukirID == jukir.ID {
			sessions = append(sessions, session)
		}
	}

	// Create 15-minute intervals (only from 6am to 10pm)
	intervals := u.create15MinuteIntervalsWorkHours(actualStart, actualEnd)
	if len(intervals) == 0 {
		return nil, errors.New("no intervals created")
	}

	// Process sessions by vehicle type and interval
	mobilData := make([]map[string]interface{}, len(intervals))
	motorData := make([]map[string]interface{}, len(intervals))

	for j, interval := range intervals {
		intervalStart, ok1 := interval["start"].(time.Time)
		intervalEnd, ok2 := interval["end"].(time.Time)
		if !ok1 || !ok2 {
			// If interval parsing fails, create empty stats
			dateStr := ""
			if date, ok := interval["date"].(string); ok {
				dateStr = date
			}
			mobilData[j] = map[string]interface{}{
				"periode":       interval["label"],
				"date":          dateStr, // Add date field for frontend filtering
				"datang":        0,
				"berangkat":     0,
				"akumulasi":     0,
				"volume":        0,
				"indeks_parkir": "0%",
			}
			motorData[j] = map[string]interface{}{
				"periode":       interval["label"],
				"date":          dateStr, // Add date field for frontend filtering
				"datang":        0,
				"berangkat":     0,
				"akumulasi":     0,
				"volume":        0,
				"indeks_parkir": "0%",
			}
			continue
		}

		intervalMap := map[string]interface{}{
			"start": intervalStart,
			"end":   intervalEnd,
			"label": interval["label"],
		}

		mobilStats := u.calculateIntervalStats(sessions, intervalMap, entities.VehicleTypeMobil, jukir.Area.MaxMobil)
		motorStats := u.calculateIntervalStats(sessions, intervalMap, entities.VehicleTypeMotor, jukir.Area.MaxMotor)

		// Get date from interval for adding to stats
		dateStr := ""
		if date, ok := interval["date"].(string); ok {
			dateStr = date
		} else {
			// Fallback: extract date from interval start time
			dateStr = intervalStart.Format("2006-01-02")
		}

		// Ensure indeks_parkir is always present and add date field
		if mobilStats == nil {
			mobilStats = map[string]interface{}{
				"periode":       interval["label"],
				"date":          dateStr,
				"datang":        0,
				"berangkat":     0,
				"akumulasi":     0,
				"volume":        0,
				"indeks_parkir": "0%",
			}
		} else {
			// Add date field to existing stats
			mobilStats["date"] = dateStr
		}
		if motorStats == nil {
			motorStats = map[string]interface{}{
				"periode":       interval["label"],
				"date":          dateStr,
				"datang":        0,
				"berangkat":     0,
				"akumulasi":     0,
				"volume":        0,
				"indeks_parkir": "0%",
			}
		} else {
			// Add date field to existing stats
			motorStats["date"] = dateStr
		}

		mobilData[j] = mobilStats
		motorData[j] = motorStats
	}

	// Calculate totals
	totalMobilDatang := 0
	totalMotorDatang := 0
	for _, mobil := range mobilData {
		if mobil != nil {
			if datang, ok := mobil["datang"].(int); ok {
				totalMobilDatang += datang
			}
		}
	}
	for _, motor := range motorData {
		if motor != nil {
			if datang, ok := motor["datang"].(int); ok {
				totalMotorDatang += datang
			}
		}
	}

	return map[string]interface{}{
		"jukir_id":           jukir.ID,
		"jukir_name":         jukir.User.Name,
		"area_id":            jukir.Area.ID,
		"area_name":          jukir.Area.Name,
		"regional":           jukir.Area.Regional,
		"kapasitas_mobil":    jukir.Area.MaxMobil,
		"kapasitas_motor":    jukir.Area.MaxMotor,
		"start_date":         actualStart.Format("2006-01-02"),
		"end_date":           actualEnd.Format("2006-01-02"),
		"mobil":              mobilData,
		"motor":              motorData,
		"total_mobil_datang": totalMobilDatang,
		"total_motor_datang": totalMotorDatang,
	}, nil
}

// ExportAreaActivityCSV exports area activity to CSV format
func (u *adminUsecase) ExportAreaActivityCSV(startTime, endTime *time.Time, areaID *uint, regional *string) (*bytes.Buffer, error) {
	activityData, err := u.GetAreaActivity(startTime, endTime, areaID, regional)
	if err != nil {
		return nil, err
	}

	data := activityData["data"].([]map[string]interface{})
	if len(data) == 0 {
		return nil, errors.New("no data to export")
	}

	var buf bytes.Buffer
	buf.WriteString("REKAPITULASI DATA PARKIR\n")
	buf.WriteString(";;;;;;;;;;;\n")
	buf.WriteString(";;;;;;;;;;;\n")

	for _, areaData := range data {
		areaName := areaData["area_name"].(string)
		kapasitasMobil := intFromAny(areaData["kapasitas_mobil"])
		kapasitasMotor := intFromAny(areaData["kapasitas_motor"])

		// Header with area name and capacity
		buf.WriteString(fmt.Sprintf("%s;;;Kapasitas ;%d;;%s;;;Kapasitas;%d\n",
			areaName, kapasitasMobil, areaName, kapasitasMotor))

		// Column headers
		buf.WriteString("Mobil Penumpang;Jumlah Kendaraan;;Akumulasi;Volume;Indeks Parkir;;Motor;Jumlah Kendaraan;;Akumulasi;Volume;Indeks Parkir\n")
		buf.WriteString("Periode Pengamatan;Datang;Berangkat;;;;;Periode Pengamatan;Datang;Berangkat;;;\n")

		mobilIntervals := areaData["mobil"].([]map[string]interface{})
		motorIntervals := areaData["motor"].([]map[string]interface{})

		maxLen := len(mobilIntervals)
		if len(motorIntervals) > maxLen {
			maxLen = len(motorIntervals)
		}

		for i := 0; i < maxLen; i++ {
			mobilLine := ";"
			if i < len(mobilIntervals) {
				m := mobilIntervals[i]
				mobilLine = fmt.Sprintf("%s;%d;%d;%d;%d;%s",
					m["periode"], m["datang"], m["berangkat"], m["akumulasi"], m["volume"], m["indeks_parkir"])
			}

			motorLine := ";"
			if i < len(motorIntervals) {
				m := motorIntervals[i]
				motorLine = fmt.Sprintf(";%s;%d;%d;%d;%d;%s",
					m["periode"], m["datang"], m["berangkat"], m["akumulasi"], m["volume"], m["indeks_parkir"])
			}

			buf.WriteString(mobilLine + motorLine + "\n")
		}

		// Total row
		totalMobil := intFromAny(areaData["total_mobil_datang"])
		totalMotor := intFromAny(areaData["total_motor_datang"])
		buf.WriteString(fmt.Sprintf(";%d;;;;;;%d;;;;\n", totalMobil, totalMotor))
		buf.WriteString("\n")
	}

	return &buf, nil
}

// ExportJukirActivityCSV exports jukir activity to CSV format
func (u *adminUsecase) ExportJukirActivityCSV(startTime, endTime *time.Time, jukirID *uint, regional *string) (*bytes.Buffer, error) {
	activityData, err := u.GetJukirActivity(startTime, endTime, jukirID, regional)
	if err != nil {
		return nil, err
	}

	data := activityData["data"].([]map[string]interface{})
	if len(data) == 0 {
		return nil, errors.New("no data to export")
	}

	var buf bytes.Buffer
	buf.WriteString("REKAPITULASI DATA PARKIR - JUKIR\n")
	buf.WriteString(";;;;;;;;;;;\n")
	buf.WriteString(";;;;;;;;;;;\n")

	for _, jukirData := range data {
		jukirName := jukirData["jukir_name"].(string)
		areaName := jukirData["area_name"].(string)
		kapasitasMobil := intFromAny(jukirData["kapasitas_mobil"])
		kapasitasMotor := intFromAny(jukirData["kapasitas_motor"])

		// Header with jukir name, area name and capacity
		buf.WriteString(fmt.Sprintf("%s - %s;;;Kapasitas ;%d;;%s - %s;;;Kapasitas;%d\n",
			jukirName, areaName, kapasitasMobil, jukirName, areaName, kapasitasMotor))

		// Column headers
		buf.WriteString("Mobil Penumpang;Jumlah Kendaraan;;Akumulasi;Volume;Indeks Parkir;;Motor;Jumlah Kendaraan;;Akumulasi;Volume;Indeks Parkir\n")
		buf.WriteString("Periode Pengamatan;Datang;Berangkat;;;;;Periode Pengamatan;Datang;Berangkat;;;\n")

		mobilIntervals := jukirData["mobil"].([]map[string]interface{})
		motorIntervals := jukirData["motor"].([]map[string]interface{})

		maxLen := len(mobilIntervals)
		if len(motorIntervals) > maxLen {
			maxLen = len(motorIntervals)
		}

		for i := 0; i < maxLen; i++ {
			mobilLine := ";"
			if i < len(mobilIntervals) {
				m := mobilIntervals[i]
				mobilLine = fmt.Sprintf("%s;%d;%d;%d;%d;%s",
					m["periode"], m["datang"], m["berangkat"], m["akumulasi"], m["volume"], m["indeks_parkir"])
			}

			motorLine := ";"
			if i < len(motorIntervals) {
				m := motorIntervals[i]
				motorLine = fmt.Sprintf(";%s;%d;%d;%d;%d;%s",
					m["periode"], m["datang"], m["berangkat"], m["akumulasi"], m["volume"], m["indeks_parkir"])
			}

			buf.WriteString(mobilLine + motorLine + "\n")
		}

		// Total row
		totalMobil := intFromAny(jukirData["total_mobil_datang"])
		totalMotor := intFromAny(jukirData["total_motor_datang"])
		buf.WriteString(fmt.Sprintf(";%d;;;;;;%d;;;;\n", totalMobil, totalMotor))
		buf.WriteString("\n")
	}

	return &buf, nil
}

// ExportAreaActivityDetailXLSX exports area activity detail to XLSX format
func (u *adminUsecase) ExportAreaActivityDetailXLSX(areaID uint, startTime, endTime *time.Time) (*bytes.Buffer, error) {
	loc := getGMT7Location()

	var start, end time.Time
	if startTime != nil && endTime != nil {
		start = startTime.In(loc)
		end = endTime.In(loc)
	} else {
		now := time.Now().In(loc)
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end = start
	}

	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, loc)

	if end.Before(start) {
		return nil, errors.New("end date must be after start date")
	}

	f := excelize.NewFile()
	sheetNames := make([]string, 0)
	dayCounter := 0

	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		dayCounter++
		if dayCounter > 7 {
			return nil, errors.New("date range too large (maximum 7 days)")
		}

		dayStart := day
		dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)
		ds := dayStart
		de := dayEnd

		activityData, err := u.GetAreaActivityDetail(areaID, &ds, &de)
		if err != nil {
			return nil, err
		}

		sheetName := dayStart.Format("2006-01-02")
		if dayCounter == 1 {
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				return nil, errors.New("failed to rename sheet")
			}
		} else {
			if _, err := f.NewSheet(sheetName); err != nil {
				return nil, errors.New("failed to create sheet")
			}
		}

		if err := writeAreaActivitySheet(f, sheetName, activityData); err != nil {
			return nil, err
		}

		sheetNames = append(sheetNames, sheetName)
	}

	if len(sheetNames) > 0 {
		if idx, err := f.GetSheetIndex(sheetNames[0]); err == nil {
			f.SetActiveSheet(idx)
		}
	}

	defer func() {
		_ = f.Close()
	}()

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, errors.New("failed to write Excel file")
	}

	return &buf, nil
}

func writeAreaActivitySheet(f *excelize.File, sheetName string, activityData map[string]interface{}) error {
	row := 1

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "REKAPITULASI DATA PARKIR")
	row += 2

	areaName, _ := activityData["area_name"].(string)
	kapasitasMobil := intFromAny(activityData["kapasitas_mobil"])
	kapasitasMotor := intFromAny(activityData["kapasitas_motor"])

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), areaName)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "Kapasitas")
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), kapasitasMobil)
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), areaName)
	f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), "Kapasitas")
	f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), kapasitasMotor)
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Mobil Penumpang")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Jumlah Kendaraan")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "Akumulasi")
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "Volume")
	f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), "Indeks Parkir")

	f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "Motor")
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "Jumlah Kendaraan")
	f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), "Akumulasi")
	f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), "Volume")
	f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), "Indeks Parkir")
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Periode Pengamatan")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Datang")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), "Berangkat")
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "Periode Pengamatan")
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "Datang")
	f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), "Berangkat")
	row++

	mobilIntervals, _ := activityData["mobil"].([]map[string]interface{})
	motorIntervals, _ := activityData["motor"].([]map[string]interface{})

	maxLen := len(mobilIntervals)
	if len(motorIntervals) > maxLen {
		maxLen = len(motorIntervals)
	}

	for i := 0; i < maxLen; i++ {
		if i < len(mobilIntervals) {
			m := mobilIntervals[i]
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), m["periode"])
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), m["datang"])
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), m["berangkat"])
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), m["akumulasi"])
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), m["volume"])
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), m["indeks_parkir"])
		}

		if i < len(motorIntervals) {
			m := motorIntervals[i]
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), m["periode"])
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), m["datang"])
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), m["berangkat"])
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), m["akumulasi"])
			f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), m["volume"])
			f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), m["indeks_parkir"])
		}

		row++
	}

	totalMobil := intFromAny(activityData["total_mobil_datang"])
	totalMotor := intFromAny(activityData["total_motor_datang"])
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), totalMobil)
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), totalMotor)

	return nil
}

// ExportJukirActivityDetailXLSX exports jukir activity detail to XLSX format
func (u *adminUsecase) ExportJukirActivityDetailXLSX(jukirID uint, startTime, endTime *time.Time) (*bytes.Buffer, error) {
	loc := getGMT7Location()

	var start, end time.Time
	if startTime != nil && endTime != nil {
		start = startTime.In(loc)
		end = endTime.In(loc)
	} else {
		now := time.Now().In(loc)
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end = start
	}

	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, loc)

	if end.Before(start) {
		return nil, errors.New("end date must be after start date")
	}

	f := excelize.NewFile()
	sheetNames := make([]string, 0)
	dayCounter := 0

	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		dayCounter++
		if dayCounter > 7 {
			return nil, errors.New("date range too large (maximum 7 days)")
		}

		dayStart := day
		dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)
		ds := dayStart
		de := dayEnd

		activityData, err := u.GetJukirActivityDetail(jukirID, &ds, &de)
		if err != nil {
			return nil, err
		}

		sheetName := dayStart.Format("2006-01-02")
		if dayCounter == 1 {
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				return nil, errors.New("failed to rename sheet")
			}
		} else {
			if _, err := f.NewSheet(sheetName); err != nil {
				return nil, errors.New("failed to create sheet")
			}
		}

		if err := writeJukirActivitySheet(f, sheetName, activityData); err != nil {
			return nil, err
		}

		sheetNames = append(sheetNames, sheetName)
	}

	if len(sheetNames) > 0 {
		if idx, err := f.GetSheetIndex(sheetNames[0]); err == nil {
			f.SetActiveSheet(idx)
		}
	}

	defer func() {
		_ = f.Close()
	}()

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, errors.New("failed to write Excel file")
	}

	return &buf, nil
}

func writeJukirActivitySheet(f *excelize.File, sheetName string, activityData map[string]interface{}) error {
	row := 1

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "REKAPITULASI DATA PARKIR - JUKIR")
	row += 2

	jukirName, _ := activityData["jukir_name"].(string)
	areaName, _ := activityData["area_name"].(string)
	kapasitasMobil := intFromAny(activityData["kapasitas_mobil"])
	kapasitasMotor := intFromAny(activityData["kapasitas_motor"])
	title := fmt.Sprintf("%s - %s", jukirName, areaName)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), title)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "Kapasitas")
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), kapasitasMobil)
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), title)
	f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), "Kapasitas")
	f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), kapasitasMotor)
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Mobil Penumpang")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Jumlah Kendaraan")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "Akumulasi")
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "Volume")
	f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), "Indeks Parkir")

	f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "Motor")
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "Jumlah Kendaraan")
	f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), "Akumulasi")
	f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), "Volume")
	f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), "Indeks Parkir")
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Periode Pengamatan")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Datang")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), "Berangkat")
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "Periode Pengamatan")
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "Datang")
	f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), "Berangkat")
	row++

	mobilIntervals, _ := activityData["mobil"].([]map[string]interface{})
	motorIntervals, _ := activityData["motor"].([]map[string]interface{})

	maxLen := len(mobilIntervals)
	if len(motorIntervals) > maxLen {
		maxLen = len(motorIntervals)
	}

	for i := 0; i < maxLen; i++ {
		if i < len(mobilIntervals) {
			m := mobilIntervals[i]
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), m["periode"])
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), m["datang"])
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), m["berangkat"])
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), m["akumulasi"])
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), m["volume"])
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), m["indeks_parkir"])
		}

		if i < len(motorIntervals) {
			m := motorIntervals[i]
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), m["periode"])
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), m["datang"])
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), m["berangkat"])
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), m["akumulasi"])
			f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), m["volume"])
			f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), m["indeks_parkir"])
		}

		row++
	}

	totalMobil := intFromAny(activityData["total_mobil_datang"])
	totalMotor := intFromAny(activityData["total_motor_datang"])
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), totalMobil)
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), totalMotor)

	return nil
}

func intFromAny(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case *int:
		if v != nil {
			return *v
		}
	case int32:
		return int(v)
	case *int32:
		if v != nil {
			return int(*v)
		}
	case int64:
		return int(v)
	case *int64:
		if v != nil {
			return int(*v)
		}
	case uint:
		return int(v)
	case *uint:
		if v != nil {
			return int(*v)
		}
	case float32:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// ImportAreasAndJukirsFromCSV imports parking areas and jukirs from CSV file
func (u *adminUsecase) ImportAreasAndJukirsFromCSV(reader io.Reader, regional string) (map[string]interface{}, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ','
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	headers, err := csvReader.Read()
	if err != nil {
		return nil, errors.New("failed to read CSV header")
	}

	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[strings.TrimSpace(h)] = i
	}

	requiredCols := []string{"NAMA JUKIR", "LOKASI PARKIR", "SEGMENTASI", "LATITUDINAL", "LONGITUDINAL"}
	for _, col := range requiredCols {
		if _, exists := colMap[col]; !exists {
			return nil, fmt.Errorf("required column missing: %s", col)
		}
	}

	areaMap := make(map[string]*entities.ParkingArea)
	var createdAreas []*entities.ParkingArea
	var createdJukirs []*entities.Jukir
	var errorList []string

	rowNum := 1
	for {
		rowNum++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("Row %d: %v", rowNum, err))
			continue
		}

		if len(record) < len(headers) {
			errorList = append(errorList, fmt.Sprintf("Row %d: insufficient columns", rowNum))
			continue
		}

		namaJukir := strings.TrimSpace(getColumnValue(record, colMap, "NAMA JUKIR"))
		lokasiParkir := strings.TrimSpace(getColumnValue(record, colMap, "LOKASI PARKIR"))
		segmentasi := strings.TrimSpace(getColumnValue(record, colMap, "SEGMENTASI"))
		latStr := strings.TrimSpace(getColumnValue(record, colMap, "LATITUDINAL"))
		lngStr := strings.TrimSpace(getColumnValue(record, colMap, "LONGITUDINAL"))
		fotoDokumentasi := strings.TrimSpace(getColumnValue(record, colMap, "FOTO DOKUMENTASI"))
		srpMotorStr := strings.TrimSpace(getColumnValue(record, colMap, "SRP MOTOR"))
		srpMobilStr := strings.TrimSpace(getColumnValue(record, colMap, "SRP MOBIL"))

		if namaJukir == "" || lokasiParkir == "" || segmentasi == "" || latStr == "" || lngStr == "" {
			errorList = append(errorList, fmt.Sprintf("Row %d: missing required fields", rowNum))
			continue
		}

		latStr = strings.ReplaceAll(latStr, ".", "")
		if len(latStr) > 2 {
			latStr = latStr[:2] + "." + latStr[2:]
		}
		lngStr = strings.ReplaceAll(lngStr, ".", "")
		if strings.HasPrefix(lngStr, "1") && len(lngStr) > 6 {
			lngStr = lngStr[:3] + "." + lngStr[3:]
		} else if len(lngStr) > 3 {
			lngStr = lngStr[:3] + "." + lngStr[3:]
		}

		latitude, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("Row %d: invalid latitude", rowNum))
			continue
		}

		longitude, err := strconv.ParseFloat(lngStr, 64)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("Row %d: invalid longitude", rowNum))
			continue
		}

		hourlyRateMobil := 2000.0
		hourlyRateMotor := 2000.0
		if srpMobilStr != "" {
			if rate, err := strconv.ParseFloat(strings.ReplaceAll(srpMobilStr, ",", "."), 64); err == nil {
				hourlyRateMobil = rate
			}
		}
		if srpMotorStr != "" {
			if rate, err := strconv.ParseFloat(strings.ReplaceAll(srpMotorStr, ",", "."), 64); err == nil {
				hourlyRateMotor = rate
			}
		}

		area, exists := areaMap[segmentasi]
		if !exists {
			allAreas, _, err := u.areaRepo.List(1000, 0)
			if err == nil {
				for i := range allAreas {
					if allAreas[i].Name == segmentasi && allAreas[i].Regional == regional {
						area = &allAreas[i]
						areaMap[segmentasi] = area
						break
					}
				}
			}

			if area == nil {
				areaName := segmentasi
				if areaName == "" {
					areaName = lokasiParkir
				}

				area = &entities.ParkingArea{
					Name:              areaName,
					Address:           lokasiParkir,
					Latitude:          latitude,
					Longitude:         longitude,
					Regional:          regional,
					HourlyRateMobil:   hourlyRateMobil,
					HourlyRateMotor:   hourlyRateMotor,
					Status:            entities.AreaStatusActive,
					StatusOperasional: "buka",
					JenisArea:         entities.JenisAreaOutdoor,
				}

				if fotoDokumentasi != "" {
					if strings.Contains(fotoDokumentasi, ",") {
						urls := strings.Split(fotoDokumentasi, ",")
						fotoDokumentasi = strings.TrimSpace(urls[0])
					}
					area.Image = &fotoDokumentasi
				}

				if err := u.areaRepo.Create(area); err != nil {
					errorList = append(errorList, fmt.Sprintf("Row %d: failed to create area", rowNum))
					continue
				}

				createdAreas = append(createdAreas, area)
				areaMap[segmentasi] = area
			}
		}

		// Generate jukir code first (email will be same as jukir_code)
		jukirCode := generateJukirCode(area.Name, time.Now())
		for i := 0; i < 10; i++ {
			_, err = u.jukirRepo.GetByJukirCode(jukirCode)
			if err != nil {
				break
			}
			jukirCode = generateJukirCode(area.Name, time.Now())
		}

		// Email is same as jukir_code (lowercase)
		email := strings.ToLower(jukirCode)
		password := generateSimplePassword(4)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("Row %d: failed to hash password", rowNum))
			continue
		}

		user := &entities.User{
			Name:            namaJukir,
			Email:           email,
			Phone:           "",
			Password:        string(hashedPassword),
			DisplayPassword: &password, // Store password for display
			Role:            entities.RoleJukir,
			Status:          entities.UserStatusActive,
		}

		if err := u.userRepo.Create(user); err != nil {
			errorList = append(errorList, fmt.Sprintf("Row %d: failed to create user", rowNum))
			continue
		}

		qrToken := "QR_" + jukirCode + "_" + time.Now().Format("20060102150405")
		jukir := &entities.Jukir{
			UserID:    user.ID,
			JukirCode: jukirCode,
			AreaID:    area.ID,
			QRToken:   qrToken,
			Status:    entities.JukirStatusActive,
		}

		if err := u.jukirRepo.Create(jukir); err != nil {
			u.userRepo.Delete(user.ID)
			errorList = append(errorList, fmt.Sprintf("Row %d: failed to create jukir", rowNum))
			continue
		}

		createdJukirs = append(createdJukirs, jukir)
	}

	return map[string]interface{}{
		"areas_created":  len(createdAreas),
		"jukirs_created": len(createdJukirs),
		"errors":         errorList,
		"total_rows":     rowNum - 1,
	}, nil
}

func getColumnValue(record []string, colMap map[string]int, colName string) string {
	if idx, exists := colMap[colName]; exists && idx < len(record) {
		return record[idx]
	}
	return ""
}

func calculateActualRevenue(sessionRepo repository.ParkingSessionRepository, areaRepo repository.ParkingAreaRepository, start, end time.Time, regional *string) float64 {
	areas, _ := areaRepo.GetActiveAreas()
	actualRevenue := 0.0

	for _, area := range areas {
		if regional != nil && *regional != "" && area.Regional != *regional {
			continue
		}

		sessions, _ := sessionRepo.GetSessionsByArea(area.ID, start, end)
		for _, session := range sessions {
			if session.TotalCost == nil {
				continue
			}

			if session.PaymentStatus == entities.PaymentStatusPaid || session.SessionStatus == entities.SessionStatusCompleted {
				actualRevenue += *session.TotalCost
			}
		}
	}

	return actualRevenue
}

func (u *adminUsecase) GetActivityLogs(jukirID *uint, areaID *uint, startTime, endTime time.Time, limit, offset int) (*entities.ActivityLogResponse, error) {
	if jukirID == nil && areaID == nil {
		return nil, errors.New("either jukir_id or area_id is required")
	}

	if limit <= 0 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	sessions, err := u.sessionRepo.GetSessionsForActivityLog(jukirID, areaID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	location := startTime.Location()
	events := make([]entities.ActivityLogItem, 0, len(sessions)*2)

	for _, session := range sessions {
		var areaInfo *entities.ActivityLogArea
		if session.Area.ID != 0 {
			areaCopy := entities.ActivityLogArea{
				ID:   session.Area.ID,
				Name: session.Area.Name,
			}
			areaInfo = &areaCopy
		}

		var jukirInfo *entities.ActivityLogJukir
		if session.Jukir != nil {
			jukirName := session.Jukir.User.Name
			jukirCopy := entities.ActivityLogJukir{
				ID:        session.Jukir.ID,
				Name:      jukirName,
				JukirCode: session.Jukir.JukirCode,
			}
			jukirInfo = &jukirCopy
		}

		var sessionID *uint
		if !session.IsManualRecord {
			sid := session.ID
			sessionID = &sid
		}

		platNomor := session.PlatNomor

		if !session.CheckinTime.Before(startTime) && session.CheckinTime.Before(endTime) {
			events = append(events, entities.ActivityLogItem{
				EventTime:   session.CheckinTime.In(location),
				EventType:   entities.ActivityEventCheckin,
				SessionID:   sessionID,
				PlatNomor:   platNomor,
				VehicleType: string(session.VehicleType),
				IsManual:    session.IsManualRecord,
				Jukir:       jukirInfo,
				Area:        areaInfo,
			})
		}

		if session.CheckoutTime != nil && !session.CheckoutTime.Before(startTime) && session.CheckoutTime.Before(endTime) {
			events = append(events, entities.ActivityLogItem{
				EventTime:   session.CheckoutTime.In(location),
				EventType:   entities.ActivityEventCheckout,
				SessionID:   sessionID,
				PlatNomor:   platNomor,
				VehicleType: string(session.VehicleType),
				IsManual:    session.IsManualRecord,
				Jukir:       jukirInfo,
				Area:        areaInfo,
			})
		}
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].EventTime.Equal(events[j].EventTime) {
			return string(events[i].EventType) < string(events[j].EventType)
		}
		return events[i].EventTime.After(events[j].EventTime)
	})

	total := len(events)
	if offset > total {
		offset = total
	}

	endIndex := offset + limit
	if endIndex > total {
		endIndex = total
	}

	paged := events[offset:endIndex]

	endInclusive := endTime.Add(-time.Nanosecond)
	if endInclusive.Before(startTime) {
		endInclusive = startTime
	}

	meta := entities.ActivityLogMeta{
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		StartDate: startTime.In(location).Format("2006-01-02"),
		EndDate:   endInclusive.In(location).Format("2006-01-02"),
		JukirID:   jukirID,
		AreaID:    areaID,
	}

	return &entities.ActivityLogResponse{
		Activities: paged,
		Meta:       meta,
	}, nil
}
