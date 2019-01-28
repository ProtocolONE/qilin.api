package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// Media is ...
type Media struct {
	ID uuid.UUID `gorm:"type:uuid; primary_key"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// localized cover image of game
	CoverImage JSONB `gorm:"type:jsonb"`
	CoverVideo JSONB `gorm:"type:jsonb"`
	// localized cover video of game
	Trailers JSONB `gorm:"type:jsonb"`
	// localized cover video of game
	Store   JSONB `gorm:"type:jsonb"`
	Capsule JSONB `gorm:"type:jsonb"`
}

type MediaService interface {
	Get(id uuid.UUID) (*Media, error)
	Update(id uuid.UUID, media *Media) error
}

func (Media) TableName() string {
	return "games"
}
