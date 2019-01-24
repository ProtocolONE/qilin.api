package orm

import (
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	if err := suite.db.DB().DropTable(model.Price{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.DB().DropTable(model.BasePrice{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}
func (suite *PriceServiceTestSuite) TestCreatePriceShouldChangeGameInDB() {
	service, err := NewPriceService(suite.db)
	updatedAt, _ := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")

	assert.Nil(suite.T(), err, "Unable to media service")

	id, _ := uuid.FromString(ID)
	game := model.BasePrice{
		ID: uuid.NewV4(),
		Common: model.JSONB{
			"currency": "USD",
			"price":    100.0,
		},
		PreOrder: model.JSONB{
			"date":    "2019-01-22T07:53:16Z",
			"enabled": false,
		},
		Prices: []model.Price{
			model.Price{BasePriceID: id, Price: 100.0, Vat: 32, Currency: "EUR"},
			model.Price{BasePriceID: id, Price: 93.23, Vat: 10, Currency: "RUR"},
		},
		UpdatedAt: &updatedAt,
	}

	err = service.Update(id, &game)
	assert.Nil(suite.T(), err, "Unable to update media for game")

	gameFromDb, err := service.Get(id)
	assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	assert.NotNil(suite.T(), gameFromDb, "Unable to get game: %v", id)
	assert.Equal(suite.T(), game.ID, gameFromDb.ID, "Incorrect Game ID from DB")
	assert.Equal(suite.T(), game.Common["currency"], gameFromDb.Common["currency"], "Incorrect Common from DB")
	assert.Equal(suite.T(), game.Common["price"], gameFromDb.Common["price"], "Incorrect Common from DB")
	assert.Equal(suite.T(), game.PreOrder, gameFromDb.PreOrder, "Incorrect PreOrder from DB")
	for i := 0; i < 2; i++ {
		assert.Equal(suite.T(), game.Prices[i].BasePriceID, gameFromDb.Prices[i].BasePriceID, "Incorrect Prices from DB")
		assert.Equal(suite.T(), game.Prices[i].Price, gameFromDb.Prices[i].Price, "Incorrect Prices from DB")
		assert.Equal(suite.T(), game.Prices[i].Currency, gameFromDb.Prices[i].Currency, "Incorrect Prices from DB")
	}
}
