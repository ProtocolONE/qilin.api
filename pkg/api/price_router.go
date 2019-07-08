package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"qilin-api/pkg/utils"
	"strings"
)

type (
	PriceRouter struct {
		service     model.PriceService
		gameService model.GameService
	}

	pricesDTO struct {
		Common   basePrice        `json:"common" validate:"required,dive"`
		PreOrder preOrder         `json:"preOrder" validate:"required,dive"`
		Prices   []pricesInternal `json:"prices" validate:"-"`
	}

	pricesInternal struct {
		Currency string  `json:"currency" validate:"required"`
		Price    float32 `json:"price" validate:"min=0"`
		Vat      int32   `json:"vat" validate:"min=0"`
	}

	preOrder struct {
		Date    string `json:"date" validate:"required"`
		Enabled bool   `json:"enabled"`
	}

	basePrice struct {
		Currency        string `json:"currency" validate:"required"`
		NotifyRateJumps bool   `json:"notifyRateJumps"`
	}
)

//InitPriceRouter is initialization method for group
func InitPriceRouter(group *echo.Group, service model.PriceService, gameService model.GameService) (router *PriceRouter, err error) {
	priceRouter := PriceRouter{service, gameService}

	packageGroup := rbac_echo.Group(group, "/packages", &priceRouter, []string{"packageId", model.PackageType, model.VendorDomain})
	packageGroup.GET("/:packageId/prices", priceRouter.getBase, nil)
	packageGroup.PUT("/:packageId/prices", priceRouter.putBase, nil)
	packageGroup.PUT("/:packageId/prices/:currency", priceRouter.updatePrice, nil)
	packageGroup.DELETE("/:packageId/prices/:currency", priceRouter.deletePrice, nil)

	gameGroup := rbac_echo.Group(group, "/games", &priceRouter, []string{"gameId", model.GameType, model.VendorDomain})
	gameGroup.GET("/:gameId/prices", priceRouter.getBase, nil)
	gameGroup.PUT("/:gameId/prices", priceRouter.putBase, nil)
	gameGroup.PUT("/:gameId/prices/:currency", priceRouter.updatePrice, nil)
	gameGroup.DELETE("/:gameId/prices/:currency", priceRouter.deletePrice, nil)

	return &priceRouter, nil
}

func (router *PriceRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	path := ctx.Path()
	if strings.Contains(path, "/games/:gameId") {
		return GetOwnerForGame(ctx)
	}
	return GetOwnerForPackage(ctx)
}

// Func made for backward compatibility with games
func (router *PriceRouter) GetPackageID(ctx *echo.Context) (packageId uuid.UUID, err error) {
	gameId_str := (*ctx).Param("gameId")
	if gameId_str != "" {
		gameId, err := uuid.FromString(gameId_str)
		if err != nil {
			return uuid.Nil, orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
		}
		game, err := router.gameService.GetInfo(gameId)
		if err != nil {
			return uuid.Nil, err
		}
		packageId = game.DefaultPackageID
	} else {
		packageId, err = uuid.FromString((*ctx).Param("packageId"))
		if err != nil {
			return uuid.Nil, orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
		}
	}
	return
}

func (router *PriceRouter) getBase(ctx echo.Context) (err error) {

	id, err := router.GetPackageID(&ctx)
	if err != nil {
		return err
	}

	price, err := router.service.GetBase(id)
	if err != nil {
		return err
	}

	result := pricesDTO{}
	err = mapper.Map(price.PackagePrices, &result)

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Can't decode price from domain to DTO. Error: "+err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (router *PriceRouter) putBase(ctx echo.Context) error {
	id, err := router.GetPackageID(&ctx)
	if err != nil {
		return err
	}

	dto := new(pricesDTO)

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
	err = mapper.Map(dto, &basePrice.PackagePrices)

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	if err := router.service.UpdateBase(id, &basePrice); err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "")
}

func (router *PriceRouter) deletePrice(ctx echo.Context) error {
	id, err := router.GetPackageID(&ctx)
	if err != nil {
		return err
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
	id, err := router.GetPackageID(&ctx)
	if err != nil {
		return err
	}

	dto := new(pricesInternal)

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
