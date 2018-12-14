package api

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"net/http"
	"qilin-api/pkg/model"
)

type UserRouter struct {
	service model.UserService
}

func InitUserRoutes(api *Server, service model.UserService) error {
	userRouter := UserRouter{service: service}

	api.AuthRouter.POST("/login", userRouter.login)

	return nil
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
		return err
	}
	result, err := api.service.Login(vals.Get("login"), vals.Get("password"))
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return ctx.JSON(http.StatusNotFound, false)
		default:
			return ctx.JSON(http.StatusInternalServerError, false)
		}
	}

	return ctx.JSON(http.StatusOK, result)
}
