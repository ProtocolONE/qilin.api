package orm_test

import (
	"github.com/ProtocolONE/rbac"
	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"qilin-api/pkg/api/mock"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"

	"github.com/satori/go.uuid"

	"github.com/stretchr/testify/suite"
)

type packageServiceTestSuite struct {
	suite.Suite
	db          *orm.Database
	service     model.PackageService
	gameService model.GameService
	vendorId    uuid.UUID
	userId      string
}

func Test_PackageService(t *testing.T) {
	suite.Run(t, new(packageServiceTestSuite))
}

func (suite *packageServiceTestSuite) SetupTest() {
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
		ID:       random.String(8, "0123456789"),
		Login:    "test@protocol.one",
		Password: "megapass",
		Nickname: "Test",
		Lang:     "ru",
	}
	err = db.DB().Create(&user).Error
	suite.Nil(err, "Unable to create user")
	suite.userId = user.ID

	// Create vendor
	ownProvider := orm.NewOwnerProvider(suite.db)
	enf := rbac.NewEnforcer()
	membershipService := orm.NewMembershipService(suite.db, ownProvider, enf, mock.NewMailer(), "")
	vendorService, err := orm.NewVendorService(db, membershipService)
	suite.Nil(err, "Unable make vendor service")
	vendor := model.Vendor{
		ID:              uuid.NewV4(),
		Name:            "domino",
		Domain3:         "domino",
		Email:           "domino@proto.com",
		HowManyProducts: "+1000",
		ManagerID:       suite.userId,
	}
	_, err = vendorService.Create(&vendor)
	suite.Nil(err, "Must create new vendor")
	suite.vendorId = vendor.ID
	suite.gameService, _ = orm.NewGameService(db)
	suite.service, _ = orm.NewPackageService(db, suite.gameService)
}

func (suite *packageServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}

	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *packageServiceTestSuite) TestPackages() {
	should := require.New(suite.T())

	gameA, err := suite.gameService.Create(suite.userId, suite.vendorId, "GameA")
	should.Nil(err)
	should.Len(gameA.DefaultPackageID, 16)
	gameB, err := suite.gameService.Create(suite.userId, suite.vendorId, "GameB")
	should.Nil(err)
	should.Len(gameB.DefaultPackageID, 16)

	pkg, err := suite.service.Create(suite.vendorId, suite.userId, "Mega package", []uuid.UUID{gameA.ID, gameB.ID})
	should.Nil(err)
	should.Equal(2, len(pkg.Products))
	should.Equal(gameA.ID, pkg.Products[0].GetID())

	pkgEmpty, err := suite.service.Create(suite.vendorId, suite.userId, "Empty package", []uuid.UUID{})
	should.NotNil(err)
	should.Nil(pkgEmpty)

	total, list, err := suite.service.GetList(suite.userId, suite.vendorId, "", "-date", 0, 20, nil)
	should.Nil(err)
	should.Equal(3, total)
	should.Equal(3, len(list)) // includes 2 default game packages
	should.Equal("Mega package", list[0].Name.EN)

	total, list2, err := suite.service.GetList(suite.userId, suite.vendorId, "", "+name", 1, 1, nil)
	should.Nil(err)
	should.Equal(1, len(list2))
	should.Equal("GameB", list2[0].Name.EN)

	total, list3, err := suite.service.GetList(suite.userId, suite.vendorId, "", "-date", 0, 20, func(packageId uuid.UUID) (bool, error) {
		return packageId != pkg.ID, nil
	})
	should.Nil(err)
	should.Equal(2, total)
	should.Equal(2, len(list3))
	should.Equal("GameB", list3[0].Name.EN)
	should.Equal("GameA", list3[1].Name.EN)

	total, list5, err := suite.service.GetList(suite.userId, suite.vendorId, "", "-date", 1, 20, func(packageId uuid.UUID) (bool, error) {
		return packageId != pkg.ID, nil
	})
	should.Nil(err)
	should.Equal(2, total)
	should.Equal(1, len(list5))
	should.Equal("GameA", list5[0].Name.EN)

	gameC, err := suite.gameService.Create(suite.userId, suite.vendorId, "GameC")
	should.Nil(err)

	pkg, err = suite.service.AddProducts(pkg.ID, []uuid.UUID{gameC.ID})
	should.Nil(err)
	should.Equal(3, len(pkg.Products))
	should.Equal(gameA.ID, pkg.Products[0].GetID())
	should.Equal(gameB.ID, pkg.Products[1].GetID())
	should.Equal(gameC.ID, pkg.Products[2].GetID())

	pkg.Name = utils.LocalizedString{EN: "Saved package"}
	pkg.Discount = 12
	pkgC, err := suite.service.Update(pkg)
	should.Nil(err)
	should.Equal("Saved package", pkgC.Name.EN)
	should.Equal(suite.userId, pkg.CreatorID)
	should.Equal(12, int(pkg.Discount))

	pkgG, err := suite.service.Get(pkg.ID)
	should.Nil(err)
	should.Equal("Saved package", pkgG.Name.EN)
	should.Equal(suite.userId, pkgG.CreatorID)
	should.Equal(12, int(pkgG.Discount))
	should.Equal(3, len(pkgG.Products))
	should.Equal(gameA.ID, pkgG.Products[0].GetID())
	should.Equal(gameB.ID, pkgG.Products[1].GetID())
	should.Equal(gameC.ID, pkgG.Products[2].GetID())

	pkgD, err := suite.service.RemoveProducts(pkgG.ID, []uuid.UUID{pkgG.Products[0].GetID()})
	should.Nil(err)
	should.Equal(pkg.ID, pkgD.ID)
	should.Equal(2, len(pkgD.Products))
	should.Equal(gameB.ID, pkgD.Products[0].GetID())
	should.Equal(gameC.ID, pkgD.Products[1].GetID())

	err = suite.service.Remove(pkg.ID)
	should.Nil(err)

	err = suite.service.Remove(gameB.DefaultPackageID)
	should.NotNil(err, "Try to remove default package")

	err = suite.service.Remove(pkg.ID)
	should.NotNil(err, "Package already removed")

	pkgX, err := suite.service.Update(pkg)
	should.Nil(pkgX)
	should.NotNil(err, "Nothing to update, package is deleted")

	pkgX, err = suite.service.AddProducts(pkg.ID, []uuid.UUID{gameA.ID})
	should.Nil(pkgX)
	should.NotNil(err, "Nothing to add, package is deleted")

	pkgX, err = suite.service.RemoveProducts(pkg.ID, []uuid.UUID{gameC.ID})
	should.Nil(pkgX)
	should.NotNil(err, "Nothing to remove, package is deleted")

	pkgX, err = suite.service.Get(pkg.ID)
	should.Nil(pkgX)
	should.NotNil(err, "Nothing to read, package is deleted")
}
