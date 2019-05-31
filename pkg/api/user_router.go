package api

import (
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"regexp"
)

type UserRouter struct {
	service   model.UserService
	verifier  *jwtverifier.JwtVerifier
	Imaginary *conf.Imaginary
}

func InitUserRoutes(api *Server, service model.UserService, verifier *jwtverifier.JwtVerifier, imaginary *conf.Imaginary) error {
	userRouter := UserRouter{service: service, verifier: verifier, Imaginary: imaginary}

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

	if err := api.service.UpdateLastSeen(userObj); err != nil {
		return err
	}

	claims := jwt.MapClaims{"user": userObj.ID}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(api.Imaginary.Secret))
	if err != nil {
		zap.L().Error("Could not generate Imaginary token", zap.Error(err))
		return err
	}

	result := model.AppState{
		User: model.UserInfo{
			Id: userObj.ID,
		},
		ImaginaryJwt: token,
	}

	return ctx.JSON(http.StatusOK, result)
}
