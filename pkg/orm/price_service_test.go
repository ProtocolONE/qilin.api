package orm

import (
	"qilin-api/pkg/model"
	"qilin-api/pkg/test"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/satori/go.uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}
	_ = db.DropAllTables()
	db.Init()

	id, _ := uuid.FromString(ID)
	err = db.DB().Save(&model.Game{
		ID:             id,
		InternalName:   "Test_game_2",
		ReleaseDate:    time.Now(),
		Genre:          pq.StringArray{},
		Tags:           pq.StringArray{},
		FeaturesCommon: pq.StringArray{},
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	suite.db = db
}

func (suite *PriceServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
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
			{BasePriceID: id, Price: 100.0, Vat: 32, Currency: "EUR"},
			{BasePriceID: id, Price: 93.23, Vat: 10, Currency: "RUR"},
		},
		UpdatedAt: &updatedAt,
	}

	err = service.UpdateBase(id, &game)
	assert.Nil(suite.T(), err, "Unable to update media for game")

	gameFromDb, err := service.GetBase(id)
	assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	assert.NotNil(suite.T(), gameFromDb, "Unable to get game: %v", id)
	assert.Equal(suite.T(), game.ID, gameFromDb.ID, "Incorrect Game ID from DB")
	assert.Equal(suite.T(), game.Common["currency"], gameFromDb.Common["currency"], "Incorrect Common from DB")
	assert.Equal(suite.T(), game.Common["price"], gameFromDb.Common["price"], "Incorrect Common from DB")
	assert.Equal(suite.T(), game.PreOrder, gameFromDb.PreOrder, "Incorrect PreOrder from DB")
}

func (suite *PriceServiceTestSuite) TestChangePrices() {
	service, err := NewPriceService(suite.db)

	assert.Nil(suite.T(), err, "Unable to media service")

	id, _ := uuid.FromString(ID)

	price1 := model.Price{
		Currency: "USD",
		Price:    123.32,
		Vat:      10,
	}

	price2 := model.Price{
		Currency: "RUB",
		Price:    666.0,
		Vat:      99,
	}

	err = service.Update(id, &price1)
	assert.Nil(suite.T(), err, "Unable to update price for game")

	err = service.Update(id, &price2)
	assert.Nil(suite.T(), err, "Unable to update price for game")

	gameFromDb, err := service.GetBase(id)
	assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	assert.NotNil(suite.T(), gameFromDb, "Unable to get game: %v", id)

	assert.Equal(suite.T(), 2, len(gameFromDb.Prices), "Incorrect Prices from DB")
	assert.Equal(suite.T(), price1.BasePriceID, gameFromDb.Prices[0].BasePriceID, "Incorrect Prices from DB")
	assert.Equal(suite.T(), price1.Price, gameFromDb.Prices[0].Price, "Incorrect Prices from DB")
	assert.Equal(suite.T(), price1.Currency, gameFromDb.Prices[0].Currency, "Incorrect Prices from DB")

	assert.Equal(suite.T(), price2.BasePriceID, gameFromDb.Prices[1].BasePriceID, "Incorrect Prices from DB")
	assert.Equal(suite.T(), price2.Price, gameFromDb.Prices[1].Price, "Incorrect Prices from DB")
	assert.Equal(suite.T(), price2.Currency, gameFromDb.Prices[1].Currency, "Incorrect Prices from DB")

	err = service.Delete(id, &price1)
	assert.Nil(suite.T(), err, "Unable to delete price: %v", err)
	gameFromDb, err = service.GetBase(id)
	assert.Equal(suite.T(), 1, len(gameFromDb.Prices), "Incorrect Prices from DB")

	err = service.Delete(id, &price2)
	assert.Nil(suite.T(), err, "Unable to delete price: %v", err)
	gameFromDb, err = service.GetBase(id)
	assert.Equal(suite.T(), 0, len(gameFromDb.Prices), "Incorrect Prices from DB")

}
