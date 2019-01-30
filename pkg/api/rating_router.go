package api

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
)

type (
	RatingsRouter struct {
		service *orm.RatingService
	}

	RatingsDTO struct {
		PEGI PEGI         `json:"PEGI"`
		ESRB CommonRating `json:"ESRB"`
		BBFC CommonRating `json:"BBFC"`
		USK  CommonRating `json:"USK"`
		CERO CommonRating `json:"CERO"`
	}

	CommonRating struct {
		DisplayOnlineNotice bool    `json:"displayOnlineNotice"`
		ShowAgeRestrict     bool    `json:"showAgeRestrict"`
		AgeRestrict         int8    `json:"ageRestrict"`
		Descriptors         []int32 `json:"descriptors"`
		Rating              string  `json:"rating"`
	}

	PEGI struct {
		DisplayOnlineNotice bool    `json:"displayOnlineNotice"`
		ShowAgeRestrict     bool    `json:"showAgeRestrict"`
		AgeRestrict         int8    `json:"ageRestrict"`
		Descriptors         []int32 `json:"descriptors"`
		Rating              int8    `json:"rating"`
	}
)

//InitRatingsRouter is initialization method for router
func InitRatingsRouter(group *echo.Group, service *orm.RatingService) (*RatingsRouter, error) {
	ratingRouter := RatingsRouter{
		service: service,
	}

	r := group.Group("/games/:id")

	r.GET("/ratings", ratingRouter.getRatings)
	r.POST("/ratings", ratingRouter.postRatings)

	return &ratingRouter, nil
}

func (router *RatingsRouter) getRatings(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	gameRating, err := router.service.GetRatingsForGame(id)

	if err != nil {
		return err
	}

	result := RatingsDTO{}
	err = mapper.Map(gameRating, &result)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Can't decode gameRating from domain to DTO. Error: "+err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (router *RatingsRouter) postRatings(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}
	dto := new(RatingsDTO)

	if err := ctx.Bind(dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	result := model.GameRating{}
	err = mapper.Map(dto, &result)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := router.service.SaveRatingsForGame(id, &result); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, "")
}
