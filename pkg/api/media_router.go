package api

import (
	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
	"time"
)

type (
	//MediaRouter is group struct
	MediaRouter struct {
		mediaService model.MediaService
	}

	//Media is DTO object with full information about media for game
	Media struct {

		// localized cover image of game
		CoverImage *utils.LocalizedString `json:"coverImage" validate:"required"`

		// localized cover video of game
		CoverVideo *utils.LocalizedString `json:"coverVideo" validate:"required"`

		// localized trailer video of game
		Trailers *utils.LocalizedStringArray `json:"trailers" validate:"required"`

		// localized screenshots video of game
		Screenshots *utils.LocalizedStringArray `json:"screenshots" validate:"required"`

		// localized cover video of game
		Store *Store `json:"store" validate:"required,dive"`

		Capsule *Capsule `json:"capsule" validate:"required,dive"`
	}

	//Capsule is DTO object with information about capsule media for game
	Capsule struct {
		Generic *utils.LocalizedString `json:"generic" validate:"required"`

		Small *utils.LocalizedString `json:"small" validate:"required"`
	}

	//Store is DTO object with information about store media for game
	Store struct {
		Special *utils.LocalizedString `json:"special" validate:"required"`

		Friends *utils.LocalizedString `json:"friends" validate:"required"`
	}
)

//InitMediaRouter is initializing group method
func InitMediaRouter(group *echo.Group, service model.MediaService) (*MediaRouter, error) {
	mediaRouter := MediaRouter{
		mediaService: service,
	}

	router := rbac_echo.Group(group, "/games/:gameId", &mediaRouter, []string{"gameId", model.GameType, model.VendorDomain})
	router.GET("/media", mediaRouter.get, nil)
	router.PUT("/media", mediaRouter.put, nil)

	return &mediaRouter, nil
}

func (api *MediaRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	return GetOwnerForGame(ctx)
}

// @Summary Change media for game
// @Description Change media data about game
// @Success 200 {object} "OK"
// @Failure 401 {object} "Unauthorized"
// @Failure 403 {object} "Forbidden"
// @Failure 404 {object} "Not found"
// @Failure 422 {object} "Unprocessable object"
// @Failure 500 {object} "Internal server error"
// @GameRouter /api/v1/games/:gameId/media [put]
func (api *MediaRouter) put(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("gameId"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	mediaDto := new(Media)

	if err := ctx.Bind(mediaDto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(mediaDto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	media := model.Media{}
	err = mapper.Map(mediaDto, &media)

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	media.UpdatedAt = time.Now()

	if err := api.mediaService.Update(id, &media); err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "")
}

// @Summary Get media for game
// @Description Get media data about game
// @Success 200 {object} Media "OK"
// @Failure 401 {object} "Unauthorized"
// @Failure 403 {object} "Forbidden"
// @Failure 404 {object} "Not found"
// @Failure 500 {object} "Internal server error"
// @GameRouter /api/v1/games/:gameId/media [get]
func (api *MediaRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("gameId"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	media, err := api.mediaService.Get(id)

	if err != nil {
		return err
	}

	result := Media{}
	err = mapper.Map(media, &result)

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, result)
}
