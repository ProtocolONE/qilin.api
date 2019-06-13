package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm/utils"
)

type keyPackageService struct {
	db *Database
	keyStreamService model.KeyStreamService
}

func NewKeyPackageService(db *Database, keyStreamService model.KeyStreamService) model.KeyPackageService {
	return &keyPackageService{db: db, keyStreamService: keyStreamService}
}

func (service *keyPackageService) Get(keyPackageId uuid.UUID) (*model.KeyPackage, error) {
	keyPackage := &model.KeyPackage{}

	err := service.db.DB().Model(model.KeyPackage{}).Where("id = ?", keyPackageId).First(keyPackage).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, NewServiceError(http.StatusNotFound, err)
		}
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	return keyPackage, nil
}

func (service *keyPackageService) Create(packageId uuid.UUID, name string, providerType model.KeyStreamType) (*model.KeyPackage, error) {
	if exist, err := utils.CheckExists(service.db.DB(), model.Package{}, packageId); !exist || err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, NewServiceError(http.StatusNotFound, err)
		}
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	if len(name) == 0 {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "Name must be not null")
	}

	keyPackage := &model.KeyPackage{
		Name:          name,
		KeyStreamType: providerType,
		PackageID:     packageId,
	}
	keyPackage.ID = uuid.NewV4()

	err := service.db.DB().Model(model.KeyPackage{}).Create(keyPackage).Error
	if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	_, err = service.keyStreamService.Create(providerType)
	if err != nil {
		return nil, err
	}

	return keyPackage, nil
}

func (service *keyPackageService) Update(keyPackageId uuid.UUID, name string) (*model.KeyPackage, error) {
	keyPackage := &model.KeyPackage{}

	err := service.db.DB().Model(model.KeyPackage{}).Where("id = ?", keyPackageId).First(keyPackage).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, NewServiceError(http.StatusNotFound, err)
		}
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	keyPackage.Name = name
	err = service.db.DB().Save(keyPackage).Error
	if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	return keyPackage, nil
}

func (service *keyPackageService) List(packageId uuid.UUID) ([]model.KeyPackage, error) {
	var keyPackages []model.KeyPackage
	err := service.db.DB().Model(model.KeyPackage{}).Where("package_id = ?", packageId).Order("created_at desc").Find(&keyPackages).Error
	if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, err)
	}

	return keyPackages, nil
}
