package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
)

type ownerProvider struct {
	db *Database
}

//GetOwnerForVendor is method for getting owner of vendor
func (provider *ownerProvider) GetOwnerForVendor(vendorId uuid.UUID) (string, error) {
	vendor := model.Vendor{}
	err := provider.db.DB().Model(&model.Vendor{}).Where("id = ?", vendorId).First(&vendor).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found ", vendorId)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get vendor"))
	}

	return vendor.ManagerID, nil
}

func (provider *ownerProvider) GetOwnerForGame(gameId uuid.UUID) (string, error) {
	vendor := model.Vendor{}
	game := model.Game{}
	err := provider.db.DB().Model(&model.Game{}).Where("id = ?", gameId).First(&game).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "Game `%s` not found ", gameId)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get game"))
	}

	if err := provider.db.DB().Model(&model.Vendor{}).Where("id = ?", game.VendorID).First(&vendor).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found ", gameId)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get vendor"))
	}

	return vendor.ManagerID, nil
}

func (provider *ownerProvider) GetOwnerForPackage(packageId uuid.UUID) (string, error) {
	vendor := model.Vendor{}
	pkg := model.Package{}
	err := provider.db.DB().Model(&model.Package{}).Where("id = ?", packageId).First(&pkg).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "Package `%s` not found", packageId)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get package"))
	}

	if err := provider.db.DB().Model(&model.Vendor{}).Where("id = ?", pkg.VendorID).First(&vendor).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found", packageId)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get vendor"))
	}

	return vendor.ManagerID, nil
}

func (provider *ownerProvider) GetOwnerForBundle(bundleId uuid.UUID) (string, error) {
	vendor := model.Vendor{}
	bundle := model.StoreBundle{}
	err := provider.db.DB().Model(&model.StoreBundle{}).Where("id = ?", bundleId).First(&bundle).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "Bundle `%s` not found", bundleId)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get bundle"))
	}

	if err := provider.db.DB().Model(&model.Vendor{}).Where("id = ?", bundle.VendorID).First(&vendor).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found", bundleId)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get vendor"))
	}

	return vendor.ManagerID, nil
}

func NewOwnerProvider(db *Database) model.OwnerProvider {
	return &ownerProvider{db}
}
