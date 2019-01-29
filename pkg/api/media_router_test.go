package api

import (
	"github.com/labstack/echo"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strings"
	"testing"
	"time"
)

type MediaRouterTestSuite struct {
	suite.Suite
	db      *orm.Database
	service *orm.MediaService
	echo 	*echo.Echo
	router  *MediaRouter
}

func Test_MediaRouter(t *testing.T) {
	suite.Run(t, new(MediaRouterTestSuite))
}

var (
	ID = "029ce039-888a-481a-a831-cde7ff4e50b8"
	emptyObject = `{"coverImage":null,"coverVideo":null,"trailers":null,"store":null,"capsule":null}`
	partialObject = `{"coverImage":{"en":"123", "ru":"321"},"coverVideo":{"en":"123", "ru":"321"},"trailers":{"en":"123", "ru":"321"},"store":null,"capsule":null}`
)

func (suite *MediaRouterTestSuite) SetupTest() {
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
		InternalName: "Test_game_1",
		ReleaseDate: time.Now(),
		Genre: pq.StringArray{},
		Tags: pq.StringArray{},
		FeaturesCommon: pq.StringArray{},
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	echo := echo.New()
	service, err := orm.NewMediaService(db)
	router, err := InitMediaRouter(echo.Group("/api/v1"), service)

	echo.Validator = &QilinValidator{validator: validator.New()}

	suite.db = db
	suite.service = service
	suite.router = router
	suite.echo = echo
}

func (suite *MediaRouterTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Media{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *MediaRouterTestSuite) TestGetMediaShouldReturnEmptyObject() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/media")
	c.SetParamNames("id")
	c.SetParamValues(ID)
		
	// Assertions
	if assert.NoError(suite.T(), suite.router.get(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), emptyObject, rec.Body.String())
	}
}

func (suite *MediaRouterTestSuite) TestGetMediaShouldReturnNotFound() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/media")
	c.SetParamNames("id")
	c.SetParamValues("00000000-0000-0000-0000-000000000000")
	
	// Assertions
	
	he := suite.router.get(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}

func (suite *MediaRouterTestSuite) TestPutMediaShouldUpdateGame() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(partialObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/media")
	c.SetParamNames("id")
	c.SetParamValues(ID)
	
	// Assertions
	if assert.NoError(suite.T(), suite.router.get(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), emptyObject, rec.Body.String())
	}
}

func (suite *MediaRouterTestSuite) TestPutMediaShouldreturnNotFound() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games/:id/media")
	c.SetParamNames("id")
	c.SetParamValues("00000000-0000-0000-0000-000000000000")
	
	// Assertions
	
	he := suite.router.get(c).(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}