package api

import (
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	validator "gopkg.in/go-playground/validator.v9"
)

type RatingRouterTestSuite struct {
	suite.Suite
	db     *orm.Database
	echo   *echo.Echo
	router *RatingsRouter
}

var (
	emptyRatings = `{"PEGI":{"displayOnlineNotice":false,"showAgeRestrict":false,"ageRestrict":0,"descriptors":null,"rating":0},"ESRB":{"displayOnlineNotice":false,"showAgeRestrict":false,"ageRestrict":0,"descriptors":null,"rating":""},"BBFC":{"displayOnlineNotice":false,"showAgeRestrict":false,"ageRestrict":0,"descriptors":null,"rating":""},"USK":{"displayOnlineNotice":false,"showAgeRestrict":false,"ageRestrict":0,"descriptors":null,"rating":""},"CERO":{"displayOnlineNotice":false,"showAgeRestrict":false,"ageRestrict":0,"descriptors":null,"rating":""}}`
	fullRatings  = `{"PEGI":{"displayOnlineNotice":true,"showAgeRestrict":false,"ageRestrict":3,"descriptors":null,"rating":0},"ESRB":{"displayOnlineNotice":false,"showAgeRestrict":false,"ageRestrict":15,"descriptors":null,"rating":""},"BBFC":{"displayOnlineNotice":true,"showAgeRestrict":true,"ageRestrict":10,"descriptors":null,"rating":""},"USK":{"displayOnlineNotice":false,"showAgeRestrict":true,"ageRestrict":5,"descriptors":null,"rating":""},"CERO":{"displayOnlineNotice":false,"showAgeRestrict":false,"ageRestrict":21,"descriptors":null,"rating":""}}`
)

func Test_RatingRouter(t *testing.T) {
	suite.Run(t, new(RatingRouterTestSuite))
}

func (suite *RatingRouterTestSuite) SetupTest() {
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
		ID:             id,
		InternalName:   "Test_game_2",
		ReleaseDate:    time.Now(),
		Genre:          pq.StringArray{},
		Tags:           pq.StringArray{},
		FeaturesCommon: pq.StringArray{},
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	e := echo.New()
	service, err := orm.NewRatingService(db)
	router, err := InitRatingsRouter(e.Group("/api/v1"), service)

	e.Validator = &QilinValidator{validator: validator.New()}

	suite.db = db
	suite.router = router
	suite.echo = e
}

func (suite *RatingRouterTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *RatingRouterTestSuite) TestGetRatingsShouldReturnEmptyObject() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyRatings))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/ratings")
	c.SetParamNames("id")
	c.SetParamValues(ID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getRatings(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), emptyRatings, rec.Body.String())
	}
}


func (suite *RatingRouterTestSuite) TestPostRatingsShouldReturnOk() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(fullRatings))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/ratings")
	c.SetParamNames("id")
	c.SetParamValues(ID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.postRatings(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
	}
}


func (suite *RatingRouterTestSuite) TestPostBadObjectShouldReturnError() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("bad object"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/ratings")
	c.SetParamNames("id")
	c.SetParamValues(ID)

	// Assertions
	he := suite.router.postRatings(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *RatingRouterTestSuite) TestGetRatingsShouldWithNilIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyRatings))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/ratings")
	c.SetParamNames("id")
	c.SetParamValues(uuid.Nil.String())

	// Assertions
	he := suite.router.getRatings(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}

func (suite *RatingRouterTestSuite) TestPostRatingsShouldWithNilIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emptyRatings))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/ratings")
	c.SetParamNames("id")
	c.SetParamValues(uuid.Nil.String())

	// Assertions
	he := suite.router.getRatings(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}


func (suite *RatingRouterTestSuite) TestGetRatingsShouldWithBadIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyRatings))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/ratings")
	c.SetParamNames("id")
	c.SetParamValues("XXX")

	// Assertions
	he := suite.router.getRatings(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *RatingRouterTestSuite) TestPostRatingsShouldWithBadIdShouldReturnError() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emptyRatings))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/ratings")
	c.SetParamNames("id")
	c.SetParamValues("XXX")

	// Assertions
	he := suite.router.postRatings(c).(*echo.HTTPError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *RatingRouterTestSuite) TestGetRatingsShouldReturnRightObject() {
	id, _ := uuid.FromString(ID)
	testModel := &model.GameRating{
		BBFC: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10.0,
		},
		CERO: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     false,
			"ageRestrict":         21.0,
		},
		ESRB: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     false,
			"ageRestrict":         15.0,
		},
		PEGI: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     false,
			"ageRestrict":         3.0,
		},
		USK: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     true,
			"ageRestrict":         5.0,
		},
		GameID: id,
	}
	assert.NoError(suite.T(), suite.db.DB().Create(&testModel).Error)

	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(fullRatings))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/ratings")
	c.SetParamNames("id")
	c.SetParamValues(ID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getRatings(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), fullRatings, rec.Body.String())
	}
}
