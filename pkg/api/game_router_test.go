package api

import (
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/ProtocolONE/rbac"
	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/mock"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"strings"
	"testing"
)

type GamesRouterTestSuite struct {
	suite.Suite
	db       *orm.Database
	echo     *echo.Echo
	router   *GameRouter
	enforcer *rbac.Enforcer
}

func Test_GamesRouter(t *testing.T) {
	suite.Run(t, new(GamesRouterTestSuite))
}

var (
	userId             = uuid.NewV4().String()
	vendorId           = uuid.NewV4().String()
	createGamesPayload = `{"InternalName":"new_game"}`
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
		suite.T().Log(err)
	}

	err = db.DB().Save(&model.User{
		ID:       userId,
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
		ManagerID:       userId,
		Users:           []model.User{{ID: userId}},
	}).Error
	require.Nil(suite.T(), err, "Unable to make user")

	echoObj := echo.New()
	echoObj.Validator = &QilinValidator{validator: validator.New()}

	ownerProvider := orm.NewOwnerProvider(db)
	enforcer := rbac.NewEnforcer()
	membership := orm.NewMembershipService(db, ownerProvider, enforcer, mock.NewMailer(), "")
	err = membership.Init()
	if err != nil {
		suite.FailNow("Membership fail", "%v", err)
	}

	service, err := orm.NewGameService(db)
	echoObj.Use(rbac_echo.NewAppContextMiddleware(ownerProvider, enforcer))

	packageService, err := orm.NewPackageService(db)
	if err != nil {
		suite.FailNow("Package fail", "%v", err)
	}

	groupApi := echoObj.Group("/api/v1")
	userService, err := orm.NewUserService(db, nil)
	router, err := InitRoutes(groupApi, service, userService, packageService)
	if err != nil {
		suite.FailNow("Init routes fail", "%v", err)
	}
	suite.db = db
	suite.router = router
	suite.echo = echoObj
	suite.enforcer = enforcer
}

func (suite *GamesRouterTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *GamesRouterTestSuite) TestShouldCreateGame() {
	err := suite.db.DB().Save(&model.User{
		ID:       userId,
		Nickname: "admin",
		Login:    "admin@protocol.one",
		Password: "123456",
		Lang:     "en",
		Currency: "usd",
	}).Error

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createGamesPayload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/games")
	c.SetParamNames("vendorId")
	c.SetParamValues(vendorId)
	c.Set(context.TokenKey, &jwtverifier.UserInfo{UserID: userId})

	err = suite.router.Create(c)
	require.Nil(suite.T(), err, "Error while create game")

	game := model.Game{}
	err = suite.db.DB().First(&game).Error
	require.Equal(suite.T(), game.InternalName, "new_game", "Incorrect game creates")
}
