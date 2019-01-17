package orm_test

import (
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MediaServiceTestSuite struct {
	suite.Suite
	db *orm.Database
}

func Test_MediaService(t *testing.T) {
	suite.Run(t, new(MediaServiceTestSuite))
}

func (suite *MediaServiceTestSuite) SetupTest() {
	dbConfig := conf.Database{
		Host:     "localhost",
		Port:     "5432",
		Database: "test_qilin",
		User:     "postgres",
		Password: "postgres",
	}

	db, err := orm.NewDatabase(&dbConfig)
	if err != nil {
		suite.Fail("Unable to connect to database: %s", err)
	}

	db.Init()

	suite.db = db
}

func (suite *MediaServiceTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Media{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *MediaServiceTestSuite) TestCreateGameShouldInsertIntoMongo() {
	// gameService, err := orm.NewMediaService(suite.db)

	// testUsername := "integration_test_user"
	// game := model.Media{
	// 	ID: uuid.NewV4(),
	// 	,
	// }

	// err = gameService.CreateGame(&game)
	// assert.Nil(suite.T(), err, "Unable to create game")
	// assert.NotEmpty(suite.T(), game.ID, "Wrong ID for created game")

	// gameFromDb, err := gameService.FindByID(game.ID)
	// assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	// assert.Equal(suite.T(), game.ID, gameFromDb.ID, "Incorrect Game ID from DB")
	// assert.Equal(suite.T(), game.Name, gameFromDb.Name, "Incorrect Game Name from DB")
}
