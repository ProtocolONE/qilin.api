package orm_test

import (
	"github.com/ProtocolONE/rbac"
	"github.com/stretchr/testify/assert"
	"net/http"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/test"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AdminOnboardingServiceTestSuite struct {
	suite.Suite
	db      *orm.Database
	service *orm.AdminOnboardingService
}

func Test_AdminOnboardingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AdminOnboardingServiceTestSuite))
}

func (suite *AdminOnboardingServiceTestSuite) SetupTest() {
	config, err := qilin_test.LoadTestConfig()
	if err != nil {
		suite.FailNow("Unable to load config", "%v", err)
	}

	db, err := orm.NewDatabase(&config.Database)
	if err != nil {
		suite.Fail("Unable to connect to database:", "%v", err)
	}

	if err := db.DropAllTables(); err != nil {
		assert.FailNow(suite.T(), "Unable to drop tables", err)
	}
	if err := db.Init(); err != nil {
		assert.FailNow(suite.T(), "Unable to init tables", err)
	}

	suite.db = db

	service, err := orm.NewAdminOnboardingService(suite.db)
	if err != nil {
		suite.Fail("Unable to create service", "%v", err)
	}

	suite.service = service

	user := model.User{
		ID:       uuid.NewV4().String(),
		Login:    "test@protocol.one",
		Password: "megapass",
		Nickname: "Test",
		Lang:     "ru",
	}

	err = db.DB().Create(&user).Error
	suite.Nil(err, "Unable to create user")

	userId := user.ID

	ownProvider := orm.NewOwnerProvider(suite.db)
	enf := rbac.NewEnforcer()
	membershipService := orm.NewMembershipService(suite.db, ownProvider, enf)

	vendorService, err := orm.NewVendorService(db, membershipService)
	suite.Nil(err, "Unable make vendor service")

	vendor := model.Vendor{
		ID:              uuid.NewV4(),
		Name:            "domino",
		Domain3:         "domino",
		Email:           "domino@proto.com",
		HowManyProducts: "+1000",
		ManagerID:       userId,
	}
	_, err = vendorService.Create(&vendor)
	suite.Nil(err, "Must create new vendor")

	vendor2 := model.Vendor{
		ID:              uuid.FromStringOrNil("5862ead5-acf5-4092-a7bc-a645f279096d"),
		Name:            "domino1",
		Domain3:         "domino2",
		Email:           "domino3@proto.com",
		HowManyProducts: "+10004",
		ManagerID:       userId,
	}
	_, err = vendorService.Create(&vendor2)

	id, _ := uuid.FromString(GameID)
	game := model.Game{}
	game.ID = id
	game.InternalName = "internalName"
	game.FeaturesCtrl = ""
	game.FeaturesCommon = []string{}
	game.Platforms = bto.Platforms{}
	game.Requirements = bto.GameRequirements{}
	game.Languages = bto.GameLangs{}
	game.FeaturesCommon = []string{}
	game.GenreMain = 1
	game.GenreAddition = []int64{1, 2}
	game.Tags = []int64{1, 2}
	game.VendorID = vendor.ID
	game.CreatorID = userId

	err = db.DB().Create(&game).Error

	if err != nil {
		suite.Fail("Unable to create game", "%v", err)
	}

	vendorDocumentsDraft := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "MEGA TEST",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Эдуард Никифоров",
				"Position": "Руководитель",
			},
			"Technical": model.JSONB{
				"FullName": "Роман Обрамович",
				"Position": "Батрак",
			},
		},
		Status:       model.StatusDraft,
		ReviewStatus: model.ReviewNew,
		Banking: model.JSONB{
			"Currency": "USD",
		},
	}
	vendorDocumentsDraft.ID = uuid.NewV4()

	vendorDocuments := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "MEGA TEST",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Эдуард Никифоров",
				"Position": "Руководитель",
			},
			"Technical": model.JSONB{
				"FullName": "Роман Обрамович",
				"Position": "Батрак",
			},
		},
		Status:       model.StatusOnReview,
		ReviewStatus: model.ReviewChecking,
		Banking: model.JSONB{
			"Currency": "USD",
		},
	}
	vendorDocuments.ID = uuid.NewV4()

	vendorDocuments2 := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "PUBG TEST",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Филимонов Андрей",
				"Position": "IT Director",
			},
		},
		Status:       model.StatusApproved,
		ReviewStatus: model.ReviewApproved,
		Banking: model.JSONB{
			"Currency": "EUR",
		},
	}
	vendorDocuments2.ID = uuid.NewV4()

	vendorDocuments3 := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "Ash of Evils ",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Lucifer",
				"Position": "CEO",
			},
		},
		Status:       model.StatusDeclined,
		ReviewStatus: model.ReviewReturned,
		Banking: model.JSONB{
			"Currency": "USD",
		},
	}
	vendorDocuments3.ID = uuid.NewV4()

	for i := 0; i < 10; i++ {
		vendorDocuments4 := model.DocumentsInfo{
			VendorID: id,
			Company: model.JSONB{
				"Name":            "ZTEST2",
				"AlternativeName": "Alt MEGA NAME",
				"Country":         "RUSSIA",
			},
			Contact: model.JSONB{
				"Authorized": model.JSONB{
					"FullName": "Test Name",
					"Position": "Test Position",
				},
				"Technical": model.JSONB{
					"FullName": "Test Name",
					"Position": "Test Position",
				},
			},
			Status:       model.StatusOnReview,
			ReviewStatus: model.ReviewNew,
			Banking: model.JSONB{
				"Currency": "USD",
			},
		}
		vendorDocuments4.ID = uuid.NewV4()
		suite.Nil(db.DB().Create(&vendorDocuments4).Error)
	}

	suite.Nil(db.DB().Create(&vendorDocuments).Error)
	suite.Nil(db.DB().Create(&vendorDocuments2).Error)
	suite.Nil(db.DB().Create(&vendorDocuments3).Error)
	suite.Nil(db.DB().Create(&vendorDocumentsDraft).Error)
}

func (suite *AdminOnboardingServiceTestSuite) TearDownTest() {
	if err := suite.db.DropAllTables(); err != nil {
		panic(err)
	}

	if err := suite.db.Close(); err != nil {
		panic(err)
	}
}

func (suite *AdminOnboardingServiceTestSuite) TestChangeStatus() {
	should := require.New(suite.T())
	id := uuid.FromStringOrNil("5862ead5-acf5-4092-a7bc-a645f279096d")

	vendorDocuments := model.DocumentsInfo{
		VendorID: id,
		Company: model.JSONB{
			"Name":            "MEGA TEST",
			"AlternativeName": "Alt MEGA NAME",
			"Country":         "RUSSIA",
		},
		Contact: model.JSONB{
			"Authorized": model.JSONB{
				"FullName": "Эдуард Никифоров",
				"Position": "Руководитель",
			},
			"Technical": model.JSONB{
				"FullName": "Роман Обрамович",
				"Position": "Батрак",
			},
		},
		Status:       model.StatusOnReview,
		ReviewStatus: model.ReviewNew,
		Banking: model.JSONB{
			"Currency": "USD",
		},
	}
	vendorDocuments.ID = uuid.NewV4()
	should.Nil(suite.db.DB().Create(&vendorDocuments).Error)
	fromDb := model.DocumentsInfo{}

	err := suite.service.ChangeStatus(id, model.ReviewApproved)
	should.Nil(err)
	should.Nil(suite.db.DB().Model(&vendorDocuments).Where("id = ?", vendorDocuments.ID).First(&fromDb).Error)
	should.Equal(model.StatusApproved, fromDb.Status)
	should.Equal(model.ReviewApproved, fromDb.ReviewStatus)

	err = suite.service.ChangeStatus(id, model.ReviewReturned)
	should.Nil(err)
	should.Nil(suite.db.DB().Model(&vendorDocuments).First(&fromDb).Error)
	should.Equal(model.StatusDeclined, fromDb.Status)
	should.Equal(model.ReviewReturned, fromDb.ReviewStatus)

	err = suite.service.ChangeStatus(id, model.ReviewChecking)
	should.Nil(err)
	should.Nil(suite.db.DB().Model(&vendorDocuments).First(&fromDb).Error)
	should.Equal(model.StatusOnReview, fromDb.Status)
	should.Equal(model.ReviewChecking, fromDb.ReviewStatus)

	err = suite.service.ChangeStatus(id, model.ReviewArchived)
	should.Nil(err)
	should.Nil(suite.db.DB().Model(&vendorDocuments).First(&fromDb).Error)
	should.Equal(model.StatusArchived, fromDb.Status)
	should.Equal(model.ReviewArchived, fromDb.ReviewStatus)

	err = suite.service.ChangeStatus(id, model.ReviewUndefined)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)
	should.Nil(suite.db.DB().Model(&vendorDocuments).First(&fromDb).Error)
	should.Equal(model.StatusArchived, fromDb.Status)
	should.Equal(model.ReviewArchived, fromDb.ReviewStatus)

	err = suite.service.ChangeStatus(id, model.ReviewNew)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)
	should.Nil(suite.db.DB().Model(&vendorDocuments).First(&fromDb).Error)
	should.Equal(model.StatusArchived, fromDb.Status)
	should.Equal(model.ReviewArchived, fromDb.ReviewStatus)

	err = suite.service.ChangeStatus(uuid.NewV4(), model.ReviewNew)
	should.NotNil(err)
	should.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)

	vendorDocuments.Status = model.StatusDraft
	vendorDocuments.ReviewStatus = model.ReviewNew
	should.Nil(suite.db.DB().Save(&vendorDocuments).Error)
	err = suite.service.ChangeStatus(id, model.ReviewNew)
	should.NotNil(err)
	should.Equal(http.StatusBadRequest, err.(*orm.ServiceError).Code)
	should.Nil(suite.db.DB().Model(&vendorDocuments).First(&fromDb).Error)
	should.Equal(model.StatusDraft, fromDb.Status)
	should.Equal(model.ReviewNew, fromDb.ReviewStatus)
}

func (suite *AdminOnboardingServiceTestSuite) TestSearching() {
	should := require.New(suite.T())

	requests, count, err := suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))

	requests, count, err = suite.service.GetRequests(100, 10, "", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(3, len(requests))

	requests, count, err = suite.service.GetRequests(100, 100, "", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(0, len(requests))

	for i := 1; i <= 10; i++ {
		requests, count, err = suite.service.GetRequests(i, 0, "", model.ReviewUndefined, "")
		should.Nil(err)
		should.NotNil(requests)
		should.Equal(i, len(requests))
	}

	requests, count, err = suite.service.GetRequests(100, 0, "MEGA", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(1, len(requests))
	should.Equal("MEGA TEST", requests[0].Company["Name"])

	requests, count, err = suite.service.GetRequests(100, 0, "mega", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(1, len(requests))
	should.Equal("MEGA TEST", requests[0].Company["Name"])

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "-status")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))
	should.Equal(model.ReviewReturned, requests[0].ReviewStatus)
	should.Equal(model.ReviewChecking, requests[1].ReviewStatus)
	should.Equal(model.ReviewApproved, requests[2].ReviewStatus)
	for i := 3; i < len(requests); i++ {
		should.Equal(model.ReviewNew, requests[i].ReviewStatus)
	}

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "+status")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))
	for i := 0; i < 10; i++ {
		should.Equal(model.ReviewNew, requests[i].ReviewStatus)
	}
	should.Equal(model.ReviewReturned, requests[12].ReviewStatus)
	should.Equal(model.ReviewChecking, requests[11].ReviewStatus)
	should.Equal(model.ReviewApproved, requests[10].ReviewStatus)

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "+name")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))
	should.Equal("Ash of Evils ", requests[0].Company["Name"])
	should.Equal("MEGA TEST", requests[1].Company["Name"])
	should.Equal("PUBG TEST", requests[2].Company["Name"])
	for i := 3; i < len(requests); i++ {
		should.Equal("ZTEST2", requests[i].Company["Name"])
	}

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "-name")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))
	for i := 0; i < 10; i++ {
		should.Equal("ZTEST2", requests[i].Company["Name"])
	}
	should.Equal("Ash of Evils ", requests[12].Company["Name"])
	should.Equal("MEGA TEST", requests[11].Company["Name"])
	should.Equal("PUBG TEST", requests[10].Company["Name"])

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "-updatedAt")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))

	for i := 0; i > len(requests)-1; i++ {
		should.True(requests[i].UpdatedAt.After(requests[i+1].UpdatedAt))
	}

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "+updatedAt")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))

	for i := 0; i > len(requests)-1; i++ {
		should.True(requests[i].UpdatedAt.Before(requests[i+1].UpdatedAt))
	}

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewNew, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(10, len(requests))

	for i := 0; i > len(requests); i++ {
		should.Equal(model.ReviewNew, requests[i].ReviewStatus)
	}

	requests, count, err = suite.service.GetRequests(100, 5, "", model.ReviewNew, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(5, len(requests))

	for i := 0; i > len(requests); i++ {
		should.Equal(model.ReviewNew, requests[i].ReviewStatus)
	}

	requests2, count, err := suite.service.GetRequests(5, 0, "", model.ReviewNew, "")
	should.Nil(err)
	should.NotNil(requests2)
	should.Equal(5, len(requests2))

	for i := 0; i > len(requests2); i++ {
		should.Equal(model.ReviewNew, requests2[i].ReviewStatus)
		should.NotEqual(requests[i].ID, requests2[i].ID)
	}

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewChecking, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(1, len(requests))

	for i := 0; i > len(requests); i++ {
		should.Equal(model.ReviewChecking, requests[i].ReviewStatus)
	}

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewApproved, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(1, len(requests))

	for i := 0; i > len(requests); i++ {
		should.Equal(model.ReviewApproved, requests[i].ReviewStatus)
	}

	requests, count, err = suite.service.GetRequests(100, 0, "", model.ReviewReturned, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(1, len(requests))
	should.Equal(1, count)

	for i := 0; i > len(requests); i++ {
		should.Equal(model.ReviewReturned, requests[i].ReviewStatus)
	}
}
