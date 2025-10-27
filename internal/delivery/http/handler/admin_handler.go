package handler

import (
	"be-parkir/internal/domain/entities"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GetAdminOverview godoc
// @Summary Get admin overview
// @Description Get system-wide statistics for admin dashboard
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vehicle_type query string false "Filter by vehicle type (mobil/motor)"
// @Param date_range query string false "Filter by date range (hari_ini, minggu_ini, bulan_ini, tahun_ini)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/overview [get]
func (h *Handlers) GetAdminOverview(c *gin.Context) {
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

	response, err := h.AdminUC.GetOverview(vehicleTypePtr, dateRangePtr)
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
			"date_range":   dateRange,
		},
	})
}

// GetJukirs godoc
// @Summary Get all jukirs
// @Description Get list of all jukirs with pagination
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
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
	jukirs, count, err := h.AdminUC.GetJukirs(limit, offset)
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
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body entities.CreateParkingAreaRequest true "Parking area data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas [post]
func (h *Handlers) CreateParkingArea(c *gin.Context) {
	var req entities.CreateParkingAreaRequest
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

	response, err := h.AdminUC.CreateParkingArea(&req)
	if err != nil {
		h.Logger.Error("Failed to create parking area:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Parking area created successfully",
		"data":    response,
	})
}

// GetParkingAreas godoc
// @Summary Get all parking areas with status
// @Description Get list of all parking areas for admin area-parkir menu
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/areas [get]
func (h *Handlers) GetParkingAreas(c *gin.Context) {
	response, err := h.AdminUC.GetParkingAreas()
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
// @Description Get transaction details by parking area
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Area ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
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

	response, count, err := h.AdminUC.GetAreaTransactions(uint(areaID), limit, offset)
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

	response, count, err := h.AdminUC.GetRevenueTable(limit, offset, areaID)
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

	var req entities.UpdateParkingAreaRequest
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
// @Param date_range query string false "Filter by date range (hari_ini, minggu_ini, bulan_ini, tahun_ini)" default(hari_ini)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/jukirs/revenue [get]
func (h *Handlers) GetAllJukirsRevenue(c *gin.Context) {
	dateRange := c.Query("date_range")

	var dateRangePtr *string
	if dateRange != "" && (dateRange == "hari_ini" || dateRange == "minggu_ini" || dateRange == "bulan_ini" || dateRange == "tahun_ini") {
		dateRangePtr = &dateRange
	}

	response, err := h.AdminUC.GetAllJukirsRevenue(dateRangePtr)
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
// @Param date_range query string false "Filter by date range (hari_ini, minggu_ini, bulan_ini)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/statistics/vehicles [get]
func (h *Handlers) GetVehicleStatistics(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")
	dateRange := c.Query("date_range")

	var vehicleTypePtr *string
	if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
		vehicleTypePtr = &vehicleType
	}

	var dateRangePtr *string
	if dateRange != "" && (dateRange == "hari_ini" || dateRange == "minggu_ini" || dateRange == "bulan_ini") {
		dateRangePtr = &dateRange
	}

	response, err := h.AdminUC.GetVehicleStatistics(dateRangePtr, vehicleTypePtr)
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
// @Param date_range query string false "Filter by date range (hari_ini, minggu_ini, bulan_ini)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/revenue/total [get]
func (h *Handlers) GetTotalRevenue(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")
	dateRange := c.Query("date_range")

	var vehicleTypePtr *string
	if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
		vehicleTypePtr = &vehicleType
	}

	var dateRangePtr *string
	if dateRange != "" && (dateRange == "hari_ini" || dateRange == "minggu_ini" || dateRange == "bulan_ini") {
		dateRangePtr = &dateRange
	}

	response, err := h.AdminUC.GetTotalRevenue(dateRangePtr, vehicleTypePtr)
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
// @Param date_range query string false "Filter by date range (minggu_ini, bulan_ini)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/chart/data [get]
func (h *Handlers) GetChartDataDetailed(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")
	dateRange := c.Query("date_range")

	var vehicleTypePtr *string
	if vehicleType != "" && (vehicleType == "mobil" || vehicleType == "motor") {
		vehicleTypePtr = &vehicleType
	}

	var dateRangePtr *string
	if dateRange != "" && (dateRange == "minggu_ini" || dateRange == "bulan_ini") {
		dateRangePtr = &dateRange
	}

	response, err := h.AdminUC.GetChartDataDetailed(dateRangePtr, vehicleTypePtr)
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
	response, err := h.AdminUC.GetParkingAreaStatistics()
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
	response, err := h.AdminUC.GetJukirStatistics()
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
	dateRange := c.DefaultQuery("date_range", "hari_ini")

	var dateRangePtr *string
	if dateRange != "" {
		if dateRange == "hari_ini" || dateRange == "minggu_ini" || dateRange == "bulan_ini" {
			dateRangePtr = &dateRange
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid date_range. Use: hari_ini, minggu_ini, or bulan_ini",
			})
			return
		}
	}

	response, count, err := h.AdminUC.GetAllJukirsListWithRevenue(dateRangePtr)
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
				"total": count,
			},
			"filter": gin.H{
				"date_range": dateRange,
			},
		},
	})
}
