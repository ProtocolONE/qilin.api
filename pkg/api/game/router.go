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

	router.POST("/games", Router.Create)
	router.GET("/games", Router.GetList)
	router.GET("/games/:id", Router.GetInfo)
	router.DELETE("/games/:id", Router.Delete)
	router.PUT("/games/:id", Router.Update)

	return nil
}

