package mongo_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2/bson"
	"qilin-api/pkg"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/mongo"
	"testing"
)

const (
	mongoUrl = "localhost:27017"
	dbName   = "test_db"
)

type GameServiceTestSuite struct {
	suite.Suite
	session *mongo.Session
}

func Test_GameService(t *testing.T) {
	suite.Run(t, new(GameServiceTestSuite))
}

func (suite *GameServiceTestSuite) SetupTest() {
	mongoConfig := conf.Database{
		Host:     mongoUrl,
		Database: dbName}

	session, err := mongo.NewSession(&mongoConfig)
	if err != nil {
		suite.Fail("Unable to connect to mongo: %s", err)
	}

	suite.session = session
}

func (suite *GameServiceTestSuite) TearDownTest() {
	if err := suite.session.DropDatabase(); err != nil {
		panic(err)
	}
	suite.session.Close()
}

func (suite *GameServiceTestSuite) TestCreateGameShouldInsertIntoMongo() {
	gameService, err := mongo.NewGameService(suite.session)

	testUsername := "integration_test_user"
	game := qilin.Game{
		Name: testUsername,
	}

	err = gameService.CreateGame(&game)
	assert.Nil(suite.T(), err, "Unable to create game")
	assert.Truef(suite.T(), bson.IsObjectIdHex(game.ID), "Wrong ID for created game")

	gameFromDb, err := gameService.FindByID(game.ID)
	assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	assert.Equal(suite.T(), game.ID, gameFromDb.ID, "Incorrect Game ID from DB")
	assert.Equal(suite.T(), game.Name, gameFromDb.Name, "Incorrect Game Name from DB")
}
