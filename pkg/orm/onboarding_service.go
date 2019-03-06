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

//SendToReview is method for sending vendor documents to review
func (p *OnboardingService) SendToReview(vendorId uuid.UUID) error {
	err := p.checkVendorExist(vendorId)
	if err != nil {
		return err
	}

	var documents model.DocumentsInfo
	err = p.db.Model(model.Vendor{ID: vendorId}).Related(&documents).Error
	if err != nil {
		return NewServiceError(http.StatusBadRequest, errors.Wrap(err, fmt.Sprintf("Can't get related documents for vendor with id %s", vendorId)))
	}

	if documents.CanBeSendToReview() == false {
		return NewServiceError(http.StatusBadRequest, fmt.Sprintf("Document has status `%s` and can not be sent to review", documents.Status.ToString()))
	}

	documents.Status = model.StatusOnReview
	documents.ReviewStatus = model.ReviewNew

	if err := p.db.Save(&documents).Error; err != nil {
		return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Can't save documents for vendor"))
	}

	return nil
}

func (p *OnboardingService) checkVendorExist(vendorId uuid.UUID) error {
	count := 0
	err := p.db.Model(&model.Vendor{}).Where("ID = ?", vendorId).Count(&count).Error

	if err != nil {
		return NewServiceError(http.StatusBadRequest, errors.Wrap(err, fmt.Sprintf("Get vendor id: %s", vendorId)))
	}

	if count == 0 {
		return NewServiceError(http.StatusNotFound, fmt.Sprintf("No vendor with id: %s", vendorId))
	}

	return nil
}

//GetForVendor is method for getting current documents info for vendor
func (p *OnboardingService) GetForVendor(id uuid.UUID) (*model.DocumentsInfo, error) {
	err := p.checkVendorExist(id)
	if err != nil {
		return nil, err
	}
	result := model.DocumentsInfo{}
	err = p.db.Where("vendor_id = ?", id).First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusBadRequest, errors.Wrap(err, fmt.Sprintf("Get vendor's documents with vendor id: %s", id)))
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
		return nil, NewServiceError(http.StatusBadRequest, errors.Wrap(err, fmt.Sprintf("Get vendor's documents with id: %s", id)).Error())
	}

	return result, err
}

//GetById is method for getting vendor documents by documentId
func (p *OnboardingService) ChangeDocument(document *model.DocumentsInfo) error {
	count := 0
	err := p.db.Model(&model.Vendor{}).Where("ID = ?", document.VendorID).Count(&count).Error

	if err != nil {
		return NewServiceError(http.StatusBadRequest, errors.Wrap(err, fmt.Sprintf("Get vendor id: %s", document.VendorID)))
	}

	if count == 0 {
		return NewServiceError(http.StatusNotFound, fmt.Sprintf("No vendor with id: %s", document.VendorID))
	}

	info := &model.DocumentsInfo{}
	result := p.db.Model(&model.Vendor{ID: document.VendorID}).Related(&info)

	if result.RecordNotFound() {
		document.ID = uuid.NewV4()
	} else {
		if result.Error != nil {
			return NewServiceError(http.StatusInternalServerError, errors.Wrap(result.Error, fmt.Sprintf("Get vendor's documents with id: %s", document.ID)).Error())
		}
		if info.CanBeChanged() == false {
			return NewServiceError(http.StatusBadRequest, fmt.Sprintf("Can't change document with status `%s`", info.Status.ToString()))
		}
		document.Status = model.StatusDraft
		document.ID = info.ID
		document.CreatedAt = info.CreatedAt
	}

	err = p.db.Save(document).Error
	if err != nil {
		return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, fmt.Sprintf("Save vendor's documents with id: %s", document.ID)).Error())
	}

	return nil
}
