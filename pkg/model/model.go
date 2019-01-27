package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// Model base model definition, including fields `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`, which could be embedded in your models
//    type User struct {
//      model.Model
//    }
type Model struct {
	ID        uuid.UUID `gorm:"type:uuid; primary_key"`
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time `sql:"index"`
}
