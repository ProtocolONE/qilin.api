package model

import (
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
		Packages []Package `gorm:"many2many:bundle_packages;jointable_foreignkey:bundle_id;"`
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

func (b *StoreBundle) GetPrice() (currency string, price float32, discount float32, err error) {
	if len(b.Packages) == 0 {
		return
	}
	currency = b.Packages[0].GetCurrency()
	var fprice float32
	for _, p := range b.Packages {
		// Search for default price
		for _, pr := range p.Prices {
			if pr.Currency == currency {
				price += pr.Price
				fprice += pr.Price - (pr.Price * float32(p.Discount) * 0.01)
				break
			}
		}
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
