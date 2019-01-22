package model_test

import (
	"qilin-api/pkg/model"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/suite"
)

type ExtensionSuite struct {
	suite.Suite
}

type TestType struct {
	ID        string
	UpdatedAt string
	Default   string
	PreOrder  string
	Prices    string
}

func Test_Extenstion(t *testing.T) {
	suite.Run(t, new(ExtensionSuite))
}

func (suite *ExtensionSuite) SetupTest() {
}

func (suite *ExtensionSuite) TearDownTest() {
}

func (suite *ExtensionSuite) TestTypeModel() {
	price := &TestType{}

	fields := model.SelectFields(price)
	expected := []string{"id", "updated_at", "default", "pre_order", "prices"}
	assert.Equal(suite.T(), expected, fields, "fields not equal")
}
