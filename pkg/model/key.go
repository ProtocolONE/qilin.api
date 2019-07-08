package model

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type Key struct {
	Model

	KeyStreamID    uuid.UUID
	ActivationCode string `gorm:"index:activation_code;not null"`
	RedeemTime     *time.Time
}
