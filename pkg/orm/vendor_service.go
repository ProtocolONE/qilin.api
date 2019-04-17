package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"strings"
)

// VendorService is service to interact with database and Vendor object.
type VendorService struct {
	db *gorm.DB
	membershipService model.MembershipService
}

const (
	errVendorConflict  = "Other vendor with the same name, domain3 or email already exists"
	errVendorNotFound  = "Vendor not found"
	errVendorManagerId = "ManagerId is invalid"
)

// NewVendorService initialize this service.
func NewVendorService(db *Database, membershipService model.MembershipService) (*VendorService, error) {
	return &VendorService{db.database, membershipService}, nil
}

func (p *VendorService) validate(item *model.Vendor) error {
	if item.Email != "" && strings.Index(item.Email, "@") == -1 {
		return NewServiceError(http.StatusBadRequest, "Invalid Email")
	}
	if len(item.Name) < 2 {
		return NewServiceError(http.StatusBadRequest, "Name is too short")
	}
	if len(item.Domain3) < 2 {
		return NewServiceError(http.StatusBadRequest, "Domain is too short")
	}
	if strings.Index("0123456789", string(item.Domain3[0])) > -1 {
		return NewServiceError(http.StatusBadRequest, "Domain is invalid")
	}
	if item.ManagerID == "" {
		return NewServiceError(http.StatusBadRequest, errVendorManagerId)
	}
	return nil
}

// CreateVendor creates new Vendor object in database
func (p *VendorService) Create(item *model.Vendor) (result *model.Vendor, err error) {
	if err := p.validate(item); err != nil {
		return nil, err
	}
	vendor := *item
	if uuid.Nil == vendor.ID {
		vendor.ID = uuid.NewV4()
	}

	tx := p.db.Begin()

	count := 0
	if err := p.db.Model(&model.Vendor{}).Where("manager_id = ?", vendor.ManagerID).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	if count > 0 {
		tx.Rollback()
		return nil, NewServiceError(http.StatusConflict, "User can be owner for one vendor only.")
	}

	err = tx.Create(&vendor).Error
	if err != nil && strings.Index(err.Error(), "duplicate key value") > -1 {
		tx.Rollback()
		return nil, NewServiceError(http.StatusConflict, errVendorConflict)
	} else if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "Insert vendor")
	}
	err = tx.Model(&vendor).Association("Users").Append(model.User{ID: vendor.ManagerID}).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "Append to association")
	}

	err = p.membershipService.AddRoleToUser(vendor.ManagerID, vendor.ManagerID, model.NotApproved)
	if err != nil {
		tx.Rollback()
		return &vendor, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Set role to owner"))
	}

	return &vendor, tx.Commit().Error
}

func (p *VendorService) Update(item *model.Vendor) (vendor *model.Vendor, err error) {
	if err := p.validate(item); err != nil {
		return nil, err
	}

	err = p.db.Model(item).
		Updates(map[string]interface{}{
			"name":            item.Name,
			"domain3":         item.Domain3,
			"email":           item.Email,
			"howmanyproducts": item.HowManyProducts,
			"manager_id":      item.ManagerID,
		}).Error

	if err != nil && strings.Index(err.Error(), "duplicate key value") > -1 {
		return nil, NewServiceError(http.StatusConflict, errVendorConflict)
	} else if err != nil {
		return nil, errors.Wrap(err, "Update vendor")
	}

	return p.FindByID(item.ID)
}

func (p *VendorService) FindByID(id uuid.UUID) (vendor *model.Vendor, err error) {
	vendor = &model.Vendor{}
	err = p.db.First(vendor, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		err = NewServiceError(404, errVendorNotFound)
	} else if err != nil {
		err = errors.Wrap(err, "Find vendor")
	}
	return
}

func (p *VendorService) GetAll(limit, offset int) (vendors []*model.Vendor, err error) {
	err = p.db.
		Offset(offset).
		Limit(limit).
		Order("created_at desc").
		Find(&vendors).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch vendors")
	}

	return vendors, err
}

