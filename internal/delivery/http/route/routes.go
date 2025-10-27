package route

import (
	"be-parkir/internal/delivery/http/handler"
	"be-parkir/internal/delivery/http/middleware"
	"be-parkir/internal/usecase"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(router *gin.Engine, handlers *handler.Handlers, jwtConfig usecase.JWTConfig, apiKeyConfig *middleware.APIKeyConfig, corsConfig *middleware.CORSConfig) {
	// Apply CORS middleware globally
	router.Use(middleware.CORS(corsConfig))

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check (no API key required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Parking Digital API is running",
		})
	})

	// API v1 routes with API key middleware
	v1 := router.Group("/api/v1")
	v1.Use(middleware.APIKeyMiddleware(apiKeyConfig))
	{
		// Auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/login-jukir", handlers.LoginJukir)
			auth.POST("/refresh", handlers.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(handlers.AuthUC), handlers.Logout)
		}

		// User profile routes (all authenticated users)
		user := v1.Group("/")
		user.Use(middleware.AuthMiddleware(handlers.AuthUC))
		{
			user.GET("/profile", handlers.GetProfile)
			user.PUT("/profile", handlers.UpdateProfile)
		}

		// Parking routes (anonymous)
		parking := v1.Group("/parking")
		{
			parking.GET("/locations", handlers.GetNearbyAreas)
			parking.POST("/checkin", handlers.Checkin)
			parking.POST("/checkout", handlers.Checkout)
			parking.GET("/active", handlers.GetActiveSession)
			parking.GET("/history", handlers.GetParkingHistory)
		}

		// Jukir routes
		jukir := v1.Group("/jukir")
		jukir.Use(middleware.AuthMiddleware(handlers.AuthUC), middleware.JukirMiddleware(handlers.JukirUC))
		{
			jukir.GET("/dashboard", handlers.GetJukirDashboard)
			jukir.GET("/pending-payments", handlers.GetPendingPayments)
			jukir.GET("/active-sessions", handlers.GetActiveSessions)
			jukir.GET("/vehicle-breakdown", handlers.GetVehicleBreakdown)
			jukir.POST("/confirm-payment", handlers.ConfirmPayment)
			jukir.GET("/qr-code", handlers.GetQRCode)
			jukir.GET("/daily-report", handlers.GetDailyReport)
			jukir.POST("/manual-checkin", handlers.ManualCheckin)
			jukir.POST("/manual-checkout", handlers.ManualCheckout)
			jukir.GET("/events", handlers.StreamJukirEvents) // SSE endpoint
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(handlers.AuthUC), middleware.AdminMiddleware())
		{
			admin.GET("/overview", handlers.GetAdminOverview)
			admin.GET("/jukirs", handlers.GetJukirs)
			admin.GET("/jukirs/revenue", handlers.GetJukirsWithRevenue)
			admin.GET("/jukirs/list", handlers.GetJukirsListWithRevenue)
			admin.GET("/jukirs/:id", handlers.GetJukirByID)
			admin.POST("/jukirs", handlers.CreateJukir)
			admin.PUT("/jukirs/:id", handlers.UpdateJukir)
			admin.DELETE("/jukirs/:id", handlers.DeleteJukir)
			admin.POST("/jukirs/manual-revenue", handlers.AddManualRevenue)
			admin.PUT("/jukirs/:id/status", handlers.UpdateJukirStatus)
			admin.GET("/statistics/vehicles", handlers.GetVehicleStatistics)
			admin.GET("/statistics/areas", handlers.GetParkingAreaStatistics)
			admin.GET("/statistics/jukirs", handlers.GetJukirStatistics)
			admin.GET("/revenue/total", handlers.GetTotalRevenue)
			admin.GET("/chart/data", handlers.GetChartDataDetailed)
			admin.GET("/reports", handlers.GetReports)
			admin.GET("/sessions", handlers.GetAllSessions)
			admin.GET("/areas", handlers.GetParkingAreas)
			admin.GET("/areas/:id", handlers.GetParkingAreaDetail)
			admin.GET("/areas/:id/status", handlers.GetParkingAreaStatus)
			admin.GET("/areas/:id/transactions", handlers.GetAreaTransactions)
			admin.POST("/areas", handlers.CreateParkingArea)
			admin.PUT("/areas/:id", handlers.UpdateParkingArea)
			admin.DELETE("/areas/:id", handlers.DeleteParkingArea)
			admin.GET("/revenue-table", handlers.GetRevenueTable)
			admin.GET("/sse-status", handlers.GetEventStreamStatus) // SSE connection status
		}
	}
}
