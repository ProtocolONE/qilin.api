package api

import (
	"github.com/labstack/echo"
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
	api.Router.GET("/vendors/findByName", vendorRouter.findByName)
	api.Router.POST("/vendors", vendorRouter.create)
	api.Router.PUT("/vendors/:id", vendorRouter.update)

	return nil
}

func (api *VendorRouter) findByName(ctx echo.Context) error {
	name := ctx.QueryParam("query")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Empty query not allowed")
	}

	vendors, err := api.vendorService.FindByName(name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Request vendors failed")
	}

	return ctx.JSON(http.StatusOK, vendors)
}

func (api *VendorRouter) getAll(ctx echo.Context) error {
	vendors, err := api.vendorService.GetAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Request vendors failed")
	}

	return ctx.JSON(http.StatusOK, vendors)
}

// @Summary Get vendor
// @Description Get full data about vendor
// @Tags Vendor
// @Accept json
// @Produce json
// @Success 200 {object} model.Merchant "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/vendor/{id} [get]
func (api *VendorRouter) get(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	vendor, err := api.vendorService.FindByID(uint(id))

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Vendor not found")
	}

	return ctx.JSON(http.StatusOK, vendor)
}

// @Summary Create vendor
// @Description Create new vendor
// @Tags Vendor
// @Accept json
// @Produce json
// @Param data body model.Vendor true "Creating vendor data"
// @Success 201 {object} model.Vendor "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/vendor [post]
func (api *VendorRouter) create(ctx echo.Context) error {
	vendor := &model.Vendor{}

	if err := ctx.Bind(vendor); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	if err := api.vendorService.CreateVendor(vendor); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Vendor create failed")
	}

	return ctx.JSON(http.StatusCreated, vendor)
}

// @Summary Update vendor
// @Description Update vendor data
// @Tags Vendor
// @Accept json
// @Produce json
// @Param data body model.Vendor true "Vendor object with new data"
// @Success 200 {object} model.Vendor "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/vendor/:id [put]
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
