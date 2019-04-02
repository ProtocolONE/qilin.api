package api

import (
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"regexp"
)

type UserRouter struct {
	service  model.UserService
	verifier *jwtverifier.JwtVerifier
}

func InitUserRoutes(api *Server, service model.UserService, verifier *jwtverifier.JwtVerifier) error {
	userRouter := UserRouter{service: service, verifier: verifier}

	api.Router.GET("/me", userRouter.getAppState)

	return nil
}

func (api *UserRouter) getAppState(ctx echo.Context) (err error) {
	externalUserId, err := context.GetAuthUserId(ctx)
	if err != nil {
		return err
	}

	userObj, err := api.service.FindByID(externalUserId)
	if err != nil || userObj.Email == "" {
		auth := ctx.Request().Header.Get("Authorization")
		r := regexp.MustCompile("Bearer ([A-z0-9_.-]{10,})")
		match := r.FindStringSubmatch(auth)

		u, err := api.verifier.GetUserInfo(ctx.Request().Context(), match[1])
		if err != nil {
			return orm.NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Get user info from oauth2"))
		}

		userObj, err = api.service.Create(externalUserId, u.Email, ctx.Request().Header.Get("Accept-Language"))
		if err != nil {
			return err
		}
	}

	result := model.AppState{User: model.UserInfo{
		Id: userObj.ID,
	}}

	return ctx.JSON(http.StatusOK, result)
}
