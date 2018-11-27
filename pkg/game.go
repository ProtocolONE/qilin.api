package qilin

import "time"

// Game is the Central object in the open gde ecosystem, which describes information about the game and all related processes and objects.
type Game struct {
	// unique merchant identifier in auth system
	ID string `json:"id" validate:"required"`

	ExternalID *map[string]string `json:"external_id"`

	// game name
	Name string `bson:"name"`

	// game description
	Description *LocalizedString `bson:"description"`

	// game price
	//Price *GamePrice `bson:"price"`

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
	FindByID(id string) (*Game, error)
	FindByName(name string) ([]*Game, error)
}
