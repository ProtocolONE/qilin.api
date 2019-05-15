package mock

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
)

type productService struct{}

func NewProductService() (model.ProductService, error) {
	return &productService{}, nil
}

func (p *productService) SpecializationIds(productIds []uuid.UUID) (games []uuid.UUID, dlcs []uuid.UUID, err error) {
	games = productIds
	dlcs = []uuid.UUID{}
	return
}
