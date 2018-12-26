package model

import "time"

type Vendor struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	Name string 				`gorm:"column:name; not null;unique"`
	// 3d level domain - example.super.com
	Domain3 string 				`gorm:"column:domain3; not null;unique"`
	// Main email for notifications and bills
	Email string 				`gorm:"column:email; not null;unique"`

	Manager User 				`gorm:"foreignkey:ManagerId; association_foreignkey:Refer"`
	ManagerId int				`gorm:"column:manager_id"`
}

type VendorService interface {
	CreateVendor(g *Vendor) error
	UpdateVendor(g *Vendor) error
	GetAll(int, int) ([]*Vendor, error)
	FindByID(id uint) (Vendor, error)
}
