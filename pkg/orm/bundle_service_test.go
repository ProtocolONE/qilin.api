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
	service     *orm.BundleService
	packages    []uuid.UUID
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

	userId := user.ID


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
		ManagerID:       userId,
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
	pkgA, err := packageService.Create(vendor.ID, "Mega package A", []uuid.UUID{gameA.ID})
	if err != nil {
		suite.Fail("Unable to create package", "%v", err)
	}
	pkgB, err := packageService.Create(vendor.ID, "Mega package B", []uuid.UUID{gameB.ID})
	if err != nil {
		suite.Fail("Unable to create package", "%v", err)
	}

	suite.packages = []uuid.UUID{pkgA.ID, pkgB.ID}
}

func (suite *bundleServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}

	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *bundleServiceTestSuite) TestCreateBundle() {
	should := require.New(suite.T())

	bundle, err := suite.service.CreateStore(suite.vendorId, "Mega bundle", suite.packages)
	should.Nil(err)
	should.Equal("Mega bundle", bundle.Name)
	should.Equal(2, len(bundle.Packages))
	should.Equal(bundle.Packages[0].ID, suite.packages[0])

	bundle2, err := suite.service.CreateStore(suite.vendorId, "Bundle Hundle", suite.packages[0:1])
	should.Nil(err)
	should.Equal("Bundle Hundle", bundle2.Name)
	should.Equal(1, len(bundle2.Packages))
	should.Equal(bundle2.Packages[0].ID, suite.packages[0])

	list, err := suite.service.GetStoreList(suite.vendorId, "", "-date", 0, 20)
	should.Nil(err)
	should.Equal(2, len(list))
	should.Equal("Bundle Hundle", list[0].Name)
}
