package api

import (
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	validator "gopkg.in/go-playground/validator.v9"
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
	emptyBasePrice = `{"common":{"currency":"","notifyRateJumps":false},"preOrder":{"date":"","enabled":false},"prices":null}`
	testPrice      = `{"price":100,"currency":"USD","vat":10}`
	testBadPrice   = `{"vat":10}`
)

func (suite *PriceRouterTestSuite) SetupTest() {
	config, err := conf.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	db.Init()

	id, _ := uuid.FromString(ID)
	err = db.DB().Save(&model.Game{
		ID: id,
		InternalName: "Test_game_2",
		ReleaseDate: time.Now(),
		Genre: pq.StringArray{},
		Tags: pq.StringArray{},
		FeaturesCommon: pq.StringArray{},
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	echo := echo.New()
	service, err := orm.NewPriceService(db)
	router, err := InitPriceRouter(echo.Group("/api/v1"), service)

	echo.Validator = &QilinValidator{validator: validator.New()}

	suite.db = db
	suite.router = router
	suite.echo = echo
}

func (suite *PriceRouterTestSuite) TearDownTest() {
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

func (suite *PriceRouterTestSuite) TestGetBasePriceShouldReturnEmptyObject() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyBasePrice))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/prices")
	c.SetParamNames("id")
	c.SetParamValues(ID)

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
	c.SetParamValues(ID)

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
	c.SetParamValues(ID, "USD")

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
	c.SetParamValues(ID, "USD")

	he := suite.router.updatePrice(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *PriceRouterTestSuite) TestPutWithIncorrectCurrencyPriceShouldReturnBadRequest() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testPrice))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/prices/:currency")
	c.SetParamNames("id", "currency")
	c.SetParamValues(ID, "EUR")

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
	c.SetParamValues(ID)

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
	he := suite.router.updatePrice(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}
