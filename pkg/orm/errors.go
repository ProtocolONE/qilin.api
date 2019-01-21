package orm

import (
	"fmt"
	"net/http"
)

type ServiceError struct {
	Code     int
	Message  interface{}
	Internal error
}

func (he *ServiceError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
}

func NewServiceError(code int, message ...interface{}) *ServiceError {
	he := &ServiceError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}
