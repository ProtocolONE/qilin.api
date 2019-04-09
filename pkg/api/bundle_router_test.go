package api

import (
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
)

var (
	bundleID = "44444444-888a-481a-a831-cde7ff4e50b8"
	emptyBaseBundle = `{
  "id": "44444444-888a-481a-a831-cde7ff4e50b8",
  "createdAt": "1970-01-01T00:00:00Z",
  "sku": "",
  "name": "Mega bundle",
  "isUpgradeAllowed": false,
  "isEnabled": false,
  "discountPolicy": {
    "discount": 0,
    "buyOption": ""
  },
  "regionalRestrinctions": {
    "allowedCountries": []
  },
  "packages": [
    {
      "id": "33333333-888a-481a-a831-cde7ff4e50b8",
      "createdAt": "1970-01-01T00:00:00Z",
      "sku": "",
      "name": "Test_package",
      "isUpgradeAllowed": false,
      "isEnabled": false,
      "products": [
        {
          "id": "029ce039-888a-481a-a831-cde7ff4e50b8",
          "name": "Test_game_1",
          "type": "games",
          "image": ""
        }
      ],
      "media": {
        "image": "",
        "cover": "",
        "thumb": ""
      },
      "discountPolicy": {
        "discount": 0,
        "buyOption": ""
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
)

type BundleRouterTestSuite struct {
	suite.Suite
	db     *orm.Database
	echo   *echo.Echo
	router *BundleRouter
}

func Test_BundleRouter(t *testing.T) {
	suite.Run(t, new(BundleRouterTestSuite))
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

	venId, err := uuid.FromString(vendorId)
	require.Nil(suite.T(), err, "Decode vendor uuid")

	id, _ := uuid.FromString(TestID)
	err = db.DB().Save(&model.Game{
		ID:             id,
		CreatedAt:      time.Unix(0, 0),
		InternalName:   "Test_game_1",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
		Product:        model.ProductEntry{EntryID: id},
		VendorID:       venId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make game")

	pkgId, _ := uuid.FromString(packageID)
	err = db.DB().Save(&model.Package{
		Model:  model.Model{
			ID: pkgId,
			CreatedAt: time.Unix(0, 0),
		},
		Name:   "Test_package",
		AllowedCountries: pq.StringArray{},
		PackagePrices: model.PackagePrices{
			Common: model.JSONB{"currency":"","NotifyRateJumps":false},
			PreOrder: model.JSONB{"date":"","enabled":false},
			Prices: []model.Price{},
		},
		VendorID:       venId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make package")
	err = db.DB().Create(&model.PackageProduct{
		PackageID: pkgId,
		ProductID: id,
	}).Error
	require.Nil(suite.T(), err, "Unable to make package product")

	buId, _ := uuid.FromString(bundleID)
	err = db.DB().Create(&model.StoreBundle{
		Model:  model.Model{
			ID: buId,
			CreatedAt: time.Unix(0, 0),
		},
		Name:   "Mega bundle",
		AllowedCountries: pq.StringArray{},
		VendorID: venId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make bundle")
	err = db.DB().Create(&model.BundlePackage{
		PackageID: pkgId,
		BundleID:  buId,
	}).Error
	require.Nil(suite.T(), err, "Unable to make bundle package")
	err = db.DB().Create(&model.BundleEntry{
		EntryID:  buId,
		EntryType: model.BundleStore,
	}).Error
	require.Nil(suite.T(), err, "Unable to make bundle entry")

	echoObj := echo.New()
	service, err := orm.NewBundleService(db)
	router, err := InitBundleRouter(echoObj.Group("/api/v1"), service)

	echoObj.Validator = &QilinValidator{validator: validator.New()}

	suite.db = db
	suite.router = router
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

func (suite *BundleRouterTestSuite) TestGetBundleShouldReturnEmptyObject() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/bundles/store/:bundleId")
	c.SetParamNames("bundleId")
	c.SetParamValues(bundleID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.GetStore(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.JSONEq(suite.T(), emptyBaseBundle, rec.Body.String())
	}
}
