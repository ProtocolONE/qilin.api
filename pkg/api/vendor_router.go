package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"strconv"
)

type VendorRouter struct {
	vendorService model.VendorService
}

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
		return echo.NewHTTPError(http.StatusInternalServerError, "Request vendors failed")
	}

	return ctx.JSON(http.StatusOK, vendors)
}

func (api *VendorRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	vendor, err := api.vendorService.FindByID(id)

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Vendor not found")
	}

	return ctx.JSON(http.StatusOK, vendor)
}

func (api *VendorRouter) create(ctx echo.Context) error {
	vendor := &model.Vendor{}
	if err := ctx.Bind(vendor); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}
	// Assign to new vendor current user id as manager
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	managerId, _ := uuid.FromBytes(claims["id"].([]byte))
	vendor.ManagerId = &managerId

	if err := api.vendorService.CreateVendor(vendor); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusCreated, vendor)
}

func (api *VendorRouter) update(ctx echo.Context) error {
	vendor := &model.Vendor{}

	if err := ctx.Bind(vendor); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	err := api.vendorService.UpdateVendor(vendor)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Vendor update failed")
	}

	return ctx.JSON(http.StatusOK, vendor)
}
