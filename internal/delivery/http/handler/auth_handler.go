package handler

import (
	"be-parkir/internal/domain/entities"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body entities.CreateUserRequest true "User registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/register [post]
func (h *Handlers) Register(c *gin.Context) {
	var req entities.CreateUserRequest
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

	response, err := h.AuthUC.Register(&req)
	if err != nil {
		h.Logger.Error("Registration failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data":    response,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body entities.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var req entities.LoginRequest
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

	response, err := h.AuthUC.Login(&req)
	if err != nil {
		h.Logger.Error("Login failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data":    response,
	})
}

// LoginJukir godoc
// @Summary Login jukir
// @Description Authenticate jukir by username and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Login credentials (username, password)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/login-jukir [post]
func (h *Handlers) LoginJukir(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	username, ok := req["username"]
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "username is required",
		})
		return
	}

	password, ok := req["password"]
	if !ok || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "password is required",
		})
		return
	}

	response, err := h.AuthUC.LoginJukir(username, password)
	if err != nil {
		h.Logger.Error("Login failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data":    response,
	})
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body entities.RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func (h *Handlers) RefreshToken(c *gin.Context) {
	var req entities.RefreshTokenRequest
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

	response, err := h.AuthUC.RefreshToken(&req)
	if err != nil {
		h.Logger.Error("Token refresh failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token refreshed successfully",
		"data":    response,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate user's refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func (h *Handlers) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Authorization header is required",
		})
		return
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.AuthUC.Logout(token)
	if err != nil {
		h.Logger.Error("Logout failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}
