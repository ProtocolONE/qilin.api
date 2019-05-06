package orm

import (
	"net/http"
	"qilin-api/pkg/model"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// priceService is service to interact with database and Media object.
type priceService struct {
	db *gorm.DB
}

// NewpriceService initialize this service.
func NewPriceService(db *Database) model.PriceService {
	return &priceService{db.database}
}

//GetBase is method for retriving base information about package pricing
func (p *priceService) GetBase(id uuid.UUID) (*model.BasePrice, error) {
	result := &model.BasePrice{ID: id}
	err := p.db.
		Select(model.SelectFields(result)).
		First(result).
		Select("*").
		Related(&result.Prices, "BasePriceID").
		Order("created_at").
		Error
	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, "Package not found")
	} else if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "search package by id"))
	}
	return result, err
}

//UpdateBase is method for updating base information about package pricing
func (p *priceService) UpdateBase(id uuid.UUID, price *model.BasePrice) error {

	domain := &model.BasePrice{ID: id}
	err := p.db.Select(model.SelectFields(domain)).First(domain).Error
	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Package not found")
	} else if err != nil {
		return errors.Wrap(err, "search package by id")
	}

	price.ID = domain.ID
	now := time.Now()
	price.UpdatedAt = &now

	err = p.db.
		Set("gorm:association_autoupdate", false).
		Set("gorm:association_autocreate", false).
		Save(price).
		Error
	if err != nil {
		return errors.Wrap(err, "save base prices for package")
	}

	return nil
}

//Delete is method for removing price with currency for package
func (p *priceService) Delete(id uuid.UUID, price *model.Price) error {
	domain := &model.BasePrice{ID: id}

	count := 0
	if err := p.db.Model(domain).Limit(1).Count(&count).Error; err != nil {
		return NewServiceError(http.StatusInternalServerError, "Package search")
	}
	if count == 0 {
		return NewServiceError(http.StatusNotFound, "Package not found")
	}

	var prices []model.Price
	err := p.db.
		Model(domain).
		Related(&prices, "BasePriceID").
		Order("created_at").
		Error
	if err != nil {
		return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "search prices for package"))
	}

	found := false
	for _, v := range prices {
		if v.Currency == price.Currency {
			p.db.Delete(&v)
			found = true
			break
		}
	}

	if found == false {
		return NewServiceError(http.StatusNotFound, "Price not found")
	}

	return nil
}

// Update is method for updating price with currency for package
func (p *priceService) Update(id uuid.UUID, price *model.Price) error {
	domain := &model.BasePrice{ID: id}
	var prices []model.Price

	count := 0
	if err := p.db.Model(domain).Where("ID = ?", id).Limit(1).Count(&count).Error; err != nil {
		return NewServiceError(http.StatusInternalServerError, "Package search")
	}

	if count == 0 {
		return NewServiceError(http.StatusNotFound, "Package not found")
	}

	err := p.db.
		Model(domain).
		Related(&prices, "BasePriceID").
		Order("created_at").
		Error
	if err != nil {
		return errors.Wrap(err, "search package by id")
	}

	price.BasePriceID = id

	for _, v := range prices {
		if v.Currency == price.Currency {
			price.ID = v.ID
			break
		}
	}

	p.db.Save(price)

	return nil
}
