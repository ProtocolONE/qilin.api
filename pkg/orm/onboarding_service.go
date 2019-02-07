package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/onboarding"
)

// OnboardingService is service to interact with database and vendor requests objects.
type OnboardingService struct {
	db *gorm.DB
}

func NewOnboardingService(db *Database) (*OnboardingService, error) {
	return &OnboardingService{db.database}, nil
}

//GetForVendor is method for getting current documents info for vendor
func (p *OnboardingService) GetForVendor(id uuid.UUID) (*onboarding.DocumentsInfo, error) {
	if p.db.NewRecord(&model.Vendor{ID: id}) {
		return nil, NewServiceError(http.StatusNotFound, fmt.Sprintf("No vendor with id: %s", id))
	}

	result := &onboarding.DocumentsInfo{}
	err := p.db.Where("ID = ?", id).First(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(err,  fmt.Sprintf("Get vendor's documents with vendor id: %s", id))
	}

	return result, err
}

//GetById is method for getting vendor documents by documentId
func (p *OnboardingService) GetById(id uuid.UUID) (*onboarding.DocumentsInfo, error) {
	result := &onboarding.DocumentsInfo{}
	err := p.db.Where("ID = ?", id).First(&result).Error

	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, fmt.Sprintf("Get vendor's documents with id: %s", id))
	} else if err != nil {
		return nil, errors.Wrap(err,  fmt.Sprintf("Get vendor's documents with id: %s", id))
	}

	return result, err
}

//GetById is method for getting vendor documents by documentId
func (p *OnboardingService) ChangeDocument(id uuid.UUID, document *onboarding.DocumentsInfo)  error {
	result := &onboarding.DocumentsInfo{}
	err := p.db.Where("ID = ?", id).First(&result).Error

	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, fmt.Sprintf("Get vendor's documents with id: %s", id))
	} else if err != nil {
		return errors.Wrap(err,  fmt.Sprintf("Get vendor's documents with id: %s", id))
	}

	return err
}



