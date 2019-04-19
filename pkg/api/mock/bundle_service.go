package mock

import (
	uuid "github.com/satori/go.uuid"
	"qilin-api/pkg/model"
)

type bundleService struct {

}

func NewBundleService() (*bundleService, error) {
	return &bundleService{}, nil
}

func (*bundleService) CreateStore(vendorId uuid.UUID, userId, name string, packages []uuid.UUID) (bundle *model.StoreBundle, err error) {
	return &model.StoreBundle{}, nil
}

func (*bundleService) GetStoreList(vendorId uuid.UUID, query, sort string, offset, limit int) (result []model.StoreBundle, err error) {
	return []model.StoreBundle{}, nil
}

func (*bundleService) Get(bundleId uuid.UUID) (bundle model.Bundle, err error) {
	return &model.StoreBundle{}, nil
}

func (*bundleService) Delete(bundleId uuid.UUID) (err error) {
	return nil
}

func (*bundleService) UpdateStore(bundle *model.StoreBundle) (result *model.StoreBundle, err error) {
	return bundle, nil
}

func (p *bundleService) AddPackages(bundleId uuid.UUID, packageIds []uuid.UUID) (err error) {
	return nil
}

func (p *bundleService) RemovePackages(bundleId uuid.UUID, packages []uuid.UUID) (err error)  {
	return nil
}