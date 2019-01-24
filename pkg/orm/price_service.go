package orm

import (
	"net/http"
	"qilin-api/pkg/model"

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

func (p *PriceService) Get(id uuid.UUID) (*model.BasePrice, error) {

	result := &model.BasePrice{}
	err := p.db.Preload("Prices").Select(model.SelectFields(result)).Where("id = ?", id).First(result).Error

	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "search game by id")
	}

	return result, err
}

func (p *PriceService) Update(id uuid.UUID, price *model.BasePrice) error {

	domain := &model.BasePrice{ID: id}
	err := p.db.Preload("Prices").Select(model.SelectFields(domain)).Where("id = ?", id).First(domain).Error

	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	price.ID = domain.ID

	err = p.db.Save(price).Error

	if err != nil {
		return errors.Wrap(err, "save prices for game")
	}

	return nil
}
