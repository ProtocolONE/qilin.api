package api

import (
	"bufio"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"io"
	"net/http"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strings"
)

type AddKeyListDTO struct {
	Keys []string `json:"keys" validate:"required"`
}

type KeyListRouter struct {
	keyListService model.KeyListService
}

func InitKeyListRouter(router *echo.Group, keyListService model.KeyListService) (*KeyListRouter, error){
	keyRouter := KeyListRouter{
		keyListService: keyListService,
	}
	r := rbac_echo.Group(router, "/packages/:packageId/keypackages/:keyPackageId", &keyRouter, []string{"*", model.PackageType, model.VendorDomain})
	r.POST("/keys", keyRouter.AddKeys, nil)
	r.POST("/file", keyRouter.AddFileKeys, nil)

	return &keyRouter, nil
}

func (router *KeyListRouter) AddFileKeys(ctx echo.Context) error {
	keyPackageIdParam := ctx.Param("keyPackageId")
	keyPackageId, err := uuid.FromString(keyPackageIdParam)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "keyPackageId is wrong")
	}

	file, err := ctx.FormFile("keys")
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	src, err := file.Open()
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}
	defer src.Close()

	reader := bufio.NewReader(src)
	var line string
	var codes []string
	shouldBreak := false
	for shouldBreak == false {
		line, err = reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return orm.NewServiceError(http.StatusBadRequest, err)
			}
			shouldBreak = true
		}
		line = strings.Trim(line, "\t\n")
		codes = append(codes, line)
	}
	if err := router.keyListService.AddKeys(keyPackageId, codes); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (router *KeyListRouter) AddKeys(ctx echo.Context) error {
	keyPackageIdParam := ctx.Param("keyPackageId")
	keyPackageId, err := uuid.FromString(keyPackageIdParam)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "keyPackageId is wrong")
	}

	dto := &AddKeyListDTO{}
	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}
	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	err = router.keyListService.AddKeys(keyPackageId, dto.Keys)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (*KeyListRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	return GetOwnerForPackage(ctx)
}
