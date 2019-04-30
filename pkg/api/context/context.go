package context

import (
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/labstack/echo/v4"
	"net/http"
	"qilin-api/pkg/orm"
)

const (
	TokenKey  = "user"
	LoggerKey = "app.logger"
)

func GetAuthUserId(ctx echo.Context) (externalUserId string, err error) {
	token := ctx.Get(TokenKey).(*jwtverifier.UserInfo)
	if token == nil {
		return "", orm.NewServiceError(http.StatusUnauthorized, "Invalid auth token")
	}
	return token.UserID, nil
}
