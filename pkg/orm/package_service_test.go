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
		ID:       uuid.NewV4().String(),
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

func (suite *packageServiceTestSuite) TestCreatePackage() {
	should := require.New(suite.T())

	gameA, err := suite.gameService.Create(suite.userId, suite.vendorId, "GameA")
	should.Nil(err)
	should.Len(gameA.DefPackageID, 16)
	gameB, err := suite.gameService.Create(suite.userId, suite.vendorId, "GameB")
	should.Nil(err)
	should.Len(gameB.DefPackageID, 16)

	pkg, err := suite.service.Create(suite.vendorId, "Mega package", []uuid.UUID{gameA.ID, gameB.ID})
	should.Nil(err)
	should.Equal(2, len(pkg.Products))
	should.Equal(gameA.ID, pkg.Products[0].GetID())

	list, err := suite.service.GetList(suite.vendorId, "", "-date", 0, 20)
	should.Nil(err)
	should.Equal(3, len(list)) // includes 2 default game packages
	should.Equal("Mega package", list[0].Name)
}
