package model

import (
	"github.com/satori/go.uuid"
	"time"
)

type Game struct {
	ID                      uuid.UUID       `gorm:"type:uuid; primary_key"`
	InternalName            string          `gorm:"column:internalName; type:text"`
	Title                   string          `gorm:"column:title; type:text"`
	Developers              string          `gorm:"column:developers; type:text"`
	Publishers              string          `gorm:"column:publishers; type:text"`
	ReleaseDate             time.Time       `gorm:"column:releaseDate; type:timestamp"`
	DisplayRemainingTime    bool            `gorm:"column:displayRemainingTime; type:boolean"`
	AchievementOnProd       bool            `gorm:"column:achievementOnProd; type:boolean"`
	Features                []string        `gorm:"column:features; type:text[]"`
	Platforms               []string        `gorm:"column:platforms; type:text[]"`
	Requirements            []string        `gorm:"column:requirements; type:text[]"`
	Languages               []string        `gorm:"column:languages; type:text[]"`
	Genre                   []string        `gorm:"column:genre; type:text[]"`
	Tags                    []string        `gorm:"column:tags; type:text[]"`
}

// GameService is a helper service class to interact with Game object.
type GameService interface {
	CreateGame(string) (*Game, error)
	UpdateGame(g *Game) error
	GetAll() ([]*Game, error)
	FindByID(id uuid.UUID) (Game, error)
	FindByName(name string) ([]*Game, error)
}
