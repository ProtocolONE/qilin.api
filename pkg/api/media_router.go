package api

import (
	"github.com/mitchellh/mapstructure"
	"net/http"
	"qilin-api/pkg/model"
	"github.com/satori/go.uuid"
	"github.com/labstack/echo"
	 maper "gopkg.in/jeevatkm/go-model.v1"
)

//MediaRouter is router struct
type MediaRouter struct {
	mediaService model.MediaService
}

type Media struct {
	
	// localized cover image of game
	CoverImage *model.LocalizedString `json:"coverImage"`

	// localized cover video of game
	CoverVideo *model.LocalizedString `json:"coverVideo"`

	// localized cover video of game
	Trailers *model.LocalizedString `json:"trailers"`

	// localized cover video of game
	Store *Store `json:"store"`

	Capsule *Capsule `json:"capsule"`
}


type Capsule struct {
	Generic *model.LocalizedString `json:"generic"`

	Small *model.LocalizedString `json:"small"`
}

type Store struct {
	Special *model.LocalizedString `json:"special"`

	Friends *model.LocalizedString `json:"friends"`
}

//InitMediaRouter is initializing router method
func InitMediaRouter(group *echo.Group, service model.MediaService) (*MediaRouter, error) {
	mediaRouter := MediaRouter{
		mediaService: service,
	}
	router := group.Group("/games/:id")
	router.GET("/media", mediaRouter.get)
	router.PUT("/media", mediaRouter.put)

	return &mediaRouter, nil
}

// @Summary Change media for game
// @Description Change media data about game
// @Success 200 {object} "OK"
// @Failure 401 {object} "Unauthorized"
// @Failure 403 {object} "Forbidden"
// @Failure 404 {object} "Not found"
// @Failure 422 {object} "Unprocessable object"
// @Failure 500 {object} "Internal server error"
// @Router /api/v1/games/:id/media [put]
func (api *MediaRouter) put(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	mediaDto := new(Media)
	if err := ctx.Bind(mediaDto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	media := model.Media{}
	input, err := maper.Map(mediaDto)
	
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

 	err = mapstructure.Decode(input, &media)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

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
// @Router /api/v1/games/:id/media [get]
func (api *MediaRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	
	media, err := api.mediaService.Get(id)

	if err != nil {
		return err
	}

	result := Media {}
	input, err := maper.Map(media)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = mapstructure.Decode(input, &result)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	
	return ctx.JSON(http.StatusOK, result)
}
