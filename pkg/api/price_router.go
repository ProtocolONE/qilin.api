package api

import (
	"fmt"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/utils"

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
		Common   BasePrice        `json:"common" validate:"required,dive"`
		PreOrder PreOrder         `json:"preOrder" validate:"required,dive"`
		Prices   []PricesInternal `json:"prices" validate:"-"`
	}

	PricesInternal struct {
		Currency string  `json:"currency" validate:"required"`
		Price    float32 `json:"price" validate:"required,gte=0"`
		Vat      int32   `json:"vat" validate:"required,gte=0"`
	}

	PreOrder struct {
		Date    string `json:"date" validate:"required"`
		Enabled bool   `json:"enabled"`
	}

	BasePrice struct {
		Currency        string `json:"currency" validate:"required"`
		NotifyRateJumps bool   `json:"notifyRateJumps"`
	}
)

//InitPriceRouter is initialization method for router
func InitPriceRouter(group *echo.Group, service *orm.PriceService) (router *PriceRouter, err error) {
	priceRouter := PriceRouter{
		service: service,
	}

	r := group.Group("/games/:id")

	r.GET("/prices", priceRouter.getBase)
	r.PUT("/prices", priceRouter.putBase)
	r.PUT("/prices/:currency", priceRouter.updatePrice)
	r.DELETE("/prices/:currency", priceRouter.deletePrice)

	return &priceRouter, nil
}

func (router *PriceRouter) getBase(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	price, err := router.service.GetBase(id)

	if err != nil {
		return err
	}

	result := PricesDTO{}
	err = mapper.Map(price, &result)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Can't decode price from domain to DTO. Error: "+err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (router *PriceRouter) putBase(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	dto := new(PricesDTO)

	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	if utils.IsCurrency(dto.Common.Currency) == false {
		return orm.NewServiceError(http.StatusUnprocessableEntity, fmt.Sprintf("Wrong currency %s", dto.Common.Currency))
	}

	basePrice := model.BasePrice{}
	err = mapper.Map(dto, &basePrice)

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	if err := router.service.UpdateBase(id, &basePrice); err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "")
}

func (router *PriceRouter) deletePrice(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	cur := ctx.Param("currency")

	if utils.IsCurrency(cur) == false {
		return orm.NewServiceError(http.StatusBadRequest, fmt.Sprintf("Wrong currency %s", cur))
	}

	price := model.Price{Currency: cur}

	if err := router.service.Delete(id, &price); err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "")
}

func (router *PriceRouter) updatePrice(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	dto := new(PricesInternal)

	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	if err := ctx.Validate(dto); err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, err)
	}

	price := model.Price{}
	err = mapper.Map(dto, &price)

	if err != nil {
		return orm.NewServiceError(http.StatusInternalServerError, err)
	}

	cur := ctx.Param("currency")

	if cur == "" || dto.Currency == "" {
		return orm.NewServiceError(http.StatusBadRequest, "Currency must be provided")
	}

	if cur != dto.Currency {
		return orm.NewServiceError(http.StatusBadRequest, fmt.Sprintf("Currency not equal. param: %v in model: %v", cur, dto.Currency))
	}

	if utils.IsCurrency(cur) == false {
		return orm.NewServiceError(http.StatusUnprocessableEntity, fmt.Sprintf("Wrong currency %s", cur))
	}

	if err := router.service.Update(id, &price); err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "")
}
