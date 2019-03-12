package game

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"strings"
	"testing"
)

type QilinValidator struct {
	validator *validator.Validate
}

func (cv *QilinValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

type GamesRouterTestSuite struct {
	suite.Suite
	db     *orm.Database
	echo   *echo.Echo
	router *Router
	token  *jwt.Token
}

func Test_GamesRouter(t *testing.T) {
	suite.Run(t, new(GamesRouterTestSuite))
}

var (
	userId             = `95a97684-9dad-11d1-80b4-00c04fd430c8`
	vendorId           = `6ba97684-9dad-11d1-80b4-00c04fd430c8`
	createGamesPayload = `{"InternalName":"new_game", "vendorId": "` + vendorId + `"}`
)

func (suite *GamesRouterTestSuite) SetupTest() {
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	if err := db.Init(); err != nil {
		fmt.Println(err)
	}

	userUuid, _ := uuid.FromString(userId)
	err = db.DB().Save(&model.User{
		ID:       userUuid,
		Nickname: "admin",
		Login:    "admin@protocol.one",
		Password: "123456",
		Lang:     "en",
		Currency: "usd",
	}).Error
	require.Nil(suite.T(), err, "Unable to make user")
	vendorUuid, _ := uuid.FromString(vendorId)
	err = db.DB().Save(&model.Vendor{
		ID:              vendorUuid,
		Name:            "domino",
		Domain3:         "domino",
		Email:           "domine@ya.ru",
		HowManyProducts: "+10",
		ManagerID:       userUuid,
		Users:           []model.User{{ID: userUuid}},
	}).Error
	require.Nil(suite.T(), err, "Unable to make user")

	echoObj := echo.New()
	echoObj.Validator = &QilinValidator{validator: validator.New()}
	groupApi := echoObj.Group("/api/v1")
	service, err := orm.NewGameService(db)
	userService, err := orm.NewUserService(db, nil)
	router, err := InitRoutes(groupApi, service, userService)
	if err != nil {
		suite.FailNow("Init routes fail", "%v", err)
	}
	suite.db = db
	suite.router = router
	suite.echo = echoObj
	//token, _ := uuid.FromString(userId)
	//suite.token = jwt.NewWithClaims(jwt.GetSigningMethod(config.Jwt.Algorithm), jwt.MapClaims{"id": base64.StdEncoding.EncodeToString(token[:])})
}

func (suite *GamesRouterTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Game{}, model.User{}, model.Vendor{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *GamesRouterTestSuite) TestShouldCreateGame() {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createGamesPayload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/games")
	c.Set(context.TokenKey, suite.token)

	err := suite.router.Create(c)
	require.Nil(suite.T(), err, "Error while create game")

	game := model.Game{}
	err = suite.db.DB().First(&game).Error
	require.Equal(suite.T(), game.InternalName, "new_game", "Incorrect game creates")
}
