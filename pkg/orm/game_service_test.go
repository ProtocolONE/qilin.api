package orm_test

import (
	"github.com/ProtocolONE/rbac"
	"github.com/stretchr/testify/assert"
	"qilin-api/pkg/api/mock"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GameServiceTestSuite struct {
	suite.Suite
	db     *orm.Database
	userId string
}

func Test_GameService(t *testing.T) {
	suite.Run(t, new(GameServiceTestSuite))
}

func (suite *GameServiceTestSuite) SetupTest() {
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

	suite.db = db

	suite.NoError(db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "PEGI",
	}).Error)

	suite.NoError(db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "ESRB",
	}).Error)

	suite.NoError(db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "USK",
	}).Error)

	suite.NoError(db.DB().Create(&model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "CERO",
	}).Error)

	suite.NoError(db.DB().Create(&model.GameGenre{
		model.GameTag{
			ID:    1,
			Title: utils.LocalizedString{EN: "Action"},
		},
	}).Error)

	suite.NoError(db.DB().Create(&model.GameGenre{
		model.GameTag{
			ID:    2,
			Title: utils.LocalizedString{EN: "Test"},
		},
	}).Error)

	suite.NoError(db.DB().Create(&model.GameGenre{
		model.GameTag{
			ID:    3,
			Title: utils.LocalizedString{EN: "Tanks"},
		},
	}).Error)
}

func (suite *GameServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *GameServiceTestSuite) TestGames() {
	should := require.New(suite.T())

	gameService, err := orm.NewGameService(suite.db)
	should.Nil(err, "Unable make game service")

	ow := orm.NewOwnerProvider(suite.db)
	enf := rbac.NewEnforcer()
	memService := orm.NewMembershipService(suite.db, ow, enf, mock.NewMailer(), "")

	vendorService, err := orm.NewVendorService(suite.db, memService)
	should.Nil(err, "Unable make vendor service")

	userService, err := orm.NewUserService(suite.db, nil)
	should.Nil(err, "Unable make user service")

	suite.T().Log("Register new user")
	user, err := userService.Create(uuid.NewV4().String(), "test@protocol.one", "ru")
	should.Nil(err, "Unable to register user1")

	suite.userId = user.ID

	suite.T().Log("Create vendor")
	vendor := model.Vendor{
		Name:            "domino",
		Domain3:         "domino",
		Email:           "domino@proto.com",
		HowManyProducts: "+1000",
		ManagerID:       user.ID,
	}
	vendor2, err := vendorService.Create(&vendor)
	should.Nil(err, "Must create new vendor")

	suite.T().Log("Makes more game-tags")
	err = gameService.CreateTags([]model.GameTag{
		{
			ID:    1,
			Title: utils.LocalizedString{EN: "Action", RU: "Стрелялки"},
		},
		{
			ID:    2,
			Title: utils.LocalizedString{EN: "Test", RU: "Тест"},
		},
		{
			ID:    3,
			Title: utils.LocalizedString{EN: "Tanks", RU: "Танки"},
		},
	})
	should.Nil(err, "Unable to create game tags")

	suite.T().Log("Makes more genres")
	err = gameService.CreateGenres([]model.GameGenre{
		{model.GameTag{
			ID:    4,
			Title: utils.LocalizedString{EN: "genre-1", RU: "Жанр-1"},
		}},
		{model.GameTag{
			ID:    5,
			Title: utils.LocalizedString{EN: "genre-2", RU: "Жанр-2"},
		}},
		{model.GameTag{
			ID:    6,
			Title: utils.LocalizedString{EN: "genre-3", RU: "Жанр-3"},
		}},
	})
	should.Nil(err, "Unable to create genres")

	suite.T().Log("Create game")
	gameName := "game1"
	game, err := gameService.Create(user.ID, vendor2.ID, gameName)
	should.Nil(err, "Unable to create game")
	should.NotEqual(game.ID, uuid.Nil, "Wrong ID for created game")
	should.Equal(game.InternalName, gameName, "Incorrect Game Name from DB")

	suite.T().Log("Fetch created game")
	game2, err := gameService.GetInfo(game.ID)
	should.Nil(err, "Get exists game")
	should.Equal(game2.InternalName, gameName, "Incorrect Game Name from DB")

	suite.T().Log("Try to create game with same name")
	_, err = gameService.Create(user.ID, vendor2.ID, gameName)
	should.NotNil(err, "Must rise error abount same internalName")

	suite.T().Log("Create anther game")
	game2Name := "game2"
	game3, err := gameService.Create(user.ID, vendor2.ID, game2Name)
	should.Nil(err, "Unable to create game")
	should.NotEqual(game3.ID, uuid.Nil, "Wrong ID for created game")
	should.Equal(game3.InternalName, game2Name, "Incorrect Game Name from DB")

	suite.T().Log("Get games list")
	games, err := gameService.GetList(user.ID, vendor2.ID, 0, 20, "", "", "", "name+", 0)
	should.Nil(err, "Unable retrive list of games")
	should.Equal(2, len(games), "Only 2 games just created")
	should.Equal(games[0].InternalName, gameName, "First game")
	should.Equal(games[1].InternalName, game2Name, "Second game")

	suite.T().Log("Check filter with offset and sort")
	games2, err := gameService.GetList(user.ID, vendor2.ID, 1, 20, "", "", "", "name-", 0)
	should.Nil(err, "Unable retrive list of games")
	should.Equal(len(games2), 1, "Only 1 retrivied")
	should.Equal(games2[0].InternalName, game2Name, "Second game name")

	suite.T().Log("Check filter with name")
	games3, err := gameService.GetList(user.ID, vendor2.ID, 0, 20, game2Name, "", "", "name-", 0)
	should.Nil(err, "Unable retrive list of games")
	should.Equal(len(games3), 1, "Only 1 retrivied")
	should.Equal(games3[0].InternalName, game2Name, "Second game name")

	suite.T().Log("Delete first game")
	err = gameService.Delete(user.ID, game.ID)
	should.Nil(err, "Game deletion must be without error")

	suite.T().Log("Try to fetch deleted game")
	game4, err := gameService.GetInfo(game.ID)
	should.NotNil(err, "Rise error because game already removed")
	should.Nil(game4, "Game must be null")

	suite.T().Log("Get games list with one game")
	games, err = gameService.GetList(user.ID, vendor2.ID, 0, 20, "", "", "", "name+", 0)
	should.Nil(err, "Unable retrive list of games")
	should.Equal(len(games), 1, "Only 1 games must be")
	should.Equal(games[0].InternalName, game2Name, "Second game")

	suite.T().Log("Update game")
	game3.Tags = []int64{1, 2}
	game3.Developers = "Developers"
	game3.InternalName = gameName + "-x"
	game3.Platforms.Windows = true
	game3.Platforms.Linux = true
	game3.Requirements.Windows.Recommended.Graphics = "4200ti"
	game3.Publishers = "Publishers"
	game3.GenreMain = 1
	game3.GenreAddition = []int64{2, 3}

	err = gameService.UpdateInfo(game3)
	should.Nil(err, "Must be save without error")

	suite.T().Log("Retrive updated game")
	game5, err := gameService.GetInfo(game3.ID)
	should.Nil(err, "Error must be null")
	should.Equal(game5.Developers, game3.Developers, "Must be same")
	should.Equal(game5.InternalName, game3.InternalName, "Must be same")
	should.Equal(game5.Platforms.Windows, game3.Platforms.Windows, "Must be same")
	should.Equal(game5.Platforms.Linux, game3.Platforms.Linux, "Must be same")
	should.Equal(game5.Requirements.Windows.Recommended.Graphics, "4200ti", "Must be same")
	should.Equal(game5.Publishers, game3.Publishers, "Must be same")
	should.Equal(len(game5.Tags), 2, "Must be same")
	should.Equal(len(game5.GenreAddition), 2, "Must be 3 extra genres")
	should.Equal(game5.GenreMain, int64(1), "Genre with id 1")

	suite.T().Log("Get game descriptions")
	gameDescr, err := gameService.GetDescr(game5.ID)
	should.Nil(err, "Error must be null")
	should.Equal(gameDescr.GameID, game5.ID, "Same as game")

	suite.T().Log("Update game descriptions")
	gameDescr.Reviews = bto.GameReviews{bto.GameReview{
		PressName: "PressName",
		Link:      "Link",
		Score:     "Score",
		Quote:     "Quote",
	}, bto.GameReview{
		PressName: "222",
		Link:      "333",
		Score:     "444",
		Quote:     "555",
	}}
	gameDescr.Socials.Facebook = "Facebook"
	gameDescr.GameSite = "GameSite"
	gameDescr.AdditionalDescription = "AdditionalDescription"
	gameDescr.Description = utils.LocalizedString{
		EN: "eng-descr",
		RU: "ru-descr",
	}
	err = gameService.UpdateDescr(gameDescr)
	should.Nil(err, "Error must be null")

	suite.T().Log("Get updated game description")
	gameDescr2, err := gameService.GetDescr(game5.ID)
	should.Nil(err, "Error must be null")
	should.Equal(len(gameDescr2.Reviews), 2, "Must be two review")
	should.Equal(gameDescr2.Reviews[0].Link, "Link", "Same value")
	should.Equal(gameDescr2.Reviews[1].Quote, "555", "Same value")
	should.Equal(gameDescr2.Socials.Facebook, gameDescr.Socials.Facebook, "Same value")
	should.Equal(gameDescr2.GameSite, gameDescr.GameSite, "Same value")
	should.Equal(gameDescr2.Description.EN, gameDescr.Description.EN, "Same value")

	suite.T().Log("Retrive tags with user", user.ID)
	tags, err := gameService.FindTags(user.ID, "Стрелялки", 20, 0)
	should.Equal(len(tags), 1, "Must be one match")
	should.Equal(tags[0].ID, 1, "Same value")
}

func (suite *GameServiceTestSuite) TestDescriptors() {
	testDescr := model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "CERO"}
	should := require.New(suite.T())

	gameService, err := orm.NewGameService(suite.db)
	should.NoError(err)

	descriptors, err := gameService.GetRatingDescriptors("")
	should.NoError(err)
	should.Equal(4, len(descriptors))

	descriptors, err = gameService.GetRatingDescriptors("CERO")
	should.NoError(err)
	should.Equal(1, len(descriptors))
	should.Equal(testDescr.System, descriptors[0].System)
	should.Equal(testDescr.Title, descriptors[0].Title)
}

func (suite *GameServiceTestSuite) TestFindAllGenres() {
	should := require.New(suite.T())

	gameService, err := orm.NewGameService(suite.db)
	should.NoError(err)

	genres, err := gameService.FindGenres(suite.userId, "", 10, 0)
	should.NoError(err)
	should.Equal(3, len(genres))
	should.Equal(1, genres[0].ID)
	should.Equal("Action", genres[0].Title.EN)

	genres2, err := gameService.FindGenres(suite.userId, "", 1, 1)
	should.NoError(err)
	should.Equal(1, len(genres2))
	should.Equal(2, genres2[0].ID)
	should.Equal("Test", genres2[0].Title.EN)
}
