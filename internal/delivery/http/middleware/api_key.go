package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyConfig represents the API key configuration
type APIKeyConfig struct {
	APIKeys    []string
	HeaderName string
	Required   bool
}

// DefaultAPIKeyConfig returns a default API key configuration
func DefaultAPIKeyConfig() *APIKeyConfig {
	return &APIKeyConfig{
		APIKeys: []string{
			"be-parkir-api-key-2025",
			"palembang-parking-secret",
			"dev-api-key-12345",
		},
		HeaderName: "X-API-Key",
		Required:   true,
	}
}

// APIKeyMiddleware validates API key from request headers
func APIKeyMiddleware(config *APIKeyConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultAPIKeyConfig()
	}

	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader(config.HeaderName)

		// If API key is required but not provided
		if config.Required && apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "API key is required",
				"error":   "Missing " + config.HeaderName + " header",
			})
			c.Abort()
			return
		}

		// If API key is provided, validate it
		if apiKey != "" {
			if !isValidAPIKey(apiKey, config.APIKeys) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Invalid API key",
					"error":   "The provided API key is not valid",
				})
				c.Abort()
				return
			}

			// Set API key info in context for logging/monitoring
			c.Set("api_key", maskAPIKey(apiKey))
		}

		c.Next()
	}
}

// isValidAPIKey checks if the provided API key is valid
func isValidAPIKey(providedKey string, validKeys []string) bool {
	providedKey = strings.TrimSpace(providedKey)

	for _, validKey := range validKeys {
		if providedKey == validKey {
			return true
		}
	}

	return false
}

// maskAPIKey masks the API key for logging (shows only first 8 characters)
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:8] + "****"
}

// OptionalAPIKeyMiddleware creates an optional API key middleware
func OptionalAPIKeyMiddleware(config *APIKeyConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultAPIKeyConfig()
	}

	// Make API key optional
	config.Required = false

	return APIKeyMiddleware(config)
}
