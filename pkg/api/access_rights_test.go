package api

import (
	"fmt"
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/ProtocolONE/rbac"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"strings"
	"testing"
)

type AccessRightsTestSuite struct {
	suite.Suite
	db            *orm.Database
	echo          *echo.Echo
	service       model.MembershipService
	enforcer      *rbac.Enforcer
	currentUser   string
	Router        *echo.Group
	ownerProvider model.OwnerProvider
}

func Test_AccessRightsTestSuite(t *testing.T) {
	assert.NotPanics(t, func() {
		suite.Run(t, new(AccessRightsTestSuite))
	}, "")
}

const shouldHaveAccessFormat string = "%s should have access to %s %s"
const shouldHaveNotAccessFormat string = "%s should NOT have access to %s %s"

func (suite *AccessRightsTestSuite) SetupTest() {
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	if err := db.Init(); err != nil {
		suite.T().Log(err)
	}

	echoObj := echo.New()
	echoObj.Validator = &QilinValidator{validator: validator.New()}
	echoObj.HTTPErrorHandler = func(e error, context echo.Context) {
		QilinErrorHandler(e, context, true)
	}

	enforcer := rbac.NewEnforcer()
	ownerProvider := orm.NewOwnerProvider(db)
	membership := orm.NewMembershipService(db, ownerProvider, enforcer)
	err = membership.Init()
	if err != nil {
		suite.FailNow("Membership fail", "%v", err)
	}

	echoObj.Use(rbac_echo.NewAppContextMiddleware(ownerProvider, enforcer))
	echoObj.Use(suite.localAuth())

	suite.Router = echoObj.Group("/api/v1")
	suite.db = db
	suite.echo = echoObj
	suite.service = membership
	suite.ownerProvider = ownerProvider
	suite.enforcer = enforcer

	if err := suite.InitRoutes(); err != nil {
		suite.T().FailNow()
	}
}

func (suite *AccessRightsTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (s *AccessRightsTestSuite) InitRoutes() error {
	s.T().Helper()

	userService, err := orm.NewUserService(s.db, nil)
	if err != nil {
		return err
	}

	mediaService, err := orm.NewMediaService(s.db)
	if err != nil {
		return err
	}
	if _, err := InitMediaRouter(s.Router, mediaService); err != nil {
		return err
	}

	priceService, err := orm.NewPriceService(s.db)
	if err != nil {
		return err
	}
	if _, err := InitPriceRouter(s.Router, priceService); err != nil {
		return err
	}

	ratingService, err := orm.NewRatingService(s.db)
	if err != nil {
		return err
	}
	if _, err := InitRatingsRouter(s.Router, ratingService); err != nil {
		return err
	}

	discountService, err := orm.NewDiscountService(s.db)
	if err != nil {
		return err
	}
	if _, err := InitDiscountsRouter(s.Router, discountService); err != nil {
		return err
	}

	clientOnboarding, err := orm.NewOnboardingService(s.db)
	if err != nil {
		return err
	}
	notificationServ, err := orm.NewNotificationService(s.db, nil, "")
	if err != nil {
		return err
	}

	if _, err := InitClientOnboardingRouter(s.Router, clientOnboarding, notificationServ); err != nil {
		return err
	}

	gameService, err := orm.NewGameService(s.db)
	if err != nil {
		return err
	}

	vendorService, err := orm.NewVendorService(s.db, s.service)
	if err != nil {
		return err
	}

	membershipService := orm.NewMembershipService(s.db, s.ownerProvider, s.enforcer)
	if err := membershipService.Init(); err != nil {
		return err
	}

	if _, err := InitClientMembershipRouter(s.Router, membershipService); err != nil {
		return err
	}

	if _, err := InitRoutes(s.Router, gameService, userService); err != nil {
		return err
	}

	if err := InitVendorRoutes(s.Router, vendorService, userService); err != nil {
		return err
	}
	return nil
}

func (suite *AccessRightsTestSuite) TestRoutes() {
	shouldBe := require.New(suite.T())

	testCases := suite.generateTestCases()

	owner := suite.createUser()
	vendor := suite.createVendor(owner)
	gameId := suite.createGame(vendor, owner).String()

	notApprovedOwner := suite.createUser()
	vendorForNotApprovedOwner := suite.createVendor(notApprovedOwner)
	gameForNotApprovedOwner := suite.createGame(vendorForNotApprovedOwner, notApprovedOwner)
	messageForNotApprovedOwner := suite.createMessage(vendorForNotApprovedOwner, notApprovedOwner)

	admin := suite.createUser()
	globalAdmin := suite.createUser()
	messageId := suite.createMessage(vendor, admin)

	shouldBe.Nil(suite.service.AddRoleToUserInGame(vendor, admin, gameId, "admin"))
	shouldBe.Nil(suite.service.AddRoleToUserInGame(vendor, globalAdmin, "*", "admin"))

	anotherOwner := suite.createUser()
	anotherVendor := suite.createVendor(anotherOwner)
	anotherGame := suite.createGame(anotherVendor, anotherOwner).String()

	vendorId = vendor.String()
	superAdmin := suite.createUser()
	suite.enforcer.AddRole(rbac.Role{Role: "super_admin", User: superAdmin, Domain: "vendor"})

	testUser := suite.createUser()
	roles := []string{"admin", "support"}

	shouldBe.True(suite.enforcer.AddRole(rbac.Role{Role: model.NotApproved, User: notApprovedOwner, Domain: "vendor"}))

	for key, values := range testCases {
		url := format(key.url, vendorId, gameId, messageId)
		method := key.method
		body := key.body

		suite.checkAccess("owner", method, url, body, owner, true)
		suite.checkAccess("anotherOwner", method, url, body, anotherOwner, contains(values, model.AnyRole))
		suite.checkAccess("superAdmin", method, url, body, superAdmin, true)

		urlUnapproved := format(key.url, vendorForNotApprovedOwner.String(), gameForNotApprovedOwner.String(), messageForNotApprovedOwner)
		suite.checkAccess("notApprovedOwner", method, urlUnapproved, body, notApprovedOwner, contains(values, model.NotApproved) || contains(values, model.AnyRole))

		for _, role := range roles {
			accept := contains(values, role) || contains(values, model.AnyRole)

			// 1. Approved owner should pass
			// 2. Another approved owner should not pass
			// 3. Super-admin should pass
			// 4. User with role X in vendor context should pass (Global role)
			// 5. User with role X in vendor context and game restriction should pass
			// 6. User with role X in vendor context and game restriction should not pass for another game
			// 7. User with role Y should not pass to action with resource needed another role

			shouldBe.Nil(suite.service.AddRoleToUserInGame(vendor, testUser, "*", role))
			suite.checkAccess(role, method, url, body, testUser, accept)
			shouldBe.Nil(suite.service.RemoveRoleToUserInGame(vendor, testUser, "*", role))

			shouldBe.Nil(suite.service.AddRoleToUserInGame(vendor, testUser, gameId, role))
			suite.checkAccess(role, method, url, body, testUser, accept)
			shouldBe.Nil(suite.service.RemoveRoleToUserInGame(vendor, testUser, gameId, role))

			shouldBe.Nil(suite.service.AddRoleToUserInGame(anotherVendor, testUser, anotherGame, role))
			suite.checkAccess(role, method, url, body, testUser, contains(values, model.AnyRole))
			shouldBe.Nil(suite.service.RemoveRoleToUserInGame(anotherVendor, testUser, anotherGame, role))
		}
	}
}

func contains(arr []string, s string) bool {
	for _, item := range arr {
		if item == s {
			return true
		}
	}
	return false
}

func format(s, vendorId, gameId, messageId string) string {
	url := strings.Replace(s, "%vendor_id", vendorId, 1)
	url = strings.Replace(url, "%game_id", gameId, 1)
	url = strings.Replace(url, "%message_id", messageId, 1)
	return url
}

func (suite *AccessRightsTestSuite) generateTestCases() map[struct {
	method string
	url    string
	body   string
}][]string {
	suite.T().Helper()

	return map[struct {
		method string
		url    string
		body   string
	}][]string{
		{http.MethodGet, "/api/v1/vendors", ""}:            {model.AnyRole},
		{http.MethodPost, "/api/v1/vendors", ""}:           {model.AnyRole},
		{http.MethodGet, "/api/v1/vendors/%vendor_id", ""}: {model.Admin, model.Manager, model.Support, model.Developer, model.Accountant, model.Store, model.Publisher, model.NotApproved},
		{http.MethodPut, "/api/v1/vendors/%vendor_id", ""}: {model.Admin, model.NotApproved},

		{http.MethodGet, "/api/v1/vendors/%vendor_id/games", ""}:  {model.Admin, model.Support},
		{http.MethodPost, "/api/v1/vendors/%vendor_id/games", ""}: {model.Admin},

		{http.MethodGet, "/api/v1/vendors/%vendor_id/documents", ""}:          {model.Admin, model.NotApproved},
		{http.MethodPut, "/api/v1/vendors/%vendor_id/documents", ""}:          {model.Admin, model.NotApproved},
		{http.MethodPost, "/api/v1/vendors/%vendor_id/documents/reviews", ""}: {model.Admin, model.NotApproved},

		{http.MethodGet, "/api/v1/vendors/%vendor_id/messages", ""}:                  {model.Admin, model.NotApproved},
		{http.MethodGet, "/api/v1/vendors/%vendor_id/messages/short", ""}:            {model.Admin, model.NotApproved},
		{http.MethodGet, "/api/v1/vendors/%vendor_id/messages/%message_id", ""}:      {model.Admin, model.NotApproved},
		{http.MethodPut, "/api/v1/vendors/%vendor_id/messages/%message_id/read", ""}: {model.Admin, model.NotApproved},

		{http.MethodGet, "/api/v1/games/%game_id", ""}:              {model.Admin, model.Support},
		{http.MethodPut, "/api/v1/games/%game_id", ""}:              {model.Admin},
		{http.MethodGet, "/api/v1/games/%game_id/descriptions", ""}: {model.Admin, model.Support},
		{http.MethodPut, "/api/v1/games/%game_id/descriptions", ""}: {model.Admin},
		{http.MethodGet, "/api/v1/games/%game_id/ratings", ""}:      {model.Admin, model.Support},
		{http.MethodPut, "/api/v1/games/%game_id/ratings", ""}:      {model.Admin},
		{http.MethodGet, "/api/v1/games/%game_id/prices", ""}:       {model.Admin, model.Support},
		{http.MethodPut, "/api/v1/games/%game_id/prices", ""}:       {model.Admin},
		{http.MethodPut, "/api/v1/games/%game_id/prices/USD", ""}:   {model.Admin},
	}
}

func (suite *AccessRightsTestSuite) checkAccess(role string, method string, path string, body string, userId string, accepted bool) {
	suite.T().Helper()
	shouldBe := require.New(suite.T())

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	suite.currentUser = userId
	suite.echo.ServeHTTP(rec, req)

	testName := ""
	if accepted {
		testName = fmt.Sprintf(shouldHaveAccessFormat, role, method, path)
	} else {
		testName = fmt.Sprintf(shouldHaveNotAccessFormat, role, method, path)
	}

	errorMsg := fmt.Sprintf("[%s] Failed: %s %s `%s` for user `%s`. Result: `%s`", testName, method, path, body, userId, rec.Body.String())
	shouldBe.NotEqual(500, rec.Code, errorMsg)

	if accepted {
		shouldBe.NotEqual(403, rec.Code, errorMsg)
	} else {
		shouldBe.Equal(403, rec.Code, errorMsg)
	}
}

func (suite *AccessRightsTestSuite) localAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Set("user", &jwtverifier.UserInfo{UserID: suite.currentUser})
			return next(c)
		}
	}
}

// ******* UTILS ****** //

func (suite *AccessRightsTestSuite) createGame(vendorUuid uuid.UUID, uId string) uuid.UUID {
	gId := uuid.NewV4()

	require.Nil(suite.T(), suite.db.DB().Create(&model.Game{
		ID:           gId,
		VendorID:     vendorUuid,
		Title:        model.RandStringRunes(10),
		InternalName: model.RandStringRunes(10),
		CreatorID:    uId,
	}).Error)

	return gId
}

func (suite *AccessRightsTestSuite) createMessage(vendorId uuid.UUID, userId string) string {
	gId := uuid.NewV4()

	notification := &model.Notification{
		VendorID: vendorId,
		Title:    model.RandStringRunes(10),
		Message:  model.RandStringRunes(10),
		UserID:   userId,
	}
	notification.ID = gId
	require.Nil(suite.T(), suite.db.DB().Create(&notification).Error)

	return gId.String()
}

func (suite *AccessRightsTestSuite) createUser() string {
	uId := uuid.NewV4().String()
	require.Nil(suite.T(), suite.db.DB().Save(&model.User{
		ID:       uId,
		Nickname: model.RandStringRunes(10),
	}).Error)

	return uId
}

func (suite *AccessRightsTestSuite) createDocuments(id uuid.UUID) {
	docId := uuid.NewV4()
	doc := model.DocumentsInfo{
		VendorID: id,
	}
	doc.ID = docId
	require.Nil(suite.T(), suite.db.DB().Save(&doc).Error)
}

func (suite *AccessRightsTestSuite) createVendor(uId string) uuid.UUID {
	vendorUuid := uuid.NewV4()
	require.Nil(suite.T(), suite.db.DB().Save(&model.Vendor{
		ID:              vendorUuid,
		Name:            model.RandStringRunes(10),
		Domain3:         model.RandStringRunes(10),
		Email:           fmt.Sprintf("%s@test.com", model.RandStringRunes(6)),
		HowManyProducts: "+10",
		ManagerID:       uId,
		Users:           []model.User{{ID: uId}},
	}).Error)

	suite.createDocuments(vendorUuid)

	return vendorUuid
}
