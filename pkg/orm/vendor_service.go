package orm

import (
	"github.com/jinzhu/gorm"
	"qilin-api/pkg/model"
)

// VendorService is service to interact with database and Vendor object.
type VendorService struct {
	db *gorm.DB
}

// NewVendorService initialize this service.
func NewVendorService(db *Database) (*VendorService, error) {
	return &VendorService{db.database}, nil
}

// CreateVendor creates new Vendor object in database
func (p *VendorService) CreateVendor(u *model.Vendor) error {
	return p.db.Create(u).Error
}

func (p *VendorService) UpdateVendor(u *model.Vendor) error {
	return p.db.Update(u).Error
}

// FindByID return Vendor object by given id
func (p *VendorService) FindByID(id uint) (vendor model.Vendor, err error) {
	err = p.db.First(&vendor, model.Vendor{ID: id}).Error
	return
}

func (p *VendorService) GetAll() ([]*model.Vendor, error) {
	var vendors []*model.Vendor
	err := p.db.Find(&vendors).Error

	return vendors, err
}

func (p *VendorService) FindByName(name string) ([]*model.Vendor, error) {
	var vendors []*model.Vendor
	err := p.db.Where("name LIKE ?", name).Find(&vendors).Error

	return vendors, err
}
