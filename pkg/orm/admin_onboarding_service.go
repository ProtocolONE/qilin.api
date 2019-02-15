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
	db *gorm.DB
}

func NewAdminOnboardingService(db *Database) (*AdminOnboardingService, error) {
	return &AdminOnboardingService{db.database}, nil
}

func (p *AdminOnboardingService) GetRequests(limit int, offset int, name string, status model.ReviewStatus, sort string) ([]model.DocumentsInfo, error) {
	var documents []model.DocumentsInfo
	query := p.db.Where("status <> ?", model.StatusDraft).Limit(limit).Offset(offset)
	if name != "" {
		query = query.Where("company->>'Name' LIKE ?", "%"+name+"%")
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

func (p *AdminOnboardingService) GetForVendor(id uuid.UUID) (*model.DocumentsInfo, error) {
	if exists, err := utils.CheckExists(p.db, &model.Vendor{}, id); !(exists && err == nil) {
		if err != nil {
			return nil, NewServiceError(http.StatusInternalServerError, err)
		}
	}

	result := model.DocumentsInfo{}
	err := p.db.Where("vendor_id = ?", id).First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, fmt.Sprintf("Get vendor's documents with vendor id: %s", id)))
	}

	return &result, nil
}
