package handler

import (
	"be-parkir/internal/usecase"

	"github.com/sirupsen/logrus"
)

type Handlers struct {
	AuthUC    usecase.AuthUsecase
	UserUC    usecase.UserUsecase
	JukirUC   usecase.JukirUsecase
	ParkingUC usecase.ParkingUsecase
	AdminUC   usecase.AdminUsecase
	Logger    *logrus.Logger
}

func NewHandlers(authUC usecase.AuthUsecase, userUC usecase.UserUsecase, jukirUC usecase.JukirUsecase, parkingUC usecase.ParkingUsecase, adminUC usecase.AdminUsecase, logger *logrus.Logger) *Handlers {
	return &Handlers{
		AuthUC:    authUC,
		UserUC:    userUC,
		JukirUC:   jukirUC,
		ParkingUC: parkingUC,
		AdminUC:   adminUC,
		Logger:    logger,
	}
}
