package api

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/model/utils"
	"qilin-api/pkg/orm"
)

type (
	RatingsRouter struct {
		service *orm.RatingService
	}

	RatingsDTO struct {
		PEGI CommonRating `json:"PEGI"`
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
)

//InitRatingsRouter is initialization method for router
func InitRatingsRouter(group *echo.Group, service *orm.RatingService) (*RatingsRouter, error) {
	ratingRouter := RatingsRouter{
		service: service,
	}

	r := group.Group("/games/:id")

	r.GET("/ratings", ratingRouter.get)
	r.PUT("/ratings", ratingRouter.put)

	return &ratingRouter, nil
}

func (router *RatingsRouter) get(ctx echo.Context) error {
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Can't decode gameRating from domain to DTO. Error: "+err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (router *RatingsRouter) put(ctx echo.Context) error {
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

//RatingStructLevelValidation is method for custom validation of game ratings
func RatingStructLevelValidation(sl validator.StructLevel) {
	rating := sl.Current().Interface().(RatingsDTO)

	validateCommonRating(rating.BBFC.Rating, "BBFC", utils.StringArray{"U", "PG", "12A", "12", "15", "18", "R18"}, sl)
	validateCommonRating(rating.CERO.Rating, "CERO", utils.StringArray{"A", "B", "C", "D", "Z"}, sl)
	validateCommonRating(rating.ESRB.Rating, "ESRB", utils.StringArray{"EC", "E", "E10+", "T", "M", "A", "RP"}, sl)
	validateCommonRating(rating.PEGI.Rating, "PEGI", utils.StringArray{"3", "7", "12", "16", "18"}, sl)
	validateCommonRating(rating.USK.Rating, "USK", utils.StringArray{"USK", "0", "6", "12", "16", "18"}, sl)
}

func validateCommonRating(field interface{}, fieldName string, values utils.StringArray, sl validator.StructLevel) {
	value := field.(string)
	if len(value) > 0 {
		exist := values.Contains(value)
		if !exist {
			sl.ReportError(field, "Rating", fieldName, "contains", values.String())
		}
	}
}
