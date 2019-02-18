package orm_test

import (
	"qilin-api/pkg/conf"
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
	db      *orm.Database
	userId  uuid.UUID
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

	_ = db.DropAllTables()
	db.Init()

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
	require := require.New(suite.T())

	gameService, err := orm.NewGameService(suite.db)
	require.Nil(err, "Unable make game service")

	vendorService, err := orm.NewVendorService(suite.db)
	require.Nil(err, "Unable make vendor service")

	pemKey := "hlF0swObPsKN+0n3FVZd9ocfhjOmcS9A8QB7aZK5cDLPmK/QSXvksgkuBTErPfMEIPBKUtD28QjQqBVD9t/nVJ92qr1tlYq7e6F+mN/lYWNYdAApbR50BNY2+GPm/Tv8w1fwaMu7z08OU0+KBxHFxq+TbRusgOeeggovS4BnQ1FhoZv5So+Tf+bSCPQcWcbdPnw4IoM055qoFwvCz4AFi5ty7eCc1GvMVqU+6N/pPA4Q1LpB9+mG8rYbwoTuy31MF6lbhKlxRHRBUdkiJRhyeRskFI6neJ0rhNd62QzU82tyyYMKZ/s4/tBTk/YxvF7QP8cWBe9/kWu/DmUdecFq6w=="

	userService, err := orm.NewUserService(suite.db, &conf.Jwt{SignatureSecret: pemKey, Algorithm: "HS256"}, nil)
	require.Nil(err, "Unable make user service")

	suite.T().Log("Register new user")
	userId, err := userService.Register("test@protocol.one", "mega123!", "ru")
	require.Nil(err, "Unable to register user1")

	suite.userId = userId

	suite.T().Log("Register second user")
	user2Id, err := userService.Register("test@protocol2.one", "mega124!", "en")
	require.Nil(err, "Unable to register user2")

	suite.T().Log("Create vendor")
	vendor := model.Vendor{
		Name:            "domino",
		Domain3:         "domino",
		Email:           "domino@proto.com",
		HowManyProducts: "+1000",
		ManagerID:       userId,
	}
	vendor2, err := vendorService.Create(&vendor)
	require.Nil(err, "Must create new vendor")

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
	require.Nil(err, "Unable to create game")

	suite.T().Log("Create game")
	gameName := "game1"
	game, err := gameService.Create(userId, vendor2.ID, gameName)
	require.Nil(err, "Unable to create game")
	require.NotEqual(game.ID, uuid.Nil, "Wrong ID for created game")
	require.Equal(game.InternalName, gameName, "Incorrect Game Name from DB")

	suite.T().Log("Fetch created game")
	game2, err := gameService.GetInfo(userId, game.ID)
	require.Nil(err, "Get exists game")
	require.Equal(game2.InternalName, gameName, "Incorrect Game Name from DB")

	suite.T().Log("Try to create game with same name")
	_, err = gameService.Create(userId, vendor2.ID, gameName)
	require.NotNil(err, "Must rise error abount same internalName")

	suite.T().Log("Create anther game")
	game2Name := "game2"
	game3, err := gameService.Create(userId, vendor2.ID, game2Name)
	require.Nil(err, "Unable to create game")
	require.NotEqual(game3.ID, uuid.Nil, "Wrong ID for created game")
	require.Equal(game3.InternalName, game2Name, "Incorrect Game Name from DB")

	suite.T().Log("Get games list")
	games, err := gameService.GetList(userId, vendor2.ID, 0, 20, "", "", "", "name+", 0)
	require.Nil(err, "Unable retrive list of games")
	require.Equal(len(games), 2, "Only 2 games just created")
	require.Equal(games[0].InternalName, gameName, "First game")
	require.Equal(games[1].InternalName, game2Name, "Second game")

	suite.T().Log("Check filter with offset and sort")
	games2, err := gameService.GetList(userId, vendor2.ID, 1, 20, "", "", "", "name-", 0)
	require.Nil(err, "Unable retrive list of games")
	require.Equal(len(games2), 1, "Only 1 retrivied")
	require.Equal(games2[0].InternalName, game2Name, "Second game name")

	suite.T().Log("Check filter with name")
	games3, err := gameService.GetList(userId, vendor2.ID, 0, 20, game2Name, "", "", "name-", 0)
	require.Nil(err, "Unable retrive list of games")
	require.Equal(len(games3), 1, "Only 1 retrivied")
	require.Equal(games3[0].InternalName, game2Name, "Second game name")

	suite.T().Log("Get game list with anther user")
	games4, err := gameService.GetList(user2Id, vendor2.ID, 0, 20, "", "", "", "name+", 0)
	require.NotNil(err, "Must be error")
	require.Nil(games4, "Retrieved games is null")

	suite.T().Log("Delete first game")
	err = gameService.Delete(userId, game.ID)
	require.Nil(err, "Game deletion must be without error")

	suite.T().Log("Try to fetch deleted game")
	game4, err := gameService.GetInfo(userId, game.ID)
	require.NotNil(err, "Rise error because game already removed")
	require.Nil(game4, "Game must be null")

	suite.T().Log("Get games list with one game")
	games, err = gameService.GetList(userId, vendor2.ID, 0, 20, "", "", "", "name+", 0)
	require.Nil(err, "Unable retrive list of games")
	require.Equal(len(games), 1, "Only 1 games must be")
	require.Equal(games[0].InternalName, game2Name, "Second game")

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

	err = gameService.UpdateInfo(userId, game3)
	require.Nil(err, "Must be save without error")

	suite.T().Log("Retrive updated game")
	game5, err := gameService.GetInfo(userId, game3.ID)
	require.Nil(err, "Error must be null")
	require.Equal(game5.Developers, game3.Developers, "Must be same")
	require.Equal(game5.InternalName, game3.InternalName, "Must be same")
	require.Equal(game5.Platforms.Windows, game3.Platforms.Windows, "Must be same")
	require.Equal(game5.Platforms.Linux, game3.Platforms.Linux, "Must be same")
	require.Equal(game5.Requirements.Windows.Recommended.Graphics, "4200ti", "Must be same")
	require.Equal(game5.Publishers, game3.Publishers, "Must be same")
	require.Equal(len(game5.Tags), 2, "Must be same")
    require.Equal(len(game5.GenreAddition), 2, "Must be 3 extra genres")
	require.Equal(game5.GenreMain, int64(1), "Genre with id 1")

	suite.T().Log("Get game descriptions")
	gameDescr, err := gameService.GetDescr(userId, game5.ID)
	require.Nil(err, "Error must be null")
	require.Equal(gameDescr.GameID, game5.ID, "Same as game")

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
	err = gameService.UpdateDescr(userId, gameDescr)
	require.Nil(err, "Error must be null")

	suite.T().Log("Get updated game description")
	gameDescr2, err := gameService.GetDescr(userId, game5.ID)
	require.Nil(err, "Error must be null")
	require.Equal(len(gameDescr2.Reviews), 2, "Must be two review")
	require.Equal(gameDescr2.Reviews[0].Link, "Link", "Same value")
	require.Equal(gameDescr2.Reviews[1].Quote, "555", "Same value")
	require.Equal(gameDescr2.Socials.Facebook, gameDescr.Socials.Facebook, "Same value")
	require.Equal(gameDescr2.GameSite, gameDescr.GameSite, "Same value")
	require.Equal(gameDescr2.Description.EN, gameDescr.Description.EN, "Same value")

	suite.T().Log("Retrive tags with user", userId.String())
	tags, err := gameService.FindTags(userId, "Стрелялки", 20, 0)
	require.Equal(len(tags), 1, "Must be one match")
	require.Equal(tags[0].ID, 1, "Same value")
}

func (suite *GameServiceTestSuite) TestDescriptors() {
	testDescr := model.Descriptor{Title: utils.LocalizedString{
		EN: "Blood",
		RU: "Кровь",
	},
		System: "CERO"}
	require := require.New(suite.T())

	gameService, err := orm.NewGameService(suite.db)
	require.NoError(err)

	descriptors, err := gameService.GetRatingDescriptors("")
	require.NoError(err)
	require.Equal(4, len(descriptors))

	descriptors, err = gameService.GetRatingDescriptors("CERO")
	require.NoError(err)
	require.Equal(1, len(descriptors))
	require.Equal(testDescr.System, descriptors[0].System)
	require.Equal(testDescr.Title, descriptors[0].Title)
}

func (suite *GameServiceTestSuite) TestFindAllGenres() {
	require := require.New(suite.T())

	gameService, err := orm.NewGameService(suite.db)
	require.NoError(err)

	genres, err := gameService.FindGenres(suite.userId, "", 10, 0)
	require.NoError(err)
	require.Equal(3, len(genres))
	require.Equal(1, genres[0].ID)
	require.Equal("Action", genres[0].Title.EN)

	genres2, err := gameService.FindGenres(suite.userId, "", 1, 1)
	require.NoError(err)
	require.Equal(1, len(genres2))
	require.Equal(2, genres2[0].ID)
	require.Equal("Test", genres2[0].Title.EN)
}
