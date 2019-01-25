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

func (api *Router) Create(ctx echo.Context) error {
    internalName := ctx.FormValue("internalName")
    vendorId, err := uuid.FromString(ctx.FormValue("vendorId"))
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid vendorId")
    }
    userId, err := context.GetAuthUUID(ctx)
    if err != nil {
        return err
    }
    game, err := api.gameService.Create(userId, vendorId, internalName)
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
    return ctx.JSON(http.StatusCreated, dto)
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
    return ctx.JSON(http.StatusCreated, "OK")
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
    return ctx.JSON(http.StatusCreated, "OK")
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
        Tagline: descr.Tagline,
        Description: descr.Description,
        AdditionalDescription: descr.AdditionalDescription,
        GameSite: descr.GameSite,
        Reviews: []DescrReview{},
        Socials: Socials{
            Facebook: descr.Socials.Facebook,
            Twitter: descr.Socials.Twitter,
        },
    }
    for _, r := range descr.Reviews {
        dto.Reviews = append(dto.Reviews, DescrReview{
            PressName: r.PressName,
            Link: r.Link,
            Score: r.Score,
            Quote: r.Quote,
        })
    }
    return ctx.JSON(http.StatusCreated, dto)
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
    for _, r := range dto.Reviews {
        reviews = append(reviews, bto.GameReview{
            PressName: r.PressName,
            Link: r.Link,
            Score: r.Score,
            Quote: r.Quote,
        })
    }
    err = api.gameService.UpdateDescr(userId, &model.GameDescr{
        GameID: gameId,
        Tagline: dto.Tagline,
        Description: dto.Description,
        AdditionalDescription: dto.AdditionalDescription,
        GameSite: dto.GameSite,
        Reviews: reviews,
        Socials: bto.Socials{
            Facebook: dto.Socials.Facebook,
            Twitter: dto.Socials.Twitter,
        },
    })
    if err != nil {
        return err
    }
    return ctx.JSON(http.StatusCreated, "OK")
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
    for _, item := range genres {
        dto = append(dto, GameTagDTO{
            Id: item.ID,
            Title: item.Title,
        })
    }
    return ctx.JSON(http.StatusCreated, dto)
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
    for _, item := range tags {
        dto = append(dto, GameTagDTO{
            Id: item.ID,
            Title: item.Title,
        })
    }
    return ctx.JSON(http.StatusCreated, dto)
}

func (api *Router) GetRatingDescriptors(ctx echo.Context) error {
    items, err := api.gameService.GetRatingDescriptors()
    if err != nil {
        return err
    }
    dto := []RatingDescriptorDTO{}
    for _, it := range items {
        dto = append(dto, RatingDescriptorDTO{
            Id: it.ID,
            Title: it.Title,
        })
    }
    return ctx.JSON(http.StatusCreated, dto)
}
