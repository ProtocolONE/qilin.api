package model

import (
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
)

type ClientDocumentStatus int8
type ReviewStatus int8

const (
	StatusDraft    ClientDocumentStatus = 0 //`draft`
	StatusOnReview ClientDocumentStatus = 1 //`on_review`
	StatusApproved ClientDocumentStatus = 2 //`approved`
	StatusDeclined ClientDocumentStatus = 3 //`declined`
	StatusArchived ClientDocumentStatus = 4 //`declined`

	ReviewUndefined ReviewStatus = -1 //`new`
	ReviewNew       ReviewStatus = 0  //`new`
	ReviewApproved  ReviewStatus = 1  //`approved`
	ReviewChecking  ReviewStatus = 2  //`checking`
	ReviewReturned  ReviewStatus = 3  //`returned`
	ReviewArchived  ReviewStatus = 4  //`archived`
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

func ReviewStatusFromString(review string) (ReviewStatus, error) {
	switch review {
	case "":
		return ReviewUndefined, nil
	case "returned":
		return ReviewReturned, nil
	case "new":
		return ReviewNew, nil
	case "approved":
		return ReviewApproved, nil
	case "checking":
		return ReviewChecking, nil
	case "archived":
		return ReviewArchived, nil
	}
	return ReviewUndefined, errors.New(fmt.Sprintf("Unknown review status `%s`", review))
}

func (status ReviewStatus) ToString() string {
	switch status {
	case ReviewReturned:
		return "returned"
	case ReviewChecking:
		return "checking"
	case ReviewApproved:
		return "approved"
	case ReviewNew:
		return "new"
	case ReviewUndefined:
		return "undefined"
	case ReviewArchived:
		return "archived"
	}

	return ""
}

func (status ClientDocumentStatus) ToString() string {
	if status == StatusDraft {
		return "draft"
	} else if status == StatusApproved {
		return "approved"
	} else if status == StatusDeclined {
		return "declined"
	} else if status == StatusOnReview {
		return "on_review"
	} else if status == StatusArchived {
		return "archived"
	}

	return ""
}

func (d DocumentsInfo) CanBeChanged() bool {
	return d.Status == StatusDraft || d.Status == StatusDeclined
}

func (d DocumentsInfo) CanBeRevokedReview() bool {
	return d.Status == StatusOnReview
}

func (d DocumentsInfo) CanBeSendToReview() bool {
	return d.Status == StatusDraft
}

func (DocumentsInfo) TableName() string {
	return "vendor_documents"
}
