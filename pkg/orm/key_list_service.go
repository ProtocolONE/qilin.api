package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
)

type keyListService struct {
	db *Database

}

func NewKeyListService(db *Database) model.KeyListService {
	return &keyListService{db: db}
}

func (service *keyListService) AddKeys(keyPackageId uuid.UUID, keys []string) error {
	keyPackage := model.KeyPackage{}
	err := service.db.DB().Model(model.KeyPackage{}).Where("id = ?", keyPackageId).First(&keyPackage).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return NewServiceErrorf(http.StatusNotFound, "Key Package with id `%s` not found", keyPackageId)
		}
		return NewServiceError(http.StatusInternalServerError, err)
	}

	streamProvider, err := NewKeyListProvider(keyPackage.KeyStreamID, service.db)
	if err != nil {
		return err
	}
	return streamProvider.AddKeys(keys)
}
