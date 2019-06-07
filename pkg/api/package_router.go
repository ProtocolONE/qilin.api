package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	pkg_utils "qilin-api/pkg/utils"
	"strconv"
	"strings"
	"time"
)

type (
	packageRouter struct {
		service         model.PackageService
		productsService model.ProductService
		priceService    model.PriceService
	}

	createPackageDTO struct {
		Name     string      `json:"name" validate:"required"`
		Products []uuid.UUID `json:"products" validate:"required"`
	}

	packageMediaDTO struct {
		Image utils.LocalizedString `json:"image" validate:"dive"`
		Cover utils.LocalizedString `json:"cover" validate:"dive"`
		Thumb utils.LocalizedString `json:"thumb" validate:"dive"`
	}

	packageDiscountPolicyDTO struct {
		Discount  uint   `json:"discount"`
		BuyOption string `json:"buyOption"`
	}

	packageRegionalRestrinctionsDTO struct {
		AllowedCountries []string `json:"allowedCountries"`
	}

	productDTO struct {
		ID    uuid.UUID              `json:"id" validate:"required"`
		Name  string                 `json:"name"`
		Type  string                 `json:"type" validate:"required"`
		Image *utils.LocalizedString `json:"image" validate:"dive"`
	}

	packageDTO struct {
		ID                    uuid.UUID                       `json:"id"`
		CreatedAt             time.Time                       `json:"createdAt"`
		Sku                   string                          `json:"sku"`
		Name                  utils.LocalizedString           `json:"name" validate:"dive"`
		IsUpgradeAllowed      bool                            `json:"isUpgradeAllowed"`
		IsEnabled             bool                            `json:"isEnabled"`
		IsDefault             bool                            `json:"isDefault"`
		DefaultProductID      uuid.UUID                       `json:"defaultProductId"`
		Products              []productDTO                    `json:"products" validate:"dive"`
		Media                 packageMediaDTO                 `json:"media" validate:"required,dive"`
		DiscountPolicy        packageDiscountPolicyDTO        `json:"discountPolicy" validate:"required,dive"`
		RegionalRestrinctions packageRegionalRestrinctionsDTO `json:"regionalRestrinctions" validate:"required,dive"`
		Commercial            pricesDTO                       `json:"commercial" validate:"-"`
	}

	packageItemDTO struct {
		ID             uuid.UUID                `json:"id"`
		CreatedAt      time.Time                `json:"createdAt"`
		Sku            string                   `json:"sku"`
		Name           utils.LocalizedString    `json:"name"`
		IsEnabled      bool                     `json:"isEnabled"`
		IsDefault      bool                     `json:"isDefault"`
		Media          packageMediaDTO          `json:"media" validate:"required,dive"`
		DiscountPolicy packageDiscountPolicyDTO `json:"discountPolicy" validate:"required,dive"`
	}
)

func InitPackageRouter(
	group *echo.Group,
	service model.PackageService,
	productService model.ProductService,
	priceService model.PriceService) (router *packageRouter, err error) {
	router = &packageRouter{
		service:         service,
		productsService: productService,
		priceService:    priceService,
	}

	vendorRouter := rbac_echo.Group(group, "/vendors/:vendorId", router, []string{"*", model.PackageListType, model.VendorDomain})
	vendorRouter.GET("/packages", router.GetList, nil)
	vendorRouter.POST("/packages", router.Create, nil)

	packageGroup := rbac_echo.Group(group, "/packages", router, []string{"packageId", model.PackageType, model.VendorDomain})
	packageGroup.GET("/:packageId", router.Get, nil)
	packageGroup.PUT("/:packageId", router.Update, nil)
	packageGroup.DELETE("/:packageId", router.Remove, nil)
	packageGroup.POST("/:packageId/products", router.AddProducts, nil)
	packageGroup.DELETE("/:packageId/products", router.RemoveProducts, nil)

	return
}

func (router *packageRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	path := ctx.Path()
	if strings.Contains(path, "/vendors/:vendorId") {
		return GetOwnerForVendor(ctx)
	}
	return GetOwnerForPackage(ctx)
}

func mapPackageItemDto(pkg *model.Package) *packageItemDTO {
	return &packageItemDTO{
		ID:        pkg.ID,
		CreatedAt: pkg.CreatedAt,
		Sku:       pkg.Sku,
		Name:      pkg.Name,
		IsEnabled: pkg.IsEnabled,
		IsDefault: pkg.DefaultProductID != uuid.Nil,
		Media: packageMediaDTO{
			Image: pkg.Image,
			Cover: pkg.ImageCover,
			Thumb: pkg.ImageThumb,
		},
		DiscountPolicy: packageDiscountPolicyDTO{
			Discount:  pkg.Discount,
			BuyOption: pkg.DiscountBuyOpt.String(),
		},
	}
}

func mapPackageDto(pkg *model.Package) (dto *packageDTO, err error) {
	dto = &packageDTO{
		ID:               pkg.ID,
		CreatedAt:        pkg.CreatedAt,
		Sku:              pkg.Sku,
		Name:             pkg.Name,
		IsUpgradeAllowed: pkg.IsUpgradeAllowed,
		IsEnabled:        pkg.IsEnabled,
		IsDefault:        pkg.DefaultProductID != uuid.Nil,
		DefaultProductID: pkg.DefaultProductID,
		Media: packageMediaDTO{
			Image: pkg.Image,
			Cover: pkg.ImageCover,
			Thumb: pkg.ImageThumb,
		},
		DiscountPolicy: packageDiscountPolicyDTO{
			Discount:  pkg.Discount,
			BuyOption: pkg.DiscountBuyOpt.String(),
		},
		RegionalRestrinctions: packageRegionalRestrinctionsDTO{
			AllowedCountries: pkg.AllowedCountries,
		},
	}
	for _, p := range pkg.Products {
		dto.Products = append(dto.Products, productDTO{
			ID:    p.GetID(),
			Name:  p.GetName(),
			Type:  string(p.GetType()),
			Image: p.GetImage(),
		})
	}
	err = mapper.Map(pkg.PackagePrices, &dto.Commercial)
	if err != nil {
		return dto, errors.Wrap(err, "Map prices")
	}
	return dto, nil
}

func mapPackageModel(dto *packageDTO) (pkg *model.Package, err error) {

	err = utils.ValidateUrls(&dto.Media.Image)
	if err != nil {
		return
	}
	err = utils.ValidateUrls(&dto.Media.Cover)
	if err != nil {
		return
	}
	err = utils.ValidateUrls(&dto.Media.Thumb)
	if err != nil {
		return
	}

	if !pkg_utils.ValidateCountryList(dto.RegionalRestrinctions.AllowedCountries) {
		return nil, orm.NewServiceError(http.StatusUnprocessableEntity, "Invalid countries")
	}

	return &model.Package{
		Model:            model.Model{ID: dto.ID},
		Sku:              dto.Sku,
		Name:             dto.Name,
		IsUpgradeAllowed: dto.IsUpgradeAllowed,
		IsEnabled:        dto.IsEnabled,
		// IsDefault:        dto.IsDefault, -- read only
		Image:            dto.Media.Image,
		ImageCover:       dto.Media.Cover,
		ImageThumb:       dto.Media.Thumb,
		Discount:         dto.DiscountPolicy.Discount,
		DiscountBuyOpt:   model.NewBuyOption(dto.DiscountPolicy.BuyOption),
		AllowedCountries: dto.RegionalRestrinctions.AllowedCountries,
	}, nil
}

func (router *packageRouter) checkRBAC(userId string, qilinCtx *rbac_echo.AppContext, productIds []uuid.UUID) error {
	// Check permissions for Games
	games, _, err := router.productsService.SpecializationIds(productIds)
	if err != nil {
		return err
	}
	for _, gameId := range games {
		owner, err := qilinCtx.GetOwnerForGame(gameId)
		if err != nil {
			return err
		}
		if qilinCtx.CheckPermissions(userId, model.VendorDomain, model.GameType, gameId.String(), owner, "read") != nil {
			return orm.NewServiceError(http.StatusForbidden, fmt.Sprintf("Access restricted for game `%s`", gameId.String()))
		}
	}
	// TODO: do same for DLC
	return nil
}

func (router *packageRouter) Create(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid vendor Id")
	}
	params := createPackageDTO{}
	err = ctx.Bind(&params)
	if err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, "Wrong parameters in body")
	}

	if errs := ctx.Validate(params); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	userId, err := context.GetAuthUserId(ctx)
	if err != nil {
		return err
	}
	qilinCtx := ctx.(rbac_echo.AppContext)

	err = router.checkRBAC(userId, &qilinCtx, params.Products)
	if err != nil {
		return err
	}

	pkg, err := router.service.Create(vendorId, userId, params.Name, params.Products)
	if err != nil {
		return err
	}
	dto, err := mapPackageDto(pkg)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, dto)
}

func (router *packageRouter) AddProducts(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("packageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid package Id")
	}
	products := []uuid.UUID{}
	err = ctx.Bind(&products)
	if err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, "Wrong products array in body")
	}

	userId, err := context.GetAuthUserId(ctx)
	if err != nil {
		return err
	}
	qilinCtx := ctx.(rbac_echo.AppContext)

	err = router.checkRBAC(userId, &qilinCtx, products)
	if err != nil {
		return err
	}

	pkg, err := router.service.AddProducts(packageId, products)
	if err != nil {
		return err
	}
	dto, err := mapPackageDto(pkg)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (router *packageRouter) RemoveProducts(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("packageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid package Id")
	}
	prods := []uuid.UUID{}
	err = ctx.Bind(&prods)
	if err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, "Wrong products array in body")
	}
	pkg, err := router.service.RemoveProducts(packageId, prods)
	if err != nil {
		return err
	}
	dto, err := mapPackageDto(pkg)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (router *packageRouter) Get(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("packageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid package Id")
	}
	pkg, err := router.service.Get(packageId)
	if err != nil {
		return err
	}
	dto, err := mapPackageDto(pkg)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (router *packageRouter) GetList(ctx echo.Context) (err error) {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid vendor Id")
	}
	offset, err := strconv.Atoi(ctx.QueryParam("offset"))
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 20
	}
	userId, err := context.GetAuthUserId(ctx)
	if err != nil {
		return err
	}
	qilinCtx := ctx.(rbac_echo.AppContext)
	filterFunc := func(packageId uuid.UUID) (grant bool, err error) {
		owner, err := qilinCtx.GetOwnerForPackage(packageId)
		if err != nil {
			return
		}
		if qilinCtx.CheckPermissions(userId, model.VendorDomain, model.PackageType, packageId.String(), owner, "read") != nil {
			return false, nil
		}
		return true, nil
	}
	query := ctx.QueryParam("query")
	sort := ctx.QueryParam("sort")
	total, packages, err := router.service.GetList(vendorId, query, sort, offset, limit, filterFunc)
	if err != nil {
		return err
	}
	dto := []*packageItemDTO{}
	for _, pkg := range packages {
		dto = append(dto, mapPackageItemDto(&pkg))
	}
	ctx.Response().Header().Add("X-Items-Count", fmt.Sprintf("%d", total))
	return ctx.JSON(http.StatusOK, dto)
}

func (router *packageRouter) Update(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("packageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid package Id")
	}

	existPkg, err := router.service.Get(packageId)
	if err != nil {
		return orm.NewServiceError(http.StatusNotFound, "Package not found")
	}

	pkgDto := &packageDTO{}
	err = ctx.Bind(pkgDto)
	if err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errors.Wrap(err, "Wrong package in body").Error())
	}
	if errs := ctx.Validate(pkgDto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	// Update package prices
	basePrice := model.BasePrice{}
	err = mapper.Map(pkgDto.Commercial, &basePrice.PackagePrices)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}
	if err := router.priceService.UpdateBase(packageId, &basePrice); err != nil {
		return err
	}
	for _, price := range basePrice.Prices {
		if err := router.priceService.Update(packageId, &price); err != nil {
			return orm.NewServiceError(http.StatusBadRequest, err)
		}
	}
	for _, existPrice := range existPkg.Prices {
		found := false
		for _, price := range basePrice.Prices {
			if price.Currency == existPrice.Currency {
				found = true
				break
			}
		}
		if !found {
			if err := router.priceService.Delete(packageId, &existPrice); err != nil {
				return orm.NewServiceError(http.StatusBadRequest, err)
			}
		}
	}

	// Update package
	pkg, err := mapPackageModel(pkgDto)
	if err != nil {
		return err
	}
	pkg.ID = packageId
	pkgRes, err := router.service.Update(pkg)
	if err != nil {
		return err
	}
	dto, err := mapPackageDto(pkgRes)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (router *packageRouter) Remove(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("packageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid package Id")
	}
	err = router.service.Remove(packageId)
	if err != nil {
		return err
	}
	return ctx.NoContent(http.StatusOK)
}
