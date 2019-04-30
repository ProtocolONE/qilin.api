package api

import (
	"encoding/json"
	"fmt"
	jwtverifier "github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/ProtocolONE/rbac"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/mock"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
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

var (
	packageVendorId = "22334455-888a-481a-a831-cde7ff4e50b8"
	packageId       = "33333333-888a-481a-a831-cde7ff4e50b8"
	packageGameId_1 = "029ce039-888a-481a-a831-cde7ff4e50b8"
	packageGameId_2 = "4444e039-888a-481a-a831-cde7ff4e50b8"
	packageJson     = `{
  "id": "33333333-888a-481a-a831-cde7ff4e50b8",
  "createdAt": "1970-01-01T00:00:00Z",
  "sku": "",
  "name": {"en": "Test_package"},
  "isUpgradeAllowed": false,
  "isEnabled": false,
  "isDefault": false,
  "products": [
    {
      "id": "029ce039-888a-481a-a831-cde7ff4e50b8",
      "name": "Test_game_1",
      "type": "games",
      "image": {"en": ""}
    }
  ],
  "media": {
    "image": {"en": ""},
    "cover": {"en": ""},
    "thumb": {"en": ""}
  },
  "discountPolicy": {
    "discount": 0,
    "buyOption": "whole"
  },
  "regionalRestrinctions": {
    "allowedCountries": []
  },
  "commercial": {
    "common": {
      "currency": "",
      "notifyRateJumps": false
    },
    "preOrder": {
      "date": "",
      "enabled": false
    },
    "prices": null
  }
}`
	updatePackageJson = `{
  "id": "33333333-888a-481a-a831-cde7ff4e50b8",
  "createdAt": "1970-01-01T00:00:00Z",
  "sku": "",
  "name": {"en": "Test_package_UPD"},
  "isUpgradeAllowed": true,
  "isEnabled": true,
  "isDefault": false,
  "products": [
    {
      "id": "029ce039-888a-481a-a831-cde7ff4e50b8",
      "name": "Test_game_1",
      "type": "games",
      "image": {"en": ""}
    }
  ],
  "media": {
    "image": {"en": ""},
    "cover": {"en": ""},
    "thumb": {"en": ""}
  },
  "discountPolicy": {
    "discount": 0,
    "buyOption": "whole"
  },
  "regionalRestrinctions": {
    "allowedCountries": []
  },
  "commercial": {
    "common": {
      "currency": "",
      "notifyRateJumps": false
    },
    "preOrder": {
      "date": "",
      "enabled": false
    },
    "prices": null
  }
}`

	updateErrorPackageJson = `{
  "id": "33333333-888a-481a-a831-cde7ff4e50b8",
  "createdAt": "1970-01-01T00:00:00Z",
  "sku": "",
  "name": {"en": "Test_package_UPD"},
  "isUpgradeAllowed": true,
  "isEnabled": true,
  "isDefault": false,
  "products": [
    {
      "id": "029ce039-888a-481a-a831-cde7ff4e50b8",
      "name": "Test_game_1",
      "type": "games",
      "image": {"en": ""}
    }
  ],
  "media": {
    "image": {"en": "THIS STRING IS NOT URL - 1"},
    "cover": {"en": "THIS STRING IS NOT URL - 2"},
    "thumb": {"en": "THIS STRING IS NOT URL - 3"}
  },
  "discountPolicy": {
    "discount": 0,
    "buyOption": "whole"
  },
  "regionalRestrinctions": {
    "allowedCountries": []
  },
  "commercial": {
    "common": {
      "currency": "",
      "notifyRateJumps": false
    },
    "preOrder": {
      "date": "",
      "enabled": false
    },
    "prices": null
  }
}`
)

type PackageRouterTestSuite struct {
	suite.Suite
	db   *orm.Database
	echo *echo.Echo
}

func Test_PackageRouter(t *testing.T) {
	suite.Run(t, new(PackageRouterTestSuite))
}

func (suite *PackageRouterTestSuite) SetupTest() {
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

	vendorId, err := uuid.FromString(packageVendorId)
	require.Nil(suite.T(), err, "Decode vendor uuid")
	err = db.DB().Save(&model.Vendor{
		ID:        vendorId,
		Name:      "Vendor",
		Domain3:   "domain",
		ManagerID: userId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	gameId_1, _ := uuid.FromString(packageGameId_1)
	err = db.DB().Save(&model.Game{
		ID:             gameId_1,
		CreatedAt:      time.Unix(0, 0),
		InternalName:   "Test_game_1",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
		Product:        model.ProductEntry{EntryID: gameId_1},
		VendorID:       vendorId,
		CreatorID:      userId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	gameId_2, _ := uuid.FromString(packageGameId_2)
	err = db.DB().Save(&model.Game{
		ID:             gameId_2,
		CreatedAt:      time.Unix(0, 0),
		InternalName:   "Test_game_2",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
		Product:        model.ProductEntry{EntryID: gameId_2},
		VendorID:       vendorId,
		CreatorID:      userId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	pkgId, _ := uuid.FromString(packageId)
	err = db.DB().Save(&model.Package{
		Model: model.Model{
			ID:        pkgId,
			CreatedAt: time.Unix(0, 0),
		},
		Name:             utils.LocalizedString{EN: "Test_package"},
		CreatorID:        userId,
		AllowedCountries: pq.StringArray{},
		PackagePrices: model.PackagePrices{
			Common:   model.JSONB{"currency": "", "NotifyRateJumps": false},
			PreOrder: model.JSONB{"date": "", "enabled": false},
			Prices:   []model.Price{},
		},
		VendorID: vendorId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make package")
	err = db.DB().Create(&model.PackageProduct{
		PackageID: pkgId,
		ProductID: gameId_1,
	}).Error
	require.Nil(suite.T(), err, "Unable to make package product")

	gameService, err := orm.NewGameService(db)
	require.Nil(suite.T(), err)

	echoObj := echo.New()
	echoObj.Validator = &QilinValidator{validator: validator.New()}
	echoObj.HTTPErrorHandler = func(e error, context echo.Context) {
		QilinErrorHandler(e, context, true)
	}

	enforcer := rbac.NewEnforcer()
	ownerProvider := orm.NewOwnerProvider(db)
	membership := orm.NewMembershipService(db, ownerProvider, enforcer, mock.NewMailer(), "")
	err = membership.Init()
	if err != nil {
		suite.FailNow("Membership fail", "%v", err)
	}

	echoObj.Use(rbac_echo.NewAppContextMiddleware(ownerProvider, enforcer))
	echoObj.Use(suite.localAuth())

	service, err := orm.NewPackageService(db, gameService)
	require.Nil(suite.T(), err)

	productService, err := orm.NewProductService(db)
	require.Nil(suite.T(), err)

	_, err = InitPackageRouter(echoObj.Group("/api/v1"), service, productService)
	require.Nil(suite.T(), err)

	suite.db = db
	suite.echo = echoObj
}

func (suite *PackageRouterTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *PackageRouterTestSuite) localAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Set(context.TokenKey, &jwtverifier.UserInfo{UserID: userId})
			return next(c)
		}
	}
}

func (suite *PackageRouterTestSuite) TestShouldReturnPackage() {
	url := fmt.Sprintf("/api/v1/packages/%s", packageId)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)
	assert.JSONEq(suite.T(), packageJson, rec.Body.String())
}

func (suite *PackageRouterTestSuite) TestShouldCreatePackage() {
	should := require.New(suite.T())

	url := fmt.Sprintf("/api/v1/vendors/%s/packages", packageVendorId)
	reader := strings.NewReader(`{"name": "New_package_2", "products": ["029ce039-888a-481a-a831-cde7ff4e50b8"]}`)
	req := httptest.NewRequest(http.MethodPost, url, reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	should.Equal(http.StatusCreated, rec.Code)
	dto := packageDTO{}
	err := json.Unmarshal(rec.Body.Bytes(), &dto)
	should.Nil(err)
	should.Equal("New_package_2", dto.Name.EN)
	should.Equal(1, len(dto.Products))
	should.Equal(packageGameId_1, dto.Products[0].ID.String())
	should.Equal("Test_game_1", dto.Products[0].Name)
	should.Equal(model.ProductGame, model.ProductType(dto.Products[0].Type))
	should.True(time.Now().Unix()-dto.CreatedAt.Unix() >= 0)
	should.True(time.Now().Unix()-dto.CreatedAt.Unix() <= 5)
}

func (suite *PackageRouterTestSuite) TestShouldReturnPackageList() {
	should := assert.New(suite.T())

	url := fmt.Sprintf("/api/v1/vendors/%s/packages", packageVendorId)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	should.Equal(http.StatusOK, rec.Code)

	dto := []packageItemDTO{}
	err := json.Unmarshal(rec.Body.Bytes(), &dto)
	should.Nil(err)
	should.Equal(len(dto), 1)
	should.Equal("Test_package", dto[0].Name.EN)
	should.Equal(packageId, dto[0].ID.String())
}

func (suite *PackageRouterTestSuite) TestShouldUpdatePackage() {
	should := assert.New(suite.T())

	url := fmt.Sprintf("/api/v1/packages/%s", packageId)
	req := httptest.NewRequest(http.MethodPut, url, strings.NewReader(updatePackageJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	should.Equal(http.StatusOK, rec.Code)
	should.JSONEq(updatePackageJson, rec.Body.String())
}

func (suite *PackageRouterTestSuite) TestShouldRiseErrorWhenUpdateWrongPackage() {
	should := assert.New(suite.T())

	url := fmt.Sprintf("/api/v1/packages/%s", packageId)
	req := httptest.NewRequest(http.MethodPut, url, strings.NewReader(updateErrorPackageJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	should.NotEqual(http.StatusOK, rec.Code)
}

func (suite *PackageRouterTestSuite) TestShouldManageGames() {
	should := assert.New(suite.T())

	{
		url := fmt.Sprintf("/api/v1/packages/%s/products/add", packageId)
		reader := strings.NewReader(fmt.Sprintf(`["%s"]`, packageGameId_2))

		req := httptest.NewRequest(http.MethodPost, url, reader)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		should.Equal(http.StatusOK, rec.Code)
		dto := packageDTO{}
		err := json.Unmarshal(rec.Body.Bytes(), &dto)
		should.Nil(err)
		should.Len(dto.Products, 2)
		should.Equal(packageGameId_2, dto.Products[1].ID.String())
	}

	{
		url := fmt.Sprintf("/api/v1/packages/%s/products/remove", packageId)
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(fmt.Sprintf(`["%s"]`, packageGameId_1)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		should.Equal(http.StatusOK, rec.Code)
		dto := packageDTO{}
		err := json.Unmarshal(rec.Body.Bytes(), &dto)
		should.Nil(err)
		should.Len(dto.Products, 1)
		should.Equal(packageGameId_2, dto.Products[0].ID.String())
	}
}

func (suite *PackageRouterTestSuite) TestGetPackageShouldDeletePackage() {
	url := fmt.Sprintf("/api/v1/packages/%s", packageId)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}
