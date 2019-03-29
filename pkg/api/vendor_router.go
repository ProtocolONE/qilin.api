package api

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strconv"
)

type (
	VendorRouter struct {
		vendorService model.VendorService
		userService   model.UserService
	}

	VendorDTO struct {
		Id              uuid.UUID `json:"id"`
		Name            string    `json:"name" validate:"required,min=2"`
		Domain3         string    `json:"domain3" validate:"required,min=2"`
		Email           string    `json:"email" validate:"required,email"`
		ManagerId       string    `json:"manager_id"`
		HowManyProducts string    `json:"howmanyproducts"`
	}
)

func InitVendorRoutes(group *echo.Group, service model.VendorService, userService model.UserService) error {
	vendorRouter := VendorRouter{
		vendorService: service,
		userService:   userService,
	}

	router := rbac_echo.Group(group, "/vendors", &vendorRouter, []string{"*", model.VendorType, model.VendorDomain})
	router.GET("/:vendorId", vendorRouter.get, nil)
	router.PUT("/:vendorId", vendorRouter.update, nil)

	group.GET("/vendors", vendorRouter.getAll)
	group.POST("/vendors", vendorRouter.create)

	return nil
}

func (api *VendorRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	return GetOwnerForVendor(ctx)
}

func (api *VendorRouter) getAll(ctx echo.Context) error {
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 20
	}
	offset, err := strconv.Atoi(ctx.QueryParam("offset"))
	if err != nil {
		offset = 0
	}

	qilinCtx := ctx.(rbac_echo.AppContext)
	userId, err := api.getUserId(ctx)
	shouldBreak := false
	localOffset := offset
	var dto []VendorDTO

	for len(dto) <= limit && shouldBreak == false {
		localLimit := limit - len(dto)
		vendors, err := api.vendorService.GetAll(localLimit, localOffset)
		if err != nil {
			return err
		}

		// we do not have enough items in DB
		shouldBreak = len(vendors) < localLimit

		for _, v := range vendors {
			owner := v.ManagerID

			// filter games that user do not have rights
			if qilinCtx.CheckPermissions(userId, model.VendorDomain, model.VendorType, "*", owner, "read") != nil {
				continue
			}

			dto = append(dto, VendorDTO{
				Id:              v.ID,
				Name:            v.Name,
				Domain3:         v.Domain3,
				Email:           v.Email,
				ManagerId:       v.ManagerID,
				HowManyProducts: v.HowManyProducts,
			})
		}

		localOffset = localOffset + len(vendors)
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (api *VendorRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}
	vendor, err := api.vendorService.FindByID(id)
	if err != nil {
		return orm.NewServiceError(http.StatusNotFound, "Vendor not found")
	}
	return ctx.JSON(http.StatusOK, VendorDTO{
		Id:              vendor.ID,
		Name:            vendor.Name,
		Domain3:         vendor.Domain3,
		Email:           vendor.Email,
		ManagerId:       vendor.ManagerID,
		HowManyProducts: vendor.HowManyProducts,
	})
}

func (api *VendorRouter) create(ctx echo.Context) error {
	dto := &VendorDTO{}
	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bind vendor obj"))
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	// Assign to new vendor current user id as manager
	userId, err := api.getUserId(ctx)
	if err != nil {
		return err
	}

	bto, err := api.vendorService.Create(&model.Vendor{
		Name:            dto.Name,
		Domain3:         dto.Domain3,
		Email:           dto.Email,
		HowManyProducts: dto.HowManyProducts,
		ManagerID:       userId,
	})
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, VendorDTO{
		Id:              bto.ID,
		Name:            bto.Name,
		Domain3:         bto.Domain3,
		Email:           bto.Email,
		HowManyProducts: bto.HowManyProducts,
		ManagerId:       bto.ManagerID,
	})
}

func (api *VendorRouter) update(ctx echo.Context) error {
	dto := &VendorDTO{}
	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bind vendor obj"))
	}
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	vendor, err := api.vendorService.Update(&model.Vendor{
		ID:              vendorId,
		Name:            dto.Name,
		Domain3:         dto.Domain3,
		Email:           dto.Email,
		HowManyProducts: dto.HowManyProducts,
		ManagerID:       dto.ManagerId,
	})
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, VendorDTO{
		Id:              vendor.ID,
		Name:            vendor.Name,
		Domain3:         vendor.Domain3,
		Email:           vendor.Email,
		HowManyProducts: vendor.HowManyProducts,
		ManagerId:       vendor.ManagerID,
	})
}

func (api *VendorRouter) getUserId(ctx echo.Context) (string, error) {
	extUserId, err := context.GetAuthUserId(ctx)
	if err != nil {
		return "", err
	}
	user, err := api.userService.FindByID(extUserId)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}
