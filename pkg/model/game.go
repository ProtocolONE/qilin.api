package model

import (
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model/game"
	"qilin-api/pkg/model/utils"
	"time"
)

type (
	GameGenre struct {
		ID                   string                 `gorm:"primary_key"`
		Title                utils.LocalizedString  `gorm:"column:languages; type:JSONB NOT NULL"`
	}

	GameTag struct {
		ID                   string                 `gorm:"primary_key"`
		Title                utils.LocalizedString  `gorm:"column:languages; type:JSONB NOT NULL"`
	}

	Game struct {
		ID                   uuid.UUID             `gorm:"type:uuid; primary_key"`
		InternalName         string                `gorm:"column:internalName; index: UNIQ"`
		Title                string                `gorm:"column:title; type:text"`
		Developers           string                `gorm:"column:developers; type:text"`
		Publishers           string                `gorm:"column:publishers; type:text"`
		ReleaseDate          time.Time             `gorm:"column:releaseDate; type:timestamp; default:now()"`
		DisplayRemainingTime bool                  `gorm:"column:displayRemainingTime; type:boolean"`
		AchievementOnProd    bool                  `gorm:"column:achievementOnProd; type:boolean"`
		FeaturesCommon       pq.StringArray        `gorm:"column:featuresCommon; type:text[] NOT NULL"`
		FeaturesCtrl         string                `gorm:"column:featuresCtrl; type:text"`
		Platforms            game.Platforms        `gorm:"column:platforms; type:JSONB NOT NULL"`
		Requirements         game.GameRequirements `gorm:"column:requirements; type:JSONB NOT NULL"`
		Languages            game.GameLangs        `gorm:"column:languages; type:JSONB NOT NULL"`
		Genre                pq.StringArray        `gorm:"column:genre"`
		Tags                 pq.StringArray        `gorm:"column:tags; type:text[];"`
	}

	// GameService is a helper service class to interact with Game object.
	GameService interface {
		GetTags([]string) ([]GameTag, error)
		GetGenres([]string) ([]GameGenre, error)
		CreateGame(string) (*Game, error)
		GetList(offset, limit int, technicalName, genre, price, releaseDate, sort string) ([]*Game, error)
	}
)