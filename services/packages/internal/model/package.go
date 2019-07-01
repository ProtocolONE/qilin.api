package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"qilin-api/services/packages/proto"
)

type Package struct {
	ID               primitive.ObjectID     `bson:"_id"`
	PackageID        string                 `json:"package_id"`
	Sku              string                 `json:"sku"`
	Name             *proto.LocalizedString `json:"name"`
	Image            *proto.LocalizedString `json:"image"`
	ImageCover       *proto.LocalizedString `json:"image_cover"`
	ImageThumb       *proto.LocalizedString `json:"image_thumb"`
	IsUpgradeAllowed bool                   `json:"is_upgrade_allowed"`
	IsEnabled        bool                   `json:"is_enabled"`
	DefaultProductID string                 `json:"default_product_id"`
	VendorID         string                 `json:"vendor_id"`
	CreatorID        string                 `json:"creator_id"`
	Discount         uint32                 `json:"discount"`
	DiscountBuyOpt   int32                  `json:"discount_buy_opt"`
	AllowedCountries []string               `json:"allowed_countries"`
	Products         []*proto.Product        `json:"products"`
}
