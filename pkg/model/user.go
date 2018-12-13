package model

import "time"

type User struct {
	// unique user identifier
	ID string `json:"id" validate:"required"`

	// User nickname for public display
	Nickname string `bson:"name"`

	// date of create user in system
	CreatedAt time.Time `json:"created_at"`

	// date of last update user in system
	UpdatedAt time.Time `json:"updated_at"`
}

// UserService is a helper service class to interact with User.
type UserService interface {
	CreateUser(g *User) error
	UpdateUser(g *User) error
	GetAll() ([]*User, error)
	FindByID(id uint) (User, error)
	FindByName(name string) ([]*User, error)
}
