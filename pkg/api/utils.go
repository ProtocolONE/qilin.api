package api

import (
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/orm"
)

func GetOwnerForGame(ctx rbac_echo.AppContext) (string, error) {
	gameIdParam := ctx.Param("gameId")
	gameId, err := uuid.FromString(gameIdParam)
	if err != nil {
		return "", orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Game id `%s` is incorrect", gameIdParam))
	}

	owner, err := ctx.GetOwnerForGame(gameId)
	if err != nil {
		return "", err
	}

	return owner, nil
}

func GetOwnerForVendor(ctx rbac_echo.AppContext) (string, error) {
	vendorIdParam := ctx.Param("vendorId")
	vendorId, err := uuid.FromString(vendorIdParam)
	if err != nil {
		return "", orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Vendor id `%s` is incorrect", vendorIdParam))
	}

	owner, err := ctx.GetOwnerForVendor(vendorId)
	if err != nil {
		return "", err
	}

	return owner, nil
}

func GetOwnerForPackage(ctx rbac_echo.AppContext) (string, error) {
	packageIdParam := ctx.Param("packageId")
	packageId, err := uuid.FromString(packageIdParam)
	if err != nil {
		return "", orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Package id `%s` is incorrect", packageIdParam))
	}

	owner, err := ctx.GetOwnerForPackage(packageId)
	if err != nil {
		return "", err
	}

	return owner, nil
}

func GetOwnerForBundle(ctx rbac_echo.AppContext) (string, error) {
	bundleIdParam := ctx.Param("bundleId")
	bundleId, err := uuid.FromString(bundleIdParam)
	if err != nil {
		return "", orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Bundle id `%s` is incorrect", bundleIdParam))
	}

	owner, err := ctx.GetOwnerForBundle(bundleId)
	if err != nil {
		return "", err
	}

	return owner, nil
}
