package model

import "github.com/satori/go.uuid"

type ProductType string

const (
	ProductGame     ProductType = "games"
	ProductDLC      ProductType = "dlcs"
)

type (
	Product interface {
		GetID() uuid.UUID
		GetType() ProductType
		GetName() string
		GetImage(lang string) string
	}

	// Model for Game and DLC generalization into Product
	ProductEntry struct {
		EntryID     uuid.UUID       `gorm:"type:uuid; primary_key"`
		EntryType   ProductType
	}
)

func (p *ProductEntry) TableName() string {
	return "products"
}
