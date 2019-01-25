package model

import (
	"github.com/satori/go.uuid"
	"time"
)

type Vendor struct {
	ID          uuid.UUID 		`gorm:"type:uuid; primary_key; default:gen_random_uuid()"`
	CreatedAt   time.Time       `gorm:"default:now()"`
	UpdatedAt   time.Time       `gorm:"default:now()"`
	DeletedAt   *time.Time 		`sql:"index"`

	Name string 				`gorm:"unique; not null"`
	// 3d level domain - example.super.com
	Domain3 string 				`gorm:"unique; not null"`
	// Main email for notifications and bills
	Email string 				`gorm:"unique; not null"`
	HowManyProducts string		`gorm:"column:howmanyproducts; not null;"`

	Manager 	*User
	ManagerID 	uuid.UUID       `gorm:"type:uuid"`

	Users       []User          `gorm:"many2many:vendor_users;"`
}

type VendorService interface {
	CreateVendor(g *Vendor) (*Vendor, error)
	UpdateVendor(g *Vendor) (*Vendor, error)
	GetAll(int, int) ([]*Vendor, error)
	FindByID(id uuid.UUID) (*Vendor, error)
}
