package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

//Discount is struct for saving and manipulating in DB
type Discount struct {
	Model

	Title       JSONB   `sql:"type:jsonb;not null"`
	Description JSONB   `sql:"type:jsonb"`
	Rate        float32 `sql:"type:float;not null"`
	DateStart   time.Time
	DateEnd     time.Time
	GameID      uuid.UUID `gorm:"type:uuid;not null"`
}
