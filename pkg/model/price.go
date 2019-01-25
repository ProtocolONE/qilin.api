package model

import (
	"time"

	"github.com/jinzhu/gorm"

	uuid "github.com/satori/go.uuid"
)

type BasePrice struct {
	ID uuid.UUID `gorm:"type:uuid; primary_key"`

	UpdatedAt *time.Time

	Common   JSONB `gorm:"type:JSONB"`
	PreOrder JSONB `gorm:"type:JSONB"`

	Prices []Price `gorm:"foreignkey:BasePriceID" field:"ignore"`
}

type Price struct {
	gorm.Model

	BasePriceID uuid.UUID

	Currency string
	Vat      int32
	Price    float32
}

//TableName is HACK method for merging this model with "games" table
func (BasePrice) TableName() string {
	return "games"
}
