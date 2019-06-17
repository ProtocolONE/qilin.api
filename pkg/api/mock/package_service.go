package mock

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
)

type packageService struct{}

func NewPackageService() (model.PackageService, error) {
	return &packageService{}, nil
}

func (*packageService) Create(vendorId uuid.UUID, userId, name string, prods []uuid.UUID) (result *model.Package, err error) {
	return &model.Package{}, nil
}

func (*packageService) AddProducts(packageId uuid.UUID, prods []uuid.UUID) (result *model.Package, err error) {
	return &model.Package{}, nil
}

func (*packageService) RemoveProducts(packageId uuid.UUID, prods []uuid.UUID) (result *model.Package, err error) {
	return &model.Package{}, nil
}

func (*packageService) Get(packageId uuid.UUID) (result *model.Package, err error) {
	return &model.Package{}, nil
}

func (*packageService) GetList(userId string, vendorId uuid.UUID, query, sort string, offset, limit int, filter model.PackageListingFilter) (total int, result []model.Package, err error) {
	return 0, []model.Package{}, nil
}

func (*packageService) Update(pkg *model.Package) (result *model.Package, err error) {
	return &model.Package{}, nil
}

func (*packageService) Remove(packageId uuid.UUID) (err error) {
	return nil
}
