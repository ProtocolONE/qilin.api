package api

import (
	"net/http"
	"net/http/httptest"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
	"qilin-api/pkg/test"
	"qilin-api/pkg/utils"
	"strings"
	"testing"

	"github.com/labstack/echo"
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

	db.DropAllTables()
	db.Init()

	id, _ := uuid.FromString(TestID)
	should.Nil(db.DB().Create(&model.Vendor{ID: id, Email: "example@example.ru", Name: "Test", Domain3: "test"}).Error, "Can't create vendor")

	e := echo.New()
	service, err := orm.NewOnboardingService(db)
	notifier, err := sys.NewNotifier(config.Notifier.ApiKey, config.Notifier.Host)
	should.Nil(err)
	notService, err := orm.NewNotificationService(db, notifier)
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

func (suite *OnboardingClientRouterTestSuite) TestGetShouldReturnEmpty() {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyDocument))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(TestID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getDocument(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), emptyDocument, rec.Body.String())
	}
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
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(TestID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getDocument(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), nonEmptyDocument, rec.Body.String())
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
	c.SetPath("/api/v1/vendors/:id/documents/reviews")
	c.SetParamNames("id")
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
	c.SetPath("/api/v1/vendors/:id/documents/reviews")
	c.SetParamNames("id")
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
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
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
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
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
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
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
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
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
		c.SetPath("/api/v1/vendors/:id/documents")
		c.SetParamNames("id")
		c.SetParamValues(TestID)

		res := suite.router.changeDocument(c)
		assert.NotNil(suite.T(), res, "Should be error. %s", tt.name)
		assert.NotPanics(suite.T(), func() {
			he := res.(*orm.ServiceError)
			assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code, "Test case failed: `%s`. Error: %#v, With body %s", tt.name, he, tt.body)
		})
	}
}
