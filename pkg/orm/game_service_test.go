package orm_test

import (
	"encoding/base64"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	"testing"
)

type GameServiceTestSuite struct {
	suite.Suite
	db *orm.Database
}

func Test_GameService(t *testing.T) {
	suite.Run(t, new(GameServiceTestSuite))
}

func (suite *GameServiceTestSuite) SetupTest() {
	dbConfig := conf.Database{
		Host:     "localhost",
		Port:     "5440",
		Database: "test_qilin",
		User:     "postgres",
		Password: "",
		LogMode: true,
	}

	db, err := orm.NewDatabase(&dbConfig)
	if err != nil {
		suite.Fail("Unable to connect to database: %s", err)
	}

	db.Init()

	suite.db = db
}

func (suite *GameServiceTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Game{}, model.Vendor{}, model.User{}, model.GameTag{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *GameServiceTestSuite) TestGames() {
	require := require.New(suite.T())

	// Go to project root directory
	_ = os.Chdir("../..")

	gameService, err := orm.NewGameService(suite.db)
	require.Nil( err, "Unable make game service")

	vendorService, err := orm.NewVendorService(suite.db)
	require.Nil( err, "Unable make vendor service")

	pemKey, err := base64.StdEncoding.DecodeString("hlF0swObPsKN+0n3FVZd9ocfhjOmcS9A8QB7aZK5cDLPmK/QSXvksgkuBTErPfMEIPBKUtD28QjQqBVD9t/nVJ92qr1tlYq7e6F+mN/lYWNYdAApbR50BNY2+GPm/Tv8w1fwaMu7z08OU0+KBxHFxq+TbRusgOeeggovS4BnQ1FhoZv5So+Tf+bSCPQcWcbdPnw4IoM055qoFwvCz4AFi5ty7eCc1GvMVqU+6N/pPA4Q1LpB9+mG8rYbwoTuy31MF6lbhKlxRHRBUdkiJRhyeRskFI6neJ0rhNd62QzU82tyyYMKZ/s4/tBTk/YxvF7QP8cWBe9/kWu/DmUdecFq6w==")
	require.Nil( err, "Unable to decode base64 secret key")

	userService, err := orm.NewUserService(suite.db, &conf.Jwt{SignatureSecret: pemKey, Algorithm: "HS256"}, nil)
	require.Nil( err, "Unable make user service")

	suite.T().Log("Register new user")
	userId, err := userService.Register("test@protocol.one", "mega123!", "ru")
	require.Nil( err, "Unable to register user1")

	suite.T().Log("Register second user")
	user2Id, err := userService.Register("test@protocol2.one", "mega124!", "en")
	require.Nil( err, "Unable to register user2")

	suite.T().Log("Create vendor")
	vendor := model.Vendor{
		Name: "domino",
		Domain3: "domino",
		Email: "domino@proto.com",
		HowManyProducts: "+1000",
		ManagerID: userId,
	}
	vendor2, err := vendorService.CreateVendor(&vendor)
	require.Nil( err, "Must create new vendor")

	suite.T().Log("Makes more game-tags")
	err = gameService.CreateTags([]model.GameTag{
		model.GameTag{
			ID: "action",
			Title: utils.LocalizedString{EN: "Action", RU: "Стрелялки"},
		},
		model.GameTag{
			ID: "test",
			Title: utils.LocalizedString{EN: "Test", RU: "Тест"},
		},
		model.GameTag{
			ID: "tank",
			Title: utils.LocalizedString{EN: "Tanks", RU: "Танки"},
		},
	})
	require.Nil( err, "Unable to create game")

	suite.T().Log("Create game")
	gameName := "game1"
	game, err := gameService.Create(userId, vendor2.ID, gameName)
	require.Nil( err, "Unable to create game")
	require.NotEqual( game.ID, uuid.Nil,"Wrong ID for created game")
	require.Equal( game.InternalName, gameName, "Incorrect Game Name from DB")

	suite.T().Log("Fetch created game")
	game2, err := gameService.GetInfo(userId, game.ID)
	require.Nil( err, "Get exists game")
	require.Equal( game2.InternalName, gameName, "Incorrect Game Name from DB")

	suite.T().Log("Try to create game with same name")
	_, err = gameService.Create(userId, vendor2.ID, gameName)
	require.NotNil( err, "Must rise error abount same internalName")

	suite.T().Log("Create anther game")
	game2Name := "game2"
	game3, err := gameService.Create(userId, vendor2.ID, game2Name)
	require.Nil( err, "Unable to create game")
	require.NotEqual( game3.ID, uuid.Nil,"Wrong ID for created game")
	require.Equal( game3.InternalName, game2Name, "Incorrect Game Name from DB")

	suite.T().Log("Get games list")
	games, err := gameService.GetList(userId, vendor2.ID, 0, 20, "", "", "", "name+", 0)
	require.Nil( err, "Unable retrive list of games")
	require.Equal( len(games), 2,"Only 2 games just created")
	require.Equal( games[0].InternalName, gameName,"First game")
	require.Equal( games[1].InternalName, game2Name,"Second game")

	suite.T().Log("Check filter with offset and sort")
	games2, err := gameService.GetList(userId, vendor2.ID, 1, 20, "", "", "", "name-", 0)
	require.Nil( err, "Unable retrive list of games")
	require.Equal( len(games2), 1,"Only 1 retrivied")
	require.Equal( games2[0].InternalName, game2Name,"Second game name")

	suite.T().Log("Check filter with name")
	games3, err := gameService.GetList(userId, vendor2.ID, 0, 20, game2Name, "", "", "name-", 0)
	require.Nil( err, "Unable retrive list of games")
	require.Equal( len(games3), 1,"Only 1 retrivied")
	require.Equal( games3[0].InternalName, game2Name,"Second game name")

	suite.T().Log("Get game list with anther user")
	games4, err := gameService.GetList(user2Id, vendor2.ID, 0, 20, "", "", "", "name+", 0)
	require.NotNil( err, "Must be error")
	require.Nil( games4,"Retrived games is null")

	suite.T().Log("Delete first game")
	err = gameService.Delete(userId, game.ID)
	require.Nil( err, "Game deletion must be without error")

	suite.T().Log("Try to fetch deleted game")
	game4, err := gameService.GetInfo(userId, game.ID)
	require.NotNil( err, "Rise error because game already removed")
	require.Nil( game4, "Game must be null")

	suite.T().Log("Get games list with one game")
	games, err = gameService.GetList(userId, vendor2.ID, 0, 20, "", "", "", "name+", 0)
	require.Nil( err, "Unable retrive list of games")
	require.Equal( len(games), 1,"Only 1 games must be")
	require.Equal( games[0].InternalName, game2Name,"Second game")

	suite.T().Log("Update game")
	game3.Tags = []string{"action", "tank"}
	game3.Developers = "Developers"
	game3.InternalName = gameName + "-x"
	game3.Platforms.Windows = true
	game3.Platforms.Linux = true
	game3.Requirements.Windows.Recommended.Graphics = "4200ti"
	game3.Publishers = "Publishers"
	err = gameService.UpdateInfo(userId, game3)
	require.Nil( err, "Must be save without error")

	suite.T().Log("Retrive updated game")
	game5, err := gameService.GetInfo(userId, game3.ID)
	require.Nil( err, "Error must be null")
	require.Equal( game5.Developers, game3.Developers,"Must be same")
	require.Equal( game5.InternalName, game3.InternalName,"Must be same")
	require.Equal( game5.Platforms.Windows, game3.Platforms.Windows,"Must be same")
	require.Equal( game5.Platforms.Linux, game3.Platforms.Linux,"Must be same")
	require.Equal( game5.Requirements.Windows.Recommended.Graphics, "4200ti","Must be same")
	require.Equal( game5.Publishers, game3.Publishers,"Must be same")
	require.Equal( len(game5.Tags), 2,"Must be same")

	suite.T().Log("Get game descriptions")
	gameDescr, err := gameService.GetDescr(userId, game5.ID)
	require.Nil( err, "Error must be null")
	require.Equal( gameDescr.GameID, game5.ID, "Same as game")

	suite.T().Log("Update game descriptions")
	gameDescr.Reviews = bto.GameReviews{bto.GameReview{
		PressName: "PressName",
		Link: "Link",
		Score: "Score",
		Quote: "Quote",
	}, bto.GameReview{
		PressName: "222",
		Link: "333",
		Score: "444",
		Quote: "555",
	}}
	gameDescr.Socials.Facebook = "Facebook"
	gameDescr.GameSite = "GameSite"
	gameDescr.AdditionalDescription = "AdditionalDescription"
	gameDescr.Description = utils.LocalizedString{
		EN: "eng-descr",
		RU: "ru-descr",
	}
	err = gameService.UpdateDescr(userId, gameDescr)
	require.Nil( err, "Error must be null")

	suite.T().Log("Get updated game description")
	gameDescr2, err := gameService.GetDescr(userId, game5.ID)
	require.Nil( err, "Error must be null")
	require.Equal( len(gameDescr2.Reviews), 2, "Must be two review")
	require.Equal( gameDescr2.Reviews[0].Link, "Link", "Same value")
	require.Equal( gameDescr2.Reviews[1].Quote, "555", "Same value")
	require.Equal( gameDescr2.Socials.Facebook, gameDescr.Socials.Facebook, "Same value")
	require.Equal( gameDescr2.GameSite, gameDescr.GameSite, "Same value")
	require.Equal( gameDescr2.Description.EN, gameDescr.Description.EN, "Same value")

	suite.T().Log("Retrive tags with user", userId.String())
	tags, err := gameService.FindTags(userId, "Стрелялки", 20, 0)
	require.Equal( len(tags), 1, "Must be one match")
	require.Equal( tags[0].ID, "action", "Same value")
}
