package orm_test

import (
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"testing"
)

type VendorServiceTestSuite struct {
	suite.Suite
	db *orm.Database
}

func Test_VendorService(t *testing.T) {
	suite.Run(t, new(VendorServiceTestSuite))
}

func (suite *VendorServiceTestSuite) SetupTest() {
	config, err := conf.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	db.Init()

	suite.db = db
}

func (suite *VendorServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *VendorServiceTestSuite) TestCreateVendorShouldPlaceInDB() {
	require := require.New(suite.T())
	
	vendorService, err := orm.NewVendorService(suite.db)

	userId := uuid.NamespaceDNS

	vendor := model.Vendor{
		ID: uuid.NewV4(),
		Name: "1C",
		Domain3: "godzilla",
		Email: "godzilla@proto.one",
		ManagerID: userId,
	}

	_, err = vendorService.Create(&vendor)
	require.Nil(err, "Unable to create vendor")
	require.NotEmpty(vendor.ID, "Wrong ID for created vendor")

	vendorFromDb, err := vendorService.FindByID(vendor.ID)
	require.Nil(err, "Unable to get vendor: %v", err)
	require.Equal(vendor.ID, vendorFromDb.ID, "Incorrect Vendor ID from DB")
	require.Equal(vendor.Name, vendorFromDb.Name, "Incorrect Vendor Name from DB")
	require.Equal(vendor.Email, vendorFromDb.Email, "Incorrect Vendor Email from DB")
	require.Equal(vendor.Domain3, vendorFromDb.Domain3, "Incorrect Vendor Domain3 from DB")

	vendor.Domain3 = "zillo"
	_, err = vendorService.Update(&vendor)
	require.Nil(err, "Unable to update vendor: %v", err)

	vendorFromDb2, err := vendorService.FindByID(vendor.ID)
	require.Nil(err, "Unable to get vendor: %v", err)
	require.Equal(vendor.Domain3, vendorFromDb2.Domain3, "Incorrect updated Vendor Domain3 from DB")

	vendor2 := model.Vendor{
		Name: "domino",
		Domain3: "2domino",
		Email: "domino@proto.com",
		ManagerID: userId,
	}
	_, err = vendorService.Create(&vendor2)
	require.NotNil(err, "Must be error cuz wrong domain name")

	vendor3 := model.Vendor{
		Name: "domino",
		Domain3: "domino",
		Email: "4456",
		ManagerID: userId,
	}
	_, err = vendorService.Create(&vendor3)
	require.NotNil(err, "Must be error cuz invalid email")
}
