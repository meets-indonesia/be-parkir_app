package handler

import (
	"be-parkir/internal/domain/entities"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GetJukirDashboard godoc
// @Summary Get jukir dashboard
// @Description Get jukir dashboard statistics
// @Tags jukir
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/jukir/dashboard [get]
func (h *Handlers) GetJukirDashboard(c *gin.Context) {
	jukirID, exists := c.Get("jukir_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Jukir not authenticated",
		})
		return
	}

	response, err := h.JukirUC.GetDashboard(jukirID.(uint))
	if err != nil {
		h.Logger.Error("Failed to get dashboard:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dashboard data retrieved successfully",
		"data":    response,
	})
}

// GetPendingPayments godoc
// @Summary Get pending payments
// @Description Get list of customers with pending payments
// @Tags jukir
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/jukir/pending-payments [get]
func (h *Handlers) GetPendingPayments(c *gin.Context) {
	jukirID, exists := c.Get("jukir_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Jukir not authenticated",
		})
		return
	}

	response, err := h.JukirUC.GetPendingPayments(jukirID.(uint))
	if err != nil {
		h.Logger.Error("Failed to get pending payments:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pending payments retrieved successfully",
		"data":    response,
	})
}

// GetActiveSessions godoc
// @Summary Get active sessions
// @Description Get list of active parking sessions in jukir's area (optional filter by vehicle_type)
// @Tags jukir
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vehicle_type query string false "Filter by vehicle type: mobil or motor"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/jukir/active-sessions [get]
func (h *Handlers) GetActiveSessions(c *gin.Context) {
	jukirID, exists := c.Get("jukir_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Jukir not authenticated",
		})
		return
	}

	// Get optional vehicle_type filter from query parameter
	var vehicleType *entities.VehicleType
	vehicleTypeStr := c.Query("vehicle_type")
	if vehicleTypeStr != "" {
		vt := entities.VehicleType(vehicleTypeStr)
		// Validate vehicle type
		if vt != entities.VehicleTypeMobil && vt != entities.VehicleTypeMotor {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid vehicle_type. Must be 'mobil' or 'motor'",
			})
			return
		}
		vehicleType = &vt
	}

	response, err := h.JukirUC.GetActiveSessions(jukirID.(uint), vehicleType)
	if err != nil {
		h.Logger.Error("Failed to get active sessions:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Active sessions retrieved successfully",
		"data":    response,
	})
}

// GetQRCode godoc
// @Summary Get QR code info
// @Description Get jukir's QR code information
// @Tags jukir
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/jukir/qr-code [get]
func (h *Handlers) GetQRCode(c *gin.Context) {
	jukirID, exists := c.Get("jukir_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Jukir not authenticated",
		})
		return
	}

	response, err := h.JukirUC.GetQRCode(jukirID.(uint))
	if err != nil {
		h.Logger.Error("Failed to get QR code:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "QR code retrieved successfully",
		"data":    response,
	})
}

// GetDailyReport godoc
// @Summary Get daily report
// @Description Get daily transaction summary for jukir
// @Tags jukir
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date query string false "Date (YYYY-MM-DD)" default(today)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/jukir/daily-report [get]
func (h *Handlers) GetDailyReport(c *gin.Context) {
	jukirID, exists := c.Get("jukir_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Jukir not authenticated",
		})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	response, err := h.JukirUC.GetDailyReport(jukirID.(uint), date)
	if err != nil {
		h.Logger.Error("Failed to get daily report:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Daily report retrieved successfully",
		"data":    response,
	})
}

// GetVehicleBreakdown godoc
// @Summary Get vehicle breakdown (in/out)
// @Description Get breakdown of vehicles entered and exited for jukir area
// @Tags jukir
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/jukir/vehicle-breakdown [get]
func (h *Handlers) GetVehicleBreakdown(c *gin.Context) {
	jukirID, exists := c.Get("jukir_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Jukir not authenticated",
		})
		return
	}

	response, err := h.JukirUC.GetVehicleBreakdown(jukirID.(uint))
	if err != nil {
		h.Logger.Error("Failed to get vehicle breakdown:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vehicle breakdown retrieved successfully",
		"data":    response,
	})
}

// ManualCheckin godoc
// @Summary Manual check-in
// @Description Create manual parking record for check-in
// @Tags jukir
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body entities.ManualCheckinRequest true "Manual check-in data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/jukir/manual-checkin [post]
func (h *Handlers) ManualCheckin(c *gin.Context) {
	jukirID, exists := c.Get("jukir_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Jukir not authenticated",
		})
		return
	}

	var req entities.ManualCheckinRequest
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

	response, err := h.ParkingUC.ManualCheckin(jukirID.(uint), &req)
	if err != nil {
		h.Logger.Error("Manual check-in failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Manual check-in successful",
		"data":    response,
	})
}

// ManualCheckout godoc
// @Summary Manual check-out
// @Description Create manual parking record for check-out
// @Tags jukir
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body entities.ManualCheckoutRequest true "Manual check-out data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/jukir/manual-checkout [post]
func (h *Handlers) ManualCheckout(c *gin.Context) {
	jukirID, exists := c.Get("jukir_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Jukir not authenticated",
		})
		return
	}

	var req entities.ManualCheckoutRequest
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

	response, err := h.ParkingUC.ManualCheckout(jukirID.(uint), &req)
	if err != nil {
		h.Logger.Error("Manual check-out failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Manual check-out successful",
		"data":    response,
	})
}
