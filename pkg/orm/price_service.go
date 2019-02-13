package orm

import (
	"net/http"
	"qilin-api/pkg/model"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// PriceService is service to interact with database and Media object.
type PriceService struct {
	db *gorm.DB
}

// NewPriceService initialize this service.
func NewPriceService(db *Database) (*PriceService, error) {
	return &PriceService{db.database}, nil
}

//GetBase is method for retriving base information about game pricing
func (p *PriceService) GetBase(id uuid.UUID) (*model.BasePrice, error) {
	result := &model.BasePrice{}
	err := p.db.Preload("Prices").Select(model.SelectFields(result)).Where("id = ?", id).First(result).Error

	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "search game by id"))
	}

	return result, err
}

//UpdateBase is method for updating base information about game pricing
func (p *PriceService) UpdateBase(id uuid.UUID, price *model.BasePrice) error {

	domain := &model.BasePrice{ID: id}
	err := p.db.Preload("Prices").Select(model.SelectFields(domain)).Where("id = ?", id).First(domain).Error

	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	price.ID = domain.ID
	now := time.Now()
	price.UpdatedAt = &now

	err = p.db.Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false).Save(price).Error

	if err != nil {
		return errors.Wrap(err, "save base prices for game")
	}

	if err != nil {
		return errors.Wrap(err, "save prices for game")
	}

	return nil
}

//Delete is method for removing price with currency for game
func (p *PriceService) Delete(id uuid.UUID, price *model.Price) error {
	domain := &model.BasePrice{ID: id}

	count := 0
	if err := p.db.Model(domain).Where("ID = ?", id).Limit(1).Count(&count).Error; err != nil {
		return NewServiceError(http.StatusInternalServerError, "Game search")
	}

	if count == 0 {
		return NewServiceError(http.StatusNotFound, "Game not found")
	}

	var prices []model.Price
	err := p.db.Model(domain).Association("Prices").Find(&prices).Error

	if err != nil {
		return NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "search prices for game"))
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

//Update is method for updating price with currency for game
func (p *PriceService) Update(id uuid.UUID, price *model.Price) error {
	domain := &model.BasePrice{ID: id}
	var prices []model.Price

	count := 0
	if err := p.db.Model(domain).Where("ID = ?", id).Limit(1).Count(&count).Error; err != nil {
		return NewServiceError(http.StatusInternalServerError, "Game search")
	}

	if count == 0 {
		return NewServiceError(http.StatusNotFound, "Game not found")
	}

	err := p.db.Model(domain).Association("Prices").Find(&prices).Error
	if err != nil {
		return errors.Wrap(err, "search game by id")
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
