package model

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)

type PackagePrices struct {
	Common   JSONB `gorm:"type:JSONB"`
	PreOrder JSONB `gorm:"type:JSONB"`

	Prices []Price `gorm:"foreignkey:BasePriceID" field:"ignore"`
}

type BasePrice struct {
	ID            uuid.UUID `gorm:"type:uuid; primary_key"`
	UpdatedAt     *time.Time
	PackagePrices `field:"extend"`
}

type Price struct {
	gorm.Model

	BasePriceID uuid.UUID

	Currency string
	Vat      int32
	Price    float32 `gorm:"type:decimal(10,2)"`
}

//TableName is HACK method for merging this model with "games" table
func (BasePrice) TableName() string {
	return "packages"
}

type PriceService interface {
	GetBase(id uuid.UUID) (*BasePrice, error)
	UpdateBase(id uuid.UUID, price *BasePrice) error
	Delete(id uuid.UUID, price *Price) error
	Update(id uuid.UUID, price *Price) error
	// TODO: For temporary purpose
	GetDefaultPackage(gameId uuid.UUID) (uuid.UUID, error)
}
