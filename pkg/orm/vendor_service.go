package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
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

func (p *VendorService) validate(item *model.Vendor) error {
	//p.db.First(&item2, "login = ? and password = ?", login, pass).Error

	if strings.Index(item.Email, "@") < 1 {
		return echo.NewHTTPError(http.StatusBadRequest,"Invalid Email")
	}
	if len(item.Name) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "Name is too short")
	}
	if len(item.Domain3) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest,"Domain is too short")
	}
	if strings.Index("0123456789", string(item.Domain3[0])) > -1 {
		return echo.NewHTTPError(http.StatusBadRequest,"Domain is invalid")
	}
	if item.ManagerId == nil || uuid.Equal(*item.ManagerId, uuid.Nil) {
		return echo.NewHTTPError(http.StatusBadRequest,"ManagerId is invalid")
	}
	return nil
}

// CreateVendor creates new Vendor object in database
func (p *VendorService) CreateVendor(item *model.Vendor) error {
	if err := p.validate(item); err != nil {
		return err
	}
	item.ID = uuid.NewV4()
	err := p.db.Create(item).Error
	if err != nil && strings.Index(err.Error(), "duplicate key value") > -1 {
		return echo.NewHTTPError(http.StatusBadRequest,"Other vendor with the same name already exists")
	} else if err != nil {
		err = errors.Wrap(err, "insert vendor")
	}
	return err
}

func (p *VendorService) UpdateVendor(item *model.Vendor) error {
	if err := p.validate(item); err != nil {
		return err
	}
	return errors.Wrap(p.db.Model(item).
		Updates(map[string]interface{}{
			"name": item.Name,
			"domain3": item.Domain3,
			"email": item.Email,
			"howmanyproducts": item.HowManyProducts}).
		Error, "insert vendor")
}

func (p *VendorService) FindByID(id uuid.UUID) (vendor model.Vendor, err error) {
	err = p.db.First(&vendor, model.Vendor{ID: id}).Error
	if err == gorm.ErrRecordNotFound {
		err = echo.NewHTTPError(404, "Vendor not found")
	} else if err != nil {
		err = errors.Wrap(err, "insert vendor")
	}
	return
}

func (p *VendorService) GetAll(limit, offset int) (vendors []*model.Vendor, err error) {

	err = errors.Wrap(p.db.
		Offset(offset).
		Limit(limit).
		Order("id desc").
		Find(&vendors).Error, "update vendor")

	return vendors, err
}
