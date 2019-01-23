package api

import (
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strconv"
)

type (
	VendorRouter struct {
		vendorService model.VendorService
	}

	VendorDTO struct {
		Id                  uuid.UUID       `json:"id"`
		Name                string          `json:"name" validate:"required,min=2"`
		Domain3             string          `json:"domain3" validate:"required,min=2"`
		Email               string          `json:"email" validate:"required,email"`
		ManagerId           uuid.UUID       `json:"manager_id"`
		HowManyProducts     string          `json:"howmanyproducts"`
	}
)

func InitVendorRoutes(api *Server, service model.VendorService) error {
	vendorRouter := VendorRouter{
		vendorService: service,
	}

	api.Router.GET("/vendors", vendorRouter.getAll)
	api.Router.GET("/vendors/:id", vendorRouter.get)
	api.Router.POST("/vendors", vendorRouter.create)
	api.Router.PUT("/vendors/:id", vendorRouter.update)

	return nil
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
	vendors, err := api.vendorService.GetAll(limit, offset)
	if err != nil {
		return err
	}

	dto := []VendorDTO{}
	for _, v := range vendors {
		dto = append(dto, VendorDTO{
			Id: v.ID,
			Name: v.Name,
			Domain3: v.Domain3,
			Email: v.Email,
			ManagerId: *v.ManagerId,
			HowManyProducts: v.HowManyProducts,
		})
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (api *VendorRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}
	vendor, err := api.vendorService.FindByID(id)
	if err != nil {
		return orm.NewServiceError(http.StatusNotFound, "Vendor not found")
	}
	return ctx.JSON(http.StatusOK, VendorDTO{
		Id: vendor.ID,
		Name: vendor.Name,
		Domain3: vendor.Domain3,
		Email: vendor.Email,
		ManagerId: *vendor.ManagerId,
		HowManyProducts: vendor.HowManyProducts,
	})
}

func (api *VendorRouter) create(ctx echo.Context) error {
	dto := &VendorDTO{}
	if err := ctx.Bind(dto); err != nil {
		return errors.Wrap(err, "Bind vendor obj")
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	// Assign to new vendor current user id as manager
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	data, _ := base64.StdEncoding.DecodeString(claims["id"].(string))
	managerId, _ := uuid.FromBytes(data)

	bto, err := api.vendorService.CreateVendor(&model.Vendor{
		Name: dto.Name,
		Domain3: dto.Domain3,
		Email: dto.Email,
		HowManyProducts: dto.HowManyProducts,
		ManagerId: &managerId,
	})
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, VendorDTO{
		Id: bto.ID,
		Name: bto.Name,
		Domain3: bto.Domain3,
		Email: bto.Email,
		HowManyProducts: bto.HowManyProducts,
		ManagerId: *bto.ManagerId,
	})
}

func (api *VendorRouter) update(ctx echo.Context) error {
	dto := &VendorDTO{}
	if err := ctx.Bind(dto); err != nil {
		return errors.Wrap(err, "Bind vendor obj")
	}
	vendorId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	vendor, err := api.vendorService.UpdateVendor(&model.Vendor{
		ID: vendorId,
		Name: dto.Name,
		Domain3: dto.Domain3,
		Email: dto.Email,
		HowManyProducts: dto.HowManyProducts,
	})
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, VendorDTO{
		Id: vendor.ID,
		Name: vendor.Name,
		Domain3: vendor.Domain3,
		Email: vendor.Email,
		HowManyProducts: vendor.HowManyProducts,
		ManagerId: *vendor.ManagerId,
	})
}
