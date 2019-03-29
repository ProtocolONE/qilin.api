package api

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strconv"
	"time"
)

type (
	PackageRouter struct {
		service *orm.PackageService
	}

	createPackageDTO struct {
		Name        string              `json:"name" validate:"required"`
		Products    []uuid.UUID         `json:"products" validate:"required"`
	}

	packageMediaDTO struct {
		Image       string              `json:"image"`
		Cover       string              `json:"cover"`
		Thumb       string              `json:"thumb"`
	}

	packageDiscountPolicyDTO struct {
		Discount    uint                `json:"discount"`
		BuyOption   model.BuyOption     `json:"buyOption"`
	}

	packageRegionalRestrinctionsDTO struct {
		AllowedCountries    []string    `json:"allowedCountries"`
	}

	productDTO struct {
		ID          uuid.UUID           `json:"id" validate:"required"`
		Name        string              `json:"name"`
		Type        string              `json:"type" validate:"required"`
		Image       string              `json:"image"`
	}

	packageDTO struct {
		ID                      uuid.UUID                            `json:"id"`
		CreatedAt               time.Time                            `json:"createdAt"`
		Sku                     string                               `json:"sku"`
		Name                    string                               `json:"name"`
		IsUpgradeAllowed        bool                                 `json:"isUpgradeAllowed"`
		IsEnabled               bool                                 `json:"isEnabled"`
		Products                []productDTO                         `json:"products" validate:"dive"`
		Media                   packageMediaDTO                      `json:"media" validate:"required,dive"`
		DiscountPolicy          packageDiscountPolicyDTO             `json:"discountPolicy" validate:"required,dive"`
		RegionalRestrinctions   packageRegionalRestrinctionsDTO      `json:"regionalRestrinctions" validate:"required,dive"`
	}
)

func InitPackageRouter(group *echo.Group, service *orm.PackageService) (router *PackageRouter, err error) {
	router = &PackageRouter{service}

	vendorGroup := group.Group("/vendor/:vendorId")

	vendorGroup.GET("/packages", router.GetList)
	vendorGroup.POST("/packages", router.Create)

	vendorGroup.GET("/packages/:id", router.Get)
	vendorGroup.PUT("/packages/:id", router.Update)
	vendorGroup.DELETE("/packages/:id", router.Remove)

	vendorGroup.POST("/packages/:id/products", router.AddProducts)
	vendorGroup.DELETE("/packages/:id/products", router.RemoveProducts)

	return
}

func mapPackageDto(pkg *model.Package, lang string) (dto packageDTO) {
	dto = packageDTO{
		ID: pkg.ID,
		CreatedAt: pkg.CreatedAt,
		Sku: pkg.Sku,
		Name: pkg.Name,
		IsUpgradeAllowed: pkg.IsUpgradeAllowed,
		IsEnabled: pkg.IsEnabled,
		Media: packageMediaDTO{
			Image: pkg.Image,
			Cover: pkg.ImageCover,
			Thumb: pkg.ImageThumb,
		},
		DiscountPolicy: packageDiscountPolicyDTO{
			Discount: pkg.Discount,
			BuyOption: pkg.DiscountBuyOpt,
		},
		RegionalRestrinctions: packageRegionalRestrinctionsDTO{
			AllowedCountries: pkg.AllowedCountries,
		},
	}
	for _, p := range pkg.Products {
		dto.Products = append(dto.Products, productDTO{
			ID: p.GetID(),
			Name: p.GetName(),
			Type: string(p.GetType()),
			Image: p.GetImage(lang),
		})
	}
	return dto
}

func mapPackageModel(dto *packageDTO) (pgk model.Package) {
	pgk = model.Package{
		Model: model.Model{ID: dto.ID},
		Sku: dto.Sku,
		Name: dto.Name,
		IsUpgradeAllowed: dto.IsUpgradeAllowed,
		IsEnabled: dto.IsEnabled,
		Image: dto.Media.Image,
		ImageCover: dto.Media.Cover,
		ImageThumb: dto.Media.Thumb,
		Discount: dto.DiscountPolicy.Discount,
		DiscountBuyOpt: dto.DiscountPolicy.BuyOption,
		AllowedCountries: dto.RegionalRestrinctions.AllowedCountries,
	}
	return
}

func (router *PackageRouter) Create(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid vendor Id")
	}
	params := createPackageDTO{}
	err = ctx.Bind(&params)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Wrong parameters in body")
	}

	lang := context.GetLang(ctx)

	if errs := ctx.Validate(params); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	pkg, err := router.service.Create(vendorId, params.Name, params.Products)
	if err != nil {
		return err
	}
	dto := mapPackageDto(pkg, lang)
	return ctx.JSON(http.StatusCreated, dto)
}

func (router *PackageRouter) AddProducts(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid package Id")
	}
	prods := []uuid.UUID{}
	err = ctx.Bind(&prods)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Wrong products array in body")
	}
	pkg, err := router.service.AddProducts(packageId, prods)
	if err != nil {
		return err
	}
	lang := context.GetLang(ctx)
	dto := mapPackageDto(pkg, lang)
	return ctx.JSON(http.StatusCreated, dto)
}

func (router *PackageRouter) RemoveProducts(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid package Id")
	}
	prods := []uuid.UUID{}
	err = ctx.Bind(&prods)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Wrong products array in body")
	}
	pkg, err := router.service.RemoveProducts(packageId, prods)
	if err != nil {
		return err
	}
	lang := context.GetLang(ctx)
	dto := mapPackageDto(pkg, lang)
	return ctx.JSON(http.StatusOK, dto)
}

func (router *PackageRouter) Get(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid package Id")
	}
	pkg, err := router.service.Get(packageId)
	if err != nil {
		return err
	}
	lang := context.GetLang(ctx)
	dto := mapPackageDto(pkg, lang)
	return ctx.JSON(http.StatusOK, dto)
}

func (router *PackageRouter) GetList(ctx echo.Context) (err error) {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid vendor Id")
	}
	offset, err := strconv.Atoi(ctx.QueryParam("offset"))
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 20
	}
	query := ctx.QueryParam("query")
	sort := ctx.QueryParam("sort")
	packages, err := router.service.GetList(vendorId, query, sort, offset, limit)
	if err != nil {
		return err
	}
	lang := context.GetLang(ctx)
	dto := []packageDTO{}
	for _, pkg := range packages {
		dto = append(dto, mapPackageDto(&pkg, lang))
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (router *PackageRouter) Update(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid package Id")
	}
	pkgDto := packageDTO{}
	err = ctx.Bind(&pkgDto)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Wrong package in body")
	}
	if errs := ctx.Validate(pkgDto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	pkg := mapPackageModel(&pkgDto)
	pkg.ID = packageId
	pkgRes, err := router.service.Update(&pkg)
	if err != nil {
		return err
	}
	lang := context.GetLang(ctx)
	dto := mapPackageDto(pkgRes, lang)
	return ctx.JSON(http.StatusOK, dto)
}

func (router *PackageRouter) Remove(ctx echo.Context) (err error) {
	packageId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid package Id")
	}
	err = router.service.Remove(packageId)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, "Ok")
}