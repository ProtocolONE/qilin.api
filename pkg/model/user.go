package model

import "time"

type User struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	// User nickname for public display
	Nickname string `bson:"name"`
	Login string `bson:"login"`
	Password string `bson:"password"`
}

type UserInfo struct {
	Id uint				`json:"id"`
	Nickname string		`json:"nickname"`
	Avatar string		`json:"avatar"`
}

type LoginResult struct {
	AccessToken string		`json:"access_token"`
	User 		UserInfo 	`json:"user"`
}

type AppState struct {
	User UserInfo			`json:"user"`
	//...
}

// UserService is a helper service class to interact with User.
type UserService interface {
	CreateUser(g *User) error
	UpdateUser(g *User) error
	FindByID(id int) (User, error)
	FindByLoginAndPass(login, pass string) (User, error)
	Login(login, pass string) (LoginResult, error)
}
