package orm_test

import (
	"github.com/ProtocolONE/rbac"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"qilin-api/pkg/api/mock"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
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
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}
	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.FailNow("Unable to connect to database", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

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
	req := require.New(suite.T())

	ownProvider := orm.NewOwnerProvider(suite.db)
	enf := rbac.NewEnforcer()
	memServide := orm.NewMembershipService(suite.db, ownProvider, enf, mock.NewMailer(), "")
	vendorService, err := orm.NewVendorService(suite.db, memServide)

	userId := uuid.NamespaceDNS.String()

	vendor := model.Vendor{
		ID:        uuid.NewV4(),
		Name:      "1C",
		Domain3:   "godzilla",
		Email:     "godzilla@proto.one",
		ManagerID: userId,
	}

	_, err = vendorService.Create(&vendor)
	req.Nil(err, "Unable to create vendor")
	req.NotEmpty(vendor.ID, "Wrong ID for created vendor")

	vendorFromDb, err := vendorService.FindByID(vendor.ID)
	req.Nil(err, "Unable to get vendor: %v", err)
	req.Equal(vendor.ID, vendorFromDb.ID, "Incorrect Vendor ID from DB")
	req.Equal(vendor.Name, vendorFromDb.Name, "Incorrect Vendor Name from DB")
	req.Equal(vendor.Email, vendorFromDb.Email, "Incorrect Vendor Email from DB")
	req.Equal(vendor.Domain3, vendorFromDb.Domain3, "Incorrect Vendor Domain3 from DB")

	vendor.Domain3 = "zillo"
	_, err = vendorService.Update(&vendor)
	req.Nil(err, "Unable to update vendor: %v", err)

	vendorFromDb2, err := vendorService.FindByID(vendor.ID)
	req.Nil(err, "Unable to get vendor: %v", err)
	req.Equal(vendor.Domain3, vendorFromDb2.Domain3, "Incorrect updated Vendor Domain3 from DB")

	vendor2 := model.Vendor{
		Name:      "domino",
		Domain3:   "2domino",
		Email:     "domino@proto.com",
		ManagerID: userId,
	}
	_, err = vendorService.Create(&vendor2)
	req.NotNil(err, "Must be error cuz wrong domain name")

	vendor3 := model.Vendor{
		Name:      "domino",
		Domain3:   "domino",
		Email:     "4456",
		ManagerID: userId,
	}
	_, err = vendorService.Create(&vendor3)
	req.NotNil(err, "Must be error cuz invalid email")
}
