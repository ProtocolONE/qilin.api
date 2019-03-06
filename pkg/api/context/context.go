package context

import (
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/labstack/echo"
	"net/http"
)

const (
	TokenKey  = "app.token"
	LoggerKey = "app.logger"
)

func GetAuthExternalUserId(ctx echo.Context) (externalUserId string, err error) {
	token := ctx.Get("user").(*jwtverifier.UserInfo)
	if token == nil {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid auth token: "+err.Error())
	}
	return token.UserID, nil
}
