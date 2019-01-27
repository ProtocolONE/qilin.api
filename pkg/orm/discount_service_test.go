package orm_test

import (
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	uuid "github.com/satori/go.uuid"

	"github.com/stretchr/testify/suite"
)

type DiscountServiceTestSuite struct {
	suite.Suite
	db      *orm.Database
	service *orm.DiscountService
}

func Test_DiscountService(t *testing.T) {
	suite.Run(t, new(DiscountServiceTestSuite))
}

var (
	GameID = "029ce039-888a-481a-a831-cde7ff4e50b9"
)

func (suite *DiscountServiceTestSuite) SetupTest() {
	dbConfig := conf.Database{
		Host:     "localhost",
		Port:     "5432",
		Database: "test_qilin",
		User:     "postgres",
		Password: "postgres",
	}

	db, err := orm.NewDatabase(&dbConfig)
	if err != nil {
		suite.Fail("Unable to connect to database: %s", err)
	}

	db.Init()

	id, _ := uuid.FromString(GameID)
	db.DB().Save(&model.Game{ID: id, Name: "Test game"})

	suite.db = db

	service, err := orm.NewDiscountService(suite.db)
	if err != nil {
		suite.Fail("Unable to create service %s", err)
	}

	suite.service = service
}

func (suite *DiscountServiceTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Game{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.DB().DropTable(model.Discount{}).Error; err != nil {
		panic(err)
	}

	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *DiscountServiceTestSuite) TestCreateDiscountShouldInsertIntoDB() {
	id, _ := uuid.FromString(GameID)
	discount := model.Discount{
		Rate: 33,
		Title: model.JSONB{
			"en": "TEST WINTER SALE",
			"ru": "ТЕСТ САЛО",
		},
		Description: model.JSONB{
			"en": "TEST WINTER SALE",
			"ru": "ТЕСТ САЛО",
		},
	}

	end, err := time.Parse(time.RFC3339, "2019-09-22T07:53:16Z")
	start, err := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")

	discount.DateEnd = end
	discount.DateStart = start

	newId, err := suite.service.AddDiscountForGame(id, &discount)

	assert.Nil(suite.T(), err, "Unable to create discount for game")
	assert.NotEqual(suite.T(), uuid.Nil, newId)

	inDb := model.Discount{}
	inDb.ID = newId
	err = suite.db.DB().Model(&inDb).First(&inDb).Error
	assert.Nil(suite.T(), err, "Unable to get discount for game")
	assert.Equal(suite.T(), discount.Title, inDb.Title, "Title not equals")
	assert.Equal(suite.T(), discount.Description, inDb.Description, "Title not equals")
	assert.Equal(suite.T(), discount.Rate, inDb.Rate, "Title not equals")
}

func (suite *DiscountServiceTestSuite) TestUpdateDiscountShouldChangeInDB() {
	id, _ := uuid.FromString(GameID)
	discount := model.Discount{
		Rate: 33,
		Title: model.JSONB{
			"en": "TEST WINTER SALE",
			"ru": "ТЕСТ САЛО",
		},
		Description: model.JSONB{
			"en": "TEST WINTER SALE",
			"ru": "ТЕСТ САЛО",
		},
	}

	end, err := time.Parse(time.RFC3339, "2019-09-22T07:53:16Z")
	start, err := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")

	discount.DateEnd = end
	discount.DateStart = start

	newId, err := suite.service.AddDiscountForGame(id, &discount)

	assert.Nil(suite.T(), err, "Unable to create discount for game")
	assert.NotEqual(suite.T(), uuid.Nil, newId)

	newDiscount := model.Discount{
		Rate: 19,
		Title: model.JSONB{
			"en": "NEW TEST WINTER SALE",
			"ru": "NEW ТЕСТ САЛО",
		},
		Description: model.JSONB{
			"en": "NEW TEST WINTER SALE",
			"ru": "NEW ТЕСТ САЛО",
		},
		DateEnd:   end,
		DateStart: start,
	}
	newDiscount.ID = newId

	err = suite.service.UpdateDiscountForGame(&newDiscount)
	assert.Nil(suite.T(), err, "Unable to update discount for game")

	inDb := model.Discount{}
	inDb.ID = newId
	err = suite.db.DB().Model(&inDb).First(&inDb).Error
	assert.Nil(suite.T(), err, "Unable to get discount for game")
	assert.Equal(suite.T(), newDiscount.Rate, inDb.Rate, "Rate not equals")
	assert.Equal(suite.T(), newDiscount.Title, inDb.Title, "Rate not equals")
	assert.Equal(suite.T(), newDiscount.Description, inDb.Description, "Rate not equals")
	assert.Equal(suite.T(), newDiscount.DateEnd, inDb.DateEnd, "Rate not equals")
	assert.Equal(suite.T(), newDiscount.DateStart, inDb.DateStart, "Rate not equals")
	assert.Equal(suite.T(), id, inDb.GameID, "Rate not equals")
}

func (suite *DiscountServiceTestSuite) TestRemoveDiscountShouldDeleteIntoDB() {
	id, _ := uuid.FromString(GameID)
	discount := model.Discount{
		Rate: 33,
		Title: model.JSONB{
			"en": "TEST WINTER SALE",
			"ru": "ТЕСТ САЛО",
		},
		Description: model.JSONB{
			"en": "TEST WINTER SALE",
			"ru": "ТЕСТ САЛО",
		},
	}

	end, err := time.Parse(time.RFC3339, "2019-09-22T07:53:16Z")
	start, err := time.Parse(time.RFC3339, "2019-01-22T07:53:16Z")

	discount.DateEnd = end
	discount.DateStart = start

	newId, err := suite.service.AddDiscountForGame(id, &discount)
	inDb := model.Discount{}
	inDb.ID = newId
	err = suite.db.DB().Model(&inDb).First(&inDb).Error
	assert.Nil(suite.T(), err, "Unable to get discount for game")

	err = suite.service.RemoveDiscountForGame(newId)
	assert.Nil(suite.T(), err, "Unable to remove discount for game")

	count := 1
	err = suite.db.DB().Model(&inDb).Count(&count).Error
	assert.Nil(suite.T(), err, "Unable to get discount for game")
	assert.Equal(suite.T(), 0, count, "Count not equal")
}
