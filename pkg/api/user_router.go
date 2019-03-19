package api

import (
	"github.com/labstack/echo"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/model"
)

type UserRouter struct {
	service model.UserService
}

func InitUserRoutes(api *Server, service model.UserService) error {
	userRouter := UserRouter{service: service}

	api.Router.GET("/me", userRouter.getAppState)

	return nil
}

func (api *UserRouter) getAppState(ctx echo.Context) (err error) {
	externalUserId, err := context.GetAuthUserId(ctx)
	if err != nil {
		return err
	}

	userObj, err := api.service.FindByID(externalUserId)
	if err != nil {
		userObj, err = api.service.Create(externalUserId, ctx.Request().Header.Get("Accept-Language"))
		if err != nil {
			return err
		}
	}

	result := model.AppState{User: model.UserInfo{
		Id: userObj.ID,
	}}

	return ctx.JSON(http.StatusOK, result)
}
