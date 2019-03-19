package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
)

//GetUserId is method for retrieving ID by externalID
func GetUserId(db *gorm.DB, id string) (uuid.UUID, error) {
	user := model.User{}
	err := db.Model(&model.User{}).Where("external_id = ?", id).First(&user).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return uuid.Nil, NewServiceErrorf(http.StatusNotFound, "User with external `%s` not found ", id)
		}
		return uuid.Nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get user by external id"))
	}

	return user.ID, nil
}

//GetOwnerForVendor is method for getting owner of vendor
func GetOwnerForVendor(db *gorm.DB, vendorId uuid.UUID) (uuid.UUID, error) {
	vendor := model.Vendor{}
	err := db.Model(&model.Vendor{}).Where("id = ?", vendorId).First(&vendor).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return uuid.Nil, NewServiceErrorf(http.StatusNotFound, "Vendor `%s` not found ", vendorId)
		}
		return uuid.Nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get vendor"))
	}

	return vendor.ManagerID, nil
}
