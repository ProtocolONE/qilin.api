package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm/utils"
	"qilin-api/pkg/sys"
	"strings"
	"time"
)

type notificationService struct {
	db *gorm.DB
	notifier sys.Notifier
}

const notificationMask string = "qilin:%s"

//NewNotificationService is method for creating new instance of service
func NewNotificationService(db *Database, notifier sys.Notifier) (model.NotificationService, error) {
	return &notificationService{db.database, notifier}, nil
}

//GetNotifications is method for retrieving
func (p *notificationService) GetNotifications(id uuid.UUID, limit int, offset int, search string, sort string) ([]model.Notification, error) {
	if exist, err := utils.CheckExists(p.db, &model.Vendor{}, id); exist == false || err != nil {
		if err != nil {
			return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Checking vendor existing"))
		}

		return nil, NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found", id)
	}

	query := p.db.Model(&model.Notification{}).Where("vendor_id = ?", id).Limit(limit).Offset(offset)

	if search != "" {
		search = "%" + search + "%%"
		query = query.Where("title ilike ? OR message ilike ?", search, search)
	}

	if sort == "" {
		query = query.Order("created_at DESC")
	} else {
		sorts := strings.Split(sort, ",")
		for _, cur := range sorts {
			switch cur {
			case "-createdDate":
				query = query.Order("created_at DESC")
			case "+createdDate":
				query = query.Order("created_at ASC")
			case "-message":
				query = query.Order("message DESC")
			case "+message":
				query = query.Order("message ASC")
			case "-title":
				query = query.Order("title DESC")
			case "+title":
				query = query.Order("title ASC")
			case "-unread":
				query = query.Order("is_read DESC")
			case "+unread":
				query = query.Order("is_read ASC")
			default:
				return nil, NewServiceErrorf(http.StatusBadRequest, "Unknown sort `%s`", cur)
			}
		}
	}

	var notifications []model.Notification
	err := query.Find(&notifications).Error
	if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Searching notifications"))
	}

	if notifications == nil {
		notifications = make([]model.Notification, 0)
	}

	return notifications, nil
}

//MarkAsRead is method for marking notification as read
func (p *notificationService) MarkAsRead(id uuid.UUID) error {
	notification := model.Notification{}
	if res := p.db.Model(&model.Notification{}).Where("id = ?", id).First(&notification); res.Error != nil {
		if res.RecordNotFound() {
			return NewServiceErrorf(http.StatusNotFound, "Can't find notification with id `%s`", id)
		}
		return NewServiceError(http.StatusInternalServerError, errors.Wrap(res.Error, "Getting notification from db"))
	}

	notification.IsRead = true
	err := p.db.Save(notification).Error
	if err != nil {
		return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Updating notification from db"))
	}

	return nil
}

//SendNotification is method for sending notification via web socket and saving to db
func (p *notificationService) SendNotification(notification *model.Notification) (*model.Notification, error) {
	if exist, err := utils.CheckExists(p.db, model.Vendor{}, notification.VendorID); !(exist && err == nil) {
		if err != nil {
			return nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Checking existing vendor"))
		}
		return nil, NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found", notification.VendorID)
	}

	notification.ID = uuid.NewV4()
	res := p.db.Create(notification)
	err := res.Error
	if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Creating of notification. %#v", notification))
	}

	message := sys.NotifyMessage{
		ID: notification.ID.String(),
		Title: notification.Title,
		Body: notification.Message,
		DateTime: time.Now().UTC().String(),
	}

	_ = p.notifier.SendMessage(fmt.Sprintf(notificationMask, notification.VendorID), message)

	return res.Value.(*model.Notification), nil
}

func (p *notificationService) GetNotification(id uuid.UUID) (*model.Notification, error) {
	notification := model.Notification{}
	res := p.db.Model(model.Notification{}).Where("id = ?", id).First(&notification)
	if res.Error != nil {
		if res.RecordNotFound() {
			return nil, NewServiceError(http.StatusNotFound, "Get notification")
		}
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(res.Error, "Get notification"))
	}
	return &notification, nil
}