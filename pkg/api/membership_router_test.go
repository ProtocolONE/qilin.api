package api

import (
	"fmt"
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/ProtocolONE/rbac"
	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/mock"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"qilin-api/pkg/utils"
	"strings"
	"testing"
)

type MembershipRouterTestSuite struct {
	suite.Suite
	db      *orm.Database
	service model.MembershipService
	echo    *echo.Echo
	router  *MembershipRouter
}

var adminId string
var ownerId string

func Test_MembershipRouter(t *testing.T) {
	suite.Run(t, new(MembershipRouterTestSuite))
}

func (suite *MembershipRouterTestSuite) SetupTest() {
	shouldBe := require.New(suite.T())
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

	ownerId = uuid.NewV4().String()
	adminId = uuid.NewV4().String()
	shouldBe.Nil(db.DB().Create(&model.User{
		ID:       adminId,
		FullName: "Admin Tester",
	}).Error)

	shouldBe.Nil(db.DB().Create(&model.User{
		ID:       ownerId,
		FullName: "Owner",
	}).Error)

	shouldBe.Nil(db.DB().Create(&model.Vendor{
		ID:        uuid.FromStringOrNil(vendorId),
		ManagerID: ownerId,
		Domain3:   "qwe",
		Email:     "asd@asd.as",
		Name:      "xcxc",
	}).Error)

	shouldBe.Nil(db.DB().Create(&model.Game{
		ID:        uuid.FromStringOrNil(TestID),
		VendorID:  uuid.FromStringOrNil(vendorId),
		CreatorID: ownerId,
	}).Error)

	e := echo.New()
	v := validator.New()
	e.Validator = &QilinValidator{validator: v}
	shouldBe.NoError(utils.RegisterCustomValidations(v))
	enf := rbac.NewEnforcer()
	ownerProvider := orm.NewOwnerProvider(db)

	service := orm.NewMembershipService(db, ownerProvider, enf, mock.NewMailer(), "127.0.0.1")
	shouldBe.Nil(service.Init())
	enf.AddRole(rbac.Role{Role: "admin", User: adminId, Domain: "vendor", Owner: ownerId, RestrictedResourceId: []string{"*"}})

	router, err := InitClientMembershipRouter(e.Group("/api/v1"), service)
	shouldBe.Nil(err)

	suite.db = db
	suite.service = service
	suite.router = router
	suite.echo = e
}

func (suite *MembershipRouterTestSuite) TestGetMemberships() {
	shouldBe := require.New(suite.T())

	testCases := []struct {
		testName string
		vendorId string
		success  bool
		code     int
	}{
		{testName: "Normal", vendorId: vendorId, code: 200, success: true},
		{testName: "Not found", vendorId: uuid.NewV4().String(), code: 404, success: false},
		{testName: "Bad vendor", vendorId: "BAD_ID", code: 400, success: false},
	}

	for _, testCase := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := suite.echo.NewContext(req, rec)
		c.SetPath("/api/v1/vendors/:vendorId/memberships")
		c.SetParamNames("vendorId")
		c.SetParamValues(testCase.vendorId)

		res := suite.router.getUsers(c)
		msg := fmt.Sprintf("[%s] %v. %v", testCase.testName, testCase, res)
		if testCase.success == false {
			shouldBe.NotNil(res, msg)
			he := res.(*orm.ServiceError)
			shouldBe.Equal(testCase.code, he.Code, msg)
		} else {
			shouldBe.Nil(res, msg)
			shouldBe.Equal(testCase.code, rec.Code, msg)
		}
	}
}

func (suite *MembershipRouterTestSuite) TestGetMembership() {
	shouldBe := require.New(suite.T())

	testCases := []struct {
		testName string
		vendorId string
		success  bool
		code     int
		userId   string
	}{
		{testName: "Normal", vendorId: vendorId, code: 200, userId: adminId, success: true},
		{testName: "Not found vendor", vendorId: uuid.NewV4().String(), userId: adminId, code: 404, success: false},
		{testName: "Bad vendor", vendorId: "BAD_ID", code: 400, userId: adminId, success: false},
		{testName: "Not found user", vendorId: vendorId, userId: uuid.NewV4().String(), code: 404, success: false},
	}

	for _, testCase := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := suite.echo.NewContext(req, rec)
		c.SetPath("/api/v1/vendors/:vendorId/memberships/:userId")
		c.SetParamNames("vendorId", "userId")
		c.SetParamValues(testCase.vendorId, testCase.userId)

		res := suite.router.getUser(c)
		msg := fmt.Sprintf("[%s] %v. %v", testCase.testName, testCase, res)
		if testCase.success == false {
			shouldBe.NotNil(res, msg)
			he := res.(*orm.ServiceError)
			shouldBe.Equal(testCase.code, he.Code, msg)
		} else {
			shouldBe.Nil(res, msg)
			shouldBe.Equal(testCase.code, rec.Code, msg)
		}
	}
}

func (suite *MembershipRouterTestSuite) TestPutMembership() {
	shouldBe := require.New(suite.T())
	changeRoles := fmt.Sprintf(`{"added":[{"id":"%s","roles":["manager"]}]}`, TestID)
	notFoundChangeRoles := fmt.Sprintf(`{"added":[{"id":"%s","roles":["manager"]}]}`, uuid.NewV4())
	badChangeRoles := `<"added":[{"id":"test","roles":["manager"]}]>`
	testCases := []struct {
		testName string
		vendorId string
		success  bool
		code     int
		body     string
		userId   string
	}{
		{testName: "Normal", vendorId: vendorId, code: 200, userId: adminId, body: changeRoles, success: true},
		{testName: "Not found vendor", vendorId: uuid.NewV4().String(), userId: adminId, body: changeRoles, code: 404, success: false},
		{testName: "Bad vendor", vendorId: "BAD_ID", code: 400, userId: adminId, body: changeRoles, success: false},
		{testName: "Not found user", vendorId: vendorId, userId: uuid.NewV4().String(), body: changeRoles, code: 404, success: false},
		{testName: "Bad request", vendorId: vendorId, userId: adminId, body: badChangeRoles, code: 400, success: false},
		{testName: "Game not found", vendorId: vendorId, userId: adminId, body: notFoundChangeRoles, code: 404, success: false},
	}

	for _, testCase := range testCases {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(testCase.body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := suite.echo.NewContext(req, rec)
		c.SetPath("/api/v1/vendors/:vendorId/memberships/:userId")
		c.SetParamNames("vendorId", "userId")
		c.SetParamValues(testCase.vendorId, testCase.userId)

		res := suite.router.changeUserRoles(c)
		msg := fmt.Sprintf("[%s] %v. %v", testCase.testName, testCase, res)
		if testCase.success == false {
			shouldBe.NotNil(res, msg)
			he := res.(*orm.ServiceError)
			shouldBe.Equal(testCase.code, he.Code, msg)
		} else {
			shouldBe.Nil(res, msg)
			shouldBe.Equal(testCase.code, rec.Code, msg)
		}
	}
}

func (suite *MembershipRouterTestSuite) TestGetPermissions() {
	shouldBe := require.New(suite.T())
	testCases := []struct {
		testName string
		vendorId string
		success  bool
		code     int
		userId   string
	}{
		{testName: "Normal", vendorId: vendorId, code: 200, userId: adminId, success: true},
		{testName: "Not found vendor", vendorId: uuid.NewV4().String(), userId: adminId, code: 404, success: false},
		{testName: "Bad vendor", vendorId: "BAD_ID", code: 400, userId: adminId, success: false},
		{testName: "Not found user", vendorId: vendorId, userId: uuid.NewV4().String(), code: 404, success: false},
	}

	for _, testCase := range testCases {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := suite.echo.NewContext(req, rec)
		c.SetPath("/api/v1/vendors/:vendorId/memberships/:userId")
		c.SetParamNames("vendorId", "userId")
		c.SetParamValues(testCase.vendorId, testCase.userId)
		c.Set(context.TokenKey, &jwtverifier.UserInfo{UserID: testCase.userId})

		res := suite.router.getUserPermissions(c)
		msg := fmt.Sprintf("[%s] %v. %v", testCase.testName, testCase, res)
		if testCase.success == false {
			shouldBe.NotNil(res, msg)
			he := res.(*orm.ServiceError)
			shouldBe.Equal(testCase.code, he.Code, msg)
		} else {
			shouldBe.Nil(res, msg)
			shouldBe.Equal(testCase.code, rec.Code, msg)
			shouldBe.NotEmpty(rec.Body, msg)
		}
	}
}

func (suite *MembershipRouterTestSuite) TestAcceptInvite() {
	shouldBe := require.New(suite.T())

	invitedUser := model.User{}
	invitedUser.ID = uuid.NewV4().String()
	invitedUser.Email = "invite@test.com"
	shouldBe.Nil(suite.db.DB().Create(&invitedUser).Error)

	invite := &model.Invite{}
	invite.ID = uuid.NewV4()
	invite.Email = invitedUser.Email
	invite.Roles = model.Roles{
		model.Role{Role: model.Support, Resource: model.ResourceRole{Id: TestID, Domain: "vendor"}},
	}
	invite.VendorId = uuid.FromStringOrNil(vendorId)

	shouldBe.Nil(suite.db.DB().Create(invite).Error)
	inviteId := invite.ID.String()

	testCases := []struct {
		testName string
		vendorId string
		inviteId string
		userId   string
		success  bool
		code     int
	}{
		{testName: "Invite for another user", vendorId: vendorId, code: 403, inviteId: inviteId, userId: ownerId, success: false},
		{testName: "Normal", vendorId: vendorId, code: 200, inviteId: inviteId, userId: invitedUser.ID, success: true},
		{testName: "Already accepted", vendorId: vendorId, code: 409, inviteId: inviteId, userId: invitedUser.ID, success: false},
		{testName: "Bad vendor id", vendorId: "SOME_BAD_UUID", code: 400, inviteId: inviteId, userId: invitedUser.ID, success: false},
		{testName: "Bad invite id", vendorId: vendorId, code: 400, inviteId: "SOME_BAD_UUID", userId: invitedUser.ID, success: false},
		{testName: "Unknown user", vendorId: vendorId, code: 404, inviteId: inviteId, userId: uuid.NewV4().String(), success: false},
		{testName: "Not existing invite", vendorId: vendorId, code: 404, inviteId: uuid.NewV4().String(), userId: invitedUser.ID, success: false},
	}

	for _, testCase := range testCases {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := suite.echo.NewContext(req, rec)
		c.Set("user", &jwtverifier.UserInfo{UserID: testCase.userId})
		c.SetPath("/api/v1/vendors/:vendorId/memberships/invites/:inviteId")
		c.SetParamNames("vendorId", "inviteId")
		c.SetParamValues(testCase.vendorId, testCase.inviteId)

		res := suite.router.acceptInvite(c)
		msg := fmt.Sprintf("[%s] %v. %v", testCase.testName, testCase, res)
		if testCase.success == false {
			shouldBe.NotNil(res, msg)
			he := res.(*orm.ServiceError)
			shouldBe.Equal(testCase.code, he.Code, msg)
		} else {
			shouldBe.Nil(res, msg)
			shouldBe.Equal(testCase.code, rec.Code, msg)
		}
	}
}

func (suite *MembershipRouterTestSuite) TestSendInvite() {
	shouldBe := require.New(suite.T())
	normalBody := `{"email":"tester@protocol.one", "roles":[{"role":"manager","resource":{"id":"*","domain":"vendor"}}]}`
	badRoleBody := `{"email":"tester@protocol.one", "roles":[{"role":"admin","resource":{"id":"*","domain":"vendor"}}]}`
	normalBodyWithGameId := fmt.Sprintf(`{"email":"tester2@protocol.one", "roles":[{"role":"manager","resource":{"id":"%s","domain":"vendor"}}]}`, TestID)
	bodyWithUnknownGameId := fmt.Sprintf(`{"email":"tester3@protocol.one", "roles":[{"role":"manager","resource":{"id":"%s","domain":"vendor"}}]}`, uuid.NewV4())
	noEmailBody := `{"roles":[{"role":"manager","resource":{"id":"*"}}]}`
	badBody := `<"email":"roman.golenok@protocol.one", "roles":[{"role":"manager","resource":{"id":"*"}}]>`
	testCases := []struct {
		testName string
		vendorId string
		success  bool
		code     int
		body     string
	}{
		{testName: "Normal", vendorId: vendorId, code: 201, body: normalBody, success: true},
		{testName: "Normal with specified game", vendorId: vendorId, code: 201, body: normalBodyWithGameId, success: true},
		{testName: "Unknown game", vendorId: vendorId, code: 422, body: bodyWithUnknownGameId, success: false},
		{testName: "Duplicate", vendorId: vendorId, code: 409, body: normalBody, success: false},
		{testName: "Bad body", vendorId: vendorId, code: 400, body: badBody, success: false},
		{testName: "Without email", vendorId: vendorId, code: 422, body: noEmailBody, success: false},
		{testName: "Unknown vendor", vendorId: uuid.NewV4().String(), code: 404, body: normalBody, success: false},
		{testName: "Bad vendor id", vendorId: "SOME_BAD_UUID", code: 400, body: noEmailBody, success: false},
		{testName: "Bad role ", vendorId: vendorId, code: 422, body: badRoleBody, success: false},
	}

	for _, testCase := range testCases {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := suite.echo.NewContext(req, rec)
		c.SetPath("/api/v1/vendors/:vendorId/memberships/invites")
		c.SetParamNames("vendorId")
		c.SetParamValues(testCase.vendorId)

		res := suite.router.sendInvite(c)
		msg := fmt.Sprintf("[%s] %v. %v", testCase.testName, testCase, res)
		if testCase.success == false {
			shouldBe.NotNil(res, msg)
			he := res.(*orm.ServiceError)
			shouldBe.Equal(testCase.code, he.Code, msg)
		} else {
			shouldBe.Nil(res, msg)
			shouldBe.Equal(testCase.code, rec.Code, msg)
			shouldBe.NotEmpty(rec.Body, msg)
		}
	}
}
