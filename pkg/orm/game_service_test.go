package orm_test

import (
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"testing"
)

type GameServiceTestSuite struct {
	suite.Suite
	db *orm.Database
}

func Test_GameService(t *testing.T) {
	suite.Run(t, new(GameServiceTestSuite))
}

func (suite *GameServiceTestSuite) SetupTest() {
	dbConfig := conf.Database{
		Host:     "localhost",
		Port:     "5440",
		Database: "test_qilin",
		User:     "postgres",
		Password: "",
	}

	db, err := orm.NewDatabase(&dbConfig)
	if err != nil {
		suite.Fail("Unable to connect to database: %s", err)
	}

	db.Init()

	suite.db = db
}

func (suite *GameServiceTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Game{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *GameServiceTestSuite) TestCreateGameShouldInsertIntoMongo() {
	gameService, err := orm.NewGameService(suite.db)

	testUsername := "integration_test_user"
	game := model.Game{
		ID: uuid.NewV4(),
		Name: testUsername,
		Prices: model.Prices{
			USD: 10,
		},
	}

	err = gameService.CreateGame(&game)
	assert.Nil(suite.T(), err, "Unable to create game")
	assert.NotEmpty(suite.T(), game.ID, "Wrong ID for created game")

	gameFromDb, err := gameService.FindByID(game.ID)
	assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	assert.Equal(suite.T(), game.ID, gameFromDb.ID, "Incorrect Game ID from DB")
	assert.Equal(suite.T(), game.Name, gameFromDb.Name, "Incorrect Game Name from DB")
}
