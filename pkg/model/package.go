package model

import "time"

type Asset interface {
	UniqueID() string
}

type Package struct {
	// unique merchant identifier in auth system
	ID string `json:"id" validate:"required"`

	// game name
	Name string `bson:"name"`

	Assets []Asset `bson:"assets"`

	// date of create merchant in system
	CreatedAt time.Time `json:"created_at"`

	// date of last update merchant in system
	UpdatedAt time.Time `json:"updated_at"`
}

// GameService is a helper service class to interact with Game object.
type PackageService interface {
	CreatePackage(g *Game) error
	UpdatePackage(g *Game) error
	DeletePackage(g *Game) error
	GetAll() ([]*Game, error)
	FindByID(id string) (*Game, error)
	FindByName(name string) ([]*Game, error)
}
