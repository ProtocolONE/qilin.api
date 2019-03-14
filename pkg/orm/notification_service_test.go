package orm_test

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/sys"
	"qilin-api/pkg/test"
	"strings"
	"testing"
)

type NotificationServiceTestSuite struct {
	suite.Suite
	db      *orm.Database
	service model.NotificationService
}

func Test_NotificationService(t *testing.T) {
	suite.Run(t, new(NotificationServiceTestSuite))
}

var vendorId = "54702e34-dff7-46b0-abbd-570eec5f92fb"

func (suite *NotificationServiceTestSuite) SetupTest() {
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

	notifier, err := sys.NewNotifier(config.Notifier.ApiKey, config.Notifier.Host)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), notifier)

	assert.Nil(suite.T(), db.DB().Create(&model.Vendor{ID: uuid.FromStringOrNil(GameID), Name: "Test vendor", Domain3: "domain", Email: "email@email.com"}).Error)
	assert.Nil(suite.T(), db.DB().Create(&model.Vendor{ID: uuid.FromStringOrNil(vendorId), Name: "Test vendor2", Domain3: "domain2", Email: "email2@email.com"}).Error)

	suite.db = db
	suite.service, err = orm.NewNotificationService(db, notifier, config.Notifier.Secret)
	assert.Nil(suite.T(), err)

}

func (suite *NotificationServiceTestSuite) TestNotificationsForWrongVendor() {
	shouldBe := require.New(suite.T())
	id, err := uuid.FromString(GameID)
	anotherVendorId, err := uuid.FromString(vendorId)
	shouldBe.Nil(err)

	notification := &model.Notification{VendorID: id, Title: "Some title", Message: "ZZZ"}
	notification.ID = uuid.NewV4()
	notification.IsRead = true
	shouldBe.Nil(suite.db.DB().Create(notification).Error)

	res, err := suite.service.GetNotification(anotherVendorId, notification.ID)
	shouldBe.Nil(res)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)

	err = suite.service.MarkAsRead(anotherVendorId, notification.ID)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)

	res, err = suite.service.GetNotification(uuid.NewV4(), notification.ID)
	shouldBe.Nil(res)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)

	err = suite.service.MarkAsRead(uuid.NewV4(), notification.ID)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)

	err = suite.service.MarkAsRead(id, uuid.NewV4())
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)

	res, err = suite.service.GetNotification(id, uuid.NewV4())
	shouldBe.Nil(res)
	shouldBe.NotNil(err)
	shouldBe.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)
}

func (suite *NotificationServiceTestSuite) TestGetNotifications() {
	should := require.New(suite.T())
	id, err := uuid.FromString(GameID)
	should.Nil(err)
	suite.generateNotifications(id)

	notifications, count, err := suite.service.GetNotifications(id, 10, 0, "", "")
	should.Nil(err)
	should.NotNil(notifications)
	should.Equal(10, len(notifications))
	should.Equal(101, count)
	for _, n := range notifications {
		should.Equal(id, n.VendorID)
	}

	notifications, count, err = suite.service.GetNotifications(uuid.NewV4(), 1000, 0, "", "")
	should.NotNil(err)
	should.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)
	should.Nil(notifications)
	should.Equal(0, len(notifications))

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "")
	should.Nil(err)
	should.NotNil(notifications)
	should.Equal(101, len(notifications))

	notifications, count, err = suite.service.GetNotifications(id, 1000, 90, "", "")
	should.Nil(err)
	should.NotNil(notifications)
	should.Equal(11, len(notifications))

	notifications, count, err = suite.service.GetNotifications(id, 10, 0, "Some", "")
	should.Nil(err)
	should.NotNil(notifications)
	should.Equal(1, len(notifications))
	should.Equal("Some title", notifications[0].Title)

	notifications, count, err = suite.service.GetNotifications(id, 10, 0, "Test", "")
	should.Nil(err)
	should.NotNil(notifications)
	should.Equal(10, len(notifications))

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "Test", "")
	should.Nil(err)
	should.NotNil(notifications)
	should.Equal(100, len(notifications))

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "-createdDate")
	should.Nil(err)
	should.NotNil(notifications)
	for i := 0; i < len(notifications)-1; i++ {
		should.True(notifications[i].CreatedAt.After(notifications[i+1].CreatedAt) || notifications[i].CreatedAt.Equal(notifications[i+1].CreatedAt))
	}

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "+createdDate")
	should.Nil(err)
	should.NotNil(notifications)
	for i := 0; i < len(notifications)-1; i++ {
		should.True(notifications[i].CreatedAt.Before(notifications[i+1].CreatedAt) || notifications[i].CreatedAt.Equal(notifications[i+1].CreatedAt))
	}

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "+title")
	should.Nil(err)
	should.NotNil(notifications)
	for i := 0; i < len(notifications)-1; i++ {
		should.Equal(-1, strings.Compare(notifications[i].Title, notifications[i+1].Title), "%d %s > %s", i, notifications[i].Title, notifications[i+1].Title)
	}

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "-title")
	should.Nil(err)
	should.NotNil(notifications)
	for i := 0; i < len(notifications)-1; i++ {
		should.Equal(1, strings.Compare(notifications[i].Title, notifications[i+1].Title), "%d %s > %s", i, notifications[i].Title, notifications[i+1].Title)
	}

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "+message")
	should.Nil(err)
	should.NotNil(notifications)
	for i := 0; i < len(notifications)-1; i++ {
		should.Equal(-1, strings.Compare(notifications[i].Message, notifications[i+1].Message), "%d %s > %s", i, notifications[i].Message, notifications[i+1].Message)
	}

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "-message")
	should.Nil(err)
	should.NotNil(notifications)
	for i := 0; i < len(notifications)-1; i++ {
		should.Equal(1, strings.Compare(notifications[i].Message, notifications[i+1].Message), "%d %s > %s", i, notifications[i].Message, notifications[i+1].Message)
	}

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "+unread")
	should.Nil(err)
	should.NotNil(notifications)
	for i := 0; i < 100; i++ {
		should.False(notifications[i].IsRead, "%d %s %b", i, notifications[i].ID, notifications[i].IsRead)
	}

	notifications, count, err = suite.service.GetNotifications(id, 1000, 0, "", "-unread")
	should.Nil(err)
	should.NotNil(notifications)
	should.True(notifications[0].IsRead)
	for i := 1; i < 101; i++ {
		should.False(notifications[i].IsRead)
	}
}

func (suite *NotificationServiceTestSuite) generateNotifications(id uuid.UUID) {
	should := require.New(suite.T())
	notification := &model.Notification{VendorID: id, Title: "Some title", Message: "ZZZ"}
	notification.ID = uuid.NewV4()
	notification.IsRead = true
	should.Nil(suite.db.DB().Create(notification).Error)

	notification = &model.Notification{VendorID: uuid.NewV4(), Title: "Some title", Message: "YYY"}
	notification.IsRead = true
	notification.ID = uuid.NewV4()
	should.Nil(suite.db.DB().Create(notification).Error)

	for i := 0; i < 100; i++ {
		notification = &model.Notification{VendorID: id, Title: fmt.Sprintf("Test title %d", i), Message: fmt.Sprintf("%d", i)}
		notification.ID = uuid.NewV4()
		notification.IsRead = false
		should.Nil(suite.db.DB().Create(notification).Error)
	}
}

func (suite *NotificationServiceTestSuite) TestMarkAsRead() {
	should := require.New(suite.T())
	id, err := uuid.FromString(GameID)
	should.Nil(err)
	notification := &model.Notification{VendorID: id, Title: "Test notification", Message: "Body notification"}
	notification.ID = uuid.NewV4()
	should.Nil(suite.db.DB().Create(notification).Error)
	should.Nil(suite.service.MarkAsRead(id, notification.ID))
	inDb := model.Notification{}
	should.Nil(suite.db.DB().Model(inDb).Where("id = ?", notification.ID).First(&inDb).Error)
	should.True(inDb.IsRead)

	should.Nil(suite.service.MarkAsRead(id, notification.ID))

	err = suite.service.MarkAsRead(id, uuid.NewV4())
	should.NotNil(err)
	should.Equal(http.StatusNotFound, err.(*orm.ServiceError).Code)
}

func (suite *NotificationServiceTestSuite) TestSendNotification() {
	should := require.New(suite.T())
	id, err := uuid.FromString(GameID)
	should.Nil(err)
	notification, err := suite.service.SendNotification(&model.Notification{VendorID: id, Title: "Test notification", Message: "Body notification"})
	should.Nil(err)
	should.NotNil(notification)
	should.NotEqual(uuid.Nil, notification.ID)

	inDb := model.Notification{}
	should.Nil(suite.db.DB().Model(model.Notification{}).Where("id = ? ", notification.ID).First(&inDb).Error)
	should.Equal("Test notification", inDb.Title)
	should.Equal("Body notification", inDb.Message)
}
