package context

import (
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/labstack/echo/v4"
	"net/http"
)

const (
	TokenKey  = "user"
	LoggerKey = "app.logger"
)

func GetAuthUserId(ctx echo.Context) (externalUserId string, err error) {
	token := ctx.Get(TokenKey).(*jwtverifier.UserInfo)
	if token == nil {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid auth token: "+err.Error())
	}
	return token.UserID, nil
}
