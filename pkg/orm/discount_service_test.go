package orm_test

import (
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
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
	config, err := conf.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}

	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.Fail("Unable to connect to database:", "%v", err)
	}

	db.Init()

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
	game.Genre = []string{}
	game.Tags = []string{}
	game.VendorID = vendor.ID
	game.CreatorID = userId

	err = db.DB().Create(&game).Error

	if err != nil {
		suite.Fail("Unable to create game", "%v", err)
	}
}

func (suite *DiscountServiceTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Game{}, model.Vendor{}, model.User{}, model.GameTag{}, model.Discount{}, model.Price{}).Error; err != nil {
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
