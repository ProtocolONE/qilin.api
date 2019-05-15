package model

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model/utils"
)

type (
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
)

func (b *StoreBundle) GetID() uuid.UUID {
	return b.ID
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

func (b *StoreBundle) findPrice(p *Package, currency string) (*Price, error) {
	for _, pr := range p.Prices {
		if pr.Currency == currency {
			return &pr, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Price for currency `%s` is not set in package `%s`", currency, p.Name))
}

func (b *StoreBundle) GetPrice() (currency string, price float32, discount float32, err error) {
	if len(b.Packages) == 0 {
		return
	}
	currency = b.Packages[0].Common["currency"].(string)
	if currency == "" {
		return "", 0, 0, errors.New("Default currency for package is not set")
	}
	var fprice float32
	for _, p := range b.Packages {
		pkgPrice, err := b.findPrice(&p, currency)
		if err != nil {
			return "", 0, 0, err
		}
		price += pkgPrice.Price
		fprice += pkgPrice.Price - (pkgPrice.Price * float32(p.Discount) * 0.01)
	}
	if price != 0 {
		discount = 100 * (price - fprice) / price
	}
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
			if pr.GetType() == ProductDLC {
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
