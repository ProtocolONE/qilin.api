package orm_test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/satori/go.uuid"

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
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}

	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.Fail("Unable to connect to database:", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

	suite.db = db

	service, err := orm.NewDiscountService(suite.db)
	if err != nil {
		suite.Fail("Unable to create service", "%v", err)
	}

	suite.service = service

	user := model.User{
		ID:       uuid.NewV4(),
		Login:    "test@protocol.one",
		Password: "megapass",
		Nickname: "Test",
		Lang:     "ru",
	}

	err = db.DB().Create(&user).Error
	suite.Nil(err, "Unable to create user")

	userId := user.ID

	vendorService, err := orm.NewVendorService(db)
	suite.Nil(err, "Unable make vendor service")

	vendor := model.Vendor{
		ID:              uuid.NewV4(),
		Name:            "domino",
		Domain3:         "domino",
		Email:           "domino@proto.com",
		HowManyProducts: "+1000",
		ManagerID:       userId,
	}
	_, err = vendorService.Create(&vendor)
	suite.Nil(err, "Must create new vendor")

	id, _ := uuid.FromString(GameID)
	game := model.Game{}
	game.ID = id
	game.InternalName = "internalName"
	game.FeaturesCtrl = ""
	game.FeaturesCommon = []string{}
	game.Platforms = bto.Platforms{}
	game.Requirements = bto.GameRequirements{}
	game.Languages = bto.GameLangs{}
	game.FeaturesCommon = []string{}
	game.GenreMain = 0
	game.GenreAddition = []int64{}
	game.Tags = []int64{}
	game.VendorID = vendor.ID
	game.CreatorID = userId

	err = db.DB().Create(&game).Error

	if err != nil {
		suite.Fail("Unable to create game", "%v", err)
	}
}

func (suite *DiscountServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}

	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *DiscountServiceTestSuite) TestGetDiscountShouldReturnObject() {
	should := require.New(suite.T())
	id, _ := uuid.FromString(GameID)

	discounts, err := suite.service.GetDiscountsForGame(id)
	should.Nil(err)
	should.Equal([]model.Discount{}, discounts)

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
	discount.ID = uuid.NewV4()
	discount.GameID = id

	should.Nil(suite.db.DB().Create(&discount).Error)

	discounts, err = suite.service.GetDiscountsForGame(id)
	should.Nil(err)
	should.Equal(1, len(discounts))
	should.Equal(discount.Title, discounts[0].Title)
	should.Equal(id, discounts[0].GameID)
	should.True(discount.DateStart.Equal(discounts[0].DateStart))
	should.True(discount.DateEnd.Equal(discounts[0].DateEnd))
	should.Equal(discount.Rate, discounts[0].Rate)
}

func (suite *DiscountServiceTestSuite) TestDiscountShouldReturnNotFoundError() {
	should := require.New(suite.T())

	discounts, err := suite.service.GetDiscountsForGame(uuid.NewV4())
	should.Nil(discounts)
	should.NotNil(err)

	he := err.(*orm.ServiceError)
	should.Equal(http.StatusNotFound, he.Code)

	newId, err := suite.service.AddDiscountForGame(uuid.NewV4(), &model.Discount{})
	should.Equal(uuid.Nil, newId)
	should.NotNil(err)

	he = err.(*orm.ServiceError)
	should.Equal(http.StatusNotFound, he.Code)

	err = suite.service.RemoveDiscountForGame(uuid.NewV4())
	should.NotNil(err)

	he = err.(*orm.ServiceError)
	should.Equal(http.StatusNotFound, he.Code)

	discount := model.Discount{}
	discount.ID = uuid.NewV4()
	err = suite.service.UpdateDiscountForGame(&discount)
	should.NotNil(err)

	he = err.(*orm.ServiceError)
	should.Equal(http.StatusNotFound, he.Code)
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
	newDiscount.GameID = id

	err = suite.service.UpdateDiscountForGame(&newDiscount)
	assert.Nil(suite.T(), err, "Unable to update discount for game")

	inDb := model.Discount{}
	inDb.ID = newId
	err = suite.db.DB().Model(&inDb).First(&inDb).Error
	assert.Nil(suite.T(), err, "Unable to get discount for game")
	assert.Equal(suite.T(), newDiscount.Rate, inDb.Rate, "Rate not equals")
	assert.Equal(suite.T(), newDiscount.Title, inDb.Title, "Rate not equals")
	assert.Equal(suite.T(), newDiscount.Description, inDb.Description, "Rate not equals")
	assert.True(suite.T(), newDiscount.DateStart.Equal(inDb.DateStart), "Rate not equals")
	assert.True(suite.T(), newDiscount.DateEnd.Equal(inDb.DateEnd), "Rate not equals")
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
	err = suite.db.DB().Model(&inDb).Where("game_id = ?", id).First(&inDb).Error
	assert.Nil(suite.T(), err, "Unable to get discount for game")

	err = suite.service.RemoveDiscountForGame(newId)
	assert.Nil(suite.T(), err, "Unable to remove discount for game")

	count := 1
	err = suite.db.DB().Model(&inDb).Where("game_id = ?", id).Count(&count).Error
	assert.Nil(suite.T(), err, "Unable to get discount for game")
	assert.Equal(suite.T(), 0, count, "Count not equal")
}
