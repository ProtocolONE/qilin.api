package api

import (
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/middleware"
	"qilin-api/pkg/orm"
)

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
