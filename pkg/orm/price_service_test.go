package orm

import (
	"time"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"
)

type PriceServiceTestSuite struct {
	suite.Suite
	db *Database
}

var (
	ID = "029ce039-888a-481a-a831-cde7ff4e50b9"
)

func Test_PriceService(t *testing.T) {
	suite.Run(t, new(PriceServiceTestSuite))
}

func (suite *PriceServiceTestSuite) SetupTest() {
	dbConfig := conf.Database{
		Host:     "localhost",
		Port:     "5432",
		Database: "test_qilin",
		User:     "postgres",
		Password: "postgres",
	}

	db, err := NewDatabase(&dbConfig)
	if err != nil {
		suite.Fail("Unable to connect to database: %s", err)
	}

	db.Init()

	id, _ := uuid.FromString(ID)
	db.DB().Save(&model.Game{ID: id, Name: "Test game"})

	suite.db = db
}

func (suite *PriceServiceTestSuite) TearDownTest() {
	// if err := suite.db.DB().DropTable(model.Media{}).Error; err != nil {
	// 	panic(err)
	// }
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}
func (suite *PriceServiceTestSuite) TestCreatePriceShouldChangeGameInDB() {
	service, err := NewPriceService(suite.db)
	updatedAt, _ := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")

	assert.Nil(suite.T(), err, "Unable to media service")

	id, _ := uuid.FromString(ID)
	game := model.Price{
		ID: uuid.NewV4(),
		Normal: model.JSONB {
			"currency": "USD",
			"price": 100,
		},
		PreOrder: model.JSONB {
			"date": "2019-01-22T07:53:16Z",
			"enabled": false,
		},
		Prices: []model.JSONB {
			model.JSONB { 
				"currency": "USD",
				"vat": 10,
				"price": 29.99,
			},
		},
		UpdatedAt: &updatedAt,
	}

	err = service.Update(id, &game);
	assert.Nil(suite.T(), err, "Unable to update media for game")

	gameFromDb, err := service.Get(id)
	assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	assert.NotNil(suite.T(), gameFromDb, "Unable to get game: %v", id)
	assert.Equal(suite.T(), game.ID, gameFromDb.ID, "Incorrect Game ID from DB")
	assert.Equal(suite.T(), game.Normal, gameFromDb.Normal, "Incorrect Normal from DB")
	assert.Equal(suite.T(), game.PreOrder, gameFromDb.PreOrder, "Incorrect PreOrder from DB")
	assert.Equal(suite.T(), game.Prices, gameFromDb.Prices, "Incorrect Prices from DB")
	assert.Equal(suite.T(), game.UpdatedAt, gameFromDb.UpdatedAt, "Incorrect UpdatedAt from DB")
}
