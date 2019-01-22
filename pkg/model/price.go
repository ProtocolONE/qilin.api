package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Price struct {
	ID uuid.UUID `gorm:"type:uuid; primary_key"`

	UpdatedAt *time.Time

	Normal   JSONB      `gorm:"type:JSONB"`
	PreOrder JSONB      `gorm:"type:JSONB"`
	Prices   JSONBArray `gorm:"type:JSONB[]"`
}

//TableName is HACK method for merging this model with "games" table
func (Price) TableName() string {
	return "games"
}
