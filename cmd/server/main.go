package main

import (
	"log"

	_ "be-parkir/docs" // Import swagger docs
	"be-parkir/internal/config"
	"be-parkir/internal/delivery/http"
	"be-parkir/internal/delivery/http/handler"
	"be-parkir/internal/delivery/http/middleware"
	"be-parkir/internal/repository"
	"be-parkir/internal/storage"
	"be-parkir/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// @title Parking Digital API
// @version 1.0
// @description REST API for Palembang Digital Parking Application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Initialize database
	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database:", err)
	}

	// Initialize Redis
	redisClient, err := repository.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	jukirRepo := repository.NewJukirRepository(db)
	areaRepo := repository.NewParkingAreaRepository(db)
	sessionRepo := repository.NewParkingSessionRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	// Initialize Event Manager for SSE
	eventManager := usecase.NewEventManager()
	logger.Info("Event Manager initialized for SSE")

	// Initialize use cases
	authUC := usecase.NewAuthUsecase(userRepo, redisClient, usecase.JWTConfig{
		SecretKey:     cfg.JWT.SecretKey,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	})
	userUC := usecase.NewUserUsecase(userRepo)
	jukirUC := usecase.NewJukirUsecase(jukirRepo, areaRepo, sessionRepo, paymentRepo, eventManager)
	parkingUC := usecase.NewParkingUsecase(sessionRepo, areaRepo, userRepo, jukirRepo, paymentRepo, eventManager)
	adminUC := usecase.NewAdminUsecase(userRepo, jukirRepo, areaRepo, sessionRepo, paymentRepo)

	// Initialize MinIO storage client
	minioClient, err := storage.NewMinIOClient(cfg.MinIO)
	if err != nil {
		logger.Fatal("Failed to initialize MinIO:", err)
	}

	// Initialize HTTP handlers
	handlers := handler.NewHandlers(authUC, userUC, jukirUC, parkingUC, adminUC, eventManager, logger, minioClient)

	// Setup middleware configurations
	apiKeyConfig := &middleware.APIKeyConfig{
		APIKeys:    []string{viper.GetString("API_KEY")},
		HeaderName: viper.GetString("API_KEY_HEADER"),
		Required:   viper.GetBool("API_KEY_REQUIRED"),
	}

	corsConfig := &middleware.CORSConfig{
		AllowOrigins:     []string{viper.GetString("CORS_ALLOW_ORIGINS")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Total-Count"},
		AllowCredentials: viper.GetBool("CORS_ALLOW_CREDENTIALS"),
		MaxAge:           viper.GetInt("CORS_MAX_AGE"),
	}

	// Setup routes
	router := gin.Default()
	http.SetupRoutes(router, handlers, usecase.JWTConfig{
		SecretKey:     cfg.JWT.SecretKey,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	}, apiKeyConfig, corsConfig)

	// Start server
	logger.Info("Starting server on port :8080")
	if err := router.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server:", err)
	}
}
