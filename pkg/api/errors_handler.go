package api

import (
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"net/http"
	"qilin-api/pkg/orm"
)

func (s *Server) QilinErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*echo.HTTPError); ok {
		msg = echo.Map{"message": he.Message, "code": he.Code}
		code = he.Code
	} else if se, ok := err.(*orm.ServiceError); ok {
		msg = echo.Map{"message": se.Message, "code": se.Code}
		code = se.Code
	} else if s.echo.Debug {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg}
	}

	zap.L().Error("Http error handler", zap.Error(err))

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD { // Issue #608
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}

		if err != nil {
			zap.L().Error("Http error handler send response failed", zap.Error(err))
		}
	}
}
