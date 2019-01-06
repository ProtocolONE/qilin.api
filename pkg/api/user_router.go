package api

import (
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
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

	token := ctx.Get("user").(*jwt.Token)
	userId := uuid.UUID{}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		data, _ := base64.StdEncoding.DecodeString(claims["id"].(string))
		userId, _ = uuid.FromBytes(data)
	}
	if userId == uuid.Nil {
		return ctx.JSON(http.StatusNotFound, "Invalid JWT Token")
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
			return ctx.JSON(http.StatusNotFound, "User not found")
		default:
			return ctx.JSON(http.StatusInternalServerError, "Server error")
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


func (api *UserRouter) register(ctx echo.Context) error {
	vals, err := ctx.FormParams()
	if err != nil {
		return err
	}
	result, err := api.service.Register(vals.Get("login"), vals.Get("password"))
	if err != nil {
		switch err {
		case orm.ErrLoginAlreadyTaken:
			return ctx.JSON(http.StatusBadRequest, err.Error())
		default:
			return ctx.JSON(http.StatusInternalServerError, "Server error")
		}
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
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, true)
}
