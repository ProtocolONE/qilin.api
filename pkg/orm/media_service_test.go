package orm_test

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"math/rand"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		fmt.Println(err)
	}
	if err := db.Init(); err != nil {
		fmt.Println(err)
	}

	id, _ := uuid.FromString(Id)
	err = db.DB().Save(&model.Game{
		ID:             id,
		InternalName:   "Test_game_3",
		ReleaseDate:    time.Now(),
		GenreAddition:  pq.Int64Array{},
		Tags:           pq.Int64Array{},
		FeaturesCommon: pq.StringArray{},
	}).Error
	assert.Nil(suite.T(), err, "Unable to make game")

	suite.db = db
}

func (suite *MediaServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
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
		CoverImage: model.JSONB{
			"ru": RandStringRunes(10),
			"en": RandStringRunes(10),
		},
		Trailers: model.JSONB{
			"ru": []string{RandStringRunes(10), RandStringRunes(10)},
			"en": []string{RandStringRunes(10), RandStringRunes(10)},
		},
		Screenshots: model.JSONB{
			"ru": []string{RandStringRunes(10), RandStringRunes(10)},
			"en": []string{RandStringRunes(10), RandStringRunes(10)},
		},
		Store: model.JSONB{
			"ru": RandStringRunes(10),
			"en": RandStringRunes(10),
		},
		CoverVideo: model.JSONB{
			"ru": RandStringRunes(10),
			"en": RandStringRunes(10),
		},
		Capsule: model.JSONB{
			"generic": map[string]interface{}{
				"ru": RandStringRunes(10),
				"en": RandStringRunes(10),
			},
			"small": map[string]interface{}{
				"ru": RandStringRunes(10),
				"en": RandStringRunes(10),
			},
		},
		UpdatedAt: time.Now(),
	}

	err = mediaService.Update(id, &game)
	assert.Nil(suite.T(), err, "Unable to update media for game")

	gameFromDb, err := mediaService.Get(id)
	assert.Nil(suite.T(), err, "Unable to get game: %v", err)
	assert.Equal(suite.T(), game.ID, gameFromDb.ID, "Incorrect Game ID from DB")
	assert.Equal(suite.T(), game.Capsule, gameFromDb.Capsule, "Incorrect capsule from DB")
	assert.Equal(suite.T(), game.CoverImage, gameFromDb.CoverImage, "Incorrect CoverImage from DB")
	assert.Equal(suite.T(), game.CoverVideo, gameFromDb.CoverVideo, "Incorrect CoverVideo from DB")
	assert.Equal(suite.T(), game.Store, gameFromDb.Store, "Incorrect Store from DB")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
