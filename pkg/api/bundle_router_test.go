package api

import (
	"encoding/json"
	"fmt"
	jwtverifier "github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/ProtocolONE/rbac"
	"github.com/jinzhu/gorm"
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
	bundleVendorId       = "11112222-888a-481a-a831-cde7ff4e50b8"
	bundleID             = "44444444-888a-481a-a831-cde7ff4e50b8"
	bundleGameId_1       = "029ce039-888a-481a-a831-cde7ff4e50b8"
	bundleGameId_2       = "6666e039-888a-481a-a831-cde7ff4e50b8"
	bundlePackageId_1    = "33333333-888a-481a-a831-cde7ff4e50b8"
	bundlePackageId_2    = "00022233-888a-481a-a831-cde7ff4e50b8"
	emptyStoreBundleJson = `{
  "id": "44444444-888a-481a-a831-cde7ff4e50b8",
  "createdAt": "1970-01-01T00:00:00Z",
  "sku": "",
  "name": {"en": "Mega bundle"},
  "isUpgradeAllowed": false,
  "isEnabled": false,
  "discountPolicy": {
    "discount": 0,
    "buyOption": "whole"
  },
  "regionalRestrinctions": {
    "allowedCountries": []
  },
  "packages": [
    {
      "id": "33333333-888a-481a-a831-cde7ff4e50b8",
      "createdAt": "1970-01-01T00:00:00Z",
      "sku": "",
      "name": {"en": "Test_package"},
      "isUpgradeAllowed": false,
      "isEnabled": false,
      "isDefault": false,
      "defaultProductId": "00000000-0000-0000-0000-000000000000",
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
    }
  ]
}`
	updateStoreBundleJson = `{
  "id": "44444444-888a-481a-a831-cde7ff4e50b8",
  "createdAt": "1970-01-01T00:00:00Z",
  "sku": "555666222",
  "name": {"en": "Mega bundle 2"},
  "isUpgradeAllowed": true,
  "isEnabled": true,
  "discountPolicy": {
    "discount": 10,
    "buyOption": "part"
  },
  "regionalRestrinctions": {
    "allowedCountries": []
  },
  "packages": [
    {
      "id": "33333333-888a-481a-a831-cde7ff4e50b8",
      "createdAt": "1970-01-01T00:00:00Z",
      "sku": "",
      "name": {"en": "Test_package"},
      "isUpgradeAllowed": false,
      "isEnabled": false,
      "isDefault": false,
      "defaultProductId":"00000000-0000-0000-0000-000000000000",
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
    }
  ]
}`

	listingStoreBundle = `[
  {
    "id": "44444444-888a-481a-a831-cde7ff4e50b8",
    "createdAt": "1970-01-01T00:00:00Z",
    "sku": "",
    "name": {
      "en": "Mega bundle"
    },
    "isUpgradeAllowed": false,
    "isEnabled": false
  }
]
`
)

type BundleRouterTestSuite struct {
	suite.Suite
	db   *orm.Database
	echo *echo.Echo
}

func Test_BundleRouter(t *testing.T) {
	suite.Run(t, new(BundleRouterTestSuite))
}

func (suite *BundleRouterTestSuite) makeGame(db *gorm.DB, gameId, name string) uuid.UUID {
	vendorId, err := uuid.FromString(bundleVendorId)
	require.Nil(suite.T(), err, "Decode vendor uuid")

	id, _ := uuid.FromString(gameId)
	err = db.Save(&model.Game{
		ID:             id,
		CreatedAt:      time.Unix(0, 0),
		InternalName:   name,
		ReleaseDate:    time.Unix(0, 0),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
		Product:        model.ProductEntry{EntryID: id},
		VendorID:       vendorId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	return id
}

func (suite *BundleRouterTestSuite) makePackage(db *gorm.DB, packageId, name string, gameId uuid.UUID) uuid.UUID {
	vendorId, err := uuid.FromString(bundleVendorId)
	require.Nil(suite.T(), err, "Decode vendor uuid")

	id, _ := uuid.FromString(packageId)
	err = db.Save(&model.Package{
		Model: model.Model{
			ID:        id,
			CreatedAt: time.Unix(0, 0),
		},
		Name:             utils.LocalizedString{EN: name},
		AllowedCountries: pq.StringArray{},
		PackagePrices: model.PackagePrices{
			Common:   model.JSONB{"currency": "", "NotifyRateJumps": false},
			PreOrder: model.JSONB{"date": "", "enabled": false},
			Prices:   []model.Price{},
		},
		VendorID: vendorId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make package")
	err = db.Create(&model.PackageProduct{
		PackageID: id,
		ProductID: gameId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make package product")

	return id
}

func (suite *BundleRouterTestSuite) makeBundle(db *gorm.DB, bundleId, name string, packageId uuid.UUID) uuid.UUID {
	vendorId, err := uuid.FromString(bundleVendorId)
	require.Nil(suite.T(), err, "Decode vendor uuid")

	id, _ := uuid.FromString(bundleId)
	err = db.Create(&model.StoreBundle{
		Model: model.Model{
			ID:        id,
			CreatedAt: time.Unix(0, 0),
		},
		Name:             utils.LocalizedString{EN: name},
		AllowedCountries: pq.StringArray{},
		VendorID:         vendorId,
		Bundle:           model.BundleEntry{EntryID: id},
	}).Error
	require.Nil(suite.T(), err, "Unable to make bundle")
	err = db.Create(&model.BundlePackage{
		PackageID: packageId,
		BundleID:  id,
	}).Error
	require.Nil(suite.T(), err, "Unable to make bundle package")

	return id
}

func (suite *BundleRouterTestSuite) SetupTest() {
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

	vendorId, err := uuid.FromString(bundleVendorId)
	require.Nil(suite.T(), err, "Decode vendor uuid")

	err = db.DB().Save(&model.Vendor{
		ID:        vendorId,
		Name:      "Vendor",
		Domain3:   "domain",
		ManagerID: userId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	gameId_1 := suite.makeGame(db.DB(), bundleGameId_1, "Test_game_1")
	gameId_2 := suite.makeGame(db.DB(), bundleGameId_2, "Test_game_2")

	packageId := suite.makePackage(db.DB(), bundlePackageId_1, "Test_package", gameId_1)
	suite.makeBundle(db.DB(), bundleID, "Mega bundle", packageId)

	suite.makePackage(db.DB(), bundlePackageId_2, "Test_package", gameId_2)

	echoObj := echo.New()
	echoObj.Validator = &QilinValidator{validator: validator.New()}
	echoObj.HTTPErrorHandler = func(e error, context echo.Context) {
		QilinErrorHandler(e, context, true)
	}

	ownerProvider := orm.NewOwnerProvider(db)
	enforcer := rbac.NewEnforcer()
	membership := orm.NewMembershipService(db, ownerProvider, enforcer, mock.NewMailer(), "")
	err = membership.Init()
	if err != nil {
		suite.FailNow("Membership fail", "%v", err)
	}

	echoObj.Use(rbac_echo.NewAppContextMiddleware(ownerProvider, enforcer))
	echoObj.Use(suite.localAuth())

	gameService, err := orm.NewGameService(db)
	require.Nil(suite.T(), err)
	packageService, err := orm.NewPackageService(db, gameService)
	require.Nil(suite.T(), err)
	service, err := orm.NewBundleService(db, packageService, gameService)
	require.Nil(suite.T(), err)

	_, err = InitBundleRouter(echoObj.Group("/api/v1"), service)
	require.Nil(suite.T(), err, "Unable to init router")

	suite.db = db
	suite.echo = echoObj
}

func (suite *BundleRouterTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *BundleRouterTestSuite) localAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Set(context.TokenKey, &jwtverifier.UserInfo{UserID: userId})
			return next(c)
		}
	}
}

func (suite *BundleRouterTestSuite) TestShouldReturnStoreBundle() {
	url := fmt.Sprintf("/api/v1/bundles/%s/store", bundleID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)
	assert.JSONEq(suite.T(), emptyStoreBundleJson, rec.Body.String())
}

func (suite *BundleRouterTestSuite) TestShouldCreateBundle() {
	should := assert.New(suite.T())

	reader := strings.NewReader(`{"name": "Mega bundle 2", "packages": ["33333333-888a-481a-a831-cde7ff4e50b8"]}`)
	url := fmt.Sprintf("/api/v1/vendors/%s/bundles/store", bundleVendorId)
	req := httptest.NewRequest(http.MethodPost, url, reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	should.Equal(http.StatusCreated, rec.Code)
	dto := storeBundleDTO{}
	err := json.Unmarshal(rec.Body.Bytes(), &dto)
	should.Nil(err)
	should.Equal("Mega bundle 2", dto.Name.EN)
	should.Equal(1, len(dto.Packages))
	should.Equal("33333333-888a-481a-a831-cde7ff4e50b8", dto.Packages[0].ID.String())
	should.Equal(1, len(dto.Packages[0].Products))
	should.Equal("Test_game_1", dto.Packages[0].Products[0].Name)
}

func (suite *BundleRouterTestSuite) TestShouldReturnStoreList() {
	url := fmt.Sprintf("/api/v1/vendors/%s/bundles/store", bundleVendorId)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)
	assert.JSONEq(suite.T(), listingStoreBundle, rec.Body.String())
}

func (suite *BundleRouterTestSuite) TestShouldAppendPackages() {
	url := fmt.Sprintf("/api/v1/bundles/%s/packages", bundleID)
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(`["00022233-888a-481a-a831-cde7ff4e50b8"]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *BundleRouterTestSuite) TestShouldUpdateBundle() {
	should := assert.New(suite.T())

	url := fmt.Sprintf("/api/v1/bundles/%s/store", bundleID)
	req := httptest.NewRequest(http.MethodPut, url, strings.NewReader(updateStoreBundleJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	should.Equal(http.StatusOK, rec.Code)
	should.JSONEq(updateStoreBundleJson, rec.Body.String())
}

func (suite *BundleRouterTestSuite) TestShouldRemovePackages() {
	url := fmt.Sprintf("/api/v1/bundles/%s/packages", bundleID)
	req := httptest.NewRequest(http.MethodDelete, url, strings.NewReader(`["33333333-888a-481a-a831-cde7ff4e50b8"]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *BundleRouterTestSuite) TestShouldDeleteBundle() {
	url := fmt.Sprintf("/api/v1/bundles/%s", bundleID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	rec := httptest.NewRecorder()

	suite.echo.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}
