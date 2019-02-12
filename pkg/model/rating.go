package model

import (
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

type (

	//GameRating is domain object with all game ratings
	GameRating struct {
		gorm.Model

		PEGI JSONB `gorm:"type:jsonb"`
		ESRB JSONB `gorm:"type:jsonb"`
		BBFC JSONB `gorm:"type:jsonb"`
		USK  JSONB `gorm:"type:jsonb"`
		CERO JSONB `gorm:"type:jsonb"`

		GameID uuid.UUID `gorm:"type:uuid"`
	}
)

const (
	DescriptorsField = "descriptors"
	AgeField = "ageRestrictions"
)