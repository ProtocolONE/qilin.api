package model

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model/game"
	"time"
)

type (
	Game struct {
		ID                   uuid.UUID             `gorm:"type:uuid; primary_key"`
		InternalName         string                `gorm:"column:internalName; index: UNIQ"`
		Title                string                `gorm:"column:title; type:text"`
		Developers           string                `gorm:"column:developers; type:text"`
		Publishers           string                `gorm:"column:publishers; type:text"`
		ReleaseDate          time.Time             `gorm:"column:releaseDate; type:timestamp; default:now()"`
		DisplayRemainingTime bool                  `gorm:"column:displayRemainingTime; type:boolean"`
		AchievementOnProd    bool                  `gorm:"column:achievementOnProd; type:boolean"`
		Features             game.Features         `gorm:"column:features; type:JSONB NOT NULL"`
		Platforms            game.Platforms        `gorm:"column:platforms; type:JSONB NOT NULL"`
		Requirements         game.GameRequirements `gorm:"column:requirements; type:JSONB NOT NULL"`
		Languages            game.GameLangs        `gorm:"column:languages; type:JSONB NOT NULL"`
		Genre                game.GameTags         `gorm:"column:genre; type:JSONB NOT NULL"`
		Tags                 game.GameTags         `gorm:"column:tags; type:JSONB NOT NULL"`
	}

	// GameService is a helper service class to interact with Game object.
	GameService interface {
		CreateGame(string) (*Game, error)
		UpdateGame(g *Game) error
		GetAll() ([]*Game, error)
		FindByID(id uuid.UUID) (Game, error)
		FindByName(name string) ([]*Game, error)
	}
)