package middleware

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/usecase"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token and sets user context
func AuthMiddleware(authUC usecase.AuthUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate token
		token, err := authUC.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Get user from token
		user, err := authUC.GetUserFromToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("user", user)

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// Convert userRole to string for comparison
		userRoleStr := string(userRole.(entities.UserRole))

		if userRoleStr != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// JukirMiddleware checks if user is a jukir and sets jukir context
func JukirMiddleware(jukirUC usecase.JukirUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// Check if user has jukir role first
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// Convert userRole to string for comparison
		userRoleStr := string(userRole.(entities.UserRole))
		if userRoleStr != string(entities.RoleJukir) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "User is not a jukir",
			})
			c.Abort()
			return
		}

		// Get jukir profile for user
		jukir, err := jukirUC.GetJukirByUserID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Jukir profile not found",
			})
			c.Abort()
			return
		}

		// Check if jukir is active
		if jukir.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Jukir account is not active",
			})
			c.Abort()
			return
		}

		// Set jukir context
		c.Set("jukir_id", jukir.ID)
		c.Set("jukir", jukir)

		c.Next()
	}
}

// AdminMiddleware checks if user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// CustomerMiddleware checks if user is a customer
func CustomerMiddleware() gin.HandlerFunc {
	return RoleMiddleware("customer")
}
