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
	MarkAsRead(id uuid.UUID) error
	SendNotification(notification *Notification) (*Notification, error)
	GetNotification(id uuid.UUID) (*Notification, error)
}