package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
	"qilin-api/pkg/test"
	"qilin-api/pkg/utils"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
)

type OnboardingClientRouterTestSuite struct {
	suite.Suite
	db      *orm.Database
	service *orm.OnboardingService
	echo    *echo.Echo
	router  *OnboardingClientRouter
}

var (
	emptyDocument                 = `{"company":{"name":"","alternativeName":"","country":"","region":"","zip":"","city":"","address":"","additionalAddress":"","registrationNumber":"","taxId":""},"contact":{"authorized":{"fullName":"","email":"","phone":"","position":""},"technical":{"fullName":"","email":"","phone":""}},"banking":{"currency":"","name":"","address":"","accountNumber":"","swift":"","details":""},"status":"draft"}`
	nonEmptyDocument              = `{"company":{"name":"TestName","alternativeName":"","country":"russia","region":"Moscow","zip":"098978","city":"Moscow","address":"Some address","additionalAddress":"Some add address","registrationNumber":"1232321312","taxId":"13122414"},"contact":{"authorized":{"fullName":"TestName","email":"test@email.com","phone":"+7123456789","position":"TestPosition"},"technical":{"fullName":"","email":"","phone":""}},"banking":{"currency":"USD","name":"Bank of Baroda","address":"string","accountNumber":"12345678901234567","swift":"QWERTY","details":"NoDetails"},"status":"draft"}`
	badDocumentNoName             = `{"company":{"country":"russia","region":"Moscow","zip":"098978","city":"Moscow","address":"Some address","additionalAddress":"Some add address","registrationNumber":"1232321312","taxId":"13122414"},"contact":{},"banking":{"currency":"USD","name":"Bank of Baroda","address":"string","accountNumber":"12345678901234567","swift":"QWERTY","details":"NoDetails"},"status":"draft"}`
	badDocumentNoRegion           = `{"company":{"name":"TestName","alternativeName":"","country":"russia","zip":"098978","city":"Moscow","address":"Some address","additionalAddress":"Some add address","registrationNumber":"1232321312","taxId":"13122414"},"contact":{"authorized":{"fullName":"test","position":"testposition","email":"email@enail.com","phone":"123123124"}},"banking":{"currency":"USD","name":"Bank of Baroda","address":"string","accountNumber":"12345678901234567","swift":"QWERTY","details":"NoDetails"},"status":"draft"}`
	badDocumentNoZip              = `{"company":{"name":"TestName","alternativeName":"","country":"russia","region":"Moscow","city":"Moscow","address":"Some address","additionalAddress":"Some add address","registrationNumber":"1232321312","taxId":"13122414"},"contact":{"authorized":{"fullName":"test","position":"testposition","email":"email@enail.com","phone":"123123124"}},"banking":{"currency":"USD","name":"Bank of Baroda","address":"string","accountNumber":"12345678901234567","swift":"QWERTY","details":"NoDetails"},"status":"draft"}`
	badDocumentNoContact          = `{"company":{"name":"TestName","alternativeName":"","country":"russia","region":"Moscow","zip":"098978","city":"Moscow","address":"Some address","additionalAddress":"Some add address","registrationNumber":"1232321312","taxId":"13122414"},"banking":{"currency":"USD","name":"Bank of Baroda","address":"string","accountNumber":"12345678901234567","swift":"QWERTY","details":"NoDetails"},"status":"draft"}`
	badDocumentWrongCurrency      = `{"company":{"name":"TestName","alternativeName":"","country":"russia","region":"Moscow","zip":"098978","city":"Moscow","address":"Some address","additionalAddress":"Some add address","registrationNumber":"1232321312","taxId":"13122414"},"contact":{"authorized":{"fullName":"test","position":"testposition","email":"email@enail.com","phone":"123123124"}},"banking":{"currency":"LOL","name":"Bank of Baroda","address":"string","accountNumber":"12345678901234567","swift":"QWERTY","details":"NoDetails"},"status":"draft"}`
	badDocumentEmptyContact       = `{"company":{"name":"TestName","alternativeName":"","country":"russia","region":"Moscow","zip":"098978","city":"Moscow","address":"Some address","additionalAddress":"Some add address","registrationNumber":"1232321312","taxId":"13122414"},"contact":{"authorized":{}},"banking":{"currency":"USD","name":"Bank of Baroda","address":"string","accountNumber":"12345678901234567","swift":"QWERTY","details":"NoDetails"},"status":"draft"}`
	badDocumentContactWithoutName = `{"company":{"name":"TestName","alternativeName":"","country":"russia","region":"Moscow","zip":"098978","city":"Moscow","address":"Some address","additionalAddress":"Some add address","registrationNumber":"1232321312","taxId":"13122414"},"contact":{"authorized":{"position":"testposition","email":"email@enail.com","phone":"123123124"}},"banking":{"currency":"USD","name":"Bank of Baroda","address":"string","accountNumber":"12345678901234567","swift":"QWERTY","details":"NoDetails"},"status":"draft"}`
)

func Test_OnboardingClientRouter(t *testing.T) {
	suite.Run(t, new(OnboardingClientRouterTestSuite))
}

func (suite *OnboardingClientRouterTestSuite) SetupTest() {
	should := require.New(suite.T())
	config, err := qilin_test.LoadTestConfig()
	should.Nil(err, "Unable to load config", "%v", err)
	db, err := orm.NewDatabase(&config.Database)
	should.Nil(err, "Unable to connect to database", "%v", err)

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

	id, _ := uuid.FromString(TestID)
	should.Nil(db.DB().Create(&model.Vendor{ID: id, Email: "example@example.ru", Name: "Test", Domain3: "test"}).Error, "Can't create vendor")

	e := echo.New()
	service, err := orm.NewOnboardingService(db)
	notifier, err := sys.NewNotifier(config.Notifier.ApiKey, config.Notifier.Host)
	should.Nil(err)
	notService, err := orm.NewNotificationService(db, notifier, config.Notifier.Secret)
	should.Nil(err)
	router, err := InitClientOnboardingRouter(e.Group("/api/v1"), service, notService)
	v := validator.New()
	assert.NoError(suite.T(), utils.RegisterCustomValidations(v))
	e.Validator = &QilinValidator{validator: v}

	suite.db = db
	suite.router = router
	suite.echo = e
}

func (suite *OnboardingClientRouterTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}

	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *OnboardingClientRouterTestSuite) TestGetDocument() {
	shouldBe := require.New(suite.T())
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getDocument(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.JSONEq(suite.T(), emptyDocument, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents")
	c.SetParamNames("vendorId")
	c.SetParamValues("XXX")

	err := suite.router.getDocument(c)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents")
	c.SetParamNames("vendorId")
	c.SetParamValues(uuid.NewV4().String())

	err = suite.router.getDocument(c)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)
}

func (suite *OnboardingClientRouterTestSuite) TestGetShouldReturnObject() {
	id, _ := uuid.FromString(TestID)
	info := model.DocumentsInfo{
		VendorID: id,
		Status:   model.StatusDraft,
		Contact: model.JSONB{
			"authorized": model.JSONB{
				"fullName": "TestName",
				"position": "TestPosition",
				"email":    "test@email.com",
				"phone":    "+7123456789",
			},
		},
		Banking: model.JSONB{
			"currency":      "USD",
			"name":          "Bank of Baroda",
			"address":       "string",
			"accountNumber": "12345678901234567",
			"swift":         "QWERTY",
			"details":       "NoDetails",
		},
		Company: model.JSONB{
			"name":               "TestName",
			"country":            "russia",
			"region":             "Moscow",
			"zip":                "098978",
			"city":               "Moscow",
			"address":            "Some address",
			"additionalAddress":  "Some add address",
			"registrationNumber": "1232321312",
			"taxId":              "13122414",
		},
	}
	info.ID = uuid.NewV4()

	err := suite.db.DB().Create(&info).Error
	assert.Nil(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(nonEmptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getDocument(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.JSONEq(suite.T(), nonEmptyDocument, rec.Body.String())
	}

	err = suite.db.DB().Delete(&info).Error
	assert.Nil(suite.T(), err)
}

func (suite *OnboardingClientRouterTestSuite) TestSendToReviewShouldReturnCreated() {
	id, _ := uuid.FromString(TestID)
	info := model.DocumentsInfo{
		VendorID: id,
		Status:   model.StatusDraft,
		Contact: model.JSONB{
			"authorized": model.JSONB{
				"fullName": "TestName",
				"position": "TestPosition",
				"email":    "test@email.com",
				"phone":    "+7123456789",
			},
		},
		Banking: model.JSONB{
			"currency":      "USD",
			"name":          "Bank of Baroda",
			"address":       "string",
			"accountNumber": "12345678901234567",
			"swift":         "QWERTY",
			"details":       "NoDetails",
		},
		Company: model.JSONB{
			"name":               "TestName",
			"country":            "russia",
			"region":             "Moscow",
			"zip":                "098978",
			"city":               "Moscow",
			"address":            "Some address",
			"additionalAddress":  "Some add address",
			"registrationNumber": "1232321312",
			"taxId":              "13122414",
		},
	}
	info.ID = uuid.NewV4()

	err := suite.db.DB().Create(&info).Error
	assert.Nil(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(nonEmptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents/reviews")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.sendToReview(c)) {
		assert.Equal(suite.T(), http.StatusCreated, rec.Code)
	}
}

func (suite *OnboardingClientRouterTestSuite) TestSendToReviewShouldReturnError() {
	id, _ := uuid.FromString(TestID)
	info := model.DocumentsInfo{
		VendorID: id,
		Status:   model.StatusOnReview,
		Contact: model.JSONB{
			"authorized": model.JSONB{
				"fullName": "TestName",
				"position": "TestPosition",
				"email":    "test@email.com",
				"phone":    "+7123456789",
			},
		},
		Banking: model.JSONB{
			"currency":      "USD",
			"name":          "Bank of Baroda",
			"address":       "string",
			"accountNumber": "12345678901234567",
			"swift":         "QWERTY",
			"details":       "NoDetails",
		},
		Company: model.JSONB{
			"name":               "TestName",
			"country":            "russia",
			"region":             "Moscow",
			"zip":                "098978",
			"city":               "Moscow",
			"address":            "Some address",
			"additionalAddress":  "Some add address",
			"registrationNumber": "1232321312",
			"taxId":              "13122414",
		},
	}
	info.ID = uuid.NewV4()

	err := suite.db.DB().Create(&info).Error
	assert.Nil(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(nonEmptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents/reviews")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	res := suite.router.sendToReview(c)
	assert.NotNil(suite.T(), res)
	he := res.(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *OnboardingClientRouterTestSuite) TestChangeShouldReturnErrorWhenNotDraft() {
	id, _ := uuid.FromString(TestID)
	info := model.DocumentsInfo{
		VendorID: id,
		Status:   model.StatusOnReview,
		Contact: model.JSONB{
			"authorized": model.JSONB{
				"fullName": "TestName",
				"position": "TestPosition",
				"email":    "test@email.com",
				"phone":    "+7123456789",
			},
		},
		Banking: model.JSONB{
			"currency":      "USD",
			"name":          "Bank of Baroda",
			"address":       "string",
			"accountNumber": "12345678901234567",
			"swift":         "QWERTY",
			"details":       "NoDetails",
		},
		Company: model.JSONB{
			"name":               "TestName",
			"country":            "russia",
			"region":             "Moscow",
			"zip":                "098978",
			"city":               "Moscow",
			"address":            "Some address",
			"additionalAddress":  "Some add address",
			"registrationNumber": "1232321312",
			"taxId":              "13122414",
		},
	}
	info.ID = uuid.NewV4()

	err := suite.db.DB().Create(&info).Error
	assert.Nil(suite.T(), err)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(nonEmptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	res := suite.router.changeDocument(c)
	assert.NotNil(suite.T(), res)
	he := res.(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *OnboardingClientRouterTestSuite) TestPutShouldReturnOK() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(nonEmptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	if assert.NoError(suite.T(), suite.router.changeDocument(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
	}
}

func (suite *OnboardingClientRouterTestSuite) TestPutShouldReturnNotFound() {
	//#1
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(nonEmptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents")
	c.SetParamNames("vendorId")
	c.SetParamValues(uuid.NewV4().String())

	res := suite.router.changeDocument(c)
	assert.NotNil(suite.T(), res)
	he := res.(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusNotFound, he.Code)
}

func (suite *OnboardingClientRouterTestSuite) TestPutShouldReturnBadRequest() {
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(nonEmptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents")
	c.SetParamNames("vendorId")
	c.SetParamValues("XXX")

	res := suite.router.changeDocument(c)
	assert.NotNil(suite.T(), res)
	he := res.(*orm.ServiceError)
	assert.Equal(suite.T(), http.StatusBadRequest, he.Code)
}

func (suite *OnboardingClientRouterTestSuite) TestPutShouldReturnError422() {
	tests := []struct {
		body string
		name string
	}{
		{body: badDocumentNoContact, name: "badDocumentNoContact"},
		{body: badDocumentNoName, name: "badDocumentNoName"},
		{body: badDocumentNoRegion, name: "badDocumentNoRegion"},
		{body: badDocumentContactWithoutName, name: "badDocumentContactWithoutName"},
		{body: badDocumentWrongCurrency, name: "badDocumentWrongCurrency"},
		{body: badDocumentNoContact, name: "badDocumentNoContact"},
		{body: badDocumentEmptyContact, name: "badDocumentEmptyContact"},
		{body: badDocumentNoZip, name: "badDocumentNoZip"},
		{body: "{}", name: "empty"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := suite.echo.NewContext(req, rec)
		c.SetPath("/api/v1/vendors/:vendorId/documents")
		c.SetParamNames("vendorId")
		c.SetParamValues(TestID)

		res := suite.router.changeDocument(c)
		assert.NotNil(suite.T(), res, "Should be error. %s", tt.name)
		assert.NotPanics(suite.T(), func() {
			he := res.(*orm.ServiceError)
			assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code, "Test case failed: `%s`. Error: %#v, With body %s", tt.name, he, tt.body)
		})
	}
}

func (suite *OnboardingClientRouterTestSuite) generateNotifications(id uuid.UUID) {
	should := require.New(suite.T())
	notification := &model.Notification{VendorID: id, Title: "Some title", Message: "ZZZ"}
	notification.ID = uuid.NewV4()
	notification.IsRead = false
	should.Nil(suite.db.DB().Create(notification).Error)

	notification = &model.Notification{VendorID: uuid.NewV4(), Title: "Some title", Message: "YYY"}
	notification.IsRead = false
	notification.ID = uuid.NewV4()
	should.Nil(suite.db.DB().Create(notification).Error)

	for i := 0; i < 100; i++ {
		notification = &model.Notification{VendorID: id, Title: fmt.Sprintf("Test title %d", i), Message: fmt.Sprintf("%d", i)}
		notification.ID = uuid.NewV4()
		notification.IsRead = true
		should.Nil(suite.db.DB().Create(notification).Error)
	}
}

func (suite *OnboardingClientRouterTestSuite) TestMarkAsRead() {
	should := require.New(suite.T())
	suite.generateNotifications(uuid.FromStringOrNil(TestID))

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages/:messageId/read")
	c.SetParamNames("vendorId", "messageId")
	c.SetParamValues("XXX", TestID)

	err := suite.router.markAsRead(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages/:messageId/read")
	c.SetParamNames("vendorId", "messageId")
	c.SetParamValues(TestID, "XXXX")

	err = suite.router.markAsRead(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages/:messageId/read")
	c.SetParamNames("vendorId", "messageId")
	c.SetParamValues(TestID, TestID)

	err = suite.router.markAsRead(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusNotFound, he.Code)
	}

	var notifications []model.Notification
	should.Nil(suite.db.DB().Model(model.Notification{}).Where("vendor_id = ?", TestID).Find(&notifications).Error)
	should.True(len(notifications) > 0)

	for _, n := range notifications {
		req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{}"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = suite.echo.NewContext(req, rec)
		c.SetPath("/api/v1/vendors/:vendorId/messages/:messageId/read")
		c.SetParamNames("vendorId", "messageId")
		c.SetParamValues(TestID, n.ID.String())

		err = suite.router.markAsRead(c)
		should.Nil(err)
		should.Equal(http.StatusOK, rec.Code)
	}
}

func (suite *OnboardingClientRouterTestSuite) TestGetNotifications() {
	should := require.New(suite.T())
	suite.generateNotifications(uuid.FromStringOrNil(TestID))

	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages")
	c.SetParamNames("vendorId")
	c.SetParamValues("XXX")

	err := suite.router.getNotifications(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages")
	c.SetParamNames("vendorId")
	c.SetParamValues(uuid.NewV4().String())

	err = suite.router.getNotifications(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusNotFound, he.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	err = suite.router.getNotifications(c)
	should.Nil(err)
	should.Equal(http.StatusOK, rec.Code)
	var notifications []NotificationDTO
	should.Nil(json.Unmarshal(rec.Body.Bytes(), &notifications))
	should.Equal(20, len(notifications))
	countStr := rec.Header().Get("X-Items-Count")
	count, err := strconv.Atoi(countStr)
	should.Nil(err)
	should.Equal(101, count)

	req = httptest.NewRequest(http.MethodGet, "/?limit=100&offset=0&sort=-title&query=Some", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	err = suite.router.getNotifications(c)
	should.Nil(err)
	should.Equal(http.StatusOK, rec.Code)
	should.Nil(json.Unmarshal(rec.Body.Bytes(), &notifications))
	should.Equal(1, len(notifications))
	countStr = rec.Header().Get("X-Items-Count")
	count, err = strconv.Atoi(countStr)
	should.Nil(err)
	should.Equal(1, count)

	req = httptest.NewRequest(http.MethodGet, "/?limit=XXX", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	err = suite.router.getNotifications(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/?offset=XXX", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	err = suite.router.getNotifications(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages/short")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	err = suite.router.getLastNotifications(c)
	should.Nil(err)
	should.Equal(http.StatusOK, rec.Code)

	var result []ShortNotificationDTO
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	should.NotNil(result)
	should.Equal(3, len(result))
	for _, n := range result {
		should.NotEmpty(n.CreatedAt)
		should.NotEmpty(n.Title)
		should.NotEmpty(n.ID)
	}
}

func (suite *OnboardingClientRouterTestSuite) TestGetNotification() {
	should := require.New(suite.T())
	notification := &model.Notification{VendorID: uuid.FromStringOrNil(TestID), Title: "Some title", Message: "ZZZ"}
	notification.ID = uuid.NewV4()
	notification.IsRead = true
	should.Nil(suite.db.DB().Create(notification).Error)

	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages/:messageId")
	c.SetParamNames("vendorId", "messageId")
	c.SetParamValues(TestID, "XXX")

	err := suite.router.getNotification(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages/:messageId")
	c.SetParamNames("vendorId", "messageId")
	c.SetParamValues(TestID, uuid.NewV4().String())

	err = suite.router.getNotification(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusNotFound, he.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/messages/:messageId")
	c.SetParamNames("vendorId", "messageId")
	c.SetParamValues(TestID, notification.ID.String())

	err = suite.router.getNotification(c)
	should.Nil(err)
	should.Equal(http.StatusOK, rec.Code)
	var result NotificationDTO
	should.Nil(json.Unmarshal(rec.Body.Bytes(), &result))
	should.Equal(notification.ID.String(), result.ID)
	should.Equal("ZZZ", result.Message)
	should.Equal("Some title", result.Title)
	should.NotEmpty(result.CreatedAt)
	createdAt, err := time.Parse(time.RFC3339, result.CreatedAt)
	should.Nil(err)
	should.True(createdAt.After(time.Now().Add(-time.Duration(1) * time.Minute)))
}

func (suite *OnboardingClientRouterTestSuite) TestRevokeReview() {
	shouldBe := require.New(suite.T())

	docs := model.DocumentsInfo{
		VendorID:     uuid.FromStringOrNil(TestID),
		Status:       model.StatusOnReview,
		ReviewStatus: model.ReviewNew,
	}
	docs.ID = uuid.NewV4()

	shouldBe.Nil(suite.db.DB().Create(&docs).Error, "Can't create vendor's docs")

	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents/reviews")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	// Assertions
	shouldBe.Nil(suite.router.revokeReview(c))
	shouldBe.Equal(http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents/reviews")
	c.SetParamNames("vendorId")
	c.SetParamValues(TestID)

	// Assertions
	err := suite.router.revokeReview(c)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)

	req = httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents/reviews")
	c.SetParamNames("vendorId")
	c.SetParamValues("XXX")

	// Assertions
	err = suite.router.revokeReview(c)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)

	req = httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:vendorId/documents/reviews")
	c.SetParamNames("vendorId")
	c.SetParamValues(uuid.NewV4().String())

	// Assertions
	err = suite.router.revokeReview(c)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)
}
