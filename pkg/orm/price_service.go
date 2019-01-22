package orm

import (
	"qilin-api/pkg/model"
	"github.com/satori/go.uuid"
	"github.com/jinzhu/gorm"
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
	
	return result, nil
}
