package api

import (
	"github.com/labstack/echo/v4"
	"github.com/lunny/html2md"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"gopkg.in/russross/blackfriday.v2"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/model"
	bto "qilin-api/pkg/model/game"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type GameRouter struct {
	gameService    model.GameService
	userService    model.UserService
	eventBus       model.EventBus
	productService model.ProductService
}

type (
	MachineRequirementsDTO struct {
		System           string `json:"system"`
		Processor        string `json:"processor"`
		Graphics         string `json:"graphics"`
		Sound            string `json:"sound"`
		Ram              int32  `json:"ram"`
		RamDimension     string `json:"ramdimension"`
		Storage          int32  `json:"storage"`
		StorageDimension string `json:"storagedimension"`
		Other            string `json:"other"`
	}

	PlatformRequirementsDTO struct {
		Minimal     MachineRequirementsDTO `json:"minimal" validate:"required,dive"`
		Recommended MachineRequirementsDTO `json:"recommended" validate:"required,dive"`
	}

	GameRequirementsDTO struct {
		Windows PlatformRequirementsDTO `json:"windows" validate:"required,dive"`
		MacOs   PlatformRequirementsDTO `json:"macOs" validate:"required,dive"`
		Linux   PlatformRequirementsDTO `json:"linux" validate:"required,dive"`
	}

	GamePlatformDTO struct {
		Windows bool `json:"windows"`
		MacOs   bool `json:"macOs"`
		Linux   bool `json:"linux"`
	}

	LangsDTO struct {
		Voice     bool `json:"voice"`
		Interface bool `json:"interface"`
		Subtitles bool `json:"subtitles"`
	}

	GameLangsDTO struct {
		EN LangsDTO `json:"en" validate:"dive"`
		RU LangsDTO `json:"ru" validate:"dive"`
		FR LangsDTO `json:"fr" validate:"dive"`
		ES LangsDTO `json:"es" validate:"dive"`
		DE LangsDTO `json:"de" validate:"dive"`
		IT LangsDTO `json:"it" validate:"dive"`
		PT LangsDTO `json:"pt" validate:"dive"`
	}

	GameTagDTO struct {
		Id    int64                 `json:"id" validate:"required"`
		Title utils.LocalizedString `json:"title" validate:"dive"`
	}

	RatingDescriptorDTO struct {
		Id     uint                  `json:"id" validate:"required"`
		Title  utils.LocalizedString `json:"title" validate:"dive"`
		System string                `json:"system"`
	}

	GameFeaturesDTO struct {
		Common      []string `json:"common"`
		Controllers string   `json:"controllers"`
	}

	BaseGameDTO struct {
		InternalName         string              `json:"internalName"`
		Title                string              `json:"title"`
		Developers           string              `json:"developers"`
		Publishers           string              `json:"publishers"`
		ReleaseDate          time.Time           `json:"releaseDate" validate:"required"`
		DisplayRemainingTime bool                `json:"displayRemainingTime"`
		AchievementOnProd    bool                `json:"achievementOnProd"`
		Features             GameFeaturesDTO     `json:"features" validate:"required,dive"`
		Platforms            GamePlatformDTO     `json:"platforms" validate:"required,dive"`
		Requirements         GameRequirementsDTO `json:"requirements" validate:"required,dive"`
		Languages            GameLangsDTO        `json:"languages" validate:"required,dive"`
	}

	GameGenreDTO struct {
		Main     int64   `json:"main"`
		Addition []int64 `json:"addition" validate:"required"`
	}

	GameDTO struct {
		ID uuid.UUID `json:"id"`
		BaseGameDTO
		Genres           GameGenreDTO `json:"genres" validate:"required,dive"`
		Tags             []int64      `json:"tags" validate:"required"`
		DefaultPackageID uuid.UUID    `json:"defaultPackageId"`
	}

	UpdateGameDTO struct {
		BaseGameDTO
		Genres GameGenreDTO `json:"genres" validate:"required,dive"`
		Tags   []int64      `json:"tags" validate:"required"`
	}

	ShortGameInfoDTO struct {
		ID           uuid.UUID    `json:"id"`
		InternalName string       `json:"internalName"`
		Icon         string       `json:"icon"`
		Genres       GameGenreDTO `json:"genres"`
		ReleaseDate  time.Time    `json:"releaseDate"`
	}

	DescrReview struct {
		PressName string `json:"pressName"`
		Link      string `json:"link"`
		Score     string `json:"score"`
		Quote     string `json:"quote"`
	}

	Socials struct {
		Facebook string `json:"facebook"`
		Twitter  string `json:"twitter"`
	}

	GameDescrDTO struct {
		Tagline               utils.LocalizedString `json:"tagline" validate:"dive"`
		Description           utils.LocalizedString `json:"description" validate:"dive"`
		Reviews               []DescrReview         `json:"reviews" validate:"required,dive"`
		AdditionalDescription string                `json:"additionalDescription"`
		GameSite              string                `json:"gameSite"`
		Socials               Socials               `json:"socials"`
	}
)

func mapReqs(r *bto.MachineRequirements) MachineRequirementsDTO {
	return MachineRequirementsDTO{
		System:           r.System,
		Processor:        r.Processor,
		Graphics:         r.Graphics,
		Sound:            r.Sound,
		Ram:              r.Ram,
		RamDimension:     r.RamDimension,
		Storage:          r.Storage,
		StorageDimension: r.StorageDimension,
		Other:            r.Other,
	}
}

func mapReqsBTO(r *MachineRequirementsDTO) bto.MachineRequirements {
	return bto.MachineRequirements{
		System:           r.System,
		Processor:        r.Processor,
		Graphics:         r.Graphics,
		Sound:            r.Sound,
		Ram:              r.Ram,
		RamDimension:     r.RamDimension,
		Storage:          r.Storage,
		StorageDimension: r.StorageDimension,
		Other:            r.Other,
	}
}

func mapGameInfo(game *model.Game) *GameDTO {
	return &GameDTO{
		ID: game.ID,
		BaseGameDTO: BaseGameDTO{
			InternalName:         game.InternalName,
			Title:                game.Title,
			Developers:           game.Developers,
			Publishers:           game.Publishers,
			ReleaseDate:          game.ReleaseDate,
			DisplayRemainingTime: game.DisplayRemainingTime,
			AchievementOnProd:    game.AchievementOnProd,
			Features:             GameFeaturesDTO{Common: game.FeaturesCommon, Controllers: game.FeaturesCtrl},
			Platforms: GamePlatformDTO{
				Windows: game.Platforms.Windows,
				MacOs:   game.Platforms.MacOs,
				Linux:   game.Platforms.Linux,
			},
			Requirements: GameRequirementsDTO{
				Windows: PlatformRequirementsDTO{
					Minimal:     mapReqs(&game.Requirements.Windows.Minimal),
					Recommended: mapReqs(&game.Requirements.Windows.Recommended)},
				Linux: PlatformRequirementsDTO{
					Minimal:     mapReqs(&game.Requirements.Linux.Minimal),
					Recommended: mapReqs(&game.Requirements.Linux.Recommended)},
				MacOs: PlatformRequirementsDTO{
					Minimal:     mapReqs(&game.Requirements.MacOs.Minimal),
					Recommended: mapReqs(&game.Requirements.MacOs.Recommended)},
			},
			Languages: GameLangsDTO{
				EN: LangsDTO{game.Languages.EN.Voice, game.Languages.EN.Interface, game.Languages.EN.Subtitles},
				RU: LangsDTO{game.Languages.RU.Voice, game.Languages.RU.Interface, game.Languages.RU.Subtitles},
				FR: LangsDTO{game.Languages.FR.Voice, game.Languages.FR.Interface, game.Languages.FR.Subtitles},
				ES: LangsDTO{game.Languages.ES.Voice, game.Languages.ES.Interface, game.Languages.ES.Subtitles},
				DE: LangsDTO{game.Languages.DE.Voice, game.Languages.DE.Interface, game.Languages.DE.Subtitles},
				IT: LangsDTO{game.Languages.IT.Voice, game.Languages.IT.Interface, game.Languages.IT.Subtitles},
				PT: LangsDTO{game.Languages.PT.Voice, game.Languages.PT.Interface, game.Languages.PT.Subtitles},
			},
		},
		Genres: GameGenreDTO{
			Main:     game.GenreMain,
			Addition: game.GenreAddition,
		},
		Tags:             game.Tags,
		DefaultPackageID: game.DefaultPackageID,
	}
}

func mapGameInfoBTO(game *UpdateGameDTO) *model.Game {
	return &model.Game{
		InternalName:         game.InternalName,
		Title:                game.Title,
		Developers:           game.Developers,
		Publishers:           game.Publishers,
		ReleaseDate:          game.ReleaseDate,
		DisplayRemainingTime: game.DisplayRemainingTime,
		AchievementOnProd:    game.AchievementOnProd,
		FeaturesCommon:       game.Features.Common,
		FeaturesCtrl:         game.Features.Controllers,
		Platforms: bto.Platforms{
			Windows: game.Platforms.Windows,
			MacOs:   game.Platforms.MacOs,
			Linux:   game.Platforms.Linux,
		},
		Requirements: bto.GameRequirements{
			Windows: bto.PlatformRequirements{
				Minimal:     mapReqsBTO(&game.Requirements.Windows.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.Windows.Recommended)},
			Linux: bto.PlatformRequirements{
				Minimal:     mapReqsBTO(&game.Requirements.Linux.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.Linux.Recommended)},
			MacOs: bto.PlatformRequirements{
				Minimal:     mapReqsBTO(&game.Requirements.MacOs.Minimal),
				Recommended: mapReqsBTO(&game.Requirements.MacOs.Recommended)},
		},
		Languages: bto.GameLangs{
			EN: bto.Langs{game.Languages.EN.Voice, game.Languages.EN.Interface, game.Languages.EN.Subtitles},
			RU: bto.Langs{game.Languages.RU.Voice, game.Languages.RU.Interface, game.Languages.RU.Subtitles},
			FR: bto.Langs{game.Languages.FR.Voice, game.Languages.FR.Interface, game.Languages.FR.Subtitles},
			ES: bto.Langs{game.Languages.ES.Voice, game.Languages.ES.Interface, game.Languages.ES.Subtitles},
			DE: bto.Langs{game.Languages.DE.Voice, game.Languages.DE.Interface, game.Languages.DE.Subtitles},
			IT: bto.Langs{game.Languages.IT.Voice, game.Languages.IT.Interface, game.Languages.IT.Subtitles},
			PT: bto.Langs{game.Languages.PT.Voice, game.Languages.PT.Interface, game.Languages.PT.Subtitles},
		},
		GenreMain:     game.Genres.Main,
		GenreAddition: game.Genres.Addition,
		Tags:          game.Tags,
	}
}

func InitGameRoutes(router *echo.Group, service model.GameService, userService model.UserService, bus model.EventBus) (*GameRouter, error) {
	if service == nil {
		return nil, errors.New("service must be provided")
	}

	if userService == nil {
		return nil, errors.New("user service must be provided")
	}

	if bus == nil {
		return nil, errors.New("event bus must be provided")
	}

	Router := GameRouter{
		gameService: service,
		userService: userService,
		eventBus:    bus,
	}

	r := rbac_echo.Group(router, "/vendors/:vendorId", &Router, []string{"*", model.VendorGameType, model.VendorDomain})
	r.GET("/games", Router.GetList, nil)
	r.POST("/games", Router.Create, nil)

	gameGroup := rbac_echo.Group(router, "/games", &Router, []string{"gameId", model.GameType, model.VendorDomain})
	gameGroup.GET("/:gameId", Router.GetInfo, nil)
	gameGroup.DELETE("/:gameId", Router.Delete, nil)
	gameGroup.PUT("/:gameId", Router.UpdateInfo, nil)
	gameGroup.GET("/:gameId/descriptions", Router.GetDescr, nil)
	gameGroup.PUT("/:gameId/descriptions", Router.UpdateDescr, nil)
	gameGroup.POST("/:gameId/publications", Router.PublishGame, []string{"gameId", model.PublishGame, model.VendorDomain})
	gameGroup.GET("/:gameId/packages", Router.GetPackages, nil)

	router.GET("/genre", Router.GetGenres) // TODO: Remove after some time
	router.GET("/genres", Router.GetGenres)
	router.GET("/tags", Router.GetTags)
	router.GET("/descriptors", Router.GetRatingDescriptors)

	return &Router, nil
}

type CreateGameDTO struct {
	InternalName string
}

func (api *GameRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	path := ctx.Path()
	if strings.Contains(path, "/vendors/:vendorId") {
		return GetOwnerForVendor(ctx)
	}
	return GetOwnerForGame(ctx)
}

func (api *GameRouter) PublishGame(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("gameId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	if err := api.eventBus.PublishGameChanges(gameId); err != nil {
		return orm.NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Can't publish game changes"))
	}

	return ctx.NoContent(http.StatusOK)
}

func (api *GameRouter) GetList(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
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
	internalName := ctx.QueryParam("internalName")
	genre := ctx.QueryParam("genre")
	releaseDate := ctx.QueryParam("releaseDate")
	sort := ctx.QueryParam("sort")
	userId, err := api.getUserId(ctx)
	if err != nil {
		return err
	}

	var dto []ShortGameInfoDTO
	qilinCtx := ctx.(rbac_echo.AppContext)
	shouldBreak := false
	localOffset := offset

	//CURSOR solution
	for len(dto) <= limit && shouldBreak == false {
		localLimit := limit - len(dto)

		games, err := api.gameService.GetList(userId, vendorId, localOffset, localLimit, internalName, genre, releaseDate, sort)
		if err != nil {
			return err
		}

		// we do not have enough items in DB
		shouldBreak = len(games) <= localLimit

		for _, game := range games {
			owner, err := qilinCtx.GetOwnerForGame(game.ID)
			if err != nil {
				return err
			}

			// filter games that user do not have rights
			if qilinCtx.CheckPermissions(userId, model.VendorDomain, model.GameType, game.ID.String(), owner, "read") != nil {
				continue
			}

			dto = append(dto, ShortGameInfoDTO{
				ID:           game.Game.ID,
				InternalName: game.InternalName,
				Icon:         "",
				Genres: GameGenreDTO{
					Main:     game.GenreMain,
					Addition: game.GenreAddition,
				},
				ReleaseDate: game.ReleaseDate,
			})
		}
		localOffset = localOffset + len(games)
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (api *GameRouter) Create(ctx echo.Context) error {
	params := CreateGameDTO{}
	err := ctx.Bind(&params)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Wrong parameters in body")
	}
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid vendorId")
	}
	userId, err := api.getUserId(ctx)
	if err != nil {
		return err
	}
	game, err := api.gameService.Create(userId, vendorId, params.InternalName)
	if err != nil {
		return err
	}
	dto := mapGameInfo(game)
	return ctx.JSON(http.StatusCreated, dto)
}

func (api *GameRouter) GetInfo(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("gameId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	game, err := api.gameService.GetInfo(gameId)
	if err != nil {
		return err
	}
	dto := mapGameInfo(game)
	return ctx.JSON(http.StatusOK, dto)
}

func (api *GameRouter) Delete(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("gameId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}
	userId, err := api.getUserId(ctx)
	if err != nil {
		return err
	}
	err = api.gameService.Delete(userId, gameId)
	if err != nil {
		return err
	}

	err = api.eventBus.PublishGameDelete(gameId)
	if err != nil {
		zap.L().Error("Error during publishing game changes.", zap.Error(err))
	}

	return ctx.JSON(http.StatusOK, "OK")
}

func (api *GameRouter) UpdateInfo(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("gameId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	dto := &UpdateGameDTO{}
	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}
	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}
	game := mapGameInfoBTO(dto)
	game.ID = gameId
	err = api.gameService.UpdateInfo(game)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (api *GameRouter) GetDescr(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("gameId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid game Id")
	}

	descr, err := api.gameService.GetDescr(gameId)
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

	descrAcc := reflect.ValueOf(&dto.Description)
	descrVal := descrAcc.Elem()
	for i := 0; i < descrVal.NumField(); i++ {
		markdown := descrVal.Field(i).String()
		html := blackfriday.Run([]byte(markdown))
		descrVal.Field(i).SetString(string(html))
	}

	return ctx.JSON(http.StatusOK, dto)
}

func (api *GameRouter) UpdateDescr(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("gameId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid game Id")
	}

	dto := &GameDescrDTO{}
	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
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

	// Converts from html to markdown
	descrAcc := reflect.ValueOf(&dto.Description)
	descrVal := descrAcc.Elem()
	for i := 0; i < descrVal.NumField(); i++ {
		field := descrVal.Field(i)
		html := field.String()
		if html != "" {
			safe_html := bluemonday.UGCPolicy().SanitizeBytes([]byte(html))
			html = string(safe_html)
		}
		field.SetString(html)
		markdown := html2md.Convert(html)
		field.SetString(markdown)
	}

	err = api.gameService.UpdateDescr(&model.GameDescr{
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

func (api *GameRouter) GetGenres(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	offset, _ := strconv.Atoi(ctx.QueryParam("offset"))
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 20
	}
	userId, err := api.getUserId(ctx)
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

func (api *GameRouter) GetTags(ctx echo.Context) (err error) {
	title := ctx.QueryParam("title")
	offset, _ := strconv.Atoi(ctx.QueryParam("offset"))
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = 20
	}
	userId, err := api.getUserId(ctx)
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

func (api *GameRouter) GetRatingDescriptors(ctx echo.Context) error {
	system := ctx.QueryParam("title")
	descriptors, err := api.gameService.GetRatingDescriptors(system)
	if err != nil {
		return err
	}
	dto := []RatingDescriptorDTO{}
	for _, desc := range descriptors {
		dto = append(dto, RatingDescriptorDTO{
			Id:     desc.ID,
			Title:  desc.Title,
			System: desc.System,
		})
	}
	return ctx.JSON(http.StatusOK, dto)
}

func (api *GameRouter) getUserId(ctx echo.Context) (string, error) {
	extUserId, err := context.GetAuthUserId(ctx)
	return extUserId, err
}

func (api *GameRouter) GetPackages(ctx echo.Context) error {
	gameId, err := uuid.FromString(ctx.Param("gameId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}
	packages, err := api.productService.GetPackages(gameId)
	if err != nil {
		return err
	}
	dto := []*packageItemDTO{}
	for _, pkg := range packages {
		dto = append(dto, mapPackageItemDto(&pkg))
	}
	return ctx.JSON(http.StatusOK, dto)
}
