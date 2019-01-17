package game

import (
	"github.com/labstack/echo"
	"qilin-api/pkg/model"
)

type Router struct {
	gameService model.GameService
}

func InitRoutes(router *echo.Group, service model.GameService) error {
	Router := Router{
		gameService: service,
	}

	router.POST("/games", Router.create)

	return nil
}
