package model

import (
	"github.com/satori/go.uuid"
)

type (
	Achievement struct {
		Model
		Name      string
		ProductID uuid.UUID
	}

	AchievementService interface {
		Create(name string) (*Achievement, error)
		Delete(achId uuid.UUID) error
	}
)
