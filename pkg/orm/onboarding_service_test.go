package orm_test

import (
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"
	"time"
)

type OnbardingServiceTestSuite struct {
	suite.Suite
	db *orm.Database
	service *orm.OnboardingService
}


func Test_OnbardingService(t *testing.T) {
	suite.Run(t, new(OnbardingServiceTestSuite))
}

func (suite *OnbardingServiceTestSuite) SetupTest() {
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	db.DropAllTables()
	db.Init()
	id, _ := uuid.FromString(Id)

	err = db.DB().Create(&model.Vendor{
		ID: id,
		Email: "test@test.com",
		Name: "Test",
		HowManyProducts: "10+",
	}).Error
	assert.Nil(suite.T(), err, "Unable to make vendor")

	err = db.DB().Save(&model.Game{
		ID:             id,
		InternalName:   "Test_game_3",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
		VendorID:		id,
	}).Error
	assert.Nil(suite.T(), err, "Unable to make game")

	suite.db = db
	suite.service, err = orm.NewOnboardingService(db)
	assert.Nil(suite.T(), err, "Unable to make service")
}

func (suite *OnbardingServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *OnbardingServiceTestSuite) TestServiceMethodsWithErrors() {
	should := require.New(suite.T())
	_, err := suite.service.GetById(uuid.NewV4())
	should.NotNil(err)
	should.Equal(404, err.(*orm.ServiceError).Code)

	_, err = suite.service.GetForVendor(uuid.NewV4())
	should.NotNil(err)
	should.Equal(404, err.(*orm.ServiceError).Code)

	doc := &model.DocumentsInfo{}
	doc.ID = uuid.NewV4()
	doc.VendorID = uuid.NewV4()

	err = suite.service.ChangeDocument(doc)
	should.NotNil(err)
	should.Equal(404, err.(*orm.ServiceError).Code)

	id, _ := uuid.FromString(Id)
	doc = &model.DocumentsInfo{}
	doc.ID = id
	doc.VendorID = uuid.NewV4()
	err = suite.service.ChangeDocument(doc)
	should.NotNil(err)
	should.Equal(404, err.(*orm.ServiceError).Code)
}

func (suite *OnbardingServiceTestSuite) TestServiceMethods() {
	should := require.New(suite.T())

	id, _ := uuid.FromString(Id)
	_, err := suite.service.GetById(id)
	should.Equal(404, err.(*orm.ServiceError).Code)

	docs, err := suite.service.GetForVendor(id)
	should.Nil(err)
	should.NotNil(docs)
	docs = &model.DocumentsInfo{
		VendorID: id,
		Status: model.StatusDraft,
		ReviewStatus: model.ReviewNew,
		Contact: model.JSONB{"name": "TEST"},
	}
	docs.ID = uuid.NewV4()

	err = suite.service.ChangeDocument(docs)
	should.Nil(err)
	dbDoc, err := suite.service.GetForVendor(id)

	should.Nil(err)
	should.Equal(docs.ID, dbDoc.ID)
	should.Equal(docs.VendorID, dbDoc.VendorID)
	should.Equal(docs.Contact, dbDoc.Contact)
	should.Equal(docs.ReviewStatus, dbDoc.ReviewStatus)
	should.Equal(docs.Status, dbDoc.Status)

	dbDoc2, err := suite.service.GetById(docs.ID)

	should.Nil(err)
	should.Equal(docs.ID, dbDoc2.ID)
	should.Equal(docs.VendorID, dbDoc2.VendorID)
	should.Equal(docs.Contact, dbDoc2.Contact)
	should.Equal(docs.ReviewStatus, dbDoc2.ReviewStatus)
	should.Equal(docs.Status, dbDoc2.Status)

	err = suite.service.RevokeReviewRequest(docs.VendorID)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)

	err = suite.service.SendToReview(docs.VendorID)
	should.Nil(err)

	//twice send to review is not allowed
	err = suite.service.SendToReview(docs.VendorID)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)

	err = suite.service.RevokeReviewRequest(docs.VendorID)
	should.Nil(err)

	docs.Status = model.StatusApproved
	should.Nil(suite.db.DB().Save(docs).Error)

	err = suite.service.RevokeReviewRequest(docs.VendorID)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)
}



