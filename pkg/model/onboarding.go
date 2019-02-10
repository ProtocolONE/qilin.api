package model

import (
	"github.com/satori/go.uuid"
)

type ClientDocumentStatus int8
type ReviewStatus int8

const (
	StatusDraft    ClientDocumentStatus = 0 //`draft`
	StatusOnReview ClientDocumentStatus = 1 //`on_review`
	StatusApproved ClientDocumentStatus = 2 //`approved`
	StatusDeclined ClientDocumentStatus = 3 //`declined`

	ReviewNew      ReviewStatus = 0 //`new`
	ReviewApproved ReviewStatus = 1 //`approved`
	ReviewChecking ReviewStatus = 2 //`checking`
	ReviewReturned ReviewStatus = 3 //`returned`
)

type DocumentsInfo struct {
	Model
	Company      JSONB                `gorm:"type:jsonb;not null"`
	Contact      JSONB                `gorm:"type:jsonb;not null"`
	Banking      JSONB                `gorm:"type:jsonb;not null"`
	Status       ClientDocumentStatus `gorm:"not null"`
	ReviewStatus ReviewStatus         `gorm:"not null"`
	VendorID     uuid.UUID            `gorm:"type:uuid;not null"`
}

func (status ClientDocumentStatus) ToString() string {
	if status == StatusDraft {
		return "draft"
	}

	if status == StatusApproved {
		return "approved"
	}

	if status == StatusDeclined {
		return "declined"
	}

	if status == StatusOnReview {
		return "on_review"
	}

	return ""
}

func (d DocumentsInfo) CanBeChanged() bool {
	return d.Status == StatusDraft || d.Status == StatusDeclined
}

func (d DocumentsInfo) CanBeSendToReview() bool {
	return d.Status == StatusDraft
}

func (DocumentsInfo) TableName() string {
	return "vendor_documents"
}
