package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"strings"
	"time"
)

type packageService struct {
	db          *gorm.DB
	gameService model.GameService
}

func NewPackageService(db *Database, gameService model.GameService) (*packageService, error) {
	return &packageService{
		db:          db.database,
		gameService: gameService,
	}, nil
}

// Transaction must be manage outside this function for commit and rollback (in case of error)
func createPackage(
	transaction *gorm.DB,
	packageId,
	vendorId uuid.UUID,
	defaultProductID uuid.UUID,
	userId,
	name string,
	products []uuid.UUID) (err error) {

	if len(strings.Trim(name, " \r\n\t")) == 0 {
		return NewServiceError(http.StatusUnprocessableEntity, "Name is empty")
	}

	newPack := model.Package{
		Model:            model.Model{ID: packageId},
		Sku:              uuid.NewV4().String(),
		Name:             utils.LocalizedString{EN: name},
		VendorID:         vendorId,
		CreatorID:        userId,
		DefaultProductID: defaultProductID,
		PackagePrices: model.PackagePrices{
			Common: model.JSONB{
				"currency":        "USD",
				"notifyRateJumps": false,
			},
			PreOrder: model.JSONB{
				"date":    time.Now().String(),
				"enabled": false,
			},
			Prices: []model.Price{},
		},
	}
	err = transaction.Create(&newPack).Error
	if err != nil {
		return errors.Wrap(err, "While create new package")
	}
	for index, productId := range products {
		err = transaction.Create(model.PackageProduct{
			PackageID: packageId,
			ProductID: productId,
			Position:  index + 1,
		}).Error
		if err != nil {
			return errors.Wrap(err, "While append products into package")
		}
	}

	return
}

func (p *packageService) filterProducts(prods []uuid.UUID) (result []uuid.UUID, err error) {
	entries, err := p.getProductEntries(prods)
	if err != nil {
		return nil, err
	}
	return entries.GetUUIDs(), err
}

func (p *packageService) getProductEntries(prods []uuid.UUID) (result model.ProductEntryArray, err error) {
	result = model.ProductEntryArray{}
	if len(prods) > 0 {
		entries := []model.ProductEntry{}
		err = p.db.Where("entry_id in (?)", prods).Find(&entries).Error
		if err != nil {
			return nil, errors.Wrap(err, "Retrieve product entries")
		}
		// Save order
		for _, prodId := range prods {
			for _, entry := range entries {
				if prodId == entry.EntryID {
					result = append(result, entry)
					break
				}
			}
		}
	}
	if len(result) == 0 {
		return result, NewServiceError(http.StatusUnprocessableEntity, "No any products")
	}
	return
}

func (p *packageService) findPackageOrError(packageId uuid.UUID) (result *model.Package, err error) {
	result = &model.Package{}
	err = p.db.Where("id = ?", packageId).First(result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, NewServiceError(http.StatusNotFound, "Package not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "Retrieve package")
	}
	return
}

func (p *packageService) Create(vendorId uuid.UUID, userId, name string, prods []uuid.UUID) (result *model.Package, err error) {
	pkgId := uuid.NewV4()

	// 1. Filter products for exists
	prods, err = p.filterProducts(prods)
	if err != nil {
		return nil, err
	}

	// 2. Insert package into DB
	transaction := p.db.Begin()
	defer func() {
		if err := recover(); err != nil {
			transaction.Rollback()
		}
	}()
	err = createPackage(transaction, pkgId, vendorId, uuid.Nil, userId, name, prods)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	err = transaction.Commit().Error
	if err != nil {
		return nil, errors.Wrap(err, "Commit while create package")
	}
	return p.Get(pkgId)
}

func (p *packageService) AddProducts(packageId uuid.UUID, prods []uuid.UUID) (result *model.Package, err error) {

	_, err = p.findPackageOrError(packageId)
	if err != nil {
		return nil, err
	}

	prods, err = p.filterProducts(prods)
	if err != nil {
		return nil, err
	}

	exists := []model.PackageProduct{}
	err = p.db.Where("package_id = ?", packageId).Find(&exists).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch package contents")
	}

	position := len(exists) + 1
	transaction := p.db.Begin()
	defer func() {
		if err := recover(); err != nil {
			transaction.Rollback()
		}
	}()
	for _, prodId := range prods {
		found := false
		for _, pu := range exists {
			if prodId == pu.ProductID {
				found = true
				break
			}
		}
		if !found {
			err = transaction.Create(&model.PackageProduct{
				PackageID: packageId,
				ProductID: prodId,
				Position:  position,
			}).Error
			position += 1
			if err != nil {
				transaction.Rollback()
				return nil, errors.Wrap(err, "Make package product link")
			}
		}
	}
	err = transaction.Commit().Error
	if err != nil {
		return nil, errors.Wrap(err, "Commit append products")
	}

	return p.Get(packageId)
}

func (p *packageService) RemoveProducts(packageId uuid.UUID, prods []uuid.UUID) (result *model.Package, err error) {

	// 1. Find the package
	pkg, err := p.findPackageOrError(packageId)
	if err != nil {
		return nil, err
	}

	// 2. Filter for reject unexist's and default products
	exists := []model.PackageProduct{}
	err = p.db.Where("package_id = ?", packageId).Find(&exists).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch package contents")
	}
	for i, prodID := range prods {
		found := false
		for _, exist := range exists {
			if prodID == exist.ProductID {
				found = true
				break
			}
		}
		if !found || prodID == pkg.DefaultProductID {
			prods = append(prods[:i], prods[i+1:]...)
		}
	}

	// 3. Actual remove products from package (if any)
	if len(prods) > 0 {
		err = p.db.Delete(model.PackageProduct{}, "package_id = ? and product_id in (?)", packageId, prods).Error
		if err != nil {
			return nil, errors.Wrap(err, "Delete package products")
		}
	} else {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "No any products for remove")
	}

	return p.Get(packageId)
}

func (p *packageService) Get(packageId uuid.UUID) (result *model.Package, err error) {

	result, err = p.findPackageOrError(packageId)
	if err != nil {
		return nil, err
	}

	type PackageProductJoin struct {
		model.PackageProduct
		EntryType model.ProductType
	}
	pkgProds := []PackageProductJoin{}
	err = p.db.
		Table("package_products").
		Select("package_products.*, products.entry_type").
		Where("package_id = ?", packageId).
		Joins("left join products on package_products.product_id = products.entry_id").
		Order("position asc").
		Find(&pkgProds).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch package contents")
	}
	prods := []model.Product{}
	if len(pkgProds) > 0 {
		for _, prod := range pkgProds {
			if prod.EntryType == model.ProductGame {
				game, err := p.gameService.GetProduct(prod.ProductID)
				if err != nil {
					return nil, errors.Wrap(err, "Fetch game for package")
				}
				prods = append(prods, game)
			} else if prod.EntryType == model.ProductDLC {
				return nil, NewServiceError(http.StatusNotImplemented, "Retrieve DLC is not implemented yet")
			}
		}
	}
	result.Products = prods

	err = p.db.
		Model(model.BasePrice{ID: packageId}).
		Related(&result.Prices).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch prices for package")
	}

	return
}

func (p *packageService) GetList(userId string, vendorId uuid.UUID, query, sort string, offset, limit int, filterFunc model.PackageListingFilter) (total int, result []model.Package, err error) {

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
		case "-discount":
			orderBy = "discount DESC"
		case "+discount":
			orderBy = "discount ASC"
		}
	}

	conds := []string{}
	vals := []interface{}{}

	if query != "" {
		user := model.User{}
		err = p.db.Select("lang").Where("id = ?", userId).First(&user).Error
		if err != nil {
			return 0, nil, errors.Wrap(err, "while fetch user")
		}
		conds = append(conds, "(name ->> ? ilike ? or name ->> 'en' ilike ?)")
		vals = append(vals, "%"+query+"%", user.Lang, "%"+query+"%")
		// TODO: Add another kinds for searching
	}

	if filterFunc != nil {
		vendorPackages := []model.Package{}
		err = p.db.
			Select("id").
			Where(`vendor_id = ?`, vendorId).
			Where(strings.Join(conds, " or "), vals...).
			Find(&vendorPackages).Error
		if err != nil {
			return 0, nil, errors.Wrap(err, "Fetch package ids")
		}
		ids := []uuid.UUID{}
		for _, pkg := range vendorPackages {
			if grant, err := filterFunc(pkg.ID); grant {
				ids = append(ids, pkg.ID)
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
				Model(model.Package{}).
				Where(`id in (?)`, ids).
				Order(orderBy).
				Find(&result).Error
			if err != nil {
				return 0, nil, errors.Wrap(err, "Fetch package list from ids")
			}
		}
	} else {
		err = p.db.
			Model(model.Package{}).
			Where(`vendor_id = ?`, vendorId).
			Where(strings.Join(conds, " or "), vals...).
			Order(orderBy).
			Limit(limit).
			Offset(offset).
			Find(&result).Error
		if err != nil {
			return 0, nil, errors.Wrap(err, "Fetch package list")
		}
		err = p.db.
			Model(model.Package{}).
			Where(`vendor_id = ?`, vendorId).
			Where(strings.Join(conds, " or "), vals...).
			Count(&total).Error
		if err != nil {
			return 0, nil, errors.Wrap(err, "Fetch package total")
		}
	}

	return
}

func (p *packageService) Update(pkg *model.Package) (*model.Package, error) {

	exist, err := p.findPackageOrError(pkg.ID)
	if err != nil {
		return nil, err
	}

	pkg.CreatedAt = exist.CreatedAt
	pkg.UpdatedAt = time.Now()
	pkg.VendorID = exist.VendorID
	pkg.PackagePrices = exist.PackagePrices
	pkg.DefaultProductID = exist.DefaultProductID
	// Products also ignored

	err = p.db.Save(pkg).Error
	if err != nil {
		return nil, errors.Wrap(err, "Save package")
	}
	return p.Get(pkg.ID)
}

func (p *packageService) Remove(packageId uuid.UUID) (err error) {

	exist, err := p.findPackageOrError(packageId)
	if err != nil {
		return err
	}

	if exist.DefaultProductID != uuid.Nil {
		return NewServiceError(http.StatusForbidden, "Package is default")
	}

	err = p.db.Delete(exist).Error
	if err != nil {
		return errors.Wrap(err, "Delete package")
	}
	return
}
