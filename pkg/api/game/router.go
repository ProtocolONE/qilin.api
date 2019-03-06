package game

import (
	"github.com/labstack/echo"
	"qilin-api/pkg/model"
)

type Router struct {
	gameService model.GameService
	userService model.UserService
}

func InitRoutes(router *echo.Group, service model.GameService, userService model.UserService) (*Router, error) {
	Router := Router{
		gameService: service,
	}

	router.POST("/games", Router.Create)
	router.GET("/vendor/:id/games", Router.GetList)
	router.GET("/games/:id", Router.GetInfo)
	router.DELETE("/games/:id", Router.Delete)
	router.PUT("/games/:id", Router.UpdateInfo)
	router.GET("/games/:id/descriptions", Router.GetDescr)
	router.PUT("/games/:id/descriptions", Router.UpdateDescr)

	router.GET("/genre", Router.GetGenres) // TODO: Remove after some time
	router.GET("/genres", Router.GetGenres)
	router.GET("/tags", Router.GetTags)
	router.GET("/descriptors", Router.GetRatingDescriptors)

	return &Router, nil
}
