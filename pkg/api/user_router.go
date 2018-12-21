package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"net/http"
	"qilin-api/pkg/model"
	"time"
)

type UserRouter struct {
	service model.UserService
}

func InitUserRoutes(api *Server, service model.UserService) error {
	userRouter := UserRouter{service: service}

	api.AuthRouter.POST("/login", userRouter.login)
	api.Router.GET("/me", userRouter.getAppState)

	return nil
}

func (api *UserRouter) getAppState(ctx echo.Context) (err error) {

	token := ctx.Get("user").(*jwt.Token)
	if token == nil {
		return ctx.JSON(http.StatusUnauthorized, false)
	}
	userId := 0
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId = int(claims["user_id"].(float64))
	}
	if userId == 0 {
		return ctx.JSON(http.StatusNotFound, QilinError{"Invalid JWT Token"})
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

	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = result.AccessToken
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	ctx.SetCookie(cookie)
	return ctx.JSON(http.StatusOK, result)
}
