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

type KeyStreamServiceTestSuite struct {
	suite.Suite
	db      *Database
	service model.KeyStreamService

	platformStreamId uuid.UUID
	keyListStreamId  uuid.UUID
	wrongStreamId    uuid.UUID
	packageId        uuid.UUID
}

func Test_KeyStreamService(t *testing.T) {
	suite.Run(t, new(KeyStreamServiceTestSuite))
}

func (suite *KeyStreamServiceTestSuite) SetupTest() {
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

	keyListStreeam := model.KeyStream{
		Type:         model.ListKeyStream,
	}
	keyListStreeam.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&keyListStreeam).Error)
	suite.keyListStreamId = keyListStreeam.ID

	gamePackage := model.Package{}
	gamePackage.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Create(&gamePackage).Error)

	keyPackage := model.KeyPackage{
		KeyStreamID:   keyListStreeam.ID,
		Name:          "Test name",
		PackageID:     uuid.NewV4(),
		KeyStreamType: model.ListKeyStream,
	}
	keyPackage.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&keyPackage).Error)
	suite.packageId = keyPackage.ID

	platformListStreeam := model.KeyStream{
		Type:         model.PlatformKeysStream,
	}
	platformListStreeam.ID = uuid.NewV4()
	suite.platformStreamId = platformListStreeam.ID
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&platformListStreeam).Error)

	wrongListStreeam := model.KeyStream{
		Type:         "unknown_type",
	}
	wrongListStreeam.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&wrongListStreeam).Error)
	suite.wrongStreamId = wrongListStreeam.ID

	platformKeyPackage := model.KeyPackage{
		KeyStreamID:   platformListStreeam.ID,
		Name:          "Another name",
		PackageID:     uuid.NewV4(),
		KeyStreamType: model.PlatformKeysStream,
	}
	platformKeyPackage.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&platformKeyPackage).Error)

	suite.db = db
	suite.service = NewKeyStreamService(db)
}

func (suite *KeyStreamServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *KeyStreamServiceTestSuite) TestGet() {
	shouldBe := require.New(suite.T())

	provider, err := suite.service.Get(uuid.NewV4())
	shouldBe.NotNil(err)
	shouldBe.Nil(provider)

	provider, err = suite.service.Get(suite.keyListStreamId)
	shouldBe.NotNil(provider)
	shouldBe.Nil(err)

	provider, err = suite.service.Get(suite.platformStreamId)
	shouldBe.NotNil(provider)
	shouldBe.Nil(err)

	provider, err = suite.service.Get(suite.wrongStreamId)
	shouldBe.NotNil(err)
	shouldBe.Nil(provider)
}

func (suite *KeyStreamServiceTestSuite) TestCreate() {
	shouldBe := require.New(suite.T())

	streamId, err := suite.service.Create(model.ListKeyStream)
	shouldBe.Nil(err)
	shouldBe.NotEqual(uuid.Nil, streamId)

	streamId, err = suite.service.Create("unknown_type")
	shouldBe.NotNil(err)
	shouldBe.Equal(uuid.Nil, streamId)
}
