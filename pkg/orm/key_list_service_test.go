package orm

import (
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"qilin-api/pkg/model"
	qilintest "qilin-api/pkg/test"
	"testing"
)

type KeyListServiceTestSuite struct {
	suite.Suite
	db      *Database
	service model.KeyListService

	rightKeyPackage uuid.UUID
	wrongKeyPackage uuid.UUID
}

func Test_KeyListService(t *testing.T) {
	suite.Run(t, new(KeyListServiceTestSuite))
}

func (suite *KeyListServiceTestSuite) SetupTest() {
	shouldBe := require.New(suite.T())
	config, err := qilintest.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

	suite.rightKeyPackage = uuid.NewV4()
	suite.wrongKeyPackage = uuid.NewV4()

	keyListStreeam := model.KeyStream{
		Type: model.ListKeyStream,
	}
	keyListStreeam.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&keyListStreeam).Error)

	keyPackage := model.KeyPackage{
		KeyStreamID: keyListStreeam.ID,
		Name: "Test name",
		PackageID: uuid.NewV4(),
		KeyStreamType: model.ListKeyStream,
	}
	keyPackage.ID = suite.rightKeyPackage
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&keyPackage).Error)

	platformListStreeam := model.KeyStream{
		Type: model.PlatformKeysStream,
	}
	platformListStreeam.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&platformListStreeam).Error)

	platformKeyPackage := model.KeyPackage{
		KeyStreamID: platformListStreeam.ID,
		Name: "Another name",
		PackageID: uuid.NewV4(),
		KeyStreamType: model.PlatformKeysStream,
	}
	platformKeyPackage.ID = suite.wrongKeyPackage
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&platformKeyPackage).Error)

	suite.db = db
	suite.service = NewKeyListService(db)
}

func (suite *KeyListServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *KeyListServiceTestSuite) TestAddKeys() {
	shouldBe := require.New(suite.T())
	keys := []string{
		model.RandStringRunes(10),
		model.RandStringRunes(10),
		model.RandStringRunes(10),
		model.RandStringRunes(10),
	}

	shouldBe.Nil(suite.service.AddKeys(suite.rightKeyPackage, keys))
	// second call with same keys should not occurs error
	shouldBe.Nil(suite.service.AddKeys(suite.rightKeyPackage, keys))
	// empty keys - no error
	shouldBe.Nil(suite.service.AddKeys(suite.rightKeyPackage, []string{}))

	shouldBe.NotNil(suite.service.AddKeys(suite.wrongKeyPackage, keys))
	shouldBe.NotNil(suite.service.AddKeys(uuid.NewV4(), keys))
}
