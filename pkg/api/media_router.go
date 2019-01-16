package api

import (
	"net/http"
	"qilin-api/pkg/model"
	"github.com/satori/go.uuid"
	"github.com/fatih/structs"
	"github.com/labstack/echo"
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
func InitMediaRouter(api *Server, service model.MediaService) error {
	mediaRouter := MediaRouter{
		mediaService: service,
	}
	router := api.Router.Group("/games/:id")
	router.GET("/media", mediaRouter.get)
	router.PUT("/media", mediaRouter.put)

	return nil
}

func (api *MediaRouter) put(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	mediaDto := new(Media)
	if err := ctx.Bind(mediaDto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	media := new(model.Media)
	media.ID = id
	media.CoverImage = structs.Map(mediaDto.CoverImage)

	if err := api.mediaService.Update(id, media); err != nil {
		return err
	}

	return ctx.String(http.StatusOK, "")
}

func (api *MediaRouter) get(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	
	media, err := api.mediaService.Get(id)

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, media)
}
