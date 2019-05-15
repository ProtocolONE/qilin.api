package orm_test

import (
	"github.com/ProtocolONE/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"qilin-api/pkg/api/mock"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"

	"github.com/satori/go.uuid"

	"github.com/stretchr/testify/suite"
)

type bundleServiceTestSuite struct {
	suite.Suite
	db         *orm.Database
	service    model.BundleService
	packages   []uuid.UUID
	games      []uuid.UUID
	vendorId   uuid.UUID
	extraGames []*model.Game
	userId     string
}

func Test_BundleService(t *testing.T) {
	suite.Run(t, new(bundleServiceTestSuite))
}

func (suite *bundleServiceTestSuite) SetupTest() {
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}

	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.Fail("Unable to connect to database:", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

	suite.db = db

	// Create user
	user := model.User{
		ID:       uuid.NewV4().String(),
		Login:    "test@protocol.one",
		Password: "megapass",
		Nickname: "Test",
		Lang:     "ru",
	}
	err = db.DB().Create(&user).Error
	suite.Nil(err, "Unable to create user")

	suite.userId = user.ID

	ownProvider := orm.NewOwnerProvider(suite.db)
	enf := rbac.NewEnforcer()
	membershipService := orm.NewMembershipService(suite.db, ownProvider, enf, mock.NewMailer(), "")

	vendorService, err := orm.NewVendorService(db, membershipService)
	suite.Nil(err, "Unable make vendor service")

	// Create vendor
	vendor := model.Vendor{
		ID:              uuid.NewV4(),
		Name:            "domino",
		Domain3:         "domino",
		Email:           "domino@proto.com",
		HowManyProducts: "+1000",
		ManagerID:       user.ID,
	}
	_, err = vendorService.Create(&vendor)
	suite.Nil(err, "Must create new vendor")
	suite.vendorId = vendor.ID

	// Create game
	gameService, _ := orm.NewGameService(db)
	gameA, err := gameService.Create(user.ID, vendor.ID, "GameA")
	if err != nil {
		suite.Fail("Unable to create game", "%v", err)
	}

	gameB, err := gameService.Create(user.ID, vendor.ID, "GameB")
	if err != nil {
		suite.Fail("Unable to create game", "%v", err)
	}
	gameC, err := gameService.Create(user.ID, vendor.ID, "GameC")
	if err != nil {
		suite.Fail("Unable to create game", "%v", err)
	}
	gameD, err := gameService.Create(user.ID, vendor.ID, "GameD")
	if err != nil {
		suite.Fail("Unable to create game", "%v", err)
	}
	suite.extraGames = []*model.Game{gameC, gameD}

	packageService, err := orm.NewPackageService(db, gameService)
	if err != nil {
		suite.Fail("Unable to create package service", "%v", err)
	}
	pkgA, err := packageService.Create(vendor.ID, user.ID, "Mega package A", []uuid.UUID{gameA.ID})
	if err != nil {
		suite.Fail("Unable to create package A", "%v", err)
	}
	pkgB, err := packageService.Create(vendor.ID, user.ID, "Mega package B", []uuid.UUID{gameB.ID})
	if err != nil {
		suite.Fail("Unable to create package B", "%v", err)
	}
	pkgB.Discount = 25
	_, err = packageService.Update(pkgB)
	if err != nil {
		suite.Fail("Unable to update package B", "%v", err)
	}

	priceService := orm.NewPriceService(db)
	err = priceService.UpdateBase(pkgA.ID, &model.BasePrice{
		PackagePrices: model.PackagePrices{
			Common: model.JSONB{
				"currency":        "USD",
				"NotifyRateJumps": true,
			},
			PreOrder: model.JSONB{
				"date":    "2019-01-22T07:53:16Z",
				"enabled": false,
			},
		},
	})
	if err != nil {
		suite.Fail("Error while update base game price", "%v", err)
	}
	err = priceService.Update(pkgA.ID, &model.Price{
		Currency: "USD",
		Price:    10.25,
	})
	if err != nil {
		suite.Fail("Error while update game price", "%v", err)
	}

	err = priceService.Update(pkgB.ID, &model.Price{
		Currency: "USD",
		Price:    5,
	})
	if err != nil {
		suite.Fail("Error while update game price", "%v", err)
	}

	suite.packages = []uuid.UUID{pkgA.ID, pkgB.ID}
	suite.games = []uuid.UUID{gameA.ID, gameB.ID}

	service, err := orm.NewBundleService(suite.db, packageService, gameService)
	if err != nil {
		suite.Fail("Unable to create service", "%v", err)
	}

	suite.service = service
}

func (suite *bundleServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}

	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *bundleServiceTestSuite) TestBundles() {
	should := require.New(suite.T())

	bundleIface, err := suite.service.CreateStore(suite.vendorId, suite.userId, "Mega bundle", suite.packages)
	bundle, _ := bundleIface.(*model.StoreBundle)
	should.Nil(err)
	should.Equal("Mega bundle", bundle.Name.EN)
	should.Equal(suite.vendorId, bundle.VendorID)
	should.Equal(2, len(bundle.Packages))
	should.Equal(bundle.Packages[0].ID, suite.packages[0])

	bundleGames, err := bundle.GetGames()
	should.Nil(err)
	should.Len(bundleGames, 2)

	should.Equal("Mega bundle", bundle.GetName().EN)
	currency, price, discount, err := bundle.GetPrice()
	should.Nil(err)
	should.Equal("USD", currency)
	should.True(price-15.25 < 0.01)
	should.True(price-15.25 > -0.01)
	should.True(math.Abs(float64(price-price*discount*0.01)-14.0) < 0.01)

	bundleIface2, err := suite.service.CreateStore(suite.vendorId, suite.userId, "Bundle Humble", suite.packages[0:1])
	bundle2, _ := bundleIface2.(*model.StoreBundle)
	should.Nil(err)
	should.Equal("Bundle Humble", bundle2.Name.EN)
	should.Equal(suite.vendorId, bundle2.VendorID)
	should.Equal(1, len(bundle2.Packages))
	should.Equal(bundle2.Packages[0].ID, suite.packages[0])

	bundleErr, err := suite.service.CreateStore(suite.vendorId, suite.userId, "Empty bundle", suite.packages[:0])
	should.NotNil(err)
	should.Nil(bundleErr)

	bundleErr, err = suite.service.CreateStore(suite.vendorId, suite.userId, "", suite.packages)
	should.NotNil(err, "Empty name")
	should.Nil(bundleErr)

	bundleErr, err = suite.service.CreateStore(uuid.NamespaceDNS, suite.userId, "Error bundle", suite.packages)
	should.NotNil(err, "Vendor not found")
	should.Nil(bundleErr)

	total, list, err := suite.service.GetStoreList(suite.vendorId, "", "-date", 0, 20, nil)
	should.Nil(err)
	should.Equal(2, total)
	should.Equal(2, len(list))
	should.Equal("Bundle Humble", list[0].GetName().EN)
	should.Equal("Mega bundle", list[1].GetName().EN)

	total, list2, err := suite.service.GetStoreList(suite.vendorId, "", "+date", 1, 20, nil)
	should.Nil(err)
	should.Equal(1, len(list2))
	should.Equal("Bundle Humble", list2[0].GetName().EN)

	total, list3, err := suite.service.GetStoreList(suite.vendorId, "", "-name", 0, 1, nil)
	should.Nil(err)
	should.Equal(1, len(list3))
	should.Equal("Mega bundle", list3[0].GetName().EN)

	bundle3, err := suite.service.Get(bundle.ID)
	should.Nil(err)
	b3_pkgs, err := bundle3.GetPackages()
	should.Nil(err)
	should.Len(b3_pkgs, 2)
	should.Equal(b3_pkgs[0].ID, suite.packages[0])
	should.Equal(b3_pkgs[1].ID, suite.packages[1])
	should.Equal(bundle.Name, *bundle3.GetName())
	isIn_1, err := bundle3.IsContains(suite.games[0])
	should.Nil(err)
	should.Equal(true, isIn_1, "GameA inside bundle")
	isIn_2, err := bundle3.IsContains(suite.games[1])
	should.Nil(err)
	should.Equal(true, isIn_2, "GameB inside bundle")
	isNotIn, err := bundle3.IsContains(uuid.NamespaceDNS)
	should.Nil(err)
	should.Equal(false, isNotIn, "Is not inside bundle")

	bundle.Name = utils.LocalizedString{EN: "Updated bundle"}
	bundle.IsEnabled = true
	bundleUpdIface, err := suite.service.UpdateStore(bundle)
	should.Nil(err)
	should.NotNil(bundleUpdIface)
	bundleUpd, _ := bundleUpdIface.(*model.StoreBundle)
	should.Equal(true, bundleUpd.IsEnabled)
	should.Equal("Updated bundle", bundleUpd.Name.EN)

	bundleErr, err = suite.service.UpdateStore(&model.StoreBundle{})
	should.NotNil(err)
	should.Nil(bundleErr)

	err = suite.service.AddPackages(bundle.ID, []uuid.UUID{
		suite.extraGames[0].DefaultPackageID,
		suite.extraGames[1].DefaultPackageID,
	})
	should.Nil(err, "Add packages")

	bundleErrIface, err := suite.service.Get(bundle.ID)
	should.Nil(err)
	should.NotNil(bundleErrIface)

	bundle, ok := bundleErrIface.(*model.StoreBundle)
	should.Equal(true, ok, "Bundle must be for store")
	should.Len(bundle.Packages, 4, "Bundle must have 4 packages")
	should.Equal(bundle.Packages[0].ID, suite.packages[0])
	should.Equal(bundle.Packages[1].ID, suite.packages[1])
	should.Equal(bundle.Packages[2].ID, suite.extraGames[0].DefaultPackageID)
	should.Equal(bundle.Packages[3].ID, suite.extraGames[1].DefaultPackageID)

	err = suite.service.RemovePackages(bundle.ID, []uuid.UUID{suite.packages[1]})
	should.Nil(err, "Remove package")

	bundleErrIface, err = suite.service.Get(bundle.ID)
	should.Nil(err)
	should.NotNil(bundleErrIface)

	bundle, ok = bundleErrIface.(*model.StoreBundle)
	should.Equal(true, ok, "Bundle must be for store")
	should.Len(bundle.Packages, 3, "Bundle must have 3 packages")
	should.Equal(bundle.Packages[0].ID, suite.packages[0])
	should.Equal(bundle.Packages[1].ID, suite.extraGames[0].DefaultPackageID)
	should.Equal(bundle.Packages[2].ID, suite.extraGames[1].DefaultPackageID)

	// Removing bundle and tries do some actions with it
	err = suite.service.Delete(bundle.ID)
	should.Nil(err, "Remove bundle")

	total, list, err = suite.service.GetStoreList(suite.vendorId, "", "+date", 0, 20, nil)
	should.Nil(err)
	should.Equal(1, total)
	should.Equal(1, len(list))
	should.Equal("Bundle Humble", list[0].GetName().EN)

	bundleErrIface, err = suite.service.Get(bundle.ID)
	should.NotNil(err)
	should.Nil(bundleErrIface)

	bundleErr, err = suite.service.UpdateStore(bundle)
	should.NotNil(err, "Update deleted bundle")
	should.Nil(bundleErr)
}
