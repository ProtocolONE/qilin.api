package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	pkg_utils "qilin-api/pkg/utils"
	"strconv"
	"strings"
	"time"
)

type (
	BundleRouter struct {
		service model.BundleService
	}

	createBundleDTO struct {
		Name     string      `json:"name" validate:"required"`
		Packages []uuid.UUID `json:"packages" validate:"required"`
	}

	bundleDiscountPolicyDTO struct {
		Discount  uint   `json:"discount"`
		BuyOption string `json:"buyOption"`
	}

	bundleRegionalRestrinctionsDTO struct {
		AllowedCountries []string `json:"allowedCountries"`
	}

	storeBundleDTO struct {
		ID                    uuid.UUID                      `json:"id"`
		CreatedAt             time.Time                      `json:"createdAt"`
		Sku                   string                         `json:"sku" validate:"required"`
		Name                  utils.LocalizedString          `json:"name" validate:"dive,required"`
		IsUpgradeAllowed      bool                           `json:"isUpgradeAllowed"`
		IsEnabled             bool                           `json:"isEnabled"`
		DiscountPolicy        bundleDiscountPolicyDTO        `json:"discountPolicy" validate:"required,dive"`
		RegionalRestrinctions bundleRegionalRestrinctionsDTO `json:"regionalRestrinctions" validate:"required,dive"`
		Packages              []*packageDTO                  `json:"packages" validate:"-"`
	}

	storeBundleItemDTO struct {
		ID               uuid.UUID             `json:"id"`
		CreatedAt        time.Time             `json:"createdAt"`
		Sku              string                `json:"sku" validate:"required"`
		Name             utils.LocalizedString `json:"name" validate:"required"`
		IsUpgradeAllowed bool                  `json:"isUpgradeAllowed"`
		IsEnabled        bool                  `json:"isEnabled"`
	}
)

func mapStoreBundleDto(bundle *model.StoreBundle) (dto *storeBundleDTO, err error) {
	dto = &storeBundleDTO{
		ID:               bundle.ID,
		CreatedAt:        bundle.CreatedAt,
		Sku:              bundle.Sku,
		Name:             bundle.Name,
		IsUpgradeAllowed: bundle.IsUpgradeAllowed,
		IsEnabled:        bundle.IsEnabled,
		DiscountPolicy: bundleDiscountPolicyDTO{
			Discount:  bundle.Discount,
			BuyOption: bundle.DiscountBuyOpt.String(),
		},
		RegionalRestrinctions: bundleRegionalRestrinctionsDTO{
			AllowedCountries: bundle.AllowedCountries,
		},
	}
	for _, p := range bundle.Packages {
		pkg, err := mapPackageDto(&p)
		if err != nil {
			return nil, err
		}
		dto.Packages = append(dto.Packages, pkg)
	}
	return dto, nil
}

func mapStoreBundleItemDto(bundle *model.StoreBundle) *storeBundleItemDTO {
	return &storeBundleItemDTO{
		ID:               bundle.ID,
		CreatedAt:        bundle.CreatedAt,
		Sku:              bundle.Sku,
		Name:             bundle.Name,
		IsUpgradeAllowed: bundle.IsUpgradeAllowed,
		IsEnabled:        bundle.IsEnabled,
	}
}

func mapStoreBundleModel(dto *storeBundleDTO) (bundle *model.StoreBundle, err error) {

	if !pkg_utils.ValidateCountryList(dto.RegionalRestrinctions.AllowedCountries) {
		return nil, orm.NewServiceError(http.StatusUnprocessableEntity, "Invalid countries")
	}

	return &model.StoreBundle{
		Model:            model.Model{ID: dto.ID},
		Sku:              dto.Sku,
		Name:             dto.Name,
		IsUpgradeAllowed: dto.IsUpgradeAllowed,
		IsEnabled:        dto.IsEnabled,
		Discount:         dto.DiscountPolicy.Discount,
		DiscountBuyOpt:   model.NewBuyOption(dto.DiscountPolicy.BuyOption),
		AllowedCountries: dto.RegionalRestrinctions.AllowedCountries,
	}, nil
}

func InitBundleRouter(group *echo.Group, service model.BundleService) (router *BundleRouter, err error) {
	router = &BundleRouter{service}

	vendorRouter := rbac_echo.Group(group, "/vendors/:vendorId", router, []string{"*", model.RoleBundleList, model.VendorDomain})
	vendorRouter.POST("/bundles/store", router.CreateStore, nil)
	vendorRouter.GET("/bundles/store", router.GetStoreList, nil)

	bundleGroup := rbac_echo.Group(group, "/bundles", router, []string{"bundleId", model.RoleBundle, model.VendorDomain})
	bundleGroup.GET("/:bundleId/store", router.GetStore, nil)
	bundleGroup.PUT("/:bundleId/store", router.UpdateStore, nil)
	bundleGroup.DELETE("/:bundleId", router.Delete, nil)
	bundleGroup.POST("/:bundleId/packages", router.AddPackages, nil)
	bundleGroup.DELETE("/:bundleId/packages", router.RemovePackages, nil)

	return
}

func (router *BundleRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	path := ctx.Path()
	if strings.Contains(path, "/vendors/:vendorId") {
		return GetOwnerForVendor(ctx)
	}
	return GetOwnerForBundle(ctx)
}

func (router *BundleRouter) CreateStore(ctx echo.Context) (err error) {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid vendor Id")
	}

	params := createBundleDTO{}
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
	err = router.checkRBAC(userId, &qilinCtx, params.Packages)
	if err != nil {
		return err
	}

	bundle, err := router.service.CreateStore(vendorId, userId, params.Name, params.Packages)
	if err != nil {
		return err
	}
	dto, err := mapStoreBundleDto(bundle.(*model.StoreBundle))
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, dto)
}

func (router *BundleRouter) GetStoreList(ctx echo.Context) (err error) {
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
	query := ctx.QueryParam("query")
	sort := ctx.QueryParam("sort")
	total, bundles, err := router.service.GetStoreList(vendorId, query, sort, offset, limit)
	if err != nil {
		return err
	}
	dto := []*storeBundleItemDTO{}
	for _, bundle := range bundles {
		dto = append(dto, mapStoreBundleItemDto(bundle.(*model.StoreBundle)))
	}
	ctx.Response().Header().Add("X-Items-Count", fmt.Sprintf("%d", total))
	return ctx.JSON(http.StatusOK, dto)
}

func (router *BundleRouter) GetStore(ctx echo.Context) (err error) {
	bundleId, err := uuid.FromString(ctx.Param("bundleId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid vendor Id")
	}

	bundle, err := router.service.Get(bundleId)
	if err != nil {
		return err
	}

	bundleStore, ok := bundle.(*model.StoreBundle)
	if !ok {
		return orm.NewServiceError(http.StatusBadRequest, "Bundle not for store")
	}
	dto, err := mapStoreBundleDto(bundleStore)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (router *BundleRouter) UpdateStore(ctx echo.Context) (err error) {
	bundleId, err := uuid.FromString(ctx.Param("bundleId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid bundle Id")
	}
	storeDto := &storeBundleDTO{}
	err = ctx.Bind(storeDto)
	if err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errors.Wrap(err, "Wrong store bundle in body").Error())
	}
	if errs := ctx.Validate(storeDto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	store, err := mapStoreBundleModel(storeDto)
	if err != nil {
		return err
	}
	store.ID = bundleId
	storeRes, err := router.service.UpdateStore(store)
	if err != nil {
		return err
	}
	dto, err := mapStoreBundleDto(storeRes.(*model.StoreBundle))
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (router *BundleRouter) Delete(ctx echo.Context) (err error) {
	bundleId, err := uuid.FromString(ctx.Param("bundleId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid vendor Id")
	}

	err = router.service.Delete(bundleId)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (router *BundleRouter) checkRBAC(userId string, qilinCtx *rbac_echo.AppContext, packagesIds []uuid.UUID) error {
	for _, packageId := range packagesIds {
		owner, err := qilinCtx.GetOwnerForPackage(packageId)
		if err != nil {
			return err
		}
		if qilinCtx.CheckPermissions(userId, model.VendorDomain, model.PackageType, packageId.String(), owner, "read") != nil {
			return orm.NewServiceError(http.StatusForbidden, fmt.Sprintf("Access restricted for package `%s`", packageId.String()))
		}
	}
	return nil
}

func (router *BundleRouter) AddPackages(ctx echo.Context) (err error) {
	bundleId, err := uuid.FromString(ctx.Param("bundleId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid bundle Id")
	}
	packages := []uuid.UUID{}
	err = ctx.Bind(&packages)
	if err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, "Wrong package array in body")
	}

	userId, err := context.GetAuthUserId(ctx)
	if err != nil {
		return err
	}
	qilinCtx := ctx.(rbac_echo.AppContext)

	err = router.checkRBAC(userId, &qilinCtx, packages)
	if err != nil {
		return err
	}

	err = router.service.AddPackages(bundleId, packages)
	if err != nil {
		return err
	}
	return ctx.NoContent(http.StatusOK)
}

func (router *BundleRouter) RemovePackages(ctx echo.Context) (err error) {
	bundleId, err := uuid.FromString(ctx.Param("bundleId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid bundle Id")
	}
	packages := []uuid.UUID{}
	err = ctx.Bind(&packages)
	if err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, "Wrong package array in body")
	}
	err = router.service.RemovePackages(bundleId, packages)
	if err != nil {
		return err
	}
	return ctx.NoContent(http.StatusOK)
}
