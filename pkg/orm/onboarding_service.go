package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
)

// OnboardingService is service to interact with database and vendor requests objects.
type OnboardingService struct {
	db *gorm.DB
}

func NewOnboardingService(db *Database) (*OnboardingService, error) {
	return &OnboardingService{db.database}, nil
}

//GetForVendor is method for getting current documents info for vendor
func (p *OnboardingService) GetForVendor(id uuid.UUID) (*model.DocumentsInfo, error) {
	count := 0
	err := p.db.Model(&model.Vendor{}).Where("ID = ?", id).Count(&count).Error

	if err != nil {
		return nil, NewServiceError(http.StatusBadRequest, errors.Wrap(err,  fmt.Sprintf("Get vendor id: %s", id)))
	}

	if count == 0 {
		return nil, NewServiceError(http.StatusNotFound, fmt.Sprintf("No vendor with id: %s", id))
	}

	result := model.DocumentsInfo{}
	err = p.db.Where("vendor_id = ?", id).First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusBadRequest, errors.Wrap(err,  fmt.Sprintf("Get vendor's documents with vendor id: %s", id)))
	}

	return &result, nil
}

//GetById is method for getting vendor documents by documentId
func (p *OnboardingService) GetById(id uuid.UUID) (*model.DocumentsInfo, error) {
	result := &model.DocumentsInfo{}
	err := p.db.Model(&model.DocumentsInfo{}).Where("ID = ?", id).First(result).Error

	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, fmt.Sprintf("Get vendor's documents with id: %s", id))
	} else if err != nil {
		return nil, NewServiceError(http.StatusBadRequest, errors.Wrap(err,  fmt.Sprintf("Get vendor's documents with id: %s", id)).Error())
	}

	return result, err
}

//GetById is method for getting vendor documents by documentId
func (p *OnboardingService) ChangeDocument(document *model.DocumentsInfo)  error {
	count := 0
	err := p.db.Model(&model.Vendor{}).Where("ID = ?", document.VendorID).Count(&count).Error

	if err != nil {
		return NewServiceError(http.StatusBadRequest, errors.Wrap(err,  fmt.Sprintf("Get vendor id: %s", document.VendorID)))
	}

	if count == 0 {
		return NewServiceError(http.StatusNotFound, fmt.Sprintf("No vendor with id: %s", document.VendorID))
	}

	info := &model.DocumentsInfo{}
	result := p.db.Where("ID = ?", document.ID).First(info)
	if result.RecordNotFound() {
		info = document
	} else {
		if result.Error != nil {
			return NewServiceError(http.StatusBadRequest, errors.Wrap(result.Error, fmt.Sprintf("Get vendor's documents with id: %s", document.ID)).Error())
		}
		document.CreatedAt = info.CreatedAt
		document.DeletedAt = info.DeletedAt
	}

	err = p.db.Save(document).Error
	if err != nil {
		return NewServiceError(http.StatusBadRequest, errors.Wrap(err, fmt.Sprintf("Save vendor's documents with id: %s", document.ID)).Error())
	}

	return nil
}





