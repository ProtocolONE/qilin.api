package game

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"strconv"
)

func (api *Router) GetList(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid vendor Id")
	}
	offset, err := strconv.Atoi(ctx.QueryParam("offset"))
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 20
	}
	internalName := ctx.QueryParam("technicalName")
	genre := ctx.QueryParam("genre")
	price, _ := strconv.ParseFloat(ctx.QueryParam("price"), 64)
	releaseDate := ctx.QueryParam("releaseDate")
	sort := ctx.QueryParam("sort")

	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}

	games, err := api.gameService.GetList(userId, vendorId, offset, limit, internalName, genre, releaseDate, sort, price)
	if err != nil {
		return err
	}
	dto := []ShortGameInfoDTO{}
	for _, game := range games {
		prices := GamePriceDTO{
			Currency: game.Price.Currency,
			Price: float64(game.Price.Price),
		}
		dto = append(dto, ShortGameInfoDTO{
			ID:           game.Game.ID,
			InternalName: game.InternalName,
			Icon:         "",
			Genres:       GameGenreDTO{
				Main:       game.GenreMain,
				Addition:   game.GenreAddition,
			},
			ReleaseDate:  game.ReleaseDate,
			Prices:       prices,
		})
	}

	return ctx.JSON(http.StatusOK, dto)
}
