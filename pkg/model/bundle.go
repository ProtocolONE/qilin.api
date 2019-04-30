package model

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model/utils"
)

type BundleType string

const (
	BundleStore   BundleType = "store_bundles"
	BundleLootbox BundleType = "lootbox"
)

type (
	// Main Bundle interface
	Bundle interface {
		GetName() *utils.LocalizedString
		IsContains(productId uuid.UUID) (bool, error)
		GetPrice() (string, float32, error)
		GetPackages() ([]Package, error)
		GetGames() ([]*ProductGameImpl, error)
		GetDlc() ([]Dlc, error)
		Buy(customerId uuid.UUID) error
	}

	// Model for StoreBundle and LootBox generalization into Bundle
	BundleEntry struct {
		EntryID   uuid.UUID `gorm:"type:uuid; primary_key"`
		EntryType BundleType
	}

	// Model to link packages and bundle
	BundlePackage struct {
		BundleID  uuid.UUID `gorm:"type:uuid; column:bundle_id"`
		PackageID uuid.UUID `gorm:"type:uuid; column:package_id"`
		Position  int
	}

	StoreBundle struct {
		Model
		Bundle           BundleEntry `gorm:"polymorphic:Entry;"`
		Sku              string
		Name             utils.LocalizedString `gorm:"type:jsonb; not null; default:'{}'"`
		IsUpgradeAllowed bool
		IsEnabled        bool
		VendorID         uuid.UUID
		Vendor           Vendor
		CreatorID        string
		// DiscountPolicy
		Discount       uint
		DiscountBuyOpt BuyOption
		// RegionalRestrinctions
		AllowedCountries pq.StringArray `gorm:"type:text[]"`
		// Bundle payload
		Packages []Package `gorm:"many2many:bundle_packages"`
	}

	BundleService interface {
		CreateStore(vendorId uuid.UUID, userId, name string, packages []uuid.UUID) (bundle *StoreBundle, err error)
		GetStoreList(vendorId uuid.UUID, query, sort string, offset, limit int) (total int, bundles []StoreBundle, err error)
		UpdateStore(bundle *StoreBundle) (result *StoreBundle, err error)
		Get(bundleId uuid.UUID) (bundle Bundle, err error)
		Delete(bundleId uuid.UUID) (err error)
		AddPackages(bundleId uuid.UUID, packages []uuid.UUID) (err error)
		RemovePackages(bundleId uuid.UUID, packages []uuid.UUID) (err error)
	}
)

func (p *BundleEntry) TableName() string {
	return "bundles"
}

func (b *StoreBundle) GetName() *utils.LocalizedString {
	return &b.Name
}

func (b *StoreBundle) IsContains(productId uuid.UUID) (contains bool, err error) {
	for _, pkg := range b.Packages {
		for _, pr := range pkg.Products {
			if pr.GetID() == productId {
				return true, nil
			}
		}
	}
	return
}

func (b *StoreBundle) findPrice(p *Package, currency string) (float32, error) {
	for _, pr := range p.Prices {
		if pr.Currency == currency {
			return pr.Price, nil
		}
	}
	return 0, errors.New(fmt.Sprintf("Price for currency `%s` is not set in package `%s`", currency, p.Name))
}

func (b *StoreBundle) GetPrice() (currency string, price float32, err error) {
	if len(b.Packages) == 0 {
		return
	}
	currency = b.Packages[0].Common["currency"].(string)
	if currency == "" {
		return "", 0, errors.New("Default currency for package is not set")
	}
	for _, p := range b.Packages {
		pkgPrice, err := b.findPrice(&p, currency)
		if err != nil {
			return "", 0, err
		}
		price += pkgPrice
	}
	// Discounts are no count here
	return
}

func (b *StoreBundle) GetPackages() (packages []Package, err error) {
	return b.Packages, nil
}

func (b *StoreBundle) GetGames() (games []*ProductGameImpl, err error) {
	games = []*ProductGameImpl{}
	for _, pkg := range b.Packages {
		for _, pr := range pkg.Products {
			if pr.GetType() == ProductGame {
				game, ok := pr.(*ProductGameImpl)
				if !ok {
					return nil, errors.New("Incorrect product type")
				}
				games = append(games, game)
			}
		}
	}
	return
}

func (b *StoreBundle) GetDlc() (dlcs []Dlc, err error) {
	dlcs = []Dlc{}
	for _, pkg := range b.Packages {
		for _, pr := range pkg.Products {
			if pr.GetType() == ProductGame {
				dlc, ok := pr.(*Dlc)
				if !ok {
					return nil, errors.New("Incorrect product type")
				}
				dlcs = append(dlcs, *dlc)
			}
		}
	}
	return
}

func (b *StoreBundle) Buy(customerId uuid.UUID) (err error) {
	return
}
