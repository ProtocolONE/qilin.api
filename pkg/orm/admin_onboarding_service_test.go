package orm_test

import (
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

	_ = db.DropAllTables()
	db.Init()

	suite.db = db

	service, err := orm.NewAdminOnboardingService(suite.db)
	if err != nil {
		suite.Fail("Unable to create service", "%v", err)
	}

	suite.service = service

	user := model.User{
		ID:       uuid.NewV4(),
		Login:    "test@protocol.one",
		Password: "megapass",
		Nickname: "Test",
		Lang:     "ru",
	}

	err = db.DB().Create(&user).Error
	suite.Nil(err, "Unable to create user")

	userId := user.ID

	vendorService, err := orm.NewVendorService(db)
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
	game.Genre = []string{}
	game.Tags = []string{}
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

func (suite *AdminOnboardingServiceTestSuite) TestSearching() {
	should := require.New(suite.T())

	requests, err := suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))

	requests, err = suite.service.GetRequests(100, 10, "", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(3, len(requests))

	requests, err = suite.service.GetRequests(100, 100, "", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(0, len(requests))

	for i := 1; i <= 10; i++ {
		requests, err = suite.service.GetRequests(i, 0, "", model.ReviewUndefined, "")
		should.Nil(err)
		should.NotNil(requests)
		should.Equal(i, len(requests))
	}

	requests, err = suite.service.GetRequests(100, 0, "MEGA", model.ReviewUndefined, "")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(1, len(requests))

	requests, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "-status")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))
	should.Equal(model.ReviewReturned, requests[0].ReviewStatus)
	should.Equal(model.ReviewChecking, requests[1].ReviewStatus)
	should.Equal(model.ReviewApproved, requests[2].ReviewStatus)
	for i := 3; i < len(requests); i++ {
		should.Equal(model.ReviewNew, requests[i].ReviewStatus)
	}

	requests, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "+status")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))
	for i := 0; i < 10; i++ {
		should.Equal(model.ReviewNew, requests[i].ReviewStatus)
	}
	should.Equal(model.ReviewReturned, requests[12].ReviewStatus)
	should.Equal(model.ReviewChecking, requests[11].ReviewStatus)
	should.Equal(model.ReviewApproved, requests[10].ReviewStatus)

	requests, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "+name")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))
	should.Equal("Ash of Evils ", requests[0].Company["Name"])
	should.Equal("MEGA TEST", requests[1].Company["Name"])
	should.Equal("PUBG TEST", requests[2].Company["Name"])
	for i := 3; i < len(requests); i++ {
		should.Equal("ZTEST2", requests[i].Company["Name"])
	}

	requests, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "-name")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))
	for i := 0; i < 10; i++ {
		should.Equal("ZTEST2", requests[i].Company["Name"])
	}
	should.Equal("Ash of Evils ", requests[12].Company["Name"])
	should.Equal("MEGA TEST", requests[11].Company["Name"])
	should.Equal("PUBG TEST", requests[10].Company["Name"])

	requests, err = suite.service.GetRequests(100, 0, "", model.ReviewUndefined, "+updatedAt")
	should.Nil(err)
	should.NotNil(requests)
	should.Equal(13, len(requests))

	for i := 0; i > len(requests); i++ {
		should.True(requests[i].UpdatedAt.Before())
	}

	should.Equal("MEGA TEST", requests[0].Company["Name"])
	should.Equal("PUBG TEST", requests[1].Company["Name"])
	should.Equal("Ash of Evils ", requests[2].Company["Name"])
	for i := 3; i < len(requests); i++ {
		should.Equal("ZTEST2", requests[i].Company["Name"])
	}
}
