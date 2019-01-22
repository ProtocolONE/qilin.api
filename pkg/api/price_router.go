package api

import (
	"qilin-api/pkg/orm"

	"github.com/labstack/echo"
	"net/http"
	"qilin-api/pkg/mapper"
	"github.com/satori/go.uuid"
)

type (
	PriceRouter struct {
		gameService *orm.PriceService
	}

	Price struct {

	}
)

//InitPriceRouter is initialization method for router
func InitPriceRouter(group *echo.Group, service *orm.PriceService) (router *echo.Group, err error) {
	priceRouter := PriceRouter{
		gameService: service,
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
	
	price, err := router.gameService.Get(id)

	if err != nil {
		return err
	}

	result := Price{}
	err = mapper.Map(price, &result)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Can't decode price from domain to DTO. Error: " + err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (router *PriceRouter) createPrice(ctx echo.Context) error {
	panic("")
}

func (router *PriceRouter) updatePrice(ctx echo.Context) error {
	panic("")
}
