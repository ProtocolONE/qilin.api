package api

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
)

type PriceRouterTestSuite struct {
	suite.Suite
	db     *orm.Database
	echo   *echo.Echo
	router *PriceRouter
}

func Test_PriceRouter(t *testing.T) {
	suite.Run(t, new(PriceRouterTestSuite))
}

var (
	packageID                         = "33333333-888a-481a-a831-cde7ff4e50b8"

	testObject                        = `{"common":{"currency":"USD","NotifyRateJumps":true},"preOrder":{"date":"2019-01-22T07:53:16Z","enabled":false}}`
	testBadObject                     = `{"common":{"NotifyRateJumps":true},"preOrder":{"enabled":false}}`
	testBadObjectWithUnwknowsCurrency = `{"common":{"currency":"XXX","NotifyRateJumps":true},"preOrder":{"date":"2019-01-22T07:53:16Z","enabled":false}}`
	emptyBasePrice                    = `{"common":{"currency":"","notifyRateJumps":false},"preOrder":{"date":"","enabled":false},"prices":null}`
	testPrice                         = `{"price":100,"currency":"USD","vat":10}`
	testBadPrice                      = `{"vat":10}`

	testPriceWithUnknownCurrency = `{"price":100,"currency":"XXX","vat":10}`
	testPriceWithBadPrice        = `{"price":"1a","currency":"USD","vat":10}`
	testPriceWithBadVat          = `{"price":100,"currency":"USD","vat":"qwe"}`
	testPriceWithPriceLowerZero  = `{"price":-100,"currency":"USD","vat":10}`
	testPriceWithVatLowerZero    = `{"price":100,"currency":"USD","vat":-10}`
)

func (suite *PriceRouterTestSuite) SetupTest() {
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

	id, _ := uuid.FromString(TestID)
	err = db.DB().Save(&model.Game{
		ID:             id,
		InternalName:   "Test_game_2",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
		Product:        model.ProductEntry{EntryID: id},
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	pkgId, _ := uuid.FromString(packageID)
	err = db.DB().Save(&model.Package{
		Model:  model.Model{ID: pkgId},
		Name:   "Test_package",
		AllowedCountries: pq.StringArray{},
		PackagePrices: model.PackagePrices{
			Common: model.JSONB{"currency":"","NotifyRateJumps":false},
			PreOrder: model.JSONB{"date":"","enabled":false},
			Prices: []model.Price{},
		},
	}).Error
	require.Nil(suite.T(), err, "Unable to make package")
	err = db.DB().Create(&model.PackageProduct{
		PackageID: pkgId,
		ProductID: id,
	}).Error
	require.Nil(suite.T(), err, "Unable to make package product")

	echoObj := echo.New()
	service, err := orm.NewPriceService(db)
	router, err := InitPriceRouter(echoObj.Group("/api/v1"), service)

	echoObj.Validator = &QilinValidator{validator: validator.New()}

	suite.db = db
	suite.router = router
	suite.echo = echoObj
}

func (suite *PriceRouterTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *PriceRouterTestSuite) TestGetBasePriceShouldReturnEmptyObject() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyBasePrice))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices")
	c.SetParamNames("packageId")
	c.SetParamValues(packageID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getBase(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.JSONEq(suite.T(), emptyBasePrice, rec.Body.String())
	}
}

func (suite *PriceRouterTestSuite) TestPutBasePriceShouldReturnOk() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices")
	c.SetParamNames("packageId")
	c.SetParamValues(packageID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.putBase(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), "", rec.Body.String())
	}
}

func (suite *PriceRouterTestSuite) TestPutPriceShouldReturnOk() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testPrice))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices/:currency")
	c.SetParamNames("packageId", "currency")
	c.SetParamValues(packageID, "USD")

	// Assertions
	if assert.NoError(suite.T(), suite.router.updatePrice(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), "", rec.Body.String())
	}
}

func (suite *PriceRouterTestSuite) TestPutPriceShouldReturnBadRequest() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testBadPrice))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices/:currency")
	c.SetParamNames("packageId", "currency")
	c.SetParamValues(packageID, "USD")

	he := suite.router.updatePrice(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutWithIncorrectCurrencyPriceShouldReturnBadRequest() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testPrice))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices/:currency")
	c.SetParamNames("packageId", "currency")
	c.SetParamValues(packageID, "EUR")

	he := suite.router.putBase(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutBadModelShouldReturn422() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testBadObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices")
	c.SetParamNames("packageId")
	c.SetParamValues(packageID)

	// Assertions
	err := suite.router.putBase(c)
	require.NotNil(suite.T(), err)
	he := err.(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutWithUnknownIdModelShouldReturnNotFound() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices")
	c.SetParamNames("packageId") /// with 0000 tests not pass(
	c.SetParamValues("00000000-0000-8000-0000-000000000000")

	// Assertions
	err := suite.router.putBase(c)
	require.NotNil(suite.T(), err)
	he := err.(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutWithBadIdModelShouldReturnBadRequest() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testBadObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices")
	c.SetParamNames("packageId")
	c.SetParamValues("0000")

	// Assertions
	he := suite.router.updatePrice(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *PriceRouterTestSuite) TestDeleteUnknownCurrencyShouldReturnError() {
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(testPriceWithUnknownCurrency))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices/:currency")
	c.SetParamNames("packageId", "currency")
	c.SetParamValues(packageID, "XXX")

	// Assertions
	err := suite.router.deletePrice(c)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*orm.ServiceError)
		assert.Equal(suite.T(), http.StatusBadRequest, he.Code, he.Message)
	}
}

func (suite *PriceRouterTestSuite) TestPutBasePriceWihhUnknownCurrencyShouldReturnError() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testBadObjectWithUnwknowsCurrency))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices/")
	c.SetParamNames("packageId")
	c.SetParamValues(packageID)

	// Assertions
	err := suite.router.putBase(c)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*orm.ServiceError)
		assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code, he.Message)
	}
}

func (suite *PriceRouterTestSuite) TestPutBadObjectsShouldReturnError() {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{name: "testPriceWithBadPrice", body: testPriceWithBadPrice, status: http.StatusBadRequest},
		{name: "testPriceWithBadVat", body: testPriceWithBadVat, status: http.StatusBadRequest},
		{name: "testPriceWithPriceLowerZero", body: testPriceWithPriceLowerZero, status: http.StatusUnprocessableEntity},
		{name: "testPriceWithVatLowerZero", body: testPriceWithVatLowerZero, status: http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetPath("/api/v1/packages/:packageId/prices/:currency")
		c.SetParamNames("packageId", "currency")
		c.SetParamValues(packageID, "USD")

		// Assertions
		err := suite.router.updatePrice(c)
		assert.NotNil(suite.T(), err, tt.name)
		if err != nil {
			he := err.(*orm.ServiceError)
			assert.Equal(suite.T(), tt.status, he.Code, fmt.Sprintf("Failed %s, message: %s", tt.name, he.Message))
		}
	}
}

func (suite *PriceRouterTestSuite) TestPutUnknownCurrencyShouldReturnError() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testPriceWithUnknownCurrency))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/prices/:currency")
	c.SetParamNames("packageId", "currency")
	c.SetParamValues(packageID, "XXX")

	// Assertions
	err := suite.router.updatePrice(c)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*orm.ServiceError)
		assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code, he.Message)
	}
}
