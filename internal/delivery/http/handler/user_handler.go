package handler

import (
	"be-parkir/internal/domain/entities"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/profile [get]
func (h *Handlers) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	user, err := h.UserUC.GetProfile(userID.(uint))
	if err != nil {
		h.Logger.Error("Failed to get profile:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile retrieved successfully",
		"data":    user,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body entities.UpdateUserRequest true "Profile update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/profile [put]
func (h *Handlers) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req entities.UpdateUserRequest
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

	user, err := h.UserUC.UpdateProfile(userID.(uint), &req)
	if err != nil {
		h.Logger.Error("Failed to update profile:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data":    user,
	})
}

// GetNearbyAreas godoc
// @Summary Get nearby parking areas
// @Description Get parking areas within specified radius of user's location
// @Tags parking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body entities.NearbyAreasRequest true "Location data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/parking/locations [get]
func (h *Handlers) GetNearbyAreas(c *gin.Context) {
	// Get query parameters
	latitudeStr := c.Query("latitude")
	longitudeStr := c.Query("longitude")
	radiusStr := c.DefaultQuery("radius", "1.0")

	if latitudeStr == "" || longitudeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "latitude and longitude are required",
		})
		return
	}

	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid latitude format",
		})
		return
	}

	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid longitude format",
		})
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		radius = 1.0
	}

	req := &entities.NearbyAreasRequest{
		Latitude:  latitude,
		Longitude: longitude,
		Radius:    radius,
	}

	response, err := h.ParkingUC.GetNearbyAreas(req)
	if err != nil {
		h.Logger.Error("Failed to get nearby areas:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Nearby areas retrieved successfully",
		"data":    response,
	})
}

// Checkin godoc
// @Summary Check in to parking
// @Description Start a parking session by scanning QR code (anonymous)
// @Tags parking
// @Accept json
// @Produce json
// @Param request body entities.CheckinRequest true "Check-in data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/parking/checkin [post]
func (h *Handlers) Checkin(c *gin.Context) {
	// No authentication required for anonymous parking

	var req entities.CheckinRequest
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

	response, err := h.ParkingUC.Checkin(&req)
	if err != nil {
		h.Logger.Error("Check-in failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Check-in successful",
		"data":    response,
	})
}

// Checkout godoc
// @Summary Check out from parking
// @Description End a parking session by scanning QR code (anonymous)
// @Tags parking
// @Accept json
// @Produce json
// @Param request body entities.CheckoutRequest true "Check-out data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/parking/checkout [post]
func (h *Handlers) Checkout(c *gin.Context) {
	// No authentication required for anonymous parking

	var req entities.CheckoutRequest
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

	response, err := h.ParkingUC.Checkout(&req)
	if err != nil {
		h.Logger.Error("Check-out failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Check-out successful",
		"data":    response,
	})
}

// GetActiveSession godoc
// @Summary Get active parking session
// @Description Get active parking session by session ID (anonymous)
// @Tags parking
// @Accept json
// @Produce json
// @Param id path int true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/parking/active/{id} [get]
func (h *Handlers) GetActiveSession(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Session ID is required",
		})
		return
	}

	sessionID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || sessionID64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid session ID",
		})
		return
	}

	response, err := h.ParkingUC.GetActiveSessionByID(uint(sessionID64))
	if err != nil {
		h.Logger.Error("Failed to get active session:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Active session retrieved successfully",
		"data":    response,
	})
}

// GetParkingHistory godoc
// @Summary Get parking history
// @Description Get parking session history by license plate or session ID (anonymous)
// @Tags parking
// @Accept json
// @Produce json
// @Param plat_nomor query string false "License plate number"
// @Param session_id query int false "Session ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/parking/history [get]
func (h *Handlers) GetParkingHistory(c *gin.Context) {
	platNomor := c.Query("plat_nomor")
	sessionIDStr := c.Query("session_id")

	// Require at least one: plat_nomor OR session_id
	if platNomor == "" && sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Either license plate number or session ID is required",
		})
		return
	}

	// If session_id is provided, return single session
	if sessionIDStr != "" {
		sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid session ID",
			})
			return
		}

		session, err := h.ParkingUC.GetHistoryBySession(uint(sessionID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Parking history retrieved successfully",
			"data": gin.H{
				"sessions": []entities.ParkingSession{*session},
				"count":    1,
			},
		})
		return
	}

	// Query by plat_nomor
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

	response, err := h.ParkingUC.GetHistoryByPlatNomor(platNomor, limit, offset)
	if err != nil {
		h.Logger.Error("Failed to get parking history:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking history retrieved successfully",
		"data":    response,
		"meta": gin.H{
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
				"total":  response.Count,
			},
		},
	})
}

// GetParkingHistoryByIDs godoc
// @Summary Get parking history by session IDs (bulk)
// @Description Get parking sessions by array of session IDs (anonymous, supports bulk request)
// @Tags parking
// @Accept json
// @Produce json
// @Param request body map[string][]uint true "Session IDs array" Example({"session_ids": [1, 2, 3]})
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/parking/history [post]
func (h *Handlers) GetParkingHistoryByIDs(c *gin.Context) {
	// Ensure this is a POST request
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"message": "Method not allowed. Use POST for bulk session history request.",
		})
		return
	}

	var req struct {
		SessionIDs []uint `json:"session_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body. Expected: {\"session_ids\": [1, 2, 3]}",
			"error":   err.Error(),
		})
		return
	}

	if len(req.SessionIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "session_ids array cannot be empty",
		})
		return
	}

	// Limit bulk requests to prevent abuse (max 100 sessions at once)
	if len(req.SessionIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Maximum 100 session IDs allowed per request",
		})
		return
	}

	sessions, err := h.ParkingUC.GetHistoryBySessionIDs(req.SessionIDs)
	if err != nil {
		h.Logger.Error("Failed to get parking history:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Parking history retrieved successfully",
		"data": gin.H{
			"sessions": sessions,
			"count":    len(sessions),
		},
	})
}
