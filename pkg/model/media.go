package model

import (
	"qilin-api/pkg/model/utils"
	"time"

	uuid "github.com/satori/go.uuid"
)

// Media is ...
type Media struct {
	ID uuid.UUID `gorm:"type:uuid; primary_key"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// localized cover image of game
	CoverImage utils.LocalizedString `gorm:"type:jsonb"`
	CoverVideo utils.LocalizedString `gorm:"type:jsonb"`
	// localized trailers video of game
	Trailers utils.LocalizedStringArray `gorm:"type:jsonb"`
	// localized screenshots video of game
	Screenshots utils.LocalizedStringArray `gorm:"type:jsonb"`

	// localized store of game
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
