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

	"github.com/labstack/echo"
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
	testObject     = `{"common":{"currency":"USD","NotifyRateJumps":true},"preOrder":{"date":"2019-01-22T07:53:16Z","enabled":false}}`
	testBadObject  = `{"common":{"NotifyRateJumps":true},"preOrder":{"enabled":false}}`
	testBadObjectWithUnwknowsCurrency  = `{"common":{"currency":"XXX","NotifyRateJumps":true},"preOrder":{"date":"2019-01-22T07:53:16Z","enabled":false}}`
	emptyBasePrice = `{"common":{"currency":"","notifyRateJumps":false},"preOrder":{"date":"","enabled":false},"prices":null}`
	testPrice      = `{"price":100,"currency":"USD","vat":10}`
	testBadPrice   = `{"vat":10}`

	testPriceWithUnknownCurrency = `{"price":100,"currency":"XXX","vat":10}`
	testPriceWithBadPrice = `{"price":"1a","currency":"USD","vat":10}`
	testPriceWithBadVat = `{"price":100,"currency":"USD","vat":"qwe"}`
	testPriceWithPriceLowerZero = `{"price":-100,"currency":"USD","vat":10}`
	testPriceWithVatLowerZero = `{"price":100,"currency":"USD","vat":-10}`
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

	_ = db.DropAllTables()
	db.Init()

	id, _ := uuid.FromString(TestID)
	err = db.DB().Save(&model.Game{
		ID:             id,
		InternalName:   "Test_game_2",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.StringArray{},
		Tags:           pq.StringArray{},
		FeaturesCommon: pq.StringArray{},
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

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
	c.SetPath("/api/v1/games/:id/prices")
	c.SetParamNames("id")
	c.SetParamValues(TestID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getBase(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), emptyBasePrice, rec.Body.String())
	}
}

func (suite *PriceRouterTestSuite) TestPutBasePriceShouldReturnOk() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/prices")
	c.SetParamNames("id")
	c.SetParamValues(TestID)

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
	c.SetPath("/api/v1/games/:id/prices/:currency")
	c.SetParamNames("id", "currency")
	c.SetParamValues(TestID, "USD")

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
	c.SetPath("/api/v1/games/:id/prices/:currency")
	c.SetParamNames("id", "currency")
	c.SetParamValues(TestID, "USD")

	he := suite.router.updatePrice(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutWithIncorrectCurrencyPriceShouldReturnBadRequest() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testPrice))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/prices/:currency")
	c.SetParamNames("id", "currency")
	c.SetParamValues(TestID, "EUR")

	he := suite.router.putBase(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutBadModelShouldReturn422() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testBadObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/prices")
	c.SetParamNames("id")
	c.SetParamValues(TestID)

	// Assertions
	he := suite.router.putBase(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutWithUnknownIdModelShouldReturnNotFound() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/prices")
	c.SetParamNames("id")
	c.SetParamValues("00000000-0000-0000-0000-000000000000")

	// Assertions
	he := suite.router.putBase(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutWithBadIdModelShouldReturnBadRequest() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testBadObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/prices")
	c.SetParamNames("id")
	c.SetParamValues("0000")

	// Assertions
	he := suite.router.updatePrice(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *PriceRouterTestSuite) TestDeleteUnknownCurrencyShouldReturnError () {
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(testPriceWithUnknownCurrency))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/prices/:currency")
	c.SetParamNames("id", "currency")
	c.SetParamValues(TestID, "XXX")

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
	c.SetPath("/api/v1/games/:id/prices/")
	c.SetParamNames("id")
	c.SetParamValues(TestID)

	// Assertions
	err := suite.router.putBase(c)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*orm.ServiceError)
		assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code, he.Message)
	}
}

func (suite *PriceRouterTestSuite) TestPutBadObjectsShouldReturnError () {
	tests := []struct {
		name string
		body string
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
		c.SetPath("/api/v1/games/:id/prices/:currency")
		c.SetParamNames("id", "currency")
		c.SetParamValues(TestID, "USD")

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
	c.SetPath("/api/v1/games/:id/prices/:currency")
	c.SetParamNames("id", "currency")
	c.SetParamValues(TestID, "XXX")

	// Assertions
	err := suite.router.updatePrice(c)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*orm.ServiceError)
		assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code, he.Message)
	}
}
