package onboarding

import (
	"database/sql/driver"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
)

type ClientDocumentStatus string
type ReviewStatus string

const (
	StatusDraft    ClientDocumentStatus = `draft`
	StatusOnReview ClientDocumentStatus = `on_review`
	StatusApproved ClientDocumentStatus = `approved`
	StatusDeclined ClientDocumentStatus = `declined`

	ReviewNew      ReviewStatus = `new`
	ReviewApproved ReviewStatus = `approved`
	ReviewChecking ReviewStatus = `checking`
	ReviewReturned ReviewStatus = `returned`
)

func (u *ClientDocumentStatus) Scan(value interface{}) error {
	*u = ClientDocumentStatus(value.([]byte))
	return nil
}
func (u ClientDocumentStatus) Value() (driver.Value, error) { return string(u), nil }

type DocumentsInfo struct {
	model.Model
	Company      model.JSONB          `gorm:"type:jsonb;"`
	Contact      model.JSONB          `gorm:"type:jsonb;"`
	Banking      model.JSONB          `gorm:"type:jsonb;"`
	Status       ClientDocumentStatus `sql:"not null;type:ENUM('draft', 'on_review', 'approved', 'declined')"`
	ReviewStatus ClientDocumentStatus `sql:"not null;type:ENUM('draft', 'on_review', 'approved', 'declined')"`
	VendorID     uuid.UUID            `gorm:"type:uuid;not null"`
}
