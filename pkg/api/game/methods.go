package game

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	"qilin-api/pkg/orm"
	"strconv"
)

type CreateGameDTO struct {
	InternalName string
	VendorId     string
}

func (api *Router) Create(ctx echo.Context) error {
	params := CreateGameDTO{}
	err := ctx.Bind(&params)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Wrong parameters in body")
	}
	vendorId, err := uuid.FromString(params.VendorId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid vendorId")
	}
	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	game, err := api.gameService.Create(userId, vendorId, params.InternalName)
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
	gameId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	game, err := api.gameService.GetInfo(userId, gameId)
	if err != nil {
		return err
	}
	dto, err := mapGameInfo(game, api.gameService)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (api *Router) Delete(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	err = api.gameService.Delete(userId, gameId)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, "OK")
}

func (api *Router) UpdateInfo(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	dto := &UpdateGameDTO{}
	if err := ctx.Bind(dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}
	game := mapGameInfoBTO(dto)
	game.ID = gameId
	err = api.gameService.UpdateInfo(userId, &game)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, "OK")
}

func (api *Router) GetDescr(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game Id")
	}
	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	descr, err := api.gameService.GetDescr(userId, gameId)
	if err != nil {
		return err
	}
	dto := GameDescrDTO{
		Tagline:               descr.Tagline,
		Description:           descr.Description,
		AdditionalDescription: descr.AdditionalDescription,
		GameSite:              descr.GameSite,
		Reviews:               []DescrReview{},
		Socials: Socials{
			Facebook: descr.Socials.Facebook,
			Twitter:  descr.Socials.Twitter,
		},
	}
	for _, review := range descr.Reviews {
		dto.Reviews = append(dto.Reviews, DescrReview{
			PressName: review.PressName,
			Link:      review.Link,
			Score:     review.Score,
			Quote:     review.Quote,
		})
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (api *Router) UpdateDescr(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid game Id")
	}
	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	dto := &GameDescrDTO{}
	if err := ctx.Bind(dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}
	reviews := []bto.GameReview{}
	for _, review := range dto.Reviews {
		reviews = append(reviews, bto.GameReview{
			PressName: review.PressName,
			Link:      review.Link,
			Score:     review.Score,
			Quote:     review.Quote,
		})
	}
	err = api.gameService.UpdateDescr(userId, &model.GameDescr{
		GameID:                gameId,
		Tagline:               dto.Tagline,
		Description:           dto.Description,
		AdditionalDescription: dto.AdditionalDescription,
		GameSite:              dto.GameSite,
		Reviews:               reviews,
		Socials: bto.Socials{
			Facebook: dto.Socials.Facebook,
			Twitter:  dto.Socials.Twitter,
		},
	})
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, "OK")
}

func (api *Router) GetGenre(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	offset, _ := strconv.Atoi(ctx.QueryParam("offset"))
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 20
	}
	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	genres, err := api.gameService.FindGenres(userId, title, limit, offset)
	if err != nil {
		return err
	}
	dto := []GameTagDTO{}
	for _, genre := range genres {
		dto = append(dto, GameTagDTO{
			Id:    genre.ID,
			Title: genre.Title,
		})
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (api *Router) GetTags(ctx echo.Context) (err error) {
	title := ctx.QueryParam("title")
	offset, _ := strconv.Atoi(ctx.QueryParam("offset"))
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 20
	}
	userId, err := context.GetAuthUUID(ctx)
	if err != nil {
		return err
	}
	tags, err := api.gameService.FindTags(userId, title, limit, offset)
	if err != nil {
		return err
	}
	dto := []GameTagDTO{}
	for _, tag := range tags {
		dto = append(dto, GameTagDTO{
			Id:    tag.ID,
			Title: tag.Title,
		})
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (api *Router) GetRatingDescriptors(ctx echo.Context) error {
    system := ctx.QueryParam("title")
    descriptors, err := api.gameService.GetRatingDescriptors(system)
    if err != nil {
        return err
    }
    dto := []RatingDescriptorDTO{}
    for _, desc := range descriptors {
        dto = append(dto, RatingDescriptorDTO{
            Id: desc.ID,
            Title: desc.Title,
            System: desc.System,
        })
    }
    return ctx.JSON(http.StatusOK, dto)
}
