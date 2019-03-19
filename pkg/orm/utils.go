package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
)

//GetUserId is method for retrieving ID by externalID
func GetUserId(db *gorm.DB, id string) (string, error) {
	user := model.User{}
	err := db.Model(&model.User{}).Where("external_id = ?", id).First(&user).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "User with external `%s` not found ", id)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get user by external id"))
	}

	return user.ID, nil
}

//GetOwnerForVendor is method for getting owner of vendor
func GetOwnerForVendor(db *gorm.DB, vendorId uuid.UUID) (string, error) {
	vendor := model.Vendor{}
	err := db.Model(&model.Vendor{}).Where("id = ?", vendorId).First(&vendor).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found ", vendorId)
		}
		return "", NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get vendor"))
	}

	return vendor.ManagerID, nil
}
