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
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/overview [get]
func (h *Handlers) GetAdminOverview(c *gin.Context) {
	response, err := h.AdminUC.GetOverview()
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
// @Description Create a new jukir account
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
