package model

import (
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
		GetPrice() (string, float32, float32, error)
		GetPackages() ([]Package, error)
		GetGames() ([]*ProductGameImpl, error)
		GetDlc() ([]Dlc, error)
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

	BundleListingFilter func(uuid.UUID) (bool, error)

	BundleService interface {
		CreateStore(vendorId uuid.UUID, userId, name string, packages []uuid.UUID) (bundle Bundle, err error)
		GetStoreList(vendorId uuid.UUID, query, sort string, offset, limit int, filterFunc BundleListingFilter) (total int, bundles []Bundle, err error)
		UpdateStore(bundle Bundle) (result Bundle, err error)

		Get(bundleId uuid.UUID) (bundle Bundle, err error)
		Delete(bundleId uuid.UUID) (err error)
		AddPackages(bundleId uuid.UUID, packages []uuid.UUID) (err error)
		RemovePackages(bundleId uuid.UUID, packages []uuid.UUID) (err error)
	}
)

func (p *BundleEntry) TableName() string {
	return "bundles"
}
