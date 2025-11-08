package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StreamJukirEvents streams real-time events to jukir clients via SSE
// @Summary Stream jukir events (SSE)
// @Description Get real-time updates for jukir sessions, payments, and stats via Server-Sent Events
// @Tags Jukir
// @Security BearerAuth
// @Produce text/event-stream
// @Success 200 {object} map[string]interface{} "SSE stream"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Not a jukir"
// @Router /jukir/events [get]
func (h *Handlers) StreamJukirEvents(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	// Get jukir info
	jukir, err := h.JukirUC.GetJukirByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Not a jukir",
			"error":   err.Error(),
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Register this jukir for events
	eventChan := h.EventManager.RegisterJukir(jukir.ID)
	defer h.EventManager.UnregisterJukir(jukir.ID)

	h.Logger.WithField("jukir_id", jukir.ID).Info("Jukir connected to SSE stream")

	// Send initial connection message
	c.SSEvent("connected", map[string]interface{}{
		"message":   "Connected to event stream",
		"jukir_id":  jukir.ID,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	c.Writer.Flush()

	// Stream events
	for {
		select {
		case event := <-eventChan:
			payload := map[string]interface{}{
				"type": event.Type,
				"data": event.Data,
			}
			c.SSEvent("", payload)
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			// Client disconnected
			h.Logger.WithField("jukir_id", jukir.ID).Info("Jukir disconnected from SSE stream")
			return
		}
	}
}

// GetEventStreamStatus returns the status of connected jukirs
// @Summary Get SSE connection status
// @Description Get the number of jukirs currently connected to the event stream
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Connection status"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /admin/sse-status [get]
func (h *Handlers) GetEventStreamStatus(c *gin.Context) {
	connectedJukirs := h.EventManager.GetConnectedJukirs()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"connected_jukirs": connectedJukirs,
			"timestamp":        time.Now().Format(time.RFC3339),
		},
	})
}
