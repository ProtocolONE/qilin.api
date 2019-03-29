package context

import (
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
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

func GetLang(ctx echo.Context) (lang string) {
	lang = ctx.Request().Header.Get("Accept-Language")
	dashIdx := strings.Index(lang, "-")
	if dashIdx > -1 {
		lang = lang[:dashIdx]
	}
	if lang == "" {
		lang = "en"
	}
	return
}

