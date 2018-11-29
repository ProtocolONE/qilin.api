package api

import (
	"github.com/labstack/echo"
	"qilin-api/pkg/model"
)

type PriceRouter struct {
	gameService model.GameService
}

func InitPriceRoutes(api *Server, service model.GameService) error {
	packageRouter := PriceRouter{
		gameService: service,
	}

	router := api.Router.Group("/games/:id")

	router.GET("/prices/:packageId", packageRouter.get)
	router.POST("/prices", packageRouter.createPrice)
	router.PUT("/prices", packageRouter.updatePrice)

	return nil
}

func (api *PriceRouter) get(ctx echo.Context) error {
	panic("")
}

func (api *PriceRouter) createPrice(ctx echo.Context) error {
	panic("")
}

func (api *PriceRouter) updatePrice(ctx echo.Context) error {
	panic("")
}
