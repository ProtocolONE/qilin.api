package utils_test

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"qilin-api/pkg/utils"
	"testing"
)

type JoinIntTestSuite struct {
	suite.Suite
}

func Test_JoinInt(t *testing.T) {
	suite.Run(t, new(JoinIntTestSuite))
}

func (suite *JoinIntTestSuite) TestGames() {
	require.Equal(suite.T(), utils.JoinInt([]int64{1, 2, 3, -45}, ","), "1,2,3,-45")
}
