package model

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/satori/go.uuid"
)

type Invite struct {
	Model
	Email    string
	VendorId uuid.UUID `gorm:"type:uuid"`
	Roles    Roles     `gorm:"type:jsonb;not null;default:'[]'"`
	Accepted bool
}

//Value is marshaling function
func (j Roles) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

//Scan is unmarshaling function
func (j *Roles) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type Roles []Role
type Role struct {
	Role     string
	Resource ResourceRole
}

type ResourceRole struct {
	Id     string
	Domain string
}

type InviteCreated struct {
	Id  string
	Url string
}
