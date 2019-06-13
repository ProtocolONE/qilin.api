package api

import (
	"errors"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"time"
)

type (
	KeyPackageDTO struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Created string `json:"created"`
		Updated string `json:"updated"`
	}

	CreateKeyPackageDTO struct {
		Name string `json:"name" validate:"required"`
		Type string `json:"type" validate:"required"`
	}

	ChangeKeyPackageDTO struct {
		Name string `json:"name" validate:"required"`
	}
)

type keyPackageRouter struct {
	keyPackageService model.KeyPackageService
}

func InitKeyPackageRouter(router *echo.Group, keyPackageService model.KeyPackageService) (*keyPackageRouter, error) {
	if keyPackageService == nil {
		return nil, errors.New("Key Package service must be provided")
	}

	keyRouter := keyPackageRouter{
		keyPackageService: keyPackageService,
	}

	r := rbac_echo.Group(router, "/packages/:packageId", &keyRouter, []string{"*", model.PackageType, model.VendorDomain})
	r.GET("/keypackages", keyRouter.GetList, nil)
	r.POST("/keypackages", keyRouter.Create, nil)
	r.GET("/keypackages/:keyPackageId", keyRouter.Get, nil)
	r.PUT("/keypackages/:keyPackageId", keyRouter.Change, nil)

	return &keyRouter, nil
}

func (router *keyPackageRouter) Get(ctx echo.Context) (err error) {
	packageId, err := getKeyPackageId(ctx)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	keyPackage, err := router.keyPackageService.Get(packageId)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, mapKeyPackage(keyPackage))
}

func (router *keyPackageRouter) GetList(ctx echo.Context) (err error) {
	packageId, err := getPackageId(ctx)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	keyPackages, err := router.keyPackageService.List(packageId)
	if err != nil {
		return err
	}

	result := []KeyPackageDTO{}
	for _, keyPackage := range keyPackages {
		result = append(result, mapKeyPackage(&keyPackage))
	}

	return ctx.JSON(http.StatusOK, result)
}

func (router *keyPackageRouter) Create(ctx echo.Context) (err error) {
	packageId, err := getPackageId(ctx)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	dto := &CreateKeyPackageDTO{}
	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}
	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	keyPackage, err := router.keyPackageService.Create(packageId, dto.Name, model.KeyStreamType(dto.Type))
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, mapKeyPackage(keyPackage))
}

func (router *keyPackageRouter) Change(ctx echo.Context) (err error) {
	packageId, err := getKeyPackageId(ctx)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	dto := &ChangeKeyPackageDTO{}
	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}
	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	keyPackage, err := router.keyPackageService.Update(packageId, dto.Name)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, mapKeyPackage(keyPackage))
}

func getPackageId(ctx echo.Context) (uuid.UUID, error) {
	packageIdStr := ctx.Param("packageId")
	return uuid.FromString(packageIdStr)
}

func getKeyPackageId(ctx echo.Context) (uuid.UUID, error) {
	packageIdStr := ctx.Param("keyPackageId")
	return uuid.FromString(packageIdStr)
}

func (*keyPackageRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	return GetOwnerForPackage(ctx)
}

func mapKeyPackage(keyPackage *model.KeyPackage) KeyPackageDTO {
	return KeyPackageDTO{
		Type:    keyPackage.KeyStreamType.String(),
		Name:    keyPackage.Name,
		ID:      keyPackage.ID.String(),
		Created: keyPackage.CreatedAt.Format(time.RFC3339),
		Updated: keyPackage.UpdatedAt.Format(time.RFC3339),
	}
}
