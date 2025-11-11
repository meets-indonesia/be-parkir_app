package handler

import (
	"be-parkir/internal/domain/entities"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// getGMT7Location returns Asia/Jakarta timezone (GMT+7)
func getGMT7Location() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC+7 if Asia/Jakarta is not available
		return time.FixedZone("GMT+7", 7*60*60)
	}
	return loc
}

// parseDateFilter parses start_date and end_date from query parameters (format: dd-mm-yyyy)
// Returns dates in GMT+7 timezone
func parseDateFilter(c *gin.Context) (*time.Time, *time.Time, error) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" && endDateStr == "" {
		return nil, nil, nil
	}

	if startDateStr == "" || endDateStr == "" {
		return nil, nil, errors.New("both start_date and end_date are required")
	}

	gmt7Loc := getGMT7Location()

	// Parse dates and set to GMT+7 timezone
	startDate, err := time.ParseInLocation("02-01-2006", startDateStr, gmt7Loc)
	if err != nil {
		return nil, nil, errors.New("invalid start_date format. Use DD-MM-YYYY")
	}

	endDate, err := time.ParseInLocation("02-01-2006", endDateStr, gmt7Loc)
	if err != nil {
		return nil, nil, errors.New("invalid end_date format. Use DD-MM-YYYY")
	}

	// Set start date to beginning of day in GMT+7
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, gmt7Loc)
	// Set end date to end of day in GMT+7
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, gmt7Loc)

	return &startDate, &endDate, nil
}

// GetAdminOverview godoc
// @Summary Get admin overview
// @Description Get system-wide statistics for admin dashboard
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vehicle_type query string false "Filter by vehicle type (mobil/motor)"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/overview [get]
func (h *Handlers) GetAdminOverview(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")

	var vehicleTypePtr *string
	if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
		vehicleTypePtr = &vehicleType
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response, err := h.AdminUC.GetOverview(vehicleTypePtr, startTime, endTime)
	if err != nil {
		h.Logger.Error("Failed to get overview:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Overview data retrieved successfully",
		"data":    response,
		"filter": map[string]interface{}{
			"vehicle_type": vehicleType,
			"start_date":   c.Query("start_date"),
			"end_date":     c.Query("end_date"),
		},
	})
}

// GetJukirs godoc
// @Summary Get all jukirs
// @Description Get list of all jukirs with pagination (optional filter by regional)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs [get]
func (h *Handlers) GetJukirs(c *gin.Context) {
	// Check if revenue parameter is requested
	includeRevenue := c.Query("include_revenue") == "true"

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// If revenue is requested, use the new method
	if includeRevenue {
		vehicleType := c.Query("vehicle_type")
		dateRange := c.Query("date_range")

		var vehicleTypePtr *string
		if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
			vehicleTypePtr = &vehicleType
		}

		var dateRangePtr *string
		if dateRange != "" && (dateRange == "hari_ini" || dateRange == "minggu_ini" || dateRange == "bulan_ini" || dateRange == "tahun_ini") {
			dateRangePtr = &dateRange
		}

		jukirsWithRevenue, count, err := h.AdminUC.GetJukirsWithRevenue(limit, offset, vehicleTypePtr, dateRangePtr)
		if err != nil {
			h.Logger.Error("Failed to get jukirs with revenue:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Jukirs with revenue retrieved successfully",
			"data":    jukirsWithRevenue,
			"meta": gin.H{
				"pagination": gin.H{
					"limit":  limit,
					"offset": offset,
					"total":  count,
				},
				"vehicle_type": vehicleType,
				"date_range":   dateRange,
			},
		})
		return
	}

	// Original behavior - return jukirs without revenue
	regional := c.Query("regional")

	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	jukirs, count, err := h.AdminUC.GetJukirs(limit, offset, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get jukirs:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Ensure we return an empty array instead of null if no jukirs
	if jukirs == nil {
		jukirs = []entities.Jukir{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukirs retrieved successfully",
		"data":    jukirs,
		"meta": gin.H{
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
				"total":  count,
			},
			"regional": regional,
		},
	})
}

// CreateJukir godoc
// @Summary Create new jukir
// @Description Create a new jukir account with auto-generated username and password (only requires name, area, and status)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body entities.CreateJukirRequest true "Jukir creation data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs [post]
func (h *Handlers) CreateJukir(c *gin.Context) {
	var req entities.CreateJukirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.Logger.Error("Validation failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	response, err := h.AdminUC.CreateJukir(&req)
	if err != nil {
		h.Logger.Error("Failed to create jukir:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Jukir created successfully",
		"data":    response,
	})
}

// UpdateJukir godoc
// @Summary Update jukir
// @Description Update jukir information
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Jukir ID"
// @Param request body entities.UpdateJukirRequest true "Jukir update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/{id} [put]
func (h *Handlers) UpdateJukir(c *gin.Context) {
	idStr := c.Param("id")
	jukirID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid jukir ID",
		})
		return
	}

	var req entities.UpdateJukirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.Logger.Error("Validation failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	response, err := h.AdminUC.UpdateJukir(uint(jukirID), &req)
	if err != nil {
		h.Logger.Error("Failed to update jukir:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir updated successfully",
		"data":    response,
	})
}

// DeleteJukir godoc
// @Summary Delete jukir
// @Description Delete a jukir by ID
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Jukir ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/{id} [delete]
func (h *Handlers) DeleteJukir(c *gin.Context) {
	idStr := c.Param("id")
	jukirID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid jukir ID",
		})
		return
	}

	err = h.AdminUC.DeleteJukir(uint(jukirID))
	if err != nil {
		h.Logger.Error("Failed to delete jukir:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir deleted successfully",
	})
}

// UpdateJukirStatus godoc
// @Summary Update jukir status
// @Description Update jukir status (active/inactive)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Jukir ID"
// @Param request body entities.UpdateJukirRequest true "Jukir status update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/{id}/status [put]
func (h *Handlers) UpdateJukirStatus(c *gin.Context) {
	idStr := c.Param("id")
	jukirID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid jukir ID",
		})
		return
	}

	var req entities.UpdateJukirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.Logger.Error("Validation failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	response, err := h.AdminUC.UpdateJukirStatus(uint(jukirID), &req)
	if err != nil {
		h.Logger.Error("Failed to update jukir status:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir status updated successfully",
		"data":    response,
	})
}

// GetReports godoc
// @Summary Get reports
// @Description Generate reports by date range and area
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param area_id query int false "Area ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/reports [get]
func (h *Handlers) GetReports(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	areaIDStr := c.Query("area_id")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "start_date and end_date are required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	var areaID *uint
	if areaIDStr != "" {
		id, err := strconv.ParseUint(areaIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid area_id",
			})
			return
		}
		areaIDUint := uint(id)
		areaID = &areaIDUint
	}

	response, err := h.AdminUC.GetReports(startDate, endDate, areaID)
	if err != nil {
		h.Logger.Error("Failed to get reports:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reports retrieved successfully",
		"data":    response,
	})
}

// GetAllSessions godoc
// @Summary Get all sessions
// @Description Get all parking sessions with filters
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param status query string false "Session status"
// @Param area_id query int false "Area ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/sessions [get]
func (h *Handlers) GetAllSessions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["session_status"] = status
	}
	if areaIDStr := c.Query("area_id"); areaIDStr != "" {
		if areaID, err := strconv.ParseUint(areaIDStr, 10, 32); err == nil {
			filters["area_id"] = uint(areaID)
		}
	}

	sessions, count, err := h.AdminUC.GetAllSessions(limit, offset, filters)
	if err != nil {
		h.Logger.Error("Failed to get sessions:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Sessions retrieved successfully",
		"data":    sessions,
		"meta": gin.H{
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
				"total":  count,
			},
		},
	})
}

// CreateParkingArea godoc
// @Summary Create parking area
// @Description Create a new parking area
// @Tags admin
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Name"
// @Param address formData string true "Address"
// @Param latitude formData number true "Latitude"
// @Param longitude formData number true "Longitude"
// @Param regional formData string true "Regional"
// @Param hourly_rate formData number true "Hourly Rate"
// @Param max_mobil formData integer false "Max Mobil"
// @Param max_motor formData integer false "Max Motor"
// @Param status_operasional formData string true "Status Operasional (buka/tutup/maintenance)"
// @Param jenis_area formData string true "Jenis Area (indoor/outdoor/mix)"
// @Param image formData file false "Area image"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas [post]
func (h *Handlers) CreateParkingArea(c *gin.Context) {
	// Ensure this is a POST request
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}

	// Check content type - support both JSON and multipart form-data
	contentType := c.GetHeader("Content-Type")
	var req entities.CreateParkingAreaRequest

	if contentType == "application/json" || strings.Contains(contentType, "application/json") {
		// Handle JSON request
		if err := c.ShouldBindJSON(&req); err != nil {
			h.Logger.Error("Failed to bind JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request data", "error": err.Error()})
			return
		}
	} else {
		// Handle multipart form-data or form-urlencoded
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// Parse multipart form-data
			if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB
				h.Logger.Error("Failed to parse multipart form:", err)
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid form data"})
				return
			}
		} else {
			// Parse regular form
			if err := c.Request.ParseForm(); err != nil {
				h.Logger.Error("Failed to parse form:", err)
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid form data"})
				return
			}
		}

		// Build request struct from form
		req.Name = c.PostForm("name")
		req.Address = c.PostForm("address")
		if latStr := c.PostForm("latitude"); latStr != "" {
			if v, err := strconv.ParseFloat(latStr, 64); err == nil {
				req.Latitude = v
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid latitude"})
				return
			}
		}
		if lngStr := c.PostForm("longitude"); lngStr != "" {
			if v, err := strconv.ParseFloat(lngStr, 64); err == nil {
				req.Longitude = v
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid longitude"})
				return
			}
		}
		req.Regional = c.PostForm("regional")
		if rateStr := c.PostForm("hourly_rate"); rateStr != "" {
			if v, err := strconv.ParseFloat(rateStr, 64); err == nil {
				req.HourlyRate = v
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid hourly_rate"})
				return
			}
		}
		if mm := c.PostForm("max_mobil"); mm != "" {
			if v, err := strconv.Atoi(mm); err == nil {
				req.MaxMobil = &v
			}
		}
		if mm := c.PostForm("max_motor"); mm != "" {
			if v, err := strconv.Atoi(mm); err == nil {
				req.MaxMotor = &v
			}
		}
		req.StatusOperasional = c.PostForm("status_operasional")
		req.JenisArea = entities.JenisArea(c.PostForm("jenis_area"))
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.Logger.Error("Validation failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Validation failed", "error": err.Error()})
		return
	}

	var imageURL *string
	// Handle optional image upload (only for multipart form-data)
	if strings.HasPrefix(contentType, "multipart/form-data") {
		fileHeader, err := c.FormFile("image")
		if err == nil && fileHeader != nil {
			f, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "failed to open image"})
				return
			}
			defer f.Close()
			// Upload to MinIO
			objectName := fmt.Sprintf("areas/%d_%s", time.Now().UnixNano(), fileHeader.Filename)
			url, err := h.Storage.Upload(c.Request.Context(), objectName, f, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
			if err != nil {
				h.Logger.Error("MinIO upload failed:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Upload failed"})
				return
			}
			imageURL = &url
		}
	}

	response, err := h.AdminUC.CreateParkingArea(&req)
	if err != nil {
		h.Logger.Error("Failed to create parking area:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// If image uploaded, update area with image URL
	if imageURL != nil {
		update := entities.UpdateParkingAreaRequest{Image: imageURL}
		if _, err := h.AdminUC.UpdateParkingArea(response.ID, &update); err != nil {
			h.Logger.Warn("Area created but failed to set image URL:", err)
		} else {
			response.Image = imageURL
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Parking area created successfully",
		"data":    response,
	})
}

// GetParkingAreas godoc
// @Summary Get all parking areas with status
// @Description Get list of all parking areas for admin area-parkir menu (optional filter by regional)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas [get]
func (h *Handlers) GetParkingAreas(c *gin.Context) {
	// Ensure this is a GET request
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}

	regional := c.Query("regional")

	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	response, err := h.AdminUC.GetParkingAreas(regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get parking areas:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking areas retrieved successfully",
		"data":    response,
	})
}

// GetParkingAreaDetail godoc
// @Summary Get parking area detail
// @Description Get full parking area detail with jukirs and metrics
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/{id} [get]
func (h *Handlers) GetParkingAreaDetail(c *gin.Context) {
	areaIDStr := c.Param("id")
	areaID, err := strconv.ParseUint(areaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid area ID",
		})
		return
	}

	response, err := h.AdminUC.GetParkingAreaDetail(uint(areaID))
	if err != nil {
		h.Logger.Error("Failed to get parking area detail:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking area detail retrieved successfully",
		"data":    response,
	})
}

// GetParkingAreaStatus godoc
// @Summary Get parking area status
// @Description Get parking area status including available and occupied slots for mobil and motor
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/{id}/status [get]
func (h *Handlers) GetParkingAreaStatus(c *gin.Context) {
	areaIDStr := c.Param("id")
	areaID, err := strconv.ParseUint(areaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid area ID",
		})
		return
	}

	response, err := h.AdminUC.GetParkingAreaStatus(uint(areaID))
	if err != nil {
		h.Logger.Error("Failed to get parking area status:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking area status retrieved successfully",
		"data":    response,
	})
}

// GetAreaTransactions godoc
// @Summary Get area transactions
// @Description Get transaction details by parking area (optional filter by date range)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/{id}/transactions [get]
func (h *Handlers) GetAreaTransactions(c *gin.Context) {
	areaIDStr := c.Param("id")
	areaID, err := strconv.ParseUint(areaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid area ID",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Parse date filter
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response, count, err := h.AdminUC.GetAreaTransactions(uint(areaID), limit, offset, startTime, endTime)
	if err != nil {
		h.Logger.Error("Failed to get area transactions:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Area transactions retrieved successfully",
		"data":    response,
		"meta": gin.H{
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
				"total":  count,
			},
			"start_date": c.Query("start_date"),
			"end_date":   c.Query("end_date"),
		},
	})
}

// GetAreaActivity godoc
// @Summary Get area activity monitoring
// @Description Get activity monitoring data for all parking areas - total masuk (checkin) and keluar (checkout) per area
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/activity [get]
func (h *Handlers) GetAreaActivity(c *gin.Context) {
	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Parse regional
	regional := c.Query("regional")
	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	// Get all areas activity (no area_id filter - shows all areas)
	response, err := h.AdminUC.GetAreaActivity(startTime, endTime, nil, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get area activity:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Area activity retrieved successfully",
		"data":    response,
	})
}

// GetAreaActivityDetail godoc
// @Summary Get detailed area activity monitoring by ID
// @Description Get detailed activity monitoring data for a specific parking area with 15-minute intervals breakdown
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/{id}/activity [get]
func (h *Handlers) GetAreaActivityDetail(c *gin.Context) {
	// Parse area ID from path
	areaIDStr := c.Param("id")
	areaID, err := strconv.ParseUint(areaIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid area ID",
		})
		return
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response, err := h.AdminUC.GetAreaActivityDetail(uint(areaID), startTime, endTime)
	if err != nil {
		h.Logger.Error("Failed to get area activity detail:", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "parking area not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Area activity detail retrieved successfully",
		"data":    response,
	})
}

// ExportAreaActivityDetailXLSX godoc
// @Summary Export area activity detail to XLSX
// @Description Export detailed area activity monitoring data to XLSX format (like CSV format)
// @Tags admin
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Success 200 {file} file "XLSX file"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/{id}/activity/export [get]
func (h *Handlers) ExportAreaActivityDetailXLSX(c *gin.Context) {
	// Parse area ID from path
	areaIDStr := c.Param("id")
	areaID, err := strconv.ParseUint(areaIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid area ID",
		})
		return
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	xlsxBuffer, err := h.AdminUC.ExportAreaActivityDetailXLSX(uint(areaID), startTime, endTime)
	if err != nil {
		h.Logger.Error("Failed to export area activity detail:", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "parking area not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Generate filename
	filename := fmt.Sprintf("area-activity-detail-%d.xlsx", areaID)
	if startTime != nil && endTime != nil {
		filename = fmt.Sprintf("area-activity-detail-%d-%s-to-%s.xlsx",
			areaID,
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"))
	}

	// Upload to MinIO
	objectName := fmt.Sprintf("exports/activity/%d_%s", time.Now().UnixNano(), filename)
	reader := bytes.NewReader(xlsxBuffer.Bytes())
	_, err = h.Storage.Upload(c.Request.Context(), objectName, reader, int64(xlsxBuffer.Len()), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err != nil {
		h.Logger.Error("Failed to upload XLSX to MinIO:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to store export file",
		})
		return
	}

	downloadURL := fmt.Sprintf("/api/v1/admin/files/%s", objectName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Area activity detail exported successfully",
		"data": gin.H{
			"filename":    filename,
			"url":         downloadURL,
			"object_name": objectName,
		},
	})
}

// GetJukirActivity godoc
// @Summary Get jukir activity monitoring
// @Description Get activity monitoring data for all jukirs - total masuk (checkin) and keluar (checkout) per jukir
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/activity [get]
func (h *Handlers) GetJukirActivity(c *gin.Context) {
	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Parse regional
	regional := c.Query("regional")
	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	// Get all jukirs activity (no jukir_id filter - shows all jukirs)
	response, err := h.AdminUC.GetJukirActivity(startTime, endTime, nil, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get jukir activity:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir activity retrieved successfully",
		"data":    response,
	})
}

// GetJukirActivityDetail godoc
// @Summary Get detailed jukir activity monitoring by ID
// @Description Get detailed activity monitoring data for a specific jukir with 15-minute intervals breakdown (9am-5pm)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Jukir ID"
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/{id}/activity [get]
func (h *Handlers) GetJukirActivityDetail(c *gin.Context) {
	// Parse jukir ID from path
	jukirIDStr := c.Param("id")
	jukirID, err := strconv.ParseUint(jukirIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid jukir ID",
		})
		return
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response, err := h.AdminUC.GetJukirActivityDetail(uint(jukirID), startTime, endTime)
	if err != nil {
		h.Logger.Error("Failed to get jukir activity detail:", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "jukir not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir activity detail retrieved successfully",
		"data":    response,
	})
}

// ExportJukirActivityDetailXLSX godoc
// @Summary Export jukir activity detail to XLSX
// @Description Export detailed jukir activity monitoring data to XLSX format (like CSV format)
// @Tags admin
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param id path int true "Jukir ID"
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Success 200 {file} file "XLSX file"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/{id}/activity/export [get]
func (h *Handlers) ExportJukirActivityDetailXLSX(c *gin.Context) {
	// Parse jukir ID from path
	jukirIDStr := c.Param("id")
	jukirID, err := strconv.ParseUint(jukirIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid jukir ID",
		})
		return
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	xlsxBuffer, err := h.AdminUC.ExportJukirActivityDetailXLSX(uint(jukirID), startTime, endTime)
	if err != nil {
		h.Logger.Error("Failed to export jukir activity detail:", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "jukir not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Generate filename
	filename := fmt.Sprintf("jukir-activity-detail-%d.xlsx", jukirID)
	if startTime != nil && endTime != nil {
		filename = fmt.Sprintf("jukir-activity-detail-%d-%s-to-%s.xlsx",
			jukirID,
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"))
	}

	// Upload to MinIO
	objectName := fmt.Sprintf("exports/activity/%d_%s", time.Now().UnixNano(), filename)
	reader := bytes.NewReader(xlsxBuffer.Bytes())
	_, err = h.Storage.Upload(c.Request.Context(), objectName, reader, int64(xlsxBuffer.Len()), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err != nil {
		h.Logger.Error("Failed to upload XLSX to MinIO:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to store export file",
		})
		return
	}

	downloadURL := fmt.Sprintf("/api/v1/admin/files/%s", objectName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir activity detail exported successfully",
		"data": gin.H{
			"filename":    filename,
			"url":         downloadURL,
			"object_name": objectName,
		},
	})
}

// ExportAreaActivityCSV godoc
// @Summary Export area activity to CSV
// @Description Export area activity monitoring data to CSV format
// @Tags admin
// @Accept json
// @Produce text/csv
// @Security BearerAuth
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param area_id query int false "Filter by area ID"
// @Param regional query string false "Filter by regional"
// @Success 200 {file} file "CSV file"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/activity/export [get]
func (h *Handlers) ExportAreaActivityCSV(c *gin.Context) {
	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Parse area_id
	var areaID *uint
	if areaIDStr := c.Query("area_id"); areaIDStr != "" {
		if id, err := strconv.ParseUint(areaIDStr, 10, 32); err == nil {
			areaIDVal := uint(id)
			areaID = &areaIDVal
		}
	}

	// Parse regional
	regional := c.Query("regional")
	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	csvBuffer, err := h.AdminUC.ExportAreaActivityCSV(startTime, endTime, areaID, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to export area activity:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Generate filename
	filename := "area-activity.csv"
	if startTime != nil && endTime != nil {
		filename = fmt.Sprintf("area-activity-%s-to-%s.csv",
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"))
	}

	// Upload to MinIO
	objectName := fmt.Sprintf("exports/activity/%d_%s", time.Now().UnixNano(), filename)
	reader := bytes.NewReader(csvBuffer.Bytes())
	_, err = h.Storage.Upload(c.Request.Context(), objectName, reader, int64(csvBuffer.Len()), "text/csv")
	if err != nil {
		h.Logger.Error("Failed to upload CSV to MinIO:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to store export file",
		})
		return
	}

	downloadURL := fmt.Sprintf("/api/v1/admin/files/%s", objectName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Area activity exported successfully",
		"data": gin.H{
			"filename":    filename,
			"url":         downloadURL,
			"object_name": objectName,
		},
	})
}

// ExportJukirActivityCSV godoc
// @Summary Export jukir activity to CSV
// @Description Export jukir activity monitoring data to CSV format
// @Tags admin
// @Accept json
// @Produce text/csv
// @Security BearerAuth
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param jukir_id query int false "Filter by jukir ID"
// @Param regional query string false "Filter by regional"
// @Success 200 {file} file "CSV file"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/activity/export [get]
func (h *Handlers) ExportJukirActivityCSV(c *gin.Context) {
	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Parse jukir_id
	var jukirID *uint
	if jukirIDStr := c.Query("jukir_id"); jukirIDStr != "" {
		if id, err := strconv.ParseUint(jukirIDStr, 10, 32); err == nil {
			jukirIDVal := uint(id)
			jukirID = &jukirIDVal
		}
	}

	// Parse regional
	regional := c.Query("regional")
	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	csvBuffer, err := h.AdminUC.ExportJukirActivityCSV(startTime, endTime, jukirID, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to export jukir activity:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Generate filename
	filename := "jukir-activity.csv"
	if startTime != nil && endTime != nil {
		filename = fmt.Sprintf("jukir-activity-%s-to-%s.csv",
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"))
	}

	// Upload to MinIO
	objectName := fmt.Sprintf("exports/activity/%d_%s", time.Now().UnixNano(), filename)
	reader := bytes.NewReader(csvBuffer.Bytes())
	_, err = h.Storage.Upload(c.Request.Context(), objectName, reader, int64(csvBuffer.Len()), "text/csv")
	if err != nil {
		h.Logger.Error("Failed to upload CSV to MinIO:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to store export file",
		})
		return
	}

	downloadURL := fmt.Sprintf("/api/v1/admin/files/%s", objectName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir activity exported successfully",
		"data": gin.H{
			"filename":    filename,
			"url":         downloadURL,
			"object_name": objectName,
		},
	})
}

// GetRevenueTable godoc
// @Summary Get revenue table
// @Description Get revenue table data for monitor-pendapatan page
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param area_id query int false "Area ID to filter"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/revenue-table [get]
func (h *Handlers) GetRevenueTable(c *gin.Context) {
	areaIDStr := c.Query("area_id")
	var areaID *uint

	if areaIDStr != "" {
		areaIDUint, err := strconv.ParseUint(areaIDStr, 10, 64)
		if err == nil {
			areaIDVal := uint(areaIDUint)
			areaID = &areaIDVal
		}
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Parse regional
	regional := c.Query("regional")
	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	response, count, err := h.AdminUC.GetRevenueTable(limit, offset, areaID, startTime, endTime, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get revenue table:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Revenue table retrieved successfully",
		"data":    response,
		"meta": gin.H{
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
				"total":  count,
			},
		},
	})
}

// ExportRevenueReport godoc
// @Summary Export revenue report to Excel
// @Description Export revenue report with actual vs estimated revenue by jukir and area
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/revenue/export [get]
func (h *Handlers) ExportRevenueReport(c *gin.Context) {
	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Parse regional
	regional := c.Query("regional")
	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	// Get Excel buffer
	excelBuffer, err := h.AdminUC.ExportRevenueReport(startTime, endTime, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to export revenue report:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Generate object name and upload to MinIO
	filename := "revenue-report.xlsx"
	if startTime != nil && endTime != nil {
		filename = fmt.Sprintf("revenue-report-%s-to-%s.xlsx",
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"))
	}
	objectName := fmt.Sprintf("exports/%d_%s", time.Now().UnixNano(), filename)

	// Upload file ke MinIO
	reader := bytes.NewReader(excelBuffer.Bytes())
	_, err = h.Storage.Upload(c.Request.Context(), objectName, reader, int64(reader.Len()), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err != nil {
		h.Logger.Error("Failed to upload export to MinIO:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to store export file", "error": err.Error()})
		return
	}

	// Gunakan proxy URL untuk download (lebih reliable daripada presigned URL)
	// Format: /api/v1/admin/files/exports/filename.xlsx
	downloadURL := fmt.Sprintf("/api/v1/admin/files/%s", objectName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Export generated successfully",
		"data": gin.H{
			"filename":    filename,
			"url":         downloadURL,
			"object_name": objectName,
		},
	})
}

// DownloadFile godoc
// @Summary Download file from MinIO
// @Description Download file from MinIO storage via proxy
// @Tags admin
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param path path string true "File path (e.g., exports/1234567890_report.xlsx)"
// @Success 200 {file} file "File download"
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/files/{path} [get]
func (h *Handlers) DownloadFile(c *gin.Context) {
	filePath := c.Param("path")
	if filePath == "" || filePath == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "File path is required"})
		return
	}

	// Remove leading slash if present
	if filePath[0] == '/' {
		filePath = filePath[1:]
	}

	// Get file from MinIO
	ctx := c.Request.Context()
	reader, size, err := h.Storage.GetObject(ctx, filePath)
	if err != nil {
		h.Logger.Error("Failed to get file from MinIO:", err)
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "File not found"})
		return
	}
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	}()

	// Extract filename from path
	parts := strings.Split(filePath, "/")
	filename := parts[len(parts)-1]

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Length", fmt.Sprintf("%d", size))

	// Stream file to response
	c.DataFromReader(http.StatusOK, size, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", reader, nil)
}

// UpdateParkingArea godoc
// @Summary Update parking area
// @Description Update an existing parking area
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Param request body entities.UpdateParkingAreaRequest true "Parking area update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/{id} [put]
func (h *Handlers) UpdateParkingArea(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid area ID",
		})
		return
	}

	// Check content type - support both JSON and multipart form-data
	contentType := c.GetHeader("Content-Type")
	var req entities.UpdateParkingAreaRequest

	if contentType == "application/json" || strings.Contains(contentType, "application/json") {
		// Handle JSON request
		if err := c.ShouldBindJSON(&req); err != nil {
			h.Logger.Error("Failed to bind JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request data", "error": err.Error()})
			return
		}
	} else {
		// Handle multipart form-data
		if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
			h.Logger.Error("Failed to parse multipart form:", err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid form data"})
			return
		}

		if name := c.PostForm("name"); name != "" {
			req.Name = &name
		}
		if address := c.PostForm("address"); address != "" {
			req.Address = &address
		}
		if latStr := c.PostForm("latitude"); latStr != "" {
			if v, err := strconv.ParseFloat(latStr, 64); err == nil {
				req.Latitude = &v
			}
		}
		if lngStr := c.PostForm("longitude"); lngStr != "" {
			if v, err := strconv.ParseFloat(lngStr, 64); err == nil {
				req.Longitude = &v
			}
		}
		if regional := c.PostForm("regional"); regional != "" {
			req.Regional = &regional
		}
		if rateStr := c.PostForm("hourly_rate"); rateStr != "" {
			if v, err := strconv.ParseFloat(rateStr, 64); err == nil {
				req.HourlyRate = &v
			}
		}
		if mm := c.PostForm("max_mobil"); mm != "" {
			if v, err := strconv.Atoi(mm); err == nil {
				req.MaxMobil = &v
			}
		}
		if mm := c.PostForm("max_motor"); mm != "" {
			if v, err := strconv.Atoi(mm); err == nil {
				req.MaxMotor = &v
			}
		}
		if so := c.PostForm("status_operasional"); so != "" {
			req.StatusOperasional = &so
		}
		if ja := c.PostForm("jenis_area"); ja != "" {
			jaVal := entities.JenisArea(ja)
			req.JenisArea = &jaVal
		}
	}

	// Handle optional image upload
	fileHeader, err := c.FormFile("image")
	if err == nil && fileHeader != nil {
		f, err := fileHeader.Open()
		if err == nil {
			defer f.Close()
			objectName := fmt.Sprintf("areas/%d_%s", time.Now().UnixNano(), fileHeader.Filename)
			url, err := h.Storage.Upload(c.Request.Context(), objectName, f, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
			if err == nil {
				req.Image = &url
			}
		}
	}

	// Validate request if has any fields
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.Logger.Error("Validation failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Validation failed", "error": err.Error()})
		return
	}

	response, err := h.AdminUC.UpdateParkingArea(uint(areaID), &req)
	if err != nil {
		h.Logger.Error("Failed to update parking area:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking area updated successfully",
		"data":    response,
	})
}

// DeleteParkingArea godoc
// @Summary Delete parking area
// @Description Delete a parking area by ID
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas/{id} [delete]
func (h *Handlers) DeleteParkingArea(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid area ID",
		})
		return
	}

	err = h.AdminUC.DeleteParkingArea(uint(areaID))
	if err != nil {
		h.Logger.Error("Failed to delete parking area:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking area deleted successfully",
	})
}

// GetAllJukirsRevenue godoc
// @Summary Get all jukirs with revenue
// @Description Get list of all jukirs with their revenue filtered by date range
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/revenue [get]
func (h *Handlers) GetAllJukirsRevenue(c *gin.Context) {
	regional := c.Query("regional")

	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response, err := h.AdminUC.GetAllJukirsRevenue(startTime, endTime, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get jukirs revenue:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukirs revenue retrieved successfully",
		"data":    response,
	})
}

// AddManualRevenue godoc
// @Summary Add manual revenue for jukir
// @Description Add manual revenue entry for a specific jukir on a specific date
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body entities.JukirRevenueRequest true "Manual revenue data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/manual-revenue [post]
func (h *Handlers) AddManualRevenue(c *gin.Context) {
	var req entities.JukirRevenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.Logger.Error("Validation failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	response, err := h.AdminUC.AddManualRevenue(&req)
	if err != nil {
		h.Logger.Error("Failed to add manual revenue:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Manual revenue added successfully",
		"data":    response,
	})
}

// GetVehicleStatistics godoc
// @Summary Get vehicle statistics
// @Description Get planning of total vehicles in and out with filters
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vehicle_type query string false "Filter by vehicle type (mobil/motor)"
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/statistics/vehicles [get]
func (h *Handlers) GetVehicleStatistics(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")
	regional := c.Query("regional")

	var vehicleTypePtr *string
	if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
		vehicleTypePtr = &vehicleType
	}

	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response, err := h.AdminUC.GetVehicleStatistics(startTime, endTime, vehicleTypePtr, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get vehicle statistics:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vehicle statistics retrieved successfully",
		"data":    response,
	})
}

// GetTotalRevenue godoc
// @Summary Get total revenue
// @Description Get total actual and estimated revenue from all jukirs
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vehicle_type query string false "Filter by vehicle type (mobil/motor)"
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/revenue/total [get]
func (h *Handlers) GetTotalRevenue(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")
	regional := c.Query("regional")

	var vehicleTypePtr *string
	if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
		vehicleTypePtr = &vehicleType
	}

	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response, err := h.AdminUC.GetTotalRevenue(startTime, endTime, vehicleTypePtr, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get total revenue:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Total revenue retrieved successfully",
		"data":    response,
	})
}

// GetJukirsListWithRevenue godoc
// @Summary Get all jukirs with revenue detail
// @Description Get list of all jukirs with actual and estimated revenue
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vehicle_type query string false "Filter by vehicle type (mobil/motor)"
// @Param date_range query string false "Filter by date range (hari_ini, minggu_ini, bulan_ini)"
// @Param export query string false "Export to Excel (true/false)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/list [get]
func (h *Handlers) GetJukirsListWithRevenue(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")
	dateRange := c.Query("date_range")
	includeRevenueStr := c.Query("include_revenue")
	status := c.Query("status")

	var vehicleTypePtr *string
	if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
		vehicleTypePtr = &vehicleType
	}

	var dateRangePtr *string
	if dateRange != "" && (dateRange == "hari_ini" || dateRange == "minggu_ini" || dateRange == "bulan_ini") {
		dateRangePtr = &dateRange
	}

	var includeRevenuePtr *bool
	if includeRevenueStr == "true" {
		includeRevenue := true
		includeRevenuePtr = &includeRevenue
	}

	var statusPtr *string
	if status != "" && (status == "active" || status == "inactive" || status == "pending") {
		statusPtr = &status
	}

	response, err := h.AdminUC.GetJukirsListWithRevenue(dateRangePtr, vehicleTypePtr, includeRevenuePtr, statusPtr)
	if err != nil {
		h.Logger.Error("Failed to get jukirs list:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukirs list retrieved successfully",
		"data":    response,
	})
}

// GetChartDataDetailed godoc
// @Summary Get detailed chart data
// @Description Get chart data with actual vs estimated revenue (minggu_ini or bulan_ini only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vehicle_type query string false "Filter by vehicle type (mobil/motor)"
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param regional query string false "Filter by regional"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/chart/data [get]
func (h *Handlers) GetChartDataDetailed(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")
	regional := c.Query("regional")

	var vehicleTypePtr *string
	if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
		vehicleTypePtr = &vehicleType
	}

	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	// Parse start_date and end_date
	startTime, endTime, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response, err := h.AdminUC.GetChartDataDetailed(startTime, endTime, vehicleTypePtr, regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get chart data:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chart data retrieved successfully",
		"data":    response,
	})
}

// GetParkingAreaStatistics godoc
// @Summary Get parking area statistics
// @Description Get count of active and inactive parking areas
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/statistics/areas [get]
func (h *Handlers) GetParkingAreaStatistics(c *gin.Context) {
	regional := c.Query("regional")

	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	response, err := h.AdminUC.GetParkingAreaStatistics(regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get parking area statistics:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking area statistics retrieved successfully",
		"data":    response,
	})
}

// GetJukirStatistics godoc
// @Summary Get jukir statistics
// @Description Get count of active and inactive jukirs
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/statistics/jukirs [get]
func (h *Handlers) GetJukirStatistics(c *gin.Context) {
	regional := c.Query("regional")

	var regionalPtr *string
	if regional != "" {
		regionalPtr = &regional
	}

	response, err := h.AdminUC.GetJukirStatistics(regionalPtr)
	if err != nil {
		h.Logger.Error("Failed to get jukir statistics:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir statistics retrieved successfully",
		"data":    response,
	})
}

// GetJukirByID godoc
// @Summary Get jukir by ID
// @Description Get detailed jukir information by ID
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Jukir ID"
// @Param include_revenue query boolean false "Include revenue data (true/false)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/:id [get]
func (h *Handlers) GetJukirByID(c *gin.Context) {
	jukirIDStr := c.Param("id")
	jukirID, err := strconv.ParseUint(jukirIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid jukir ID",
		})
		return
	}

	// Get include_revenue query param
	includeRevenue := c.Query("include_revenue") == "true"

	// Get date_range filter if provided
	dateRange := c.Query("date_range")
	var dateRangePtr *string
	if dateRange != "" && (dateRange == "hari_ini" || dateRange == "minggu_ini" || dateRange == "bulan_ini") {
		dateRangePtr = &dateRange
	} else if includeRevenue && dateRange == "" {
		// Default to minggu_ini if include_revenue is true but no date_range specified
		defaultDateRange := "minggu_ini"
		dateRangePtr = &defaultDateRange
	}

	response, err := h.AdminUC.GetJukirByID(uint(jukirID), dateRangePtr)
	if err != nil {
		h.Logger.Error("Failed to get jukir:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukir retrieved successfully",
		"data":    response,
	})
}

// GetJukirsWithRevenue godoc
// @Summary Get all jukirs with revenue
// @Description Get list of all jukirs with their revenue data filtered by date range
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date_range query string false "Filter by date range (hari_ini, minggu_ini, bulan_ini)" default(hari_ini)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/revenue [get]
func (h *Handlers) GetJukirsWithRevenue(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid limit",
		})
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid offset",
		})
		return
	}

	dateRange := c.DefaultQuery("date_range", "hari_ini")

	var dateRangePtr *string
	if dateRange != "" {
		switch dateRange {
		case "hari_ini", "minggu_ini", "bulan_ini":
			dateRangePtr = &dateRange
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid date_range. Use: hari_ini, minggu_ini, or bulan_ini",
			})
			return
		}
	}

	response, count, err := h.AdminUC.GetAllJukirsListWithRevenue(limit, offset, dateRangePtr)
	if err != nil {
		h.Logger.Error("Failed to get jukirs with revenue:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Jukirs with revenue retrieved successfully",
		"data":    response,
		"meta": gin.H{
			"pagination": gin.H{
				"total":  count,
				"limit":  limit,
				"offset": offset,
			},
			"filter": gin.H{
				"date_range": dateRange,
			},
		},
	})
}

// ImportAreasAndJukirsFromCSV godoc
// @Summary Import parking areas and jukirs from CSV
// @Description Import parking areas and jukirs from CSV file (format: FOLMULIR TITIK PARKIR)
// @Tags admin
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "CSV file to import"
// @Param regional formData string true "Regional (Barat, Utara, Selatan, Timur)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/import/areas-jukirs [post]
func (h *Handlers) ImportAreasAndJukirsFromCSV(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		h.Logger.Error("Failed to parse multipart form:", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid form data"})
		return
	}

	regional := c.PostForm("regional")
	if regional == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Regional is required"})
		return
	}

	validRegionals := []string{"Barat", "Utara", "Selatan", "Timur"}
	valid := false
	for _, r := range validRegionals {
		if r == regional {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid regional. Must be one of: Barat, Utara, Selatan, Timur"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.Logger.Error("Failed to get file:", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "File is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.Logger.Error("Failed to open file:", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Failed to open file"})
		return
	}
	defer file.Close()

	result, err := h.AdminUC.ImportAreasAndJukirsFromCSV(file, regional)
	if err != nil {
		h.Logger.Error("Failed to import CSV:", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "CSV imported successfully", "data": result})
}

// GetActivityLogs godoc
// @Summary Get detailed activity logs
// @Description Get checkin and checkout activity logs filtered by jukir or parking area
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param jukir_id query int false "Filter by jukir ID"
// @Param area_id query int false "Filter by parking area ID"
// @Param start_date query string false "Start date (DD-MM-YYYY)"
// @Param end_date query string false "End date (DD-MM-YYYY)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/activity-logs [get]
func (h *Handlers) GetActivityLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid limit",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid offset",
		})
		return
	}

	var jukirID *uint
	if jukirIDStr := c.Query("jukir_id"); jukirIDStr != "" {
		id, err := strconv.ParseUint(jukirIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid jukir_id",
			})
			return
		}
		parsed := uint(id)
		jukirID = &parsed
	}

	var areaID *uint
	if areaIDStr := c.Query("area_id"); areaIDStr != "" {
		id, err := strconv.ParseUint(areaIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid area_id",
			})
			return
		}
		parsed := uint(id)
		areaID = &parsed
	}

	if jukirID == nil && areaID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Either jukir_id or area_id is required",
		})
		return
	}

	startPtr, endPtr, err := parseDateFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	loc := getGMT7Location()
	var startTime, endTime time.Time
	if startPtr != nil && endPtr != nil {
		startTime = startPtr.In(loc)
		endTime = endPtr.In(loc).Add(time.Nanosecond)
	} else {
		now := time.Now().In(loc)
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		endTime = startTime.Add(24 * time.Hour)
	}

	response, err := h.AdminUC.GetActivityLogs(jukirID, areaID, startTime, endTime, limit, offset)
	if err != nil {
		h.Logger.Error("Failed to get activity logs:", err)
		status := http.StatusInternalServerError
		if err.Error() == "either jukir_id or area_id is required" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Activity logs retrieved successfully",
		"data":    response,
	})
}
