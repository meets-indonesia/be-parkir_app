package main

import (
	"log"

	"be-parkir/internal/config"
	"be-parkir/internal/delivery/http"
	"be-parkir/internal/delivery/http/handler"
	"be-parkir/internal/repository"
	"be-parkir/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

	// Initialize use cases
	authUC := usecase.NewAuthUsecase(userRepo, redisClient, usecase.JWTConfig{
		SecretKey:     cfg.JWT.SecretKey,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	})
	userUC := usecase.NewUserUsecase(userRepo)
	jukirUC := usecase.NewJukirUsecase(jukirRepo, areaRepo, sessionRepo, paymentRepo)
	parkingUC := usecase.NewParkingUsecase(sessionRepo, areaRepo, userRepo, jukirRepo, paymentRepo)
	adminUC := usecase.NewAdminUsecase(userRepo, jukirRepo, areaRepo, sessionRepo, paymentRepo)

	// Initialize HTTP handlers
	handlers := handler.NewHandlers(authUC, userUC, jukirUC, parkingUC, adminUC, logger)

	// Setup routes
	router := gin.Default()
	http.SetupRoutes(router, handlers, usecase.JWTConfig{
		SecretKey:     cfg.JWT.SecretKey,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	})

	// Start server
	logger.Info("Starting server on port :8080")
	if err := router.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server:", err)
	}
}
