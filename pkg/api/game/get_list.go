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
	offset, err := strconv.Atoi(ctx.FormValue("offset"))
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(ctx.FormValue("offset"))
	if err != nil {
		limit = 20
	}
	internalName := ctx.FormValue("technicalName")
	genre := ctx.FormValue("genre")
	price, _ := strconv.ParseFloat(ctx.FormValue("price"), 64)
	releaseDate := ctx.FormValue("releaseDate")
	sort := ctx.FormValue("sort")

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
	for _, g := range games {
		// Filter only game genres
		genres := []GameTagDTO{}
		for _, a := range g.Genre {
			for _, b := range all_genres {
				if b.ID == a {
					genres = append(genres, GameTagDTO{Id: b.ID, Title: b.Title})
					break
				}
			}
		}
		dto = append(dto, ShortGameInfoDTO{
			ID: g.ID,
			InternalName: g.InternalName,
			Icon: "",
			Genre: genres,
			ReleaseDate: g.ReleaseDate,
			Prices: GamePricesDTO{},
		})
	}

	return ctx.JSON(http.StatusCreated, dto)
}
