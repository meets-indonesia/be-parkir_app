package handler

import (
	"be-parkir/internal/storage"
	"be-parkir/internal/usecase"

	"github.com/sirupsen/logrus"
)

type Handlers struct {
	AuthUC       usecase.AuthUsecase
	UserUC       usecase.UserUsecase
	JukirUC      usecase.JukirUsecase
	ParkingUC    usecase.ParkingUsecase
	AdminUC      usecase.AdminUsecase
	EventManager *usecase.EventManager
	Logger       *logrus.Logger
	Storage      *storage.MinIOClient
}

func NewHandlers(authUC usecase.AuthUsecase, userUC usecase.UserUsecase, jukirUC usecase.JukirUsecase, parkingUC usecase.ParkingUsecase, adminUC usecase.AdminUsecase, eventManager *usecase.EventManager, logger *logrus.Logger, storage *storage.MinIOClient) *Handlers {
	return &Handlers{
		AuthUC:       authUC,
		UserUC:       userUC,
		JukirUC:      jukirUC,
		ParkingUC:    parkingUC,
		AdminUC:      adminUC,
		EventManager: eventManager,
		Logger:       logger,
		Storage:      storage,
	}
}
