package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	mutils "qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm/utils"
	"strings"
	"time"
)

type bundleService struct {
	db             *gorm.DB
	gameService    model.GameService
	packageService model.PackageService
}

func NewBundleService(db *Database, packageService model.PackageService, gameService model.GameService) (model.BundleService, error) {
	return &bundleService{db.database, gameService, packageService}, nil
}

func (p *bundleService) CreateStore(vendorId uuid.UUID, userId, name string, packageIds []uuid.UUID) (bundle model.Bundle, err error) {

	if len(strings.Trim(name, " \r\n\t")) == 0 {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "Name is empty")
	}

	packages := []model.Package{}
	if len(packageIds) > 0 {
		err = p.db.Where("id in (?)", packageIds).Find(&packages).Error
		if err != nil {
			return nil, errors.Wrap(err, "Search packages")
		}
	}
	if len(packages) == 0 {
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
		Model:     model.Model{ID: uuid.NewV4()},
		Sku:       uuid.NewV4().String(),
		Name:      mutils.LocalizedString{EN: name},
		VendorID:  vendorId,
		IsEnabled: false,
		CreatorID: userId,
	}
	newBundle.Bundle.EntryID = newBundle.ID
	err = p.db.Create(&newBundle).Error
	if err != nil {
		return nil, errors.Wrap(err, "While create new bundle")
	}

	db := p.db.Begin()
	// We walks `packageIds` first cuz want to persistent ordering
	for index, pkgID := range packageIds {
		for _, pkg := range packages {
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

func (p *bundleService) GetStoreList(vendorId uuid.UUID, query, sort string, offset, limit int, filterFunc model.BundleListingFilter) (total int, result []model.Bundle, err error) {
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

	storeBundles := []model.StoreBundle{}
	if filterFunc != nil {
		vendorBundles := []model.StoreBundle{}
		err = p.db.
			Select("id").
			Where(`vendor_id = ?`, vendorId).
			Where(strings.Join(conds, " or "), vals...).
			Find(&vendorBundles).Error
		if err != nil {
			return 0, nil, errors.Wrap(err, "Fetch store bundle ids")
		}
		ids := []uuid.UUID{}
		for _, bundle := range vendorBundles {
			if grant, err := filterFunc(bundle.ID); grant {
				ids = append(ids, bundle.ID)
			} else if err != nil {
				return 0, nil, err
			}
		}
		total = len(ids)
		if offset < 0 {
			offset = 0
		}
		if offset > len(ids) {
			offset = len(ids)
		}
		if limit < 0 {
			limit = 0
		}
		if offset+limit > len(ids) {
			limit = len(ids) - offset
		}
		ids = ids[offset : offset+limit]
		if len(ids) > 0 {
			err = p.db.
				Model(model.StoreBundle{}).
				Where(`id in (?)`, ids).
				Order(orderBy).
				Find(&storeBundles).Error
			if err != nil {
				return 0, nil, errors.Wrap(err, "Fetch store bundles from ids")
			}
		}
	} else {
		// Get store bundles
		err = p.db.
			Model(model.StoreBundle{}).
			Where(`vendor_id = ?`, vendorId).
			Where(strings.Join(conds, " or "), vals...).
			Order(orderBy).
			Limit(limit).
			Offset(offset).
			Find(&storeBundles).Error
		if err != nil {
			return 0, nil, errors.Wrap(err, "Fetch store bundle list")
		}
		// Calc total bundles for vendor
		err = p.db.
			Model(model.StoreBundle{}).
			Where(`vendor_id = ?`, vendorId).
			Where(strings.Join(conds, " or "), vals...).
			Count(&total).Error
		if err != nil {
			return 0, nil, errors.Wrap(err, "Fetch store bundle total")
		}
	}

	result = []model.Bundle{}
	for _, bundle := range storeBundles {
		copyBundle := bundle
		result = append(result, &copyBundle)
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
		bundle := model.StoreBundle{}
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
		return &bundle, nil
	}

	return nil, NewServiceError(http.StatusNotImplemented, "Only store bundle, yet")
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
	} else {
		return NewServiceError(http.StatusNotImplemented, "Only store bundle, yet")
	}

	return nil
}

func (p *bundleService) UpdateStore(storeBundle model.Bundle) (result model.Bundle, err error) {
	bundle, ok := storeBundle.(*model.StoreBundle)
	if !ok {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "Bundle isn't store bundle")
	}
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

	return bu, err
}

func (p *bundleService) checkForExists(bundleId uuid.UUID) (err error) {
	entry := model.BundleEntry{}
	err = p.db.Where(model.BundleEntry{EntryID: bundleId}).Find(&entry).Error
	if err == gorm.ErrRecordNotFound {
		return NewServiceError(http.StatusNotFound, "Bundle not found")
	} else if err != nil {
		return errors.Wrap(err, "Retrieve bundle entry")
	}
	if entry.EntryType == model.BundleStore {
		exists, err := utils.CheckExists(p.db, model.StoreBundle{}, bundleId)
		if err != nil {
			return errors.Wrap(err, "Retrieve store bundle")
		}
		if !exists {
			return NewServiceError(http.StatusNotFound, "Bundle not found")
		}
	} else {
		return NewServiceError(http.StatusNotImplemented, "Only store bundle")
	}
	return
}

func (p *bundleService) AddPackages(bundleId uuid.UUID, packageIds []uuid.UUID) (err error) {

	// 1. Check bundle for exists
	err = p.checkForExists(bundleId)
	if err != nil {
		return err
	}

	// 2. Check packages for exists
	packages := []model.Package{}
	if len(packageIds) > 0 {
		err = p.db.Where("id in (?)", packageIds).Find(&packages).Error
		if err != nil {
			return errors.Wrap(err, "Search packages")
		}
	}

	// 3. Filter already bound packages
	existsIds := []model.BundlePackage{}
	err = p.db.Where("bundle_id = ?", bundleId).Find(&existsIds).Error
	if err != nil {
		return errors.Wrap(err, "Retrieve bundle packages")
	}
	for _, exist := range existsIds {
		for i, pkg := range packages {
			if exist.PackageID == pkg.ID {
				packages = append(packages[:i], packages[i+1:]...)
				break
			}
		}
	}
	if len(packages) == 0 {
		return NewServiceError(http.StatusUnprocessableEntity, "No any package")
	}

	// 4. Append packages with defined order
	db := p.db.Begin()
	// We walks `packages` first cuz want to persistent ordering
	for index, pkgID := range packageIds {
		for _, pkg := range packages {
			if pkgID != pkg.ID {
				continue
			}
			err = db.Create(&model.BundlePackage{
				PackageID: pkg.ID,
				BundleID:  bundleId,
				Position:  len(existsIds) + index + 1,
			}).Error
			if err != nil {
				db.Rollback()
				return errors.Wrap(err, "While append packages into bundle")
			}
			break
		}
	}
	err = db.Commit().Error
	if err != nil {
		return errors.Wrap(err, "While commit packages")
	}

	return
}

func (p *bundleService) RemovePackages(bundleId uuid.UUID, packages []uuid.UUID) (err error) {

	err = p.checkForExists(bundleId)
	if err != nil {
		return err
	}

	if len(packages) > 0 {
		err = p.db.Delete(model.BundlePackage{}, "bundle_id = ? and package_id in (?)", bundleId, packages).Error
		if err != nil {
			return errors.Wrap(err, "While delete packages from bundle")
		}
	}

	return
}
