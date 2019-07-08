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

type KeyListProviderTestSuite struct {
	suite.Suite
	db *Database

	rightKeyPackage uuid.UUID
	rightKeyStream  uuid.UUID
	wrongKeyPackage uuid.UUID
	wrongKeyStream  uuid.UUID
	provider        model.KeyStreamProvider
}

func Test_KeyListProvider(t *testing.T) {
	suite.Run(t, new(KeyListProviderTestSuite))
}

func (suite *KeyListProviderTestSuite) SetupTest() {
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
		Type:         model.ListKeyStream,
	}
	keyListStreeam.ID = uuid.NewV4()
	suite.rightKeyStream = keyListStreeam.ID
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&keyListStreeam).Error)

	keyPackage := model.KeyPackage{
		Name:          "Test name",
		PackageID:     uuid.NewV4(),
		KeyStreamType: model.ListKeyStream,
	}
	keyPackage.ID = suite.rightKeyPackage
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&keyPackage).Error)

	platformListStreeam := model.KeyStream{
		Type:         model.PlatformKeysStream,
	}
	platformListStreeam.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&platformListStreeam).Error)
	suite.wrongKeyPackage = platformListStreeam.ID

	platformKeyPackage := model.KeyPackage{
		Name:          "Another name",
		PackageID:     uuid.NewV4(),
		KeyStreamType: model.PlatformKeysStream,
	}
	platformKeyPackage.ID = suite.wrongKeyPackage
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&platformKeyPackage).Error)

	suite.db = db
	suite.provider, err = NewKeyListProvider(keyListStreeam.ID, db)
	shouldBe.Nil(err)
}

func (suite *KeyListProviderTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *KeyListProviderTestSuite) TestInitNewProvider() {
	shouldBe := require.New(suite.T())
	_, err := NewKeyListProvider(uuid.NewV4(), suite.db)
	shouldBe.NotNil(err)

	_, err = NewKeyListProvider(suite.wrongKeyStream, suite.db)
	shouldBe.NotNil(err)
}

func (suite *KeyListProviderTestSuite) TestAddKeysList() {
	shouldBe := require.New(suite.T())
	keys := []string{
		model.RandStringRunes(10),
		model.RandStringRunes(10),
	}
	err := suite.provider.AddKeys(keys)
	shouldBe.Nil(err)

	err = suite.provider.AddKeys(keys)
	shouldBe.Nil(err)
}

func (suite *KeyListProviderTestSuite) TestRedeemList() {
	shouldBe := require.New(suite.T())
	// no one keys
	_, err := suite.provider.RedeemList(10)
	shouldBe.NotNil(err)

	for i := 0; i < 10; i++{
		key := model.Key{
			KeyStreamID: suite.rightKeyStream,
			ActivationCode: model.RandStringRunes(10),
		}
		key.ID = uuid.NewV4()
		shouldBe.Nil(suite.db.DB().Create(&key).Error)
	}

	keys, err := suite.provider.RedeemList(10)
	shouldBe.Nil(err)
	shouldBe.NotNil(keys)
	shouldBe.Equal(10, len(keys))

	_, err = suite.provider.RedeemList(10)
	shouldBe.NotNil(err)
}

func (suite *KeyListProviderTestSuite) TestRedeem() {
	shouldBe := require.New(suite.T())
	// no one keys
	_, err := suite.provider.Redeem()
	shouldBe.NotNil(err)

	key := model.Key{
		KeyStreamID: suite.rightKeyStream,
		ActivationCode: model.RandStringRunes(10),
	}
	key.ID = uuid.NewV4()

	shouldBe.Nil(suite.db.DB().Create(&key).Error)
	key1, err := suite.provider.Redeem()
	shouldBe.Nil(err)
	shouldBe.NotNil(key1)
	shouldBe.Equal(key.ActivationCode, key1.ActivationCode)
	shouldBe.NotNil(key1.RedeemTime)

	_, err = suite.provider.Redeem()
	shouldBe.NotNil(err)
}
