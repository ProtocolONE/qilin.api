package context

import (
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	TokenKey  = "app.token"
	LoggerKey = "app.logger"
)

func getLogger(ctx echo.Context) *logrus.Entry {
	obj := ctx.Get(LoggerKey)
	if obj == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}

	return obj.(*logrus.Entry)
}

func getToken(ctx echo.Context) *jwt.Token {
	obj := ctx.Get(TokenKey)
	if obj == nil {
		return nil
	}

	return obj.(*jwt.Token)
}

func GetAuthUUID(ctx echo.Context) (result uuid.UUID, err error) {
	token := ctx.Get(TokenKey).(*jwt.Token)
	if token == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid auth token")
	}
	claims := token.Claims.(jwt.MapClaims)
	data, err := base64.StdEncoding.DecodeString(claims["id"].(string))
	if data == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid auth token")
	}
	uuidObj, err := uuid.FromBytes(data)
	if data == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid auth token")
	}
	return uuidObj, nil
}