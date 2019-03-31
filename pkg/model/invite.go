package model

import "github.com/satori/go.uuid"

type Invite struct {
	Model
	Email    string
	VendorId uuid.UUID `gorm:"type:uuid"`
	Roles    JSONB     `gorm:"type:jsonb"`
}

type InviteCreated struct {
	Id  string
	Url string
}
