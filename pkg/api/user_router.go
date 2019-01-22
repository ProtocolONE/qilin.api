package api

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/model"
	"time"
)

type UserRouter struct {
	service model.UserService
}

func InitUserRoutes(api *Server, service model.UserService) error {
	userRouter := UserRouter{service: service}

	api.AuthRouter.POST("/login", userRouter.login)
	api.AuthRouter.POST("/register", userRouter.register)
	api.AuthRouter.POST("/reset", userRouter.reset)
	api.Router.GET("/me", userRouter.getAppState)

	return nil
}

func (api *UserRouter) getAppState(ctx echo.Context) (err error) {

	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}

	userObj, err := api.service.FindByID(userId)
	if err != nil {
		return err
	}

	result := model.AppState{User: model.UserInfo{
		Id: userObj.ID,
		Nickname: userObj.Nickname,
	}}

	return ctx.JSON(http.StatusOK, result)
}

// @Summary Login user
// @Description Login user using Qilin login/password or facebook/google/VK tokens.
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} model.Merchant "OK"
// @Failure 404 {object} model.Error "User not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /auth-api/login [post]
func (api *UserRouter) login(ctx echo.Context) error {
	vals, err := ctx.FormParams()
	if err != nil {
		return errors.Wrap(err, "Parse form")
	}
	result, err := api.service.Login(vals.Get("login"), vals.Get("password"))
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = result.AccessToken
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	ctx.SetCookie(cookie)
	return ctx.JSON(http.StatusOK, result)
}

func (api *UserRouter) register(ctx echo.Context) error {
	vals, err := ctx.FormParams()
	if err != nil {
		return errors.Wrap(err, "when parse form in register")
	}
	result, err := api.service.Register(vals.Get("login"), vals.Get("password"), ctx.Request().Header.Get("Accept-Language"))
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, result.String())
}

func (api *UserRouter) reset(ctx echo.Context) error {
	vals, err := ctx.FormParams()
	if err != nil {
		return err
	}
	err = api.service.ResetPassw(vals.Get("email"))
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, true)
}
