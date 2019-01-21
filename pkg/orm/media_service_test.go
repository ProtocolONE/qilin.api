package orm_test

import (
	"math/rand"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"
)

type MediaServiceTestSuite struct {
	suite.Suite
	db *orm.Database
}

var (
	Id = "029ce039-888a-481a-a831-cde7ff4e50b8"
)

func Test_MediaService(t *testing.T) {
	suite.Run(t, new(MediaServiceTestSuite))
}

func (suite *MediaServiceTestSuite) SetupTest() {
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

	id, _ := uuid.FromString(Id)
	db.DB().Save(&model.Game{ID: id})

	suite.db = db
}

func (suite *MediaServiceTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Media{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}
func (suite *MediaServiceTestSuite) TestCreateMediaShouldChangeGameInDB() {
	mediaService, err := orm.NewMediaService(suite.db)

	assert.Nil(suite.T(), err, "Unable to media service")

	id, _ := uuid.FromString(Id)
	game := model.Media{
		ID: uuid.NewV4(),
		CoverImage: model.JSONB {
			"ru": RandStringRunes(10),
			"en": RandStringRunes(10),
		},
		Trailers: model.JSONB {
			"ru": RandStringRunes(10),
			"en": RandStringRunes(10),
		},
		Store:  model.JSONB {
			"ru": RandStringRunes(10),
			"en": RandStringRunes(10),
		},
		CoverVideo:  model.JSONB {
			"ru": RandStringRunes(10),
			"en": RandStringRunes(10),
		},
		Capsule: model.JSONB {
			"generic": map[string]interface {} {
				"ru": RandStringRunes(10),
				"en": RandStringRunes(10),
			},
			"small": map[string]interface {} {
				"ru": RandStringRunes(10),
				"en": RandStringRunes(10),
			},
		},
	}

	err = mediaService.Update(id, &game);
	assert.Nil(suite.T(), err, "Unable to update media for game")

	gameFromDb, err := mediaService.Get(id)
	assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	assert.Equal(suite.T(), game.ID, gameFromDb.ID, "Incorrect Game ID from DB")
	assert.Equal(suite.T(), game.Capsule, gameFromDb.Capsule, "Incorrect capsule from DB")
	assert.Equal(suite.T(), game.CoverImage, gameFromDb.CoverImage, "Incorrect CoverImage from DB")
	assert.Equal(suite.T(), game.CoverVideo, gameFromDb.CoverVideo, "Incorrect CoverVideo from DB")
	assert.Equal(suite.T(), game.Store, gameFromDb.Store, "Incorrect Store from DB")
	assert.Equal(suite.T(), game.Trailers, gameFromDb.Trailers, "Incorrect Trailers from DB")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}