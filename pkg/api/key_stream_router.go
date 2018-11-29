package api

import (
	"github.com/labstack/echo"
	"qilin-api/pkg/model"
)

type KeyStreamRouter struct {
	//gameService qilin.GameService
}

func InitKeyStreamRoutes(api *Server, service model.GameService) error {
	packageRouter := KeyStreamRouter{}

	router := api.Router.Group("/games/:id/packages/:packageId")

	router.GET("/streams", packageRouter.getAll)
	router.GET("/streams/:streamId", packageRouter.get)
	router.POST("/streams", packageRouter.createKeyStream)
	router.PUT("/streams/:streamId", packageRouter.updateKeyStream)
	router.DELETE("/streams/:streamId", packageRouter.deleteKeyStream)

	return nil
}

func (api *KeyStreamRouter) getAll(ctx echo.Context) error {
	panic("")
}

func (api *KeyStreamRouter) get(ctx echo.Context) error {
	panic("")
}

func (api *KeyStreamRouter) createKeyStream(ctx echo.Context) error {
	panic("")
}

func (api *KeyStreamRouter) updateKeyStream(ctx echo.Context) error {
	panic("")
}

func (api *KeyStreamRouter) deleteKeyStream(ctx echo.Context) error {
	panic("")
}
