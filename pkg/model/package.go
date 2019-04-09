package model

import (
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
)

type BuyOption string

const (
	Whole BuyOption = "whole"
	Part BuyOption = "part"
)

type (
	Package struct {
		Model
		Sku                 string
		Name                string
		Image               string
		ImageCover          string
		ImageThumb          string
		IsUpgradeAllowed    bool
		IsEnabled           bool
		VendorID            uuid.UUID
		Vendor              Vendor
		CreatorID           string
		// DiscountPolicy
		Discount            uint
		DiscountBuyOpt      BuyOption
		// RegionalRestrinctions
		AllowedCountries    pq.StringArray  `gorm:"type:text[]"`
		// Payload
		Products            []Product       `gorm:"-"`
		// Prices in package
		PackagePrices
	}

	PackageProduct struct {
		PackageID           uuid.UUID
		ProductID           uuid.UUID
		Position            int
	}

	PackageService interface {
		Create(vendorId uuid.UUID, name string, prods []uuid.UUID) (*Package, error)
		Get(packageId uuid.UUID) (result *Package, err error)
		GetList(vendorId uuid.UUID, query, orderBy string, offset, limit int) (result []Package, err error)
		AddProducts(packageId uuid.UUID, prods []uuid.UUID) (*Package, error)
		RemoveProducts(packageId uuid.UUID, prods []uuid.UUID) (*Package, error)
		Update(pkg *Package) (result *Package, err error)
		Remove(packageId uuid.UUID) (err error)
	}
)
