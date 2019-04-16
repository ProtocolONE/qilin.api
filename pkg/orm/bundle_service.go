package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/random"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm/utils"
	"strings"
	"time"
)

type bundleService struct {
	db *gorm.DB
	gameService model.GameService
	packageService model.PackageService
}

func NewBundleService(db *Database) (model.BundleService, error) {
	gameService, _ := NewGameService(db)
	packageService, _ := NewPackageService(db, gameService)
	return &bundleService{db.database, gameService, packageService}, nil
}

func (p *bundleService) CreateStore(vendorId uuid.UUID, name string, packages []uuid.UUID) (bundle *model.StoreBundle, err error) {

	if len(strings.Trim(name, " \r\n\t")) == 0 {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "Name is empty")
	}

	pkgObjs := []model.Package{}
	if len(packages) > 0 {
		err = p.db.Where("id in (?)", packages).Find(&pkgObjs).Error
		if err != nil {
			return nil, errors.Wrap(err, "Search packages")
		}
	}
	if len(pkgObjs) == 0 {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "No any package")
	}

	vendorFound, err := utils.CheckExists(p.db, model.Vendor{}, vendorId)
	if err != nil {
		return nil, errors.Wrap(err, "Vendor exists")
	}
	if !vendorFound {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "Invalid vendor")
	}

	newBundle := model.StoreBundle{
		Model: model.Model{ID: uuid.NewV4()},
		Sku: random.String(8, "123456789"),
		Name: name,
		VendorID: vendorId,
		IsEnabled: false,
	}
	newBundle.Bundle.EntryID = newBundle.ID
	err = p.db.Create(&newBundle).Error
	if err != nil {
		return nil, errors.Wrap(err, "While create new bundle")
	}

	db := p.db.Begin()
	// We walks `packages` first cuz want to persistent ordering
	for index, pkgID := range packages {
		for _, pkg := range pkgObjs {
			if pkgID != pkg.ID {
				continue
			}
			err = db.Create(&model.BundlePackage{
				PackageID: pkg.ID,
				BundleID:  newBundle.ID,
				Position:  index + 1,
			}).Error
			if err != nil {
				db.Rollback()
				return nil, errors.Wrap(err, "While append packages into bundle")
			}
		}
	}
	err = db.Commit().Error
	if err != nil {
		return nil, errors.Wrap(err, "While commit packages")
	}

	bundleIfce, err := p.Get(newBundle.ID)

	return bundleIfce.(*model.StoreBundle), nil
}

func (p *bundleService) GetStoreList(vendorId uuid.UUID, query, sort string, offset, limit int) (result []model.StoreBundle, err error) {
	orderBy := ""
	orderBy = "created_at ASC"
	if sort != "" {
		switch sort {
		case "-date":
			orderBy = "created_at DESC"
		case "+date":
			orderBy = "created_at ASC"
		case "-name":
			orderBy = "name DESC"
		case "+name":
			orderBy = "name ASC"
		}
	}

	conds := []string{}
	vals := []interface{}{}

	if query != "" {
		conds = append(conds, `name ilike ?`)
		vals = append(vals, "%"+query+"%")
		// TODO: Add another kinds for searching
	}

	err = p.db.
		Model(model.StoreBundle{}).
		Where(`vendor_id = ?`, vendorId).
		Where(strings.Join(conds, " or "), vals...).
		Order(orderBy).
		Limit(limit).
		Offset(offset).
		Find(&result).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch store bundle list")
	}

	return
}

func (p *bundleService) Get(bundleId uuid.UUID) (bundle model.Bundle, err error) {

	entry := model.BundleEntry{}
	err = p.db.Where(model.BundleEntry{EntryID: bundleId}).Find(&entry).Error
	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, "Bundle not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "Retrieve bundle entry")
	}

	if entry.EntryType == model.BundleStore {
		bundle := &model.StoreBundle{}
		err = p.db.Where("id = ?", bundleId).Find(&bundle).Error
		if err != nil {
			return nil, errors.Wrap(err, "Retrieve store bundle")
		}
		bundlePkgs := []model.BundlePackage{}
		err = p.db.
			Where("bundle_id = ?", bundleId).
			Order("position asc").
			Find(&bundlePkgs).Error
		if err != nil {
			return nil, errors.Wrap(err, "Retrieve bundle packages")
		}
		for _, bp := range bundlePkgs {
			pkg, err := p.packageService.Get(bp.PackageID)
			if err != nil {
				return nil, errors.Wrap(err, "Retrieve bundle packages")
			}
			bundle.Packages = append(bundle.Packages, *pkg)
		}
		return bundle, nil
	}

	return
}

func (p *bundleService) Delete(bundleId uuid.UUID) (err error) {

	entry := model.BundleEntry{}
	err = p.db.Where(model.BundleEntry{EntryID: bundleId}).Find(&entry).Error
	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Bundle not found")
	} else if err != nil {
		return errors.Wrap(err, "Retrieve bundle entry")
	}

	if entry.EntryType == model.BundleStore {
		err = p.db.Delete(model.StoreBundle{}, "id = ?", bundleId).Error
		if err != nil {
			return errors.Wrap(err, "Retrieve store bundle")
		}
	}

	return nil
}

func (p *bundleService) UpdateStore(bundle *model.StoreBundle) (result *model.StoreBundle, err error) {
	exist := &model.StoreBundle{Model: model.Model{ID: bundle.ID}}
	err = p.db.First(exist).Error
	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, "Bundle not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "Retrieve bundle entry")
	}
	bundle.CreatedAt = exist.CreatedAt
	bundle.UpdatedAt = time.Now()
	bundle.VendorID = exist.VendorID
	bundle.Packages = []model.Package{}
	err = p.db.Save(bundle).Error
	if err != nil {
		return nil, errors.Wrap(err, "Save package")
	}

	bu, err := p.Get(bundle.ID)

	return bu.(*model.StoreBundle), err
}