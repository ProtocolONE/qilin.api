package model

import (
	"time"
)

type User struct {
	ID        string     `gorm:"type:varchar(64); unique_index"`
	CreatedAt time.Time  `gorm:"default:now()"`
	UpdatedAt time.Time  `gorm:"default:now()"`
	DeletedAt *time.Time `sql:"index"`

	// User nickname for public display
	Nickname string
	Login    string
	Password string
	Lang     string `gorm:"default:'ru'"`
	Currency string `gorm:"default:'usd'"`
	Email    string
	FullName string

	LastSeen *time.Time

	Vendors []Vendor `gorm:"many2many:vendor_users;"`
}

type UserInfo struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Lang     string `json:"lang"`
	Currency string `json:"currency"`
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
	FindByID(id string) (User, error)
	Create(id string, lang string) (User, error)
}
