package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"qilin-api/pkg/model"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *Database) (model.ProductService, error) {
	return &ProductService{db.database}, nil
}

func (p *ProductService) SpecializationIds(productIds []uuid.UUID) (games []uuid.UUID, dlcs []uuid.UUID, err error) {
	games = []uuid.UUID{}
	dlcs = []uuid.UUID{}
	entries := []model.ProductEntry{}
	if len(productIds) > 0 {
		err := p.db.Where("entry_id in (?)", productIds).Find(&entries).Error
		if err != nil {
			return nil, nil, errors.Wrap(err, "Retrieve products")
		}
	}
	for _, entry := range entries {
		switch entry.EntryType {
		case model.ProductGame:
			games = append(games, entry.EntryID)
			break
		case model.ProductDLC:
			dlcs = append(dlcs, entry.EntryID)
			break
		}
	}
	return
}

func (p *ProductService) GetPackages(productId uuid.UUID) (packages []model.Package, err error) {

	packages = []model.Package{}

	prodPackages := []model.PackageProduct{}
	err = p.db.
		Where("product_id = ?", productId).
		Find(&prodPackages).Error
	if err != nil {
		return nil, errors.Wrap(err, "Get packages by product")
	}

	// TODO: Return related packages...

	return
}
