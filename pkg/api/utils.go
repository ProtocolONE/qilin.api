package api

import (
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/middleware"
	"qilin-api/pkg/orm"
)

func GetOwnerForGame(ctx middleware.QilinContext) (string, error) {
	gameIdParam := ctx.Param("id")
	gameId, err := uuid.FromString(gameIdParam)
	if err != nil {
		return "", orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Game id `%s` is incorrect", gameIdParam))
	}

	owner, err := ctx.GetOwnerForVendor(gameId)
	if err != nil {
		return "", err
	}

	return owner, nil
}

func GetOwnerForVendor(ctx middleware.QilinContext) (string, error) {
	vendorIdParam := ctx.Param("id")
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
