package api

import (
	"github.com/labstack/echo"
	"qilin-api/pkg/model"
)

type PackageRouter struct {
	//gameService qilin.GameService
}

func InitPackageRoutes(api *Server, service model.GameService) error {
	packageRouter := PackageRouter{}

	router := api.Router.Group("/games/:id")

	router.GET("/packages", packageRouter.getAll)
	router.GET("/packages/:packageId", packageRouter.get)
	router.POST("/packages", packageRouter.createPackage)
	router.PUT("/packages/:packageId", packageRouter.updatePackage)
	router.DELETE("/packages/:packageId", packageRouter.deletePackage)

	return nil
}

func (api *PackageRouter) getAll(ctx echo.Context) error {
	panic("")
}

func (api *PackageRouter) get(ctx echo.Context) error {
	panic("")
}

func (api *PackageRouter) createPackage(ctx echo.Context) error {
	panic("")
}

func (api *PackageRouter) updatePackage(ctx echo.Context) error {
	panic("")
}

func (api *PackageRouter) deletePackage(ctx echo.Context) error {
	panic("")
}
