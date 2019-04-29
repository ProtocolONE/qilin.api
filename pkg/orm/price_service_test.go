package orm

import (
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
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
	gameID    = "029ce039-3333-481a-a831-cde7ff4e50b9"
	packageID = "029ce039-888a-481a-a831-cde7ff4e50b9"
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

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

	gameId, _ := uuid.FromString(gameID)
	err = db.DB().Save(&model.Game{
		ID:             gameId,
		InternalName:   "Test_game_2",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	pkgId, _ := uuid.FromString(packageID)
	err = db.DB().Save(&model.Package{
		Model: model.Model{ID: pkgId},
		Name:  utils.LocalizedString{EN: "Test_package_2"},
	}).Error
	require.Nil(suite.T(), err, "Unable to make package")

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
	service := NewPriceService(suite.db)
	updatedAt, _ := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")

	id, _ := uuid.FromString(packageID)
	pkg := model.BasePrice{
		ID: uuid.NewV4(),
		PackagePrices: model.PackagePrices{
			Common: model.JSONB{
				"currency":        "USD",
				"notifyRateJumps": false,
			},
			PreOrder: model.JSONB{
				"date":    "2019-01-22T07:53:16Z",
				"enabled": false,
			},
			Prices: []model.Price{
				{BasePriceID: id, Price: 100.0, Vat: 32, Currency: "EUR"},
				{BasePriceID: id, Price: 93.23, Vat: 10, Currency: "RUB"},
			},
		},
		UpdatedAt: &updatedAt,
	}

	err := service.UpdateBase(id, &pkg)
	assert.Nil(suite.T(), err, "Unable to update media for package")

	pkgFromDb, err := service.GetBase(id)
	assert.Nil(suite.T(), err, "Unable to get package: %v", err)
	assert.NotNil(suite.T(), pkgFromDb, "Unable to get package: %v", id)
	assert.Equal(suite.T(), pkg.ID, pkgFromDb.ID, "Incorrect Game ID from DB")
	assert.Equal(suite.T(), pkg.Common["currency"], pkgFromDb.Common["currency"], "Incorrect Common from DB")
	assert.Equal(suite.T(), pkg.Common["notifyRateJumps"], pkgFromDb.Common["notifyRateJumps"], "Incorrect Common from DB")
	assert.Equal(suite.T(), pkg.PreOrder, pkgFromDb.PreOrder, "Incorrect PreOrder from DB")
}

func (suite *PriceServiceTestSuite) TestPriceServiceShouldReturnError() {
	service := NewPriceService(suite.db)
	price1 := model.Price{
		Currency: "USD",
		Price:    123.32,
		Vat:      10,
	}

	res, err := service.GetBase(uuid.NewV4())
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), res)
	if err != nil {
		he := err.(*ServiceError)
		assert.Equal(suite.T(), http.StatusNotFound, he.Code)
	}

	err = service.Update(uuid.NewV4(), &price1)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*ServiceError)
		assert.Equal(suite.T(), http.StatusNotFound, he.Code)
	}

	err = service.Delete(uuid.NewV4(), &price1)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*ServiceError)
		assert.Equal(suite.T(), http.StatusNotFound, he.Code)
	}
}

func (suite *PriceServiceTestSuite) TestChangePrices() {
	service := NewPriceService(suite.db)

	id, _ := uuid.FromString(packageID)

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

	err := service.Update(id, &price1)
	assert.Nil(suite.T(), err, "Unable to update price for package")

	err = service.Update(id, &price2)
	assert.Nil(suite.T(), err, "Unable to update price for package")

	pkgFromDb, err := service.GetBase(id)
	assert.Nil(suite.T(), err, "Unable to get package: %v", err)
	assert.NotNil(suite.T(), pkgFromDb, "Unable to get package: %v", id)

	assert.Equal(suite.T(), 2, len(pkgFromDb.Prices), "Incorrect Prices from DB")
	assert.Equal(suite.T(), price1.BasePriceID, pkgFromDb.Prices[0].BasePriceID, "Incorrect Prices from DB")
	assert.Equal(suite.T(), price1.Price, pkgFromDb.Prices[0].Price, "Incorrect Prices from DB")
	assert.Equal(suite.T(), price1.Currency, pkgFromDb.Prices[0].Currency, "Incorrect Prices from DB")

	assert.Equal(suite.T(), price2.BasePriceID, pkgFromDb.Prices[1].BasePriceID, "Incorrect Prices from DB")
	assert.Equal(suite.T(), price2.Price, pkgFromDb.Prices[1].Price, "Incorrect Prices from DB")
	assert.Equal(suite.T(), price2.Currency, pkgFromDb.Prices[1].Currency, "Incorrect Prices from DB")

	err = service.Delete(id, &price1)
	assert.Nil(suite.T(), err, "Unable to delete price: %v", err)
	pkgFromDb, err = service.GetBase(id)
	assert.Equal(suite.T(), 1, len(pkgFromDb.Prices), "Incorrect Prices from DB")

	err = service.Delete(id, &price2)
	assert.Nil(suite.T(), err, "Unable to delete price: %v", err)
	pkgFromDb, err = service.GetBase(id)
	assert.Equal(suite.T(), 0, len(pkgFromDb.Prices), "Incorrect Prices from DB")

}
