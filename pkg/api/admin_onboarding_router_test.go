package api

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"net/url"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"qilin-api/pkg/utils"
	"strings"
	"testing"
)

type OnboardingAdminRouterTestSuite struct {
	suite.Suite
	db      *orm.Database
	service *orm.AdminOnboardingService
	echo    *echo.Echo
	router  *OnboardingAdminRouter
}

var (
	changeToAccepted     = `{"status":"approved", "message":null}`
	changeToDeclined     = `{"status":"returned", "message":null}`
	changeToChecking     = `{"status":"checking", "message":null}`
	changeToArchived     = `{"status":"archived", "message":null}`
	changeToNew          = `{"status":"new", "message":null}`
	changeToSomethingBad = `{"status":"who knows, who knows", "message":null}`
	changeWithoutStatus  = `{"message":null}`
)

func Test_OnboardingAdminRouter(t *testing.T) {
	suite.Run(t, new(OnboardingAdminRouterTestSuite))
}

func (suite *OnboardingAdminRouterTestSuite) SetupTest() {
	should := require.New(suite.T())
	config, err := qilin_test.LoadTestConfig()
	should.Nil(err, "Unable to load config", "%v", err)
	db, err := orm.NewDatabase(&config.Database)
	should.Nil(err, "Unable to connect to database", "%v", err)

	db.DropAllTables()
	db.Init()

	id, _ := uuid.FromString(TestID)
	should.Nil(db.DB().Create(&model.Vendor{ID: id, Email: "example@example.ru", Name: "Test", Domain3: "test"}).Error, "Can't create vendor")
	should.Nil(db.DB().Create(&model.Vendor{ID: uuid.FromStringOrNil("413ab3ec-91b0-43c4-8a4c-653a265288fa"), Email: "example3@example.ru", Name: "Test3", Domain3: "test3"}).Error, "Can't create vendor")

	e := echo.New()
	service, err := orm.NewAdminOnboardingService(db)
	router, err := InitAdminOnboardingRouter(e.Group("/api/v1"), service)
	v := validator.New()
	assert.NoError(suite.T(), utils.RegisterCustomValidations(v))
	e.Validator = &QilinValidator{validator: v}

	suite.db = db
	suite.router = router
	suite.echo = e
}

func (suite *OnboardingAdminRouterTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}

	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *OnboardingAdminRouterTestSuite) TestGetDocumentsShouldGetErrors() {
	suite.generateReviews(suite.db)

	should := require.New(suite.T())
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues("XXX")

	err := suite.router.getDocument(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(uuid.NewV4().String())

	err = suite.router.getDocument(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusNotFound, he.Code)
	}
}

func (suite *OnboardingAdminRouterTestSuite) TestChangeDocumentStatus() {
	suite.generateReviews(suite.db)
	should := require.New(suite.T())

	vendorDocumentsDraft := model.DocumentsInfo{
		VendorID: uuid.FromStringOrNil("413ab3ec-91b0-43c4-8a4c-653a265288fa"),
		Company: model.JSONB{
			"Name":            "MEGA TEST",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Эдуард Никифоров",
				"Position": "Руководитель",
			},
			"Technical": model.JSONB{
				"FullName": "Роман Обрамович",
				"Position": "Батрак",
			},
		},
		Status:       model.StatusOnReview,
		ReviewStatus: model.ReviewNew,
		Banking: model.JSONB{
			"Currency": "USD",
		},
	}
	vendorDocumentsDraft.ID = uuid.NewV4()
	should.Nil(suite.db.DB().Create(&vendorDocumentsDraft).Error)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(changeToAccepted))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(vendorDocumentsDraft.VendorID.String())

	err := suite.router.changeStatus(c)
	should.Nil(err)

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(changeToDeclined))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(vendorDocumentsDraft.VendorID.String())

	err = suite.router.changeStatus(c)
	should.Nil(err)

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(changeToChecking))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(vendorDocumentsDraft.VendorID.String())

	err = suite.router.changeStatus(c)
	should.Nil(err)

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(changeToArchived))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(vendorDocumentsDraft.VendorID.String())

	err = suite.router.changeStatus(c)
	should.Nil(err)

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(changeToNew))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(vendorDocumentsDraft.VendorID.String())

	err = suite.router.changeStatus(c)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(changeToSomethingBad))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(vendorDocumentsDraft.VendorID.String())

	err = suite.router.changeStatus(c)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(changeWithoutStatus))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(vendorDocumentsDraft.VendorID.String())

	err = suite.router.changeStatus(c)
	should.NotNil(err)
	should.Equal(http.StatusUnprocessableEntity, err.(*orm.ServiceError).Code)

	vendorDocumentsDraft.Status = model.StatusDraft
	should.Nil(suite.db.DB().Save(&vendorDocumentsDraft).Error)

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(changeToChecking))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(vendorDocumentsDraft.VendorID.String())

	err = suite.router.changeStatus(c)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)
}

func (suite *OnboardingAdminRouterTestSuite) TestGetDocuments() {
	suite.generateReviews(suite.db)

	should := require.New(suite.T())
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(emptyObject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/:id/documents")
	c.SetParamNames("id")
	c.SetParamValues(TestID)

	// Assertions
	if assert.NoError(suite.T(), suite.router.getDocument(c)) {
		should.Equal(http.StatusOK, rec.Code)
		docs := DocumentsInfoResponseDTO{}
		should.Nil(json.Unmarshal(rec.Body.Bytes(), &docs))
	}
}

func (suite *OnboardingAdminRouterTestSuite) TestGetReviews() {
	should := require.New(suite.T())
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("[]"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/reviews")

	// Assertions
	if assert.NoError(suite.T(), suite.router.getReviews(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), "[]", rec.Body.String())
	}

	suite.generateReviews(suite.db)

	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader("[]"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/reviews")

	// Assertions
	if assert.NoError(suite.T(), suite.router.getReviews(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.NotEqual(suite.T(), "[]", rec.Body.String())
		var reviews []DocumentsInfoResponseDTO
		should.Nil(json.Unmarshal(rec.Body.Bytes(), &reviews))
		should.Equal(13, len(reviews))
	}

	q := make(url.Values)
	q.Set("limit", "100")
	q.Set("offset", "10")
	q.Set("sort", "-name")
	q.Set("name", "test")
	q.Set("status", "new")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), strings.NewReader("[]"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/reviews")

	// Assertions
	if assert.NoError(suite.T(), suite.router.getReviews(c)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
	}

	q = make(url.Values)
	q.Set("limit", "qwe")
	q.Set("offset", "10")
	q.Set("sort", "-name")
	q.Set("name", "test")
	q.Set("status", "new")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), strings.NewReader("[]"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/reviews")

	err := suite.router.getReviews(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	q = make(url.Values)
	q.Set("limit", "100")
	q.Set("offset", "qwe")
	q.Set("sort", "-name")
	q.Set("name", "test")
	q.Set("status", "new")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), strings.NewReader("[]"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/reviews")

	err = suite.router.getReviews(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	q = make(url.Values)
	q.Set("limit", "100")
	q.Set("offset", "10")
	q.Set("sort", "-test")
	q.Set("name", "test")
	q.Set("status", "new")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), strings.NewReader("[]"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/reviews")

	err = suite.router.getReviews(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}

	q = make(url.Values)
	q.Set("limit", "100")
	q.Set("offset", "10")
	q.Set("sort", "-name")
	q.Set("name", "test")
	q.Set("status", "test")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), strings.NewReader("[]"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)
	c.SetPath("/api/v1/vendors/reviews")

	err = suite.router.getReviews(c)
	should.NotNil(err)
	if err != nil {
		he := err.(*orm.ServiceError)
		should.Equal(http.StatusBadRequest, he.Code)
	}
}

func (suite *OnboardingAdminRouterTestSuite) generateReviews(db *orm.Database) {
	id, _ := uuid.FromString("c513baaf-8cfc-4f68-921b-45a54dea741c")
	game := model.Game{}
	game.ID = id
	game.InternalName = "internalName"
	game.FeaturesCtrl = ""
	game.FeaturesCommon = []string{}
	game.Platforms = bto.Platforms{}
	game.Requirements = bto.GameRequirements{}
	game.Languages = bto.GameLangs{}
	game.FeaturesCommon = []string{}
	game.GenreMain = 1
	game.GenreAddition = []int64{1, 2}
	game.Tags = []int64{1, 2}
	game.CreatorID = id

	err := db.DB().Create(&game).Error

	if err != nil {
		suite.Fail("Unable to create game", "%v", err)
	}

	vendorDocumentsDraft := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "MEGA TEST",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Эдуард Никифоров",
				"Position": "Руководитель",
			},
			"Technical": model.JSONB{
				"FullName": "Роман Обрамович",
				"Position": "Батрак",
			},
		},
		Status:       model.StatusDraft,
		ReviewStatus: model.ReviewNew,
		Banking: model.JSONB{
			"Currency": "USD",
		},
	}
	vendorDocumentsDraft.ID = uuid.NewV4()

	vendorDocuments := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "MEGA TEST",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Эдуард Никифоров",
				"Position": "Руководитель",
			},
			"Technical": model.JSONB{
				"FullName": "Роман Обрамович",
				"Position": "Батрак",
			},
		},
		Status:       model.StatusOnReview,
		ReviewStatus: model.ReviewChecking,
		Banking: model.JSONB{
			"Currency": "USD",
		},
	}
	vendorDocuments.ID = uuid.NewV4()

	vendorDocuments2 := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "PUBG TEST",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Филимонов Андрей",
				"Position": "IT Director",
			},
		},
		Status:       model.StatusApproved,
		ReviewStatus: model.ReviewApproved,
		Banking: model.JSONB{
			"Currency": "EUR",
		},
	}
	vendorDocuments2.ID = uuid.NewV4()

	vendorDocuments3 := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "Ash of Evils ",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Lucifer",
				"Position": "CEO",
			},
		},
		Status:       model.StatusDeclined,
		ReviewStatus: model.ReviewReturned,
		Banking: model.JSONB{
			"Currency": "USD",
		},
	}
	vendorDocuments3.ID = uuid.NewV4()

	for i := 0; i < 10; i++ {
		vendorDocuments4 := model.DocumentsInfo{
			VendorID: id,
			Company: model.JSONB{
				"Name":            "ZTEST2",
				"AlternativeName": "Alt MEGA NAME",
				"Country":         "RUSSIA",
			},
			Contact: model.JSONB{
				"Authorized": model.JSONB{
					"FullName": "Test Name",
					"Position": "Test Position",
				},
				"Technical": model.JSONB{
					"FullName": "Test Name",
					"Position": "Test Position",
				},
			},
			Status:       model.StatusOnReview,
			ReviewStatus: model.ReviewNew,
			Banking: model.JSONB{
				"Currency": "USD",
			},
		}
		vendorDocuments4.ID = uuid.NewV4()
		suite.Nil(db.DB().Create(&vendorDocuments4).Error)
	}

	suite.Nil(db.DB().Create(&vendorDocuments).Error)
	suite.Nil(db.DB().Create(&vendorDocuments2).Error)
	suite.Nil(db.DB().Create(&vendorDocuments3).Error)
	suite.Nil(db.DB().Create(&vendorDocumentsDraft).Error)
}