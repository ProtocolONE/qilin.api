package game

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
)

func (api *Router) Create(ctx echo.Context) error {
	internalName := ctx.FormValue("internalName")

	vendorId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	game, err := api.gameService.Create(vendorId, internalName)
	if err != nil {
		return err
	}
	dto, err := mapGameInfo(game, api.gameService)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, dto)
}

func (api *Router) GetInfo(ctx echo.Context) error {
	game_id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	vendorId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	game, err := api.gameService.GetInfo(vendorId, &game_id)
	if err != nil {
		return err
	}
	dto, err := mapGameInfo(game, api.gameService)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, dto)
}

func (api *Router) Delete(ctx echo.Context) error {
	game_id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	vendorId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	err = api.gameService.Delete(vendorId, &game_id)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, "Ok")
}

func (api *Router) Update(ctx echo.Context) error {
	game_id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	vendorId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}

	dto := &GameDTO{ID: game_id}
	if err := ctx.Bind(dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	//if errs := ctx.Validate(dto); errs != nil {
	//	return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	//}

	game := mapGameInfoBTO(dto)

	err = api.gameService.Update(vendorId, &game)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, "Ok")
}
