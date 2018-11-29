package model

import "time"

type Owner struct {
	// unique merchant identifier in auth system
	ID string `json:"id" validate:"required"`

	// game name
	Name string `bson:"name"`

	// date of create merchant in system
	CreatedAt time.Time `json:"created_at"`

	// date of last update merchant in system
	UpdatedAt time.Time `json:"updated_at"`
}
