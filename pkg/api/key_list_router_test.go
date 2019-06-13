package api

import (
	"bytes"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	qilintest "qilin-api/pkg/test"
	"strings"
	"testing"
)

type KeyListRouterTestSuite struct {
	suite.Suite
	db              *orm.Database
	echo            *echo.Echo
	router          *KeyListRouter
	rightKeyPackage uuid.UUID
	wrongKeyPackage uuid.UUID
	rightKeyStream  uuid.UUID
	keyPackage      uuid.UUID
}

func Test_KeyListRouter(t *testing.T) {
	suite.Run(t, new(KeyListRouterTestSuite))
}

func (suite *KeyListRouterTestSuite) SetupTest() {
	shouldBe := require.New(suite.T())
	config, err := qilintest.LoadTestConfig()
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

	suite.rightKeyPackage = uuid.NewV4()
	suite.wrongKeyPackage = uuid.NewV4()
	suite.keyPackage = uuid.NewV4()

	keyListStreeam := model.KeyStream{
		Type: model.ListKeyStream,
	}
	keyListStreeam.ID = uuid.NewV4()
	suite.rightKeyStream = keyListStreeam.ID
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&keyListStreeam).Error)

	gamePackage := model.Package{Name: utils.LocalizedString{EN: "test package"}}
	gamePackage.ID = suite.keyPackage

	shouldBe.Nil(db.DB().Model(model.Package{}).Create(&gamePackage).Error)

	keyPackage := model.KeyPackage{
		Name:          "Test name",
		PackageID:     suite.keyPackage,
		KeyStreamID:   suite.rightKeyStream,
		KeyStreamType: model.ListKeyStream,
	}
	keyPackage.ID = suite.rightKeyPackage
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&keyPackage).Error)

	platformListStreeam := model.KeyStream{
		Type: model.PlatformKeysStream,
	}
	platformListStreeam.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&platformListStreeam).Error)
	suite.wrongKeyPackage = platformListStreeam.ID

	platformKeyPackage := model.KeyPackage{
		Name:          "Another name",
		PackageID:     uuid.NewV4(),
		KeyStreamType: model.PlatformKeysStream,
	}
	platformKeyPackage.ID = suite.wrongKeyPackage
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&platformKeyPackage).Error)

	e := echo.New()
	e.Validator = &QilinValidator{validator: validator.New()}

	service := orm.NewKeyListService(db)
	suite.router, err = InitKeyListRouter(e.Group("/api/v1"), service)
	suite.db = db
	suite.echo = e
	shouldBe.Nil(err)
}

func (suite *KeyListRouterTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *KeyListRouterTestSuite) TestAddKeys() {
	shouldBe := require.New(suite.T())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"keys":["QWERTY","TESTSOMECODE"]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/keypackages/:keyPackageId")
	c.SetParamNames("packageId", "keyPackageId")
	c.SetParamValues(suite.keyPackage.String(), suite.rightKeyPackage.String())

	err := suite.router.AddKeys(c)
	shouldBe.Nil(err)
	shouldBe.Equal(200, rec.Code)
}

func (suite *KeyListRouterTestSuite) TestAddKeysBadRequest() {
	shouldBe := require.New(suite.T())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"":somethingwronghere"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/keypackages/:keyPackageId")
	c.SetParamNames("packageId", "keyPackageId")
	c.SetParamValues(suite.keyPackage.String(), suite.rightKeyPackage.String())

	err := suite.router.AddKeys(c)
	shouldBe.NotNil(err)
	shouldBe.Equal(400, err.(*orm.ServiceError).Code)
}

func (suite *KeyListRouterTestSuite) TestAddKeysNotFound() {
	shouldBe := require.New(suite.T())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"keys":["QWERTY","TESTSOMECODE"]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/keypackages/:keyPackageId")
	c.SetParamNames("packageId", "keyPackageId")
	c.SetParamValues(uuid.NewV4().String(), uuid.NewV4().String())

	err := suite.router.AddKeys(c)
	shouldBe.NotNil(err)
	shouldBe.Equal(404, err.(*orm.ServiceError).Code)
}

func (suite *KeyListRouterTestSuite) TestAddFile() {
	shouldBe := require.New(suite.T())
	values := map[string]io.Reader {
		"keys": strings.NewReader("QWERTY\nTEST\nOPIUY"),
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		fw, err := w.CreateFormFile(key, key)
		shouldBe.Nil(err)
		if _, err := io.Copy(fw, r); err != nil {
			shouldBe.FailNow("Can't copy fw to r")
		}
	}
	shouldBe.Nil(w.Close())

	req := httptest.NewRequest(http.MethodPost, "/", &b)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/packages/:packageId/keypackages/:keyPackageId")
	c.SetParamNames("packageId", "keyPackageId")
	c.SetParamValues(suite.keyPackage.String(), suite.rightKeyPackage.String())

	err := suite.router.AddFileKeys(c)
	shouldBe.Nil(err)
	shouldBe.Equal(200, rec.Code)
}