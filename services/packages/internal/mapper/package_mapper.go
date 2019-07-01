package mapper

import (
	"qilin-api/services/packages/internal/model"
	"qilin-api/services/packages/proto"
)

func ToProto(p *model.Package) *proto.Package {
	if p == nil {
		return nil
	}

	return &proto.Package{
		Id:               p.PackageID,
		VendorId:         p.VendorID,
		AllowedCountries: p.AllowedCountries,
		DefaultProductId: p.DefaultProductID,
		Name:             p.Name,
		ImageCover:       p.ImageCover,
		Image:            p.Image,
		ImageThumb:       p.ImageThumb,
		IsUpgradeAllowed: p.IsUpgradeAllowed,
		Sku:              p.Sku,
		IsEnabled:        p.IsEnabled,
		Discount:         p.Discount,
		CreatorId:        p.CreatorID,
		DiscountOption:   proto.DiscountOption(p.DiscountBuyOpt),
		Products:         p.Products,
	}
}

func FromProto(p *proto.Package) model.Package {
	return model.Package{
		PackageID:        p.Id,
		DiscountBuyOpt:   int32(p.DiscountOption),
		Discount:         p.Discount,
		IsEnabled:        p.IsEnabled,
		Sku:              p.Sku,
		IsUpgradeAllowed: p.IsUpgradeAllowed,
		ImageThumb:       p.ImageThumb,
		Image:            p.Image,
		ImageCover:       p.ImageCover,
		Name:             p.Name,
		DefaultProductID: p.DefaultProductId,
		CreatorID:        p.DefaultProductId,
		AllowedCountries: p.AllowedCountries,
		VendorID:         p.VendorId,
		Products:         p.Products,
	}
}
