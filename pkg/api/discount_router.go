package api

import (
	"github.com/pkg/errors"
	"net/http"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
)

type Discount struct {
	ID          string                 `json:"id"`
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

	r := rbac_echo.Group(group,"/games/:gameId", &router, []string{"gameId", model.GameType, model.VendorDomain})
	r.GET("/discounts", router.get, nil)
	r.POST("/discounts", router.post, nil)
	r.PUT("/discounts/:discountId", router.put, nil)
	r.DELETE("/discounts/:discountId", router.delete, nil)

	return &router, nil
}

func (router *DiscountsRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	return GetOwnerForGame(ctx)
}

func (router *DiscountsRouter) post(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("gameId"))
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
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	domain.DateStart = dto.Date.Start
	domain.DateEnd = dto.Date.End

	discountID, err := router.service.AddDiscountForGame(id, &domain)
	if err != nil {
		return err
	}

	dto.ID = discountID.String()

	return ctx.JSON(http.StatusCreated, dto)
}

func (router *DiscountsRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("gameId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	discounts, err := router.service.GetDiscountsForGame(id)
	if err != nil {
		return err
	}

	dto := make([]Discount, len(discounts))
	err = mapper.Map(discounts, &dto)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "Can't decode domain to dto"))
	}

	if dto == nil {
		return ctx.JSON(http.StatusOK, make([]Discount, 0))
	}

	for i, d := range discounts {
		dto[i].ID = d.ID.String()
		dto[i].Date = DateRange{
			Start: d.DateStart,
			End:   d.DateEnd,
		}
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (router *DiscountsRouter) put(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("gameId"))
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
	domain.GameID = gameId

	if err := router.service.UpdateDiscountForGame(&domain); err != nil {
		return err
	}

	return nil
}

func (router *DiscountsRouter) delete(ctx echo.Context) error {
	_, err := uuid.FromString(ctx.Param("gameId"))
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
