package orm_test

import (
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/ProtocolONE/rbac"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MemershipServiceTestSuite struct {
	suite.Suite
	db      *orm.Database
	service model.MembershipService
}

func Test_MembershipService(t *testing.T) {
	suite.Run(t, new(MemershipServiceTestSuite))
}

var ownerId uuid.UUID

func (suite *MemershipServiceTestSuite) SetupTest() {
	shouldBe := require.New(suite.T())

	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	shouldBe.Nil(db.DropAllTables())
	shouldBe.Nil(db.Init())

	enf := rbac.NewEnforcer()

	gService, _ := orm.NewGameService(db)
	vService, _ := orm.NewVendorService(db)

	suite.db = db
	suite.service = orm.NewMembershipService(db, gService, vService, enf)
	shouldBe.Nil(suite.service.Init())

	ownerId := uuid.NewV4()
	userId2 := uuid.NewV4()
	userId3 := uuid.NewV4()

	shouldBe.Nil(db.DB().Create(&model.User{Email: "owner@example.com", ID: ownerId.String(), FullName: "Owner Test", Login: "owner", Password: "test"}).Error)
	shouldBe.Nil(db.DB().Create(&model.User{Email: "admin@example.com", ID: userId2.String(), FullName: "Admin Test", Login: "admin", Password: "test"}).Error)
	shouldBe.Nil(db.DB().Create(&model.User{Email: "support@example.com", ID: userId3.String(), FullName: "Support Test", Login: "support", Password: "test"}).Error)

	shouldBe.Nil(db.DB().Create(&model.Vendor{Name: "Test Vendor", ID: uuid.FromStringOrNil(vendorId), Email: "WTF@example.com", Domain3: "somedomain", HowManyProducts: "0", ManagerID: ownerId.String()}).Error)

	shouldBe.True(enf.AddRole(rbac.Role{Role: "admin", User: userId2.String(), Domain: "vendor", RestrictedResourceId: nil, Owner: ownerId.String()}))
	shouldBe.True(enf.AddRole(rbac.Role{Role: "support", User: userId3.String(), Domain: "vendor", RestrictedResourceId: nil, Owner: ownerId.String()}))
}

func (suite *MemershipServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *MemershipServiceTestSuite) TestAddRoleToUser() {
	shouldBe := require.New(suite.T())
	vId := uuid.FromStringOrNil(vendorId)

	userId3 := uuid.NewV4().String()

	shouldBe.Nil(suite.db.DB().Create(&model.User{Email: "new_admin@example.com", ID: userId3, FullName: "Admin New Test", Login: "new_admin", Password: "test"}).Error)

	err := suite.service.AddRoleToUserInGame(vId, userId3, "", model.Admin)
	shouldBe.Nil(err)

	err = suite.service.AddRoleToUserInGame(vId, userId3, "", model.Admin)
	shouldBe.NotNil(err)

	gameId := uuid.FromStringOrNil(Id)
	shouldBe.Nil(suite.db.DB().Create(&model.Game{ID: gameId, InternalName: "Test Internal Name", VendorID: vId, Title: "Test title"}).Error)

	err = suite.service.AddRoleToUserInGame(vId, userId3, gameId.String(), model.Support)
	shouldBe.Nil(err)

	err = suite.service.AddRoleToUserInGame(vId, userId3, uuid.NewV4().String(), model.Support)
	shouldBe.NotNil(err)

	err = suite.service.AddRoleToUserInGame(vId, uuid.NewV4().String(), gameId.String(), model.Support)
	shouldBe.NotNil(err)

	err = suite.service.AddRoleToUserInGame(uuid.NewV4(), userId3, gameId.String(), model.Support)
	shouldBe.NotNil(err)

	users, err := suite.service.GetUsers(vId)
	shouldBe.Nil(err)
	shouldBe.Equal(3, len(users))

	for _, user := range users {
		if user.Email != "new_admin@example.com" {
			continue
		}

		shouldBe.Equal(9, len(user.Roles))
		for _, role := range user.Roles {
			shouldBe.NotEmpty(role.Resource.Meta.InternalName)
		}
	}
}

func (suite *MemershipServiceTestSuite) TestGetUsers() {
	shouldBe := require.New(suite.T())
	vId := uuid.FromStringOrNil(vendorId)
	userRoles, err := suite.service.GetUsers(vId)
	shouldBe.Nil(err)
	shouldBe.NotNil(userRoles)
	shouldBe.Equal(2, len(userRoles))
	for _, u := range userRoles {
		shouldBe.NotEmpty(u.Roles)
		for _, r := range u.Roles {
			shouldBe.NotEmpty(r.Domain)
			shouldBe.NotEmpty(r.Role)
			shouldBe.NotEmpty(r.Resource)
		}
		shouldBe.NotEmpty(u.Email)
		shouldBe.NotEmpty(u.Name)
	}
}
