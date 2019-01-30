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

type DiscountRouterTestSuite struct {
	suite.Suite
	db     *orm.Database
	echo   *echo.Echo
	router *DiscountsRouter
}

func Test_DiscountsRouter(t *testing.T) {
	suite.Run(t, new(DiscountRouterTestSuite))
}

var (
	testDiscountObject         = `{"title":{"en":"WINTER SSSSALE","ru":"сейл"},"desctiption":{"en":"desct","ru":"desct"},"date":{"start":"2019-07-21T17:32:28Z","end":"2019-08-21T17:32:28Z"},"rate":30}`
	testDiscountsObject        = `{"common":{"currency":"USD","NotifyRateJumps":true},"preOrder":{"date":"2019-01-22T07:53:16Z","enabled":false}}`
	testBadDiscountObject      = `{"desctiption":{"en":"desct","ru":"desct"},"date":{"start":"2019-07-21T17:32:28Z","end":"2019-08-21T17:32:28Z"}}`
	testBadEmptyDiscountObject = `{}`
	emptyDiscounts             = `[]`
)

func (suite *DiscountRouterTestSuite) SetupTest() {
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
	service, err := orm.NewDiscountService(db)
	router, err := InitDiscountsRouter(echo.Group("/api/v1"), service)

	echo.Validator = &QilinValidator{validator: validator.New()}

	suite.db = db
	suite.router = router
	suite.echo = echo
}

func (suite *DiscountRouterTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Discount{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.DB().DropTable(model.Game{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *DiscountRouterTestSuite) TestGetDiscountsShouldReturnEmptyArray() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyDiscounts))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts")
	c.SetParamNames("id")
	c.SetParamValues(ID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.get(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), emptyDiscounts, rec.Body.String())
	}
}

func (suite *DiscountRouterTestSuite) TestPostDiscountShouldReturnId() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts")
	c.SetParamNames("id")
	c.SetParamValues(ID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.post(c)) {
		assert.Equal(suite.T(), http.StatusCreated, rec.Code)
		assert.NotEmpty(suite.T(), rec.Body.String())
	}
}

func (suite *DiscountRouterTestSuite) TestPostDiscountWithIncorrectIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts")
	c.SetParamNames("id")
	c.SetParamValues("BAD-ID")

	// Assertions
	he := suite.router.post(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *DiscountRouterTestSuite) TestPostDiscountWithIncorrectObjectShouldReturnError() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testBadDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts")
	c.SetParamNames("id")
	c.SetParamValues(ID)

	// Assertions
	he := suite.router.post(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *DiscountRouterTestSuite) TestPostDiscountWithEmptyObjectShouldReturnError() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testBadEmptyDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts")
	c.SetParamNames("id")
	c.SetParamValues(ID)

	// Assertions
	he := suite.router.post(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *DiscountRouterTestSuite) TestPutDiscountWithIncorrectIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id")
	c.SetParamValues("BAD-ID", uuid.NewV4().String())

	// Assertions
	he := suite.router.put(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *DiscountRouterTestSuite) TestPutDiscountWithCorrectObjectShouldReturnOk() {
	id, _ := uuid.FromString(ID)
	discount := model.Discount{
		Title: model.JSONB{
			"en": "asd",
		},
		GameID: id,
		Rate:   10,
	}
	discount.ID = uuid.NewV4()

	err := suite.db.DB().Create(&discount).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id", "discountId")
	c.SetParamValues(ID, discount.ID.String())

	// Assertions
	if assert.NoError(suite.T(), suite.router.put(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
	}
}

func (suite *DiscountRouterTestSuite) TestPutDiscountWithIncorrectObjectShouldReturnError() {
	id, _ := uuid.FromString(ID)
	discount := model.Discount{
		Title: model.JSONB{
			"en": "asd",
		},
		GameID: id,
		Rate:   10,
	}
	discount.ID = uuid.NewV4()

	err := suite.db.DB().Create(&discount).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testBadDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id", "discountId")
	c.SetParamValues(ID, discount.ID.String())

	// Assertions
	he := suite.router.put(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *DiscountRouterTestSuite) TestPutDiscountWithEmptyObjectShouldReturnError() {
	id, _ := uuid.FromString(ID)
	discount := model.Discount{
		Title: model.JSONB{
			"en": "asd",
		},
		GameID: id,
		Rate:   10,
	}
	discount.ID = uuid.NewV4()

	err := suite.db.DB().Create(&discount).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testBadEmptyDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id", "discountId")
	c.SetParamValues(ID, discount.ID.String())

	// Assertions
	he := suite.router.put(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code)
}

func (suite *DiscountRouterTestSuite) TestPutDiscountWithUnknownDiscountIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id", "discountId")
	c.SetParamValues(ID, "00000000-0000-0000-0000-000000000000")

	// Assertions
	he := suite.router.put(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}

func (suite *DiscountRouterTestSuite) TestGetDiscountsWithInvalidIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyDiscounts))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts")
	c.SetParamNames("id")
	c.SetParamValues("BAD-ID")

	// Assertions
	he := suite.router.get(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *DiscountRouterTestSuite) TestGetDiscountsWithUnknownIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyDiscounts))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts")
	c.SetParamNames("id")
	c.SetParamValues("00000000-0000-0000-0000-000000000000")

	// Assertions
	he := suite.router.get(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}

func (suite *DiscountRouterTestSuite) TestDeleteDiscountWithIncorrectIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id", "discountId")
	c.SetParamValues("BAD-ID", uuid.NewV4().String())

	// Assertions
	he := suite.router.delete(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *DiscountRouterTestSuite) TestDeleteDiscountWithUnknownDiscountIDShouldReturnError() {
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id", "discountId")
	c.SetParamValues(ID, uuid.NewV4().String())

	// Assertions
	he := suite.router.delete(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}

func (suite *DiscountRouterTestSuite) TestDeleteDiscountWithIncorrectDiscountIDShouldReturnError() {
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id", "discountId")
	c.SetParamValues(ID, "BAD-ID")

	// Assertions
	he := suite.router.delete(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *DiscountRouterTestSuite) TestDeleteDiscountWithCorrectIdShouldReturnOk() {
	id, _ := uuid.FromString(ID)
	discount := model.Discount{
		Title: model.JSONB{
			"en": "asd",
		},
		GameID: id,
		Rate:   10,
	}
	discount.ID = uuid.NewV4()

	err := suite.db.DB().Create(&discount).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(testDiscountObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/discounts/:discountId")
	c.SetParamNames("id", "discountId")
	c.SetParamValues(ID, discount.ID.String())

	// Assertions
	if assert.NoError(suite.T(), suite.router.delete(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
	}
}