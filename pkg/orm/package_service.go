package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/random"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"strings"
	"time"
)

type PackageService struct {
	db *gorm.DB
	gameService model.GameService
}

func NewPackageService(db *Database) (*PackageService, error) {
	gameService, _ := NewGameService(db)
	return &PackageService{db.database, gameService}, nil
}

func (p *PackageService) Create(vendorId uuid.UUID, name string, prods []uuid.UUID) (result *model.Package, err error) {
	entries := []model.ProductEntry{}
	if len(prods) > 0 {
		err = p.db.Where("entry_id in (?)", prods).Find(&entries).Error
		if err != nil {
			return nil, errors.Wrap(err, "Search products")
		}
	}
	if len(entries) == 0 {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "No any products")
	}

	newPack := model.Package{
		Model: model.Model{ID: uuid.NewV4()},
		Sku: random.String(8, "123456789"),
		Name: name,
		VendorID: vendorId,
	}
	err = p.db.Create(&newPack).Error
	if err != nil {
		return nil, errors.Wrap(err, "While create new package")
	}

	db := p.db.Begin()
	for index, prodId := range prods {
		err = db.Create(model.PackageProduct{
			PackageID: newPack.ID,
			ProductID: prodId,
			Position: index + 1,
		}).Error
		if err != nil {
			return nil, errors.Wrap(err, "While append products into package")
		}
	}
	db.Commit()

	return p.Get(newPack.ID)
}

func (p *PackageService) AddProducts(packageId uuid.UUID, prods []uuid.UUID) (result *model.Package, err error) {
	entries := []model.ProductEntry{}
	if len(prods) > 0 {
		err = p.db.Where("entry_id in (?)", prods).Find(&entries).Error
		if err != nil {
			return nil, errors.Wrap(err, "Search products")
		}
	}
	if len(entries) == 0 {
		return nil, NewServiceError(http.StatusUnprocessableEntity, "No any products")
	}

	exists := []model.PackageProduct{}
	err = p.db.Where("package_id = ?", packageId).Find(&exists).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch package contents")
	}

	position := len(exists) + 1
	batchDb := p.db.Begin()
	for _, p := range entries {
		found := false
		for _, pu := range exists {
			if p.EntryID == pu.ProductID {
				found = true
				break
			}
		}
		if !found {
			err = batchDb.Create(&model.PackageProduct{
				PackageID: packageId,
				ProductID: p.EntryID,
				Position: position,
			}).Error
			position += 1
			if err != nil {
				batchDb.Rollback()
				return nil, errors.Wrap(err, "Make package product link")
			}
		}
	}
	batchDb.Commit()

	return p.Get(packageId)
}

func (p *PackageService) RemoveProducts(packageId uuid.UUID, prods []uuid.UUID) (result *model.Package, err error) {

	if len(prods) > 0 {
		err = p.db.Delete(model.PackageProduct{}, "package_id = ? and product_id in (?)", packageId, prods).Error
		if err != nil {
			return nil, errors.Wrap(err, "Delete package products")
		}
	}

	return p.Get(packageId)
}

func (p *PackageService) Get(packageId uuid.UUID) (result *model.Package, err error)  {
	result = &model.Package{}
	err = p.db.Where("id = ?", packageId.String()).First(result).Error
	if err != nil {
		return nil, errors.Wrap(err, "Get package by id")
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
			} else
			if prod.EntryType == model.ProductDLC {
				//...
			}
		}
	}
	result.Products = prods

	err = p.db.
		Model(model.Package{}).
		Where("id = ?", packageId).
		Association("Prices").
		Find(&result.Prices).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch prices for package")
	}

	return
}

func (p *PackageService) GetList(vendorId uuid.UUID, query, sort string, offset, limit int) (result []model.Package, err error) {

	orderBy := ""
	orderBy = "created_at ASC"
	if sort != "" {
		switch sort {
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
		conds = append(conds, `name ilike ?`)
		vals = append(vals, "%"+query+"%")
		// TODO: Add another kinds for searching
	}

	err = p.db.
		Model(model.Package{}).
		Where(`vendor_id = ?`, vendorId).
		Where(strings.Join(conds, " or "), vals...).
		Order(orderBy).
		Limit(limit).
		Offset(offset).
		Find(&result).Error
	if err != nil {
		return nil, errors.Wrap(err, "Fetch package list")
	}

	return
}

func (p *PackageService) Update(pkg *model.Package) (*model.Package, error) {
	exist := &model.Package{Model: model.Model{ID: pkg.ID}}
	err := p.db.First(exist).Error
	if err != nil {
		return nil, errors.Wrap(err, "Get package by id")
	}
	pkg.CreatedAt = exist.CreatedAt
	pkg.UpdatedAt = time.Now()
	pkg.VendorID = exist.VendorID
	pkg.Products = []model.Product{}
	err = p.db.Save(pkg).Error
	if err != nil {
		return nil, errors.Wrap(err, "Save package")
	}
	return pkg, err
}

func (p *PackageService) Remove(packageId uuid.UUID) (err error) {
	exist := &model.Package{Model: model.Model{ID: packageId}}
	err = p.db.First(exist).Error
	if err != nil {
		return errors.Wrap(err, "Get package by id")
	}
	err = p.db.Delete(exist).Error
	if err != nil {
		return errors.Wrap(err, "Delete package")
	}
	return
}