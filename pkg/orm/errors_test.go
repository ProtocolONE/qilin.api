package orm_test

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"qilin-api/pkg/orm"
	"testing"
)

func TestNewServiceError(t *testing.T) {
	err := errors.New("Some error")
	serviceErr := orm.NewServiceError(404, err)

	should := require.New(t)
	should.NotNil(serviceErr)
	should.Equal(404, serviceErr.Code)
	should.Equal("Some error", serviceErr.Message)
	should.Equal("code=404, message=Some error", serviceErr.Error())

	serviceErr = orm.NewServiceError(400, "")
	should.NotNil(serviceErr)
	should.Equal(400, serviceErr.Code)
	should.Equal("Bad Request", serviceErr.Message)

	serviceErr = orm.NewServiceError(401, err, "Another error", "End error")
	should.NotNil(serviceErr)
	should.Equal(401, serviceErr.Code)
	should.Equal("Some error", serviceErr.Message)

	serviceErr = orm.NewServiceError(422, "Another error", err, "End error")
	should.NotNil(serviceErr)
	should.Equal(422, serviceErr.Code)
	should.Equal("Another error", serviceErr.Message)
}

func TestNewServiceErrorf(t *testing.T) {
	should := require.New(t)

	serviceErr := orm.NewServiceErrorf(404, "Some error %s", "Hello Test")
	should.NotNil(serviceErr)
	should.Equal(404, serviceErr.Code)
	should.Equal("Some error Hello Test", serviceErr.Message)
}
