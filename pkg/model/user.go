package model

import (
	"github.com/satori/go.uuid"
	"time"
)

type User struct {
	ID        uuid.UUID  `gorm:"type:uuid; primary_key; default:gen_random_uuid()"`
	CreatedAt time.Time  `gorm:"default:now()"`
	UpdatedAt time.Time  `gorm:"default:now()"`
	DeletedAt *time.Time `sql:"index"`

	// User nickname for public display
	Nickname string
	Login    string
	Password string
	Lang     string `gorm:"default:'ru'"`
	Currency string `gorm:"default:'usd'"`

	Vendors []Vendor `gorm:"many2many:vendor_users;"`
}

type UserInfo struct {
	Id       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"`
	Lang     string    `json:"lang"`
	Currency string    `json:"currency"`
}

type LoginResult struct {
	AccessToken string   `json:"access_token"`
	User        UserInfo `json:"user"`
}

type AppState struct {
	User UserInfo `json:"user"`
	//...
}

// UserService is a helper service class to interact with User.
type UserService interface {
	UpdateUser(g *User) error
	FindByID(id *uuid.UUID) (User, error)
	Login(login, pass string) (LoginResult, error)
	Register(login, pass, lang string) (uuid.UUID, error)
	ResetPassw(email string) error
}
