package game

import (
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func (api *Router) getList(ctx echo.Context) error {
	offset, err := strconv.Atoi(ctx.FormValue("offset"))
	if err != nil {
		offset = 0
	}
	limit, _ := strconv.Atoi(ctx.FormValue("offset"))
	if err != nil {
		limit = 20
	}
	technicalName := ctx.FormValue("technicalName")
	genre := ctx.FormValue("genre")
	price := ctx.FormValue("price")
	releaseDate := ctx.FormValue("releaseDate")
	sort := ctx.FormValue("sort")

	games, err := api.gameService.GetList(offset, limit, technicalName, genre, price, releaseDate, sort)
	if err != nil {
		return err
	}
	dto := []GameDTO{}
	for _, g := range games {
		dto = append(dto, GameDTO{
			InternalName: g.InternalName,
		})
	}

	return ctx.JSON(http.StatusCreated, dto)
}
