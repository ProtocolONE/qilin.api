package orm_test

import (
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/onboarding"
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
		Genre:          pq.StringArray{},
		Tags:           pq.StringArray{},
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

func (suite *OnbardingServiceTestSuite) TestGetById() {
	should := require.New(suite.T())

	id, _ := uuid.FromString(Id)
	docs, err := suite.service.GetById(id)
	should.Nil(err)
	should.Equal(onboarding.DocumentsInfo{}, docs)
}


