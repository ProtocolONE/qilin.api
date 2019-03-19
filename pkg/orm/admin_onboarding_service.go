package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"strings"
)

// AdminOnboardingService is service to interact with vendor requests objects with admin rights
type AdminOnboardingService struct {
	db       *gorm.DB
}

func NewAdminOnboardingService(db *Database) (*AdminOnboardingService, error) {
	return &AdminOnboardingService{db.database}, nil
}

func (p *AdminOnboardingService) GetRequests(limit int, offset int, name string, status model.ReviewStatus, sort string) ([]model.DocumentsInfo, error) {

	var documents []model.DocumentsInfo
	query := p.db.Where("status <> ?", model.StatusDraft).Limit(limit).Offset(offset)
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
				return nil, NewServiceError(http.StatusBadRequest, fmt.Sprintf("Unsupported sorting '%s'", curSort))
			}
			query = query.Order(orderBy)
		}
	}

	err := query.Find(&documents).Error

	if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	return documents, nil
}

func (p *AdminOnboardingService) GetForVendor( /*userId uuid.UUID, */ vendorId uuid.UUID) (*model.DocumentsInfo, error) {
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

	return nil
}
