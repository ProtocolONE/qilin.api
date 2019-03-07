package model

import "github.com/satori/go.uuid"

type Notification struct {
	Model

	Title    string `gorm:"not null"`
	Message  string
	IsRead   bool      `gorm:"not null;default:false"`
	VendorID uuid.UUID `gorm:"type:uuid;not null"`
	UserID   uuid.UUID `gorm:"type:uuid;not null"`
}

type NotificationService interface {
	GetNotifications(id uuid.UUID, limit int, offset int, search string, sort string) ([]Notification, error)
	MarkAsRead(vendorId uuid.UUID, messageId uuid.UUID) error
	GetUserToken(id uuid.UUID) string
	SendNotification(notification *Notification) (*Notification, error)
	GetNotification(vendorId uuid.UUID, messageId uuid.UUID) (*Notification, error)
}
