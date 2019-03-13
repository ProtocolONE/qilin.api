package orm

import (
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/test"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RatingServiceTestSuite struct {
	suite.Suite
	db      *Database
	service *RatingService
}

var (
	ESRBDescriptors []uint
	PEGIDescriptors []uint
	USKDescriptors  []uint
	CERODescriptors []uint
	BBFCDescriptors []uint
)

func Test_RatingService(t *testing.T) {
	suite.Run(t, new(RatingServiceTestSuite))
}

func (suite *RatingServiceTestSuite) SetupTest() {
	require := require.New(suite.T())

	ESRBDescriptors = []uint{}
	PEGIDescriptors = []uint{}
	USKDescriptors = []uint{}
	CERODescriptors = []uint{}
	BBFCDescriptors = []uint{}

	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		suite.T().Log(err)
	}
	if err := db.Init(); err != nil {
		suite.T().Log(err)
	}

	id, _ := uuid.FromString(ID)
	err = db.DB().Create(&model.Game{
		ID:             id,
		InternalName:   "Test_game_2",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
	}).Error

	require.Nil(err, "Unable to make game")

	suite.db = db

	res := db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "USK",
	})
	require.NoError(res.Error)
	USKDescriptors = append(USKDescriptors, res.Value.(*model.Descriptor).ID)

	res = db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "BBFC",
	})
	require.NoError(res.Error)
	BBFCDescriptors = append(BBFCDescriptors, res.Value.(*model.Descriptor).ID)

	res = db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "CERO",
	})
	require.NoError(res.Error)
	CERODescriptors = append(CERODescriptors, res.Value.(*model.Descriptor).ID)

	res = db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "ESRB",
	})
	require.NoError(res.Error)
	ESRBDescriptors = append(ESRBDescriptors, res.Value.(*model.Descriptor).ID)

	res = db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "PEGI",
	})
	require.NoError(res.Error)
	PEGIDescriptors = append(PEGIDescriptors, res.Value.(*model.Descriptor).ID)

	service, err := NewRatingService(db)

	suite.db = db
	suite.service = service
}

func (suite *RatingServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *RatingServiceTestSuite) TestGetRatingsForGameShouldReturnEmptyObject() {
	id, _ := uuid.FromString(ID)

	ratings, err := suite.service.GetRatingsForGame(id)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), &model.GameRating{}, ratings, "Ratings not equal")
}

func (suite *RatingServiceTestSuite) TestGetRatingsForGameShouldReturnFullObject() {
	id, _ := uuid.FromString(ID)
	testModel := &model.GameRating{
		BBFC: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10.0,
		},
		CERO: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     false,
			"ageRestrict":         21.0,
		},
		ESRB: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     false,
			"ageRestrict":         15.0,
		},
		PEGI: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     false,
			"ageRestrict":         3.0,
		},
		USK: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     true,
			"ageRestrict":         5.0,
		},
		GameID: id,
	}

	assert.NoError(suite.T(), suite.service.db.Create(&testModel).Error)

	ratings, err := suite.service.GetRatingsForGame(id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testModel.GameID, ratings.GameID, "GameID not equal")
	assert.Equal(suite.T(), testModel.USK["ageRestrict"], ratings.USK["ageRestrict"], "USK not equal")
	assert.Equal(suite.T(), testModel.USK["displayOnlineNotice"], ratings.USK["displayOnlineNotice"], "USK not equal")
	assert.Equal(suite.T(), testModel.USK["showAgeRestrict"], ratings.USK["showAgeRestrict"], "USK not equal")
	assert.Equal(suite.T(), testModel.PEGI["ageRestrict"], ratings.PEGI["ageRestrict"], "PEGI not equal")
	assert.Equal(suite.T(), testModel.PEGI["displayOnlineNotice"], ratings.PEGI["displayOnlineNotice"], "PEGI not equal")
	assert.Equal(suite.T(), testModel.PEGI["showAgeRestrict"], ratings.PEGI["showAgeRestrict"], "PEGI not equal")
	assert.Equal(suite.T(), testModel.ESRB["ageRestrict"], ratings.ESRB["ageRestrict"], "ESRB not equal")
	assert.Equal(suite.T(), testModel.ESRB["displayOnlineNotice"], ratings.ESRB["displayOnlineNotice"], "ESRB[ not equal")
	assert.Equal(suite.T(), testModel.ESRB["showAgeRestrict"], ratings.ESRB["showAgeRestrict"], "ESRB[ not equal")
	assert.Equal(suite.T(), testModel.CERO["ageRestrict"], ratings.CERO["ageRestrict"], "CERO not equal")
	assert.Equal(suite.T(), testModel.CERO["displayOnlineNotice"], ratings.CERO["displayOnlineNotice"], "CERO not equal")
	assert.Equal(suite.T(), testModel.CERO["showAgeRestrict"], ratings.CERO["showAgeRestrict"], "CERO not equal")
	assert.Equal(suite.T(), testModel.BBFC["ageRestrict"], ratings.BBFC["ageRestrict"], "BBFC not equal")
	assert.Equal(suite.T(), testModel.BBFC["displayOnlineNotice"], ratings.BBFC["displayOnlineNotice"], "BBFC not equal")
	assert.Equal(suite.T(), testModel.BBFC["showAgeRestrict"], ratings.BBFC["showAgeRestrict"], "BBFC not equal")

	assert.NoError(suite.T(), suite.service.db.Delete(&testModel).Error)
}

func (suite *RatingServiceTestSuite) TestChangeRatingsForGameShouldReturnChangeInDB() {
	id, _ := uuid.FromString(ID)
	testModel := &model.GameRating{
		BBFC: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10.0,
		},
		CERO: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     false,
			"ageRestrict":         21.0,
		},
		ESRB: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     false,
			"ageRestrict":         15.0,
		},
		PEGI: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     false,
			"ageRestrict":         3.0,
		},
		USK: model.JSONB{
			"displayOnlineNotice": false,
			"showAgeRestrict":     true,
			"ageRestrict":         5.0,
		},
	}

	err := suite.service.SaveRatingsForGame(id, testModel)
	assert.NoError(suite.T(), err)
	ratings, err := suite.service.GetRatingsForGame(id)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ratings, "ratings is null")
	assert.Equal(suite.T(), id, ratings.GameID, "GameID not equal")
	assert.Equal(suite.T(), testModel.USK["ageRestrict"], ratings.USK["ageRestrict"], "USK not equal")
	assert.Equal(suite.T(), testModel.USK["displayOnlineNotice"], ratings.USK["displayOnlineNotice"], "USK not equal")
	assert.Equal(suite.T(), testModel.USK["showAgeRestrict"], ratings.USK["showAgeRestrict"], "USK not equal")
	assert.Equal(suite.T(), testModel.PEGI["ageRestrict"], ratings.PEGI["ageRestrict"], "PEGI not equal")
	assert.Equal(suite.T(), testModel.PEGI["displayOnlineNotice"], ratings.PEGI["displayOnlineNotice"], "PEGI not equal")
	assert.Equal(suite.T(), testModel.PEGI["showAgeRestrict"], ratings.PEGI["showAgeRestrict"], "PEGI not equal")
	assert.Equal(suite.T(), testModel.ESRB["ageRestrict"], ratings.ESRB["ageRestrict"], "ESRB not equal")
	assert.Equal(suite.T(), testModel.ESRB["displayOnlineNotice"], ratings.ESRB["displayOnlineNotice"], "ESRB[ not equal")
	assert.Equal(suite.T(), testModel.ESRB["showAgeRestrict"], ratings.ESRB["showAgeRestrict"], "ESRB[ not equal")
	assert.Equal(suite.T(), testModel.CERO["ageRestrict"], ratings.CERO["ageRestrict"], "CERO not equal")
	assert.Equal(suite.T(), testModel.CERO["displayOnlineNotice"], ratings.CERO["displayOnlineNotice"], "CERO not equal")
	assert.Equal(suite.T(), testModel.CERO["showAgeRestrict"], ratings.CERO["showAgeRestrict"], "CERO not equal")
	assert.Equal(suite.T(), testModel.BBFC["ageRestrict"], ratings.BBFC["ageRestrict"], "BBFC not equal")
	assert.Equal(suite.T(), testModel.BBFC["displayOnlineNotice"], ratings.BBFC["displayOnlineNotice"], "BBFC not equal")
	assert.Equal(suite.T(), testModel.BBFC["showAgeRestrict"], ratings.BBFC["showAgeRestrict"], "BBFC not equal")
}

func (suite *RatingServiceTestSuite) TestChangeRatingsWithBadIdShouldReturnError() {
	testModel2 := &model.GameRating{
		BBFC: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10,
			"rating":              "U",
			"descriptors": []int{
				666, 667,
			},
		},
	}

	err := suite.service.SaveRatingsForGame(uuid.NewV4(), testModel2)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*ServiceError)
		assert.Equal(suite.T(), http.StatusNotFound, he.Code, he.Message)
	}

	err = suite.service.SaveRatingsForGame(uuid.Nil, testModel2)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*ServiceError)
		assert.Equal(suite.T(), http.StatusNotFound, he.Code, he.Message)
	}
}

func (suite *RatingServiceTestSuite) TestChangeRatingsWithBadDescriptorsShouldReturnError() {
	id, _ := uuid.FromString(ID)
	testModel := &model.GameRating{
		BBFC: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10,
			"rating":              "U",
			model.DescriptorsField: []uint{
				666, 667,
			},
		},
		PEGI: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10,
			"rating":              "U",
			model.DescriptorsField: []uint{
				666, 667,
			},
		},
		USK: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10,
			"rating":              "U",
			model.DescriptorsField: []uint{
				666, 667,
			},
		},
		ESRB: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10,
			"rating":              "U",
			model.DescriptorsField: []uint{
				666, 667,
			},
		},
		CERO: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10,
			"rating":              "U",
			model.DescriptorsField: []uint{
				666, 667,
			},
		},
	}
	err := suite.service.SaveRatingsForGame(id, testModel)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*ServiceError)
		assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code, he.Message)
	}

	testModel2 := &model.GameRating{
		BBFC: model.JSONB{
			"displayOnlineNotice": true,
			"showAgeRestrict":     true,
			"ageRestrict":         10,
			"rating":              "U",
			model.DescriptorsField: []uint{
				666, 667,
			},
		},
	}

	err = suite.service.SaveRatingsForGame(id, testModel2)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*ServiceError)
		assert.Equal(suite.T(), http.StatusUnprocessableEntity, he.Code, he.Message)
	}
}

func (suite *RatingServiceTestSuite) TestChangeRatingsShouldReturnOk() {
	id, _ := uuid.FromString(ID)
	testModel := &model.GameRating{
		BBFC: model.JSONB{
			"displayOnlineNotice":  true,
			"showAgeRestrict":      true,
			"ageRestrict":          10,
			"rating":               "U",
			model.DescriptorsField: BBFCDescriptors,
		},
		PEGI: model.JSONB{
			"displayOnlineNotice":  true,
			"showAgeRestrict":      true,
			"ageRestrict":          10,
			"rating":               "U",
			model.DescriptorsField: PEGIDescriptors,
		},
		USK: model.JSONB{
			"displayOnlineNotice":  true,
			"showAgeRestrict":      true,
			"ageRestrict":          10,
			"rating":               "U",
			model.DescriptorsField: USKDescriptors,
		},
		ESRB: model.JSONB{
			"displayOnlineNotice":  true,
			"showAgeRestrict":      true,
			"ageRestrict":          10,
			"rating":               "U",
			model.DescriptorsField: ESRBDescriptors,
		},
		CERO: model.JSONB{
			"displayOnlineNotice":  true,
			"showAgeRestrict":      true,
			"ageRestrict":          10,
			"rating":               "U",
			model.DescriptorsField: CERODescriptors,
		},
	}
	err := suite.service.SaveRatingsForGame(id, testModel)
	assert.Nil(suite.T(), err)
}

func (suite *RatingServiceTestSuite) TestGetRatingsForGameShouldReturnNotFound() {
	res, err := suite.service.GetRatingsForGame(uuid.NewV4())
	assert.Nil(suite.T(), res)
	assert.NotNil(suite.T(), err)
	if err != nil {
		he := err.(*ServiceError)
		assert.Equal(suite.T(), http.StatusNotFound, he.Code)
		assert.NotNil(suite.T(), he.Message)
	}
}
