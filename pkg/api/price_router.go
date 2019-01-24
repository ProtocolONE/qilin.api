package api

import (
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"

	"net/http"
	"qilin-api/pkg/mapper"

	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

type (
	PriceRouter struct {
		service *orm.PriceService
	}

	PricesDTO struct {
		Normal   Price            `json:"normal" validate:"required,dive" mapper:"ignore"`
		PreOrder PreOrder         `json:"preOrder" validate:"required,dive"`
		Prices   []PricesInternal `json:"prices" validate:"required,dive"`
	}

	PricesInternal struct {
		Currency string  `json:"currency" validate:"required"`
		Price    float32 `json:"price" validate:"required"`
		Vat      int32   `json:"vat" validate:"required"`
	}

	PreOrder struct {
		Date    string `json:"date" validate:"required"`
		Enabled bool   `json:"enabled" validate:"required"`
	}

	Price struct {
		Currency string  `json:"currency" validate:"required"`
		Price    float32 `json:"price" validate:"required"`
	}
)

//InitPriceRouter is initialization method for router
func InitPriceRouter(group *echo.Group, service *orm.PriceService) (router *echo.Group, err error) {
	priceRouter := PriceRouter{
		service: service,
	}

	router = group.Group("/games/:id")

	router.GET("/prices", priceRouter.get)

	return router, nil
}

func (router *PriceRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	price, err := router.service.Get(id)

	if err != nil {
		return err
	}

	result := Price{}
	err = mapper.Map(price, &result)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Can't decode price from domain to DTO. Error: "+err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (router *PriceRouter) put(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	dto := new(Price)

	if err := ctx.Bind(dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	prices := model.Price{}
	err = mapper.Map(dto, &prices)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := router.service.Update(id, &prices); err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "")
}

func (router *PriceRouter) createPrice(ctx echo.Context) error {
	panic("")
}

func (router *PriceRouter) updatePrice(ctx echo.Context) error {
	panic("")
}
