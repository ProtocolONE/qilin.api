package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"testing"
	"strings"
	"net/http"
	"net/http/httptest"
	"gopkg.in/go-playground/validator.v9"
	"github.com/stretchr/testify/suite"
)


type PriceRouterTestSuite struct {
	suite.Suite
	db      *orm.Database
	//service *orm.MediaService
	echo 	*echo.Echo
	router  *PriceRouter
}

func Test_PriceRouter(t *testing.T) {
	suite.Run(t, new(PriceRouterTestSuite))
}

var (
	ID = "029ce039-888a-481a-a831-cde7ff4e50b9"
	testObject = `{"prices":{"default":{"currency":"USD","price":0},"preOrder":{"date":"2019-01-22T07:53:16Z","enabled":false},"prices":[{"currency":"USD","vat":0,"price":0}]}}`
)

func (suite *PriceRouterTestSuite) SetupTest() {
	dbConfig := conf.Database{
		Host:     "localhost",
		Port:     "5432",
		Database: "test_qilin",
		User:     "postgres",
		Password: "postgres",
	}

	db, err := orm.NewDatabase(&dbConfig)
	if err != nil {
		suite.Fail("Unable to connect to database: %s", err)
	}

	db.Init()

	id, _ := uuid.FromString(ID)
	db.DB().Save(&model.Game{ID: id})

	echo := echo.New()
	service, err := orm.NewMediaService(db)
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
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}