package orm

import (
	"errors"
	"github.com/jinzhu/gorm"
	"qilin-api/pkg/model"
	"strings"
)

// VendorService is service to interact with database and Vendor object.
type VendorService struct {
	db *gorm.DB
}

// NewVendorService initialize this service.
func NewVendorService(db *Database) (*VendorService, error) {
	return &VendorService{db.database}, nil
}

func validate(item *model.Vendor) error {
	if strings.Index(item.Email, "@") < 1 {
		return errors.New("Invalid Email")
	}
	if len(item.Name) < 2 {
		return errors.New( "Name is too short")
	}
	if len(item.Domain3) < 2 {
		return errors.New("Domain is too short")
	}
	if strings.Index("0123456789", string(item.Domain3[0])) > -1 {
		return errors.New("Domain is invalid")
	}
	if item.ManagerId < 1 {
		return errors.New("ManagerId is invalid")
	}
	return nil
}

// CreateVendor creates new Vendor object in database
func (p *VendorService) CreateVendor(item *model.Vendor) error {
	if err := validate(item); err != nil {
		return err
	}
	return p.db.Create(item).Error
}

func (p *VendorService) UpdateVendor(item *model.Vendor) error {
	if err := validate(item); err != nil {
		return err
	}
	return p.db.Model(item).
		Updates(map[string]interface{}{
			"name": item.Name,
			"domain3": item.Domain3,
			"email": item.Email}).
		Error
}

func (p *VendorService) FindByID(id uint) (vendor model.Vendor, err error) {
	err = p.db.First(&vendor, model.Vendor{ID: id}).Error
	return
}

func (p *VendorService) GetAll(limit, offset int) (vendors []*model.Vendor, err error) {

	err = p.db.
		Offset(offset).
		Limit(limit).
		Order("id desc").
		Find(&vendors).Error

	return vendors, err
}
