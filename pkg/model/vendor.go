package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Vendor struct {
	gorm.Model

	ID uint 					`gorm:"primary_key; AUTO_INCREMENT"`
	Name string 				`gorm:"column:name; not null;unique"`
	// 3d level domain - example.super.com
	Domain3 string 				`gorm:"column:domain3; not null;unique"`
	// Main email for notifications and bills
	Email string 				`gorm:"column:email; not null;unique"`

	Manager User 				`gorm:"foreignkey:ManagerId; association_foreignkey:Refer"`
	ManagerId int				`gorm:"column:manager_id"`

	CreatedAt time.Time 		`gorm:"column:created_at"`
	UpdatedAt time.Time 		`gorm:"column:updated_at"`
}

type VendorService interface {
	CreateVendor(g *Vendor) error
	UpdateVendor(g *Vendor) error
	GetAll() ([]*Vendor, error)
	FindByID(id uint) (Vendor, error)
	FindByName(name string) ([]*Vendor, error)
}
