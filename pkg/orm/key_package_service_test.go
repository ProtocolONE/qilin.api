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

type KeyPackageServiceTestSuite struct {
	suite.Suite
	db      *Database
	service model.KeyPackageService

	rightKeyPackage uuid.UUID
	packageId       uuid.UUID
	wrongKeyPackage uuid.UUID
}

func Test_KeyPackageService(t *testing.T) {
	suite.Run(t, new(KeyPackageServiceTestSuite))
}

func (suite *KeyPackageServiceTestSuite) SetupTest() {
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
	shouldBe.Nil(db.DB().Model(model.KeyStream{}).Create(&keyListStreeam).Error)
	gamePackage := model.Package{

	}
	gamePackage.ID = uuid.NewV4()
	shouldBe.Nil(db.DB().Model(model.Package{}).Create(&gamePackage).Error)
	suite.packageId = gamePackage.ID

	keyPackage := model.KeyPackage{
		KeyStreamID:   keyListStreeam.ID,
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

	platformKeyPackage := model.KeyPackage{
		KeyStreamID:   platformListStreeam.ID,
		Name:          "Another name",
		PackageID:     uuid.NewV4(),
		KeyStreamType: model.PlatformKeysStream,
	}
	platformKeyPackage.ID = suite.wrongKeyPackage
	shouldBe.Nil(db.DB().Model(model.KeyPackage{}).Create(&platformKeyPackage).Error)

	suite.db = db
	suite.service = NewKeyPackageService(db, NewKeyStreamService(db))
}

func (suite *KeyPackageServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *KeyPackageServiceTestSuite) TestGet() {
	shouldBe := require.New(suite.T())
	keyPackage, err := suite.service.Get(suite.rightKeyPackage)
	shouldBe.Nil(err)
	shouldBe.NotNil(keyPackage)

	keyPackage, err = suite.service.Get(uuid.NewV4())
	shouldBe.NotNil(err)
	shouldBe.Nil(keyPackage)
}

func (suite *KeyPackageServiceTestSuite) TestCreate() {
	shouldBe := require.New(suite.T())
	keyPackage, err := suite.service.Create(suite.packageId, "Some name", model.ListKeyStream)
	shouldBe.Nil(err)
	shouldBe.NotNil(keyPackage)

	keyPackage, err = suite.service.Create(suite.packageId, "Some name 2", model.ListKeyStream)
	shouldBe.Nil(err)
	shouldBe.NotNil(keyPackage)

	keyPackage, err = suite.service.Create(suite.packageId, "Some name 3", model.PlatformKeysStream)
	shouldBe.Nil(err)
	shouldBe.NotNil(keyPackage)

	keyPackage, err = suite.service.Create(suite.packageId, "Ooops", "Unknown_key_stream")
	shouldBe.NotNil(err)
	shouldBe.Nil(keyPackage)

	keyPackage, err = suite.service.Create(uuid.NewV4(), "Not found", model.PlatformKeysStream)
	shouldBe.NotNil(err)
	shouldBe.Nil(keyPackage)

	keyPackage, err = suite.service.Create(suite.packageId, "", model.ListKeyStream)
	shouldBe.NotNil(err)
	shouldBe.Nil(keyPackage)
}

func (suite *KeyPackageServiceTestSuite) TestList() {
	shouldBe := require.New(suite.T())
	keyPackages, err := suite.service.List(uuid.NewV4())
	shouldBe.Nil(err)
	shouldBe.NotNil(keyPackages)
	shouldBe.Empty(keyPackages)

	keyPackage, err := suite.service.Create(suite.packageId, "Some name", model.ListKeyStream)
	shouldBe.Nil(err)
	shouldBe.NotNil(keyPackage)

	keyPackages, err = suite.service.List(suite.packageId)
	shouldBe.NotNil(keyPackages)
	shouldBe.Nil(err)
}

func (suite *KeyPackageServiceTestSuite) TestUpdate() {
	shouldBe := require.New(suite.T())
	keyPackage, err := suite.service.Update(uuid.NewV4(), "Not found")
	shouldBe.NotNil(err)
	shouldBe.Nil(keyPackage)

	keyPackage, err = suite.service.Update(suite.rightKeyPackage, "Another name")
	shouldBe.Nil(err)
	shouldBe.NotNil(keyPackage)
}