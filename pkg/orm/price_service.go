package orm

import (
	"fmt"
	"net/http"
	"qilin-api/pkg/model"
	"github.com/satori/go.uuid"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// PriceService is service to interact with database and Media object.
type PriceService struct {
	db *gorm.DB
}

// NewPriceService initialize this service.
func NewPriceService(db *Database) (*PriceService, error) {
	return &PriceService{db.database}, nil
}

func (p *PriceService) Get(id uuid.UUID) (*model.Price, error) {

	result := &model.Price{}
	err := p.db.Select(model.SelectFields(result)).Where("ID = ?", id).First(&result).Error

	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "search game by id")
	}

	return result, err
}

func (p *PriceService) Update(id uuid.UUID, price *model.Price) error {

	domain := &model.Price{}
	err := p.db.Select(model.SelectFields(domain)).Where("ID = ?", id).First(domain).Error

	fmt.Printf("%#v", domain)

	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Game not found")
	} else if err != nil {
		return errors.Wrap(err, "search game by id")
	}

	if price.UpdatedAt.Before(*domain.UpdatedAt) && !price.UpdatedAt.Equal(*domain.UpdatedAt) {
		// return NewServiceError(http.StatusConflict, "Object has new changes")
	}

	price.ID = domain.ID
	err = p.db.Save(price).Error

	if err != nil {
		return errors.Wrap(err, "save prices for game")
	}

	return nil
}

