package api

import (
	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"time"
)

type (
	BundleRouter struct {
		service *orm.BundleService
	}

	createBundleDTO struct {
		Name        string              `json:"name" validate:"required"`
		Packages    []uuid.UUID         `json:"packages" validate:"required"`
	}

	bundleDiscountPolicyDTO struct {
		Discount    uint                `json:"discount"`
		BuyOption   model.BuyOption     `json:"buyOption"`
	}

	bundleRegionalRestrinctionsDTO struct {
		AllowedCountries    []string    `json:"allowedCountries"`
	}

	storeBundleDTO struct {
		ID                      uuid.UUID                       `json:"id"`
		CreatedAt               time.Time                       `json:"createdAt"`
		Sku                     string                          `json:"sku" validate:"required"`
		Name                    string                          `json:"name" validate:"required"`
		IsUpgradeAllowed        bool                            `json:"isUpgradeAllowed"`
		IsEnabled               bool                            `json:"isEnabled"`
		DiscountPolicy          bundleDiscountPolicyDTO         `json:"discountPolicy" validate:"required,dive"`
		RegionalRestrinctions   bundleRegionalRestrinctionsDTO  `json:"regionalRestrinctions" validate:"required,dive"`
		Packages                []packageDTO                    `json:"packages" validate:"required,dive"`
	}
)

func mapStoreBundleDto(bundle *model.StoreBundle, lang string) (dto storeBundleDTO, err error) {
	dto = storeBundleDTO{
		ID: bundle.ID,
		CreatedAt: bundle.CreatedAt,
		Sku: bundle.Sku,
		Name: bundle.Name,
		IsUpgradeAllowed: bundle.IsUpgradeAllowed,
		IsEnabled: bundle.IsEnabled,
		DiscountPolicy: bundleDiscountPolicyDTO{
			Discount: bundle.Discount,
			BuyOption: bundle.DiscountBuyOpt,
		},
		RegionalRestrinctions: bundleRegionalRestrinctionsDTO{
			AllowedCountries: bundle.AllowedCountries,
		},
	}
	for _, p := range bundle.Packages {
		pkg, err := mapPackageDto(&p, lang)
		if err != nil {
			return dto, err
		}
		dto.Packages = append(dto.Packages, pkg)
	}
	return dto, nil
}

func InitBundleRouter(group *echo.Group, service *orm.BundleService) (router *BundleRouter, err error) {
	router = &BundleRouter{service}

	packageGroup := rbac_echo.Group(group,"/vendors/:vendorId", router, []string{"vendorId", "bundleId", model.VendorType})

	packageGroup.GET("/bundles/store/:bundleId", router.GetStore, nil)
	packageGroup.POST("/bundles/store", router.CreateStore, nil)
	packageGroup.DELETE("/bundles/:bundleId", router.Delete, nil)

	return
}

func (router *BundleRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	return GetOwnerForBundle(ctx)
}

func (router *BundleRouter) CreateStore(ctx echo.Context) (err error) {
	params := createBundleDTO{}
	err = ctx.Bind(&params)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Wrong parameters in body")
	}

	if errs := ctx.Validate(params); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	bundle, err := router.service.CreateStore(params.Name, params.Packages)
	if err != nil {
		return err
	}
	lang := context.GetLang(ctx)

	dto, err := mapStoreBundleDto(bundle, lang)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, dto)
}

func (router *BundleRouter) GetStore(ctx echo.Context) (err error) {
	bundleId, err := uuid.FromString(ctx.Param("bundleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid vendor Id")
	}

	bundle, err := router.service.Get(bundleId)
	if err != nil {
		return err
	}

	bundleStore, ok := bundle.(*model.StoreBundle)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Bundle not for store")
	}
	lang := context.GetLang(ctx)

	dto, err := mapStoreBundleDto(bundleStore, lang)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (router *BundleRouter) Delete(ctx echo.Context) (err error) {
	bundleId, err := uuid.FromString(ctx.Param("bundleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid vendor Id")
	}

	err = router.service.Delete(bundleId)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, "Ok")
}
