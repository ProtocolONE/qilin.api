package model

import (
	"github.com/satori/go.uuid"
	"time"
)

// Game is the Central object in the open gde ecosystem, which describes information about the game
// and all related processes and objects.
type Game struct {
	// unique merchant identifier in auth system
	ID uuid.UUID `gorm:"type:uuid; primary_key"`

	//ExternalID *map[string]string `json:"external_id"`

	// game name
	Name string `json:"name"`

	// game description
	Description *LocalizedString `json:"description"`

	// game price
	Prices Prices `json:"prices" gorm:"auto_preload"`

	// date of create merchant in system
	CreatedAt time.Time `json:"created_at"`

	// date of last update merchant in system
	UpdatedAt time.Time `json:"updated_at"`
}

// GameService is a helper service class to interact with Game object.
type GameService interface {
	CreateGame(g *Game) error
	UpdateGame(g *Game) error
	GetAll() ([]*Game, error)
	FindByID(id uuid.UUID) (Game, error)
	FindByName(name string) ([]*Game, error)
}
