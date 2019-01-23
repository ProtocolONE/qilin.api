package model

import (
	"github.com/satori/go.uuid"
	"time"
)

type Vendor struct {
	ID         uuid.UUID 		`gorm:"type:uuid; primary_key; default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time 		`sql:"index"`

	Name string 				`gorm:"column:name; not null;unique"`
	// 3d level domain - example.super.com
	Domain3 string 				`gorm:"column:domain3; not null;unique"`
	// Main email for notifications and bills
	Email string 				`gorm:"column:email; not null;unique"`
	HowManyProducts string		`gorm:"column:hawmanyproducts; not null;"`

	Manager 	User 			`gorm:"foreignkey:ManagerId; association_foreignkey:Refer"`
	ManagerId 	*uuid.UUID		`gorm:"column:manager_id; type:uuid"`
}

type VendorService interface {
	CreateVendor(g *Vendor) (*Vendor, error)
	UpdateVendor(g *Vendor) (*Vendor, error)
	GetAll(int, int) ([]*Vendor, error)
	FindByID(id uuid.UUID) (*Vendor, error)
}
