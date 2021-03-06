package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm/utils"
	"strings"
)

// AdminOnboardingService is service to interact with vendor requests objects with admin rights
type AdminOnboardingService struct {
	db                *gorm.DB
	membershipService model.MembershipService
	ownerProvider     model.OwnerProvider
}

func NewAdminOnboardingService(db *Database, membershipService model.MembershipService, ownerProvider model.OwnerProvider) (*AdminOnboardingService, error) {
	return &AdminOnboardingService{db.database, membershipService, ownerProvider}, nil
}

func (p *AdminOnboardingService) GetRequests(limit int, offset int, name string, status model.ReviewStatus, sort string) ([]model.DocumentsInfo, int, error) {
	var documents []model.DocumentsInfo
	query := p.db.Model(model.DocumentsInfo{}).Where("status <> ?", model.StatusDraft).Limit(limit).Offset(offset)
	if name != "" {
		query = query.Where("company->>'Name' ILIKE ?", "%"+name+"%")
	}

	if status != model.ReviewUndefined {
		query = query.Where("review_status = ?", status)
	}

	if sort != "" {
		sorts := strings.Split(sort, ",")
		for _, curSort := range sorts {
			orderBy := ""
			switch curSort {
			case "-updatedAt":
				orderBy = "updated_at DESC"
			case "+updatedAt":
				orderBy = "updated_at ASC"
			case "-status":
				orderBy = "review_status DESC"
			case "+status":
				orderBy = "review_status ASC"
			case "-name":
				orderBy = "company->>'Name' DESC"
			case "+name":
				orderBy = "company->>'Name' ASC"
			}
			if orderBy == "" {
				return nil, 0, NewServiceError(http.StatusBadRequest, fmt.Sprintf("Unsupported sorting '%s'", curSort))
			}
			query = query.Order(orderBy)
		}
	}

	err := query.Find(&documents).Error

	if err != nil {
		return nil, 0, NewServiceError(http.StatusInternalServerError, err)
	}

	count := 0
	err = query.Limit(nil).Offset(nil).Count(&count).Error
	if err != nil {
		return nil, 0, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Counting reviews"))
	}

	return documents, count, nil
}

func (p *AdminOnboardingService) GetForVendor(vendorId uuid.UUID) (*model.DocumentsInfo, error) {
	if exist, err := utils.CheckExists(p.db, &model.Vendor{}, vendorId); !(exist && err == nil) {
		if err != nil {
			return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Check vendor existing"))
		}
		return nil, NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found", vendorId)
	}

	result := model.DocumentsInfo{}
	err := p.db.Where("vendor_id = ?", vendorId).First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, fmt.Sprintf("Get vendor's documents with vendor vendorId: %s", vendorId)))
	}

	return &result, nil
}

func (p *AdminOnboardingService) ChangeStatus(id uuid.UUID, status model.ReviewStatus) error {
	doc, err := p.GetForVendor(id)
	if err != nil {
		return err
	}

	if doc == nil || doc.ID == uuid.Nil {
		return NewServiceError(http.StatusBadRequest, "Trying to change status for non-existing review")
	}

	if doc.Status == model.StatusDraft {
		return NewServiceError(http.StatusBadRequest, "Trying to change status for documents that have not been sent to review")
	}

	switch status {
	case model.ReviewApproved:
		doc.Status = model.StatusApproved
	case model.ReviewReturned:
		doc.Status = model.StatusDeclined
	case model.ReviewChecking:
		doc.Status = model.StatusOnReview
	case model.ReviewArchived:
		doc.Status = model.StatusArchived
	default:
		return NewServiceError(http.StatusBadRequest, fmt.Sprintf("Can't change to status `%s`", status.ToString()))
	}
	doc.ReviewStatus = status

	err = p.db.Save(doc).Error
	if err != nil {
		return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Saving document error"))
	}

	if doc.Status == model.StatusApproved {
		owner, err := p.ownerProvider.GetOwnerForVendor(doc.VendorID)
		if err != nil {
			return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Trying get owner for vendor for removing not_approved role"))
		}

		//Error is ignored here
		_ = p.membershipService.RemoveRoleToUserInGame(doc.VendorID, owner, "*", model.NotApproved)
		_ = p.membershipService.AddRoleToUserInGame(doc.VendorID, owner, "*", model.VendorOwner)
	}

	return nil
}
