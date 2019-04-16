package orm_test

import (
	"github.com/ProtocolONE/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"qilin-api/pkg/api/mock"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"

	"github.com/satori/go.uuid"

	"github.com/stretchr/testify/suite"
)

type bundleServiceTestSuite struct {
	suite.Suite
	db          *orm.Database
	service     model.BundleService
	packages    []uuid.UUID
	games       []uuid.UUID
	vendorId    uuid.UUID
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

	service, err := orm.NewBundleService(suite.db)
	if err != nil {
		suite.Fail("Unable to create service", "%v", err)
	}

	suite.service = service

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

	packageService, err := orm.NewPackageService(db, gameService)
	if err != nil {
		suite.Fail("Unable to create package service", "%v", err)
	}
	pkgA, err := packageService.Create(vendor.ID, user.ID, "Mega package A", []uuid.UUID{gameA.ID})
	if err != nil {
		suite.Fail("Unable to create package", "%v", err)
	}
	pkgB, err := packageService.Create(vendor.ID, user.ID,"Mega package B", []uuid.UUID{gameB.ID})
	if err != nil {
		suite.Fail("Unable to create package", "%v", err)
	}

	suite.packages = []uuid.UUID{pkgA.ID, pkgB.ID}
	suite.games = []uuid.UUID{gameA.ID, gameB.ID}
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

	bundle, err := suite.service.CreateStore(suite.vendorId, "Mega bundle", suite.packages)
	should.Nil(err)
	should.Equal("Mega bundle", bundle.Name)
	should.Equal(suite.vendorId, bundle.VendorID)
	should.Equal(2, len(bundle.Packages))
	should.Equal(bundle.Packages[0].ID, suite.packages[0])

	bundle2, err := suite.service.CreateStore(suite.vendorId, "Bundle Hundle", suite.packages[0:1])
	should.Nil(err)
	should.Equal("Bundle Hundle", bundle2.Name)
	should.Equal(suite.vendorId, bundle2.VendorID)
	should.Equal(1, len(bundle2.Packages))
	should.Equal(bundle2.Packages[0].ID, suite.packages[0])

	bundleErr, err := suite.service.CreateStore(suite.vendorId, "Empty bundle", suite.packages[:0])
	should.NotNil(err)
	should.Nil(bundleErr)

	bundleErr, err = suite.service.CreateStore(suite.vendorId, "", suite.packages)
	should.NotNil(err, "Empty name")
	should.Nil(bundleErr)

	bundleErr, err = suite.service.CreateStore(uuid.NamespaceDNS, "Error bundle", suite.packages)
	should.NotNil(err, "Vendor not found")
	should.Nil(bundleErr)

	list, err := suite.service.GetStoreList(suite.vendorId, "", "-date", 0, 20)
	should.Nil(err)
	should.Equal(2, len(list))
	should.Equal("Bundle Hundle", list[0].Name)
	should.Equal("Mega bundle", list[1].Name)

	list2, err := suite.service.GetStoreList(suite.vendorId, "", "+date", 1, 20)
	should.Nil(err)
	should.Equal(1, len(list2))
	should.Equal("Bundle Hundle", list2[0].Name)

	list3, err := suite.service.GetStoreList(suite.vendorId, "", "-name", 0, 1)
	should.Nil(err)
	should.Equal(1, len(list3))
	should.Equal("Mega bundle", list3[0].Name)

	bundle3, err := suite.service.Get(bundle.ID)
	should.Nil(err)
	b3_pkgs, err := bundle3.GetPackages()
	should.Nil(err)
	should.Len(b3_pkgs, 2)
	should.Equal(b3_pkgs[0].ID, suite.packages[0])
	should.Equal(b3_pkgs[1].ID, suite.packages[1])
	should.Equal(bundle.Name, bundle3.GetName())
	isIn_1, err := bundle3.IsContains(suite.games[0])
	should.Nil(err)
	should.Equal(true, isIn_1, "GameA inside bundle")
	isIn_2, err := bundle3.IsContains(suite.games[1])
	should.Nil(err)
	should.Equal(true, isIn_2, "GameB inside bundle")
	isNotIn, err := bundle3.IsContains(uuid.NamespaceDNS)
	should.Nil(err)
	should.Equal(false, isNotIn, "Is not inside bundle")

	bundle.Name = "Updated bundle"
	bundle.IsEnabled = true
	bundleUpd, err := suite.service.UpdateStore(bundle)
	should.Nil(err)
	should.NotNil(bundleUpd)
	should.Equal(true, bundleUpd.IsEnabled)
	should.Equal("Updated bundle", bundleUpd.Name)

	bundleErr, err = suite.service.UpdateStore(&model.StoreBundle{})
	should.NotNil(err)
	should.Nil(bundleErr)

	err = suite.service.Delete(bundle.ID)
	should.Nil(err, "Remove bundle")

	list, err = suite.service.GetStoreList(suite.vendorId, "", "+date", 0, 20)
	should.Nil(err)
	should.Equal(1, len(list))
	should.Equal("Bundle Hundle", list[0].Name)

	bundleErrIface, err := suite.service.Get(bundle.ID)
	should.NotNil(err)
	should.Nil(bundleErrIface)

	bundleErr, err = suite.service.UpdateStore(bundle)
	should.NotNil(err, "Update deleted bundle")
	should.Nil(bundleErr)
}
