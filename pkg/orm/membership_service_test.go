package orm_test

import (
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/shersh/rbac"
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

	suite.db = db
	suite.service = orm.NewMembershipService(db, enf)
	shouldBe.Nil(suite.service.Init())

	ownerId := uuid.NewV4()
	userId2 := uuid.NewV4()
	userId3 := uuid.NewV4()

	externalId1 := RandStringRunes(5)
	externalId2 := RandStringRunes(5)
	externalId3 := RandStringRunes(5)

	shouldBe.Nil(db.DB().Create(&model.User{ExternalID: externalId1, Email: "owner@example.com", ID: ownerId, FullName: "Owner Test", Login: "owner", Password: "test"}).Error)
	shouldBe.Nil(db.DB().Create(&model.User{ExternalID: externalId2, Email: "admin@example.com", ID: userId2, FullName: "Admin Test", Login: "admin", Password: "test"}).Error)
	shouldBe.Nil(db.DB().Create(&model.User{ExternalID: externalId3, Email: "support@example.com", ID: userId3, FullName: "Support Test", Login: "support", Password: "test"}).Error)

	shouldBe.Nil(db.DB().Create(&model.Vendor{Name: "Test Vendor", ID: uuid.FromStringOrNil(vendorId), Email: "WTF@example.com", Domain3: "somedomain", HowManyProducts: "0", ManagerID: ownerId}).Error)

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

	userId3 := uuid.NewV4()

	externalId3 := RandStringRunes(5)

	shouldBe.Nil(suite.db.DB().Create(&model.User{ExternalID: externalId3, Email: "new_admin@example.com", ID: userId3, FullName: "Admin New Test", Login: "new_admin", Password: "test"}).Error)

	err := suite.service.AddRoleToUserInGame(vId, userId3, uuid.Nil, model.Admin)
	shouldBe.Nil(err)

	err = suite.service.AddRoleToUserInGame(vId, userId3, uuid.Nil, model.Admin)
	shouldBe.NotNil(err)

	gameId := uuid.FromStringOrNil(Id)
	shouldBe.Nil(suite.db.DB().Create(&model.Game{ID: gameId, InternalName: "Test Internal Name", VendorID: vId, Title: "Test title"}).Error)

	err = suite.service.AddRoleToUserInGame(vId, userId3, gameId, model.Support)
	shouldBe.Nil(err)

	err = suite.service.AddRoleToUserInGame(vId, userId3, uuid.NewV4(), model.Support)
	shouldBe.NotNil(err)

	err = suite.service.AddRoleToUserInGame(vId, uuid.NewV4(), gameId, model.Support)
	shouldBe.NotNil(err)

	err = suite.service.AddRoleToUserInGame(uuid.NewV4(), userId3, gameId, model.Support)
	shouldBe.NotNil(err)

	users, err := suite.service.GetUsers(vId)
	shouldBe.Nil(err)
	shouldBe.Equal(3, len(users))

	for _, user := range users {
		if user.Email != "new_admin@example.com" {
			continue
		}

		shouldBe.Equal(2, len(user.Roles))
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

