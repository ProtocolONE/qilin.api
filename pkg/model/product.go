package model

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model/utils"
)

type ProductType string

const (
	ProductGame ProductType = "games"
	ProductDLC  ProductType = "dlcs"
)

type (
	Product interface {
		GetID() uuid.UUID
		GetType() ProductType
		GetName() string
		GetImage() *utils.LocalizedString
	}

	// Model for Game and DLC generalization into Product
	ProductEntry struct {
		EntryID   uuid.UUID `gorm:"type:uuid; primary_key"`
		EntryType ProductType
	}

	ProductEntryArray []ProductEntry
)

func (p *ProductEntry) TableName() string {
	return "products"
}

type ProductService interface {
	SpecializationIds([]uuid.UUID) (games []uuid.UUID, dlcs []uuid.UUID, err error)
}

func (p ProductEntryArray) GetUUIDs() (result []uuid.UUID) {
	result = []uuid.UUID{}
	for _, prod := range p {
		result = append(result, prod.EntryID)
	}
	return
}
