package http

import (
	"be-parkir/internal/delivery/http/handler"
	"be-parkir/internal/delivery/http/route"
	"be-parkir/internal/usecase"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, handlers *handler.Handlers, jwtConfig usecase.JWTConfig) {
	route.SetupRoutes(router, handlers, jwtConfig)
}
