package api

import (
	"github.com/labstack/echo"
	"net/http"
	"qilin-api/pkg"
)

type GameRouter struct {
	gameService qilin.GameService
}

func InitGameRoutes(api *Server, service qilin.GameService) error {
	gameRouter := GameRouter{
		gameService: service,
	}

	api.Router.GET("/game", gameRouter.getAll)
	api.Router.GET("/game/:id", gameRouter.get)
	api.Router.GET("/game/findByName", gameRouter.findByName)
	api.Router.POST("/game", gameRouter.create)
	api.Router.PUT("/game/:id", gameRouter.update)

	return nil
}

func (api *GameRouter) findByName(ctx echo.Context) error {
	name := ctx.QueryParam("query")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Empty query not allowed")
	}

	games, err := api.gameService.FindByName(name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Request games failed")
	}

	return ctx.JSON(http.StatusOK, games)
}

func (api *GameRouter) getAll(ctx echo.Context) error {
	games, err := api.gameService.GetAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Request games failed")
	}

	return ctx.JSON(http.StatusOK, games)
}

// @Summary Get game
// @Description Get full data about game
// @Tags Game
// @Accept json
// @Produce json
// @Success 200 {object} model.Merchant "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/game/{id} [get]
func (api *GameRouter) get(ctx echo.Context) error {
	game, err := api.gameService.FindByID(ctx.Param("id"))

	if err != nil || game == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Game not found")
	}

	return ctx.JSON(http.StatusOK, game)
}

// @Summary Create game
// @Description Create new game
// @Tags Game
// @Accept json
// @Produce json
// @Param data body model.Game true "Creating game data"
// @Success 201 {object} model.Game "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/game [post]
func (api *GameRouter) create(ctx echo.Context) error {
	game := &qilin.Game{}

	if err := ctx.Bind(game); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	if err := api.gameService.CreateGame(game); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Game create failed")
	}

	return ctx.JSON(http.StatusCreated, game)
}

// @Summary Update game
// @Description Update game data
// @Tags Game
// @Accept json
// @Produce json
// @Param data body model.Game true "Game object with new data"
// @Success 200 {object} model.Game "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/game/:id [put]
func (api *GameRouter) update(ctx echo.Context) error {
	game := &qilin.Game{}

	if err := ctx.Bind(game); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	err := api.gameService.UpdateGame(game)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Game update failed")
	}

	return ctx.JSON(http.StatusOK, game)
}
