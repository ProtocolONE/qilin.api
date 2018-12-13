package api

import (
	"github.com/labstack/echo"
	"net/http"
	"qilin-api/pkg/model"
)

type UserRouter struct {
	service model.UserService
}

func InitUserRoutes(api *Server, service model.UserService) error {
	userRouter := UserRouter{
		service: service,
	}
	
	api.Router.POST("/auth", userRouter.auth)

	return nil
}

func (api *UserRouter) auth(ctx echo.Context) error {
	/*game := &model.Game{}

	if err := ctx.Bind(game); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	if err := api.service.CreateUser(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Game create failed")
	}*/

	return ctx.JSON(http.StatusOK, user)
}
