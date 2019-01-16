package api

import (
	"github.com/labstack/echo"
	"net/http"
	"qilin-api/pkg/orm"
)

func (s *Server) QilinErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if _, ok := err.(*echo.HTTPError); ok {
		s.echo.DefaultHTTPErrorHandler(err, c)
		return
	} else
	if se, ok := err.(*orm.ServiceError); ok {
		msg = echo.Map{"message": se.Message, "code": se.Code}
	} else if s.echo.Debug {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg}
	}

	s.echo.Logger.Error(err)

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD { // Issue #608
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}
		if err != nil {
			s.echo.Logger.Error(err)
		}
	}
}