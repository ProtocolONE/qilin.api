package api

import (
	"net/http"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	"time"

	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

type Discount struct {
	Title       *utils.LocalizedString `json:"title" validate:"required"`
	Description *utils.LocalizedString `json:"description"`
	Date        DateRange              `json:"date" validate:"required,dive,required"`
	Rate        float32                `json:"rate" validate:"required,gte=0"`
}

type DateRange struct {
	Start time.Time `json:"start" validate:"required"`
	End   time.Time `json:"end" validate:"required"`
}

type DiscountsRouter struct {
	service *orm.DiscountService
}

type DiscountCreated struct {
	ID uuid.UUID `json:"id"`
}

func InitDiscountsRouter(group *echo.Group, service *orm.DiscountService) (*DiscountsRouter, error) {
	router := DiscountsRouter{
		service: service,
	}

	r := group.Group("/games/:id")
	r.GET("/discounts", router.get)
	r.PUT("/discounts/:discountId", router.put)
	r.POST("/discounts", router.post)
	r.DELETE("/discounts/:discountId", router.delete)

	return &router, nil
}

func (router *DiscountsRouter) post(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	dto := new(Discount)
	if err := ctx.Bind(dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	domain := model.Discount{}
	err = mapper.Map(dto, &domain)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	domain.DateStart = dto.Date.Start
	domain.DateEnd = dto.Date.End

	discountID, err := router.service.AddDiscountForGame(id, &domain)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, DiscountCreated{discountID})
}

func (router *DiscountsRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	discounts, err := router.service.GetDiscountsForGame(id)
	if err != nil {
		return err
	}

	var dto []Discount
	err = mapper.Map(discounts, &dto)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Can't decode domain to dto")
	}

	if dto == nil {
		return ctx.JSON(http.StatusOK, make([]Discount, 0))
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (router *DiscountsRouter) put(ctx echo.Context) error {
	_, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	discountID, err := uuid.FromString(ctx.Param("discountId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid discount Id")
	}

	dto := new(Discount)
	if err := ctx.Bind(dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	domain := model.Discount{}
	err = mapper.Map(dto, &domain)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	domain.DateStart = dto.Date.Start
	domain.DateEnd = dto.Date.End
	domain.ID = discountID

	if err := router.service.UpdateDiscountForGame(&domain); err != nil {
		return err
	}

	return nil
}

func (router *DiscountsRouter) delete(ctx echo.Context) error {
	_, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	discountID, err := uuid.FromString(ctx.Param("discountId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid discount Id")
	}

	if err = router.service.RemoveDiscountForGame(discountID); err != nil {
		return err
	}

	return nil
}
