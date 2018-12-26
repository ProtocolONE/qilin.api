package orm_test

import (
	"github.com/stretchr/testify/assert"
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
	dbConfig := conf.Database{
		Host:     "localhost",
		Port:     "5440",
		Database: "test_qilin",
		User:     "postgres",
		Password: "",
	}

	db, err := orm.NewDatabase(&dbConfig)
	if err != nil {
		suite.Fail("Unable to connect to database: %s", err)
	}

	db.Init()

	suite.db = db
}

func (suite *VendorServiceTestSuite) TearDownTest() {
	if err := suite.db.DB().DropTable(model.Vendor{}).Error; err != nil {
		panic(err)
	}
	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *VendorServiceTestSuite) TestCreateVendorShouldPlaceInDB() {
	vendorService, err := orm.NewVendorService(suite.db)

	vendor := model.Vendor{
		Name: "1C",
		Domain3: "godzilla",
		Email: "godzilla@proto.one",
		ManagerId: 1,
	}

	err = vendorService.CreateVendor(&vendor)
	assert.Nil(suite.T(), err, "Unable to create vendor")
	assert.NotEmpty(suite.T(), vendor.ID, "Wrong ID for created vendor")

	vendorFromDb, err := vendorService.FindByID(vendor.ID)
	assert.Nil(suite.T(), err, "Unable to get vendor: %v", err)
	assert.Equal(suite.T(), vendor.ID, vendorFromDb.ID, "Incorrect Vendor ID from DB")
	assert.Equal(suite.T(), vendor.Name, vendorFromDb.Name, "Incorrect Vendor Name from DB")
	assert.Equal(suite.T(), vendor.Email, vendorFromDb.Email, "Incorrect Vendor Email from DB")
	assert.Equal(suite.T(), vendor.Domain3, vendorFromDb.Domain3, "Incorrect Vendor Domain3 from DB")

	vendor.Domain3 = "zillo"
	err = vendorService.UpdateVendor(&vendor)
	assert.Nil(suite.T(), err, "Unable to update vendor: %v", err)

	vendorFromDb2, err := vendorService.FindByID(vendor.ID)
	assert.Nil(suite.T(), err, "Unable to get vendor: %v", err)
	assert.Equal(suite.T(), vendor.Domain3, vendorFromDb2.Domain3, "Incorrect updated Vendor Domain3 from DB")

	vendor2 := model.Vendor{
		Name: "domino",
		Domain3: "2domino",
		Email: "domino@proto.com",
		ManagerId: 1,
	}
	err = vendorService.CreateVendor(&vendor2)
	assert.NotNil(suite.T(), err, "Must be error cuz wrong domain name")

	vendor3 := model.Vendor{
		Name: "domino",
		Domain3: "domino",
		Email: "4456",
		ManagerId: 1,
	}
	err = vendorService.CreateVendor(&vendor3)
	assert.NotNil(suite.T(), err, "Must be error cuz invalid email")
}
