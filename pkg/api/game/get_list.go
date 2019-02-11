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
	limit, err := strconv.Atoi(ctx.QueryParam("offset"))
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

	all_genres, err := api.gameService.GetGenres(nil)
	if err != nil {
		return err
	}

	games, err := api.gameService.GetList(userId, vendorId, offset, limit, internalName, genre, releaseDate, sort, price)
	if err != nil {
		return err
	}
	dto := []ShortGameInfoDTO{}
	for _, game := range games {
		// Filter only game genres
		genres := []GameTagDTO{}
		for _, genre_id := range game.Genre {
			for _, genre := range all_genres {
				if genre.ID == genre_id {
					genres = append(genres, GameTagDTO{Id: genre.ID, Title: genre.Title})
					break
				}
			}
		}
		dto = append(dto, ShortGameInfoDTO{
			ID:           game.ID,
			InternalName: game.InternalName,
			Icon:         "",
			Genre:        genres,
			ReleaseDate:  game.ReleaseDate,
			Prices:       GamePricesDTO{},
		})
	}

	return ctx.JSON(http.StatusOK, dto)
}
