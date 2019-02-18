package api

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strconv"
)

type OnboardingAdminRouter struct {
	service *orm.AdminOnboardingService
}

type ChangeStatusRequest struct {
	Message string `json:"message"`
	Status  string `json:"status" validate:"required"`
}

func InitAdminOnboardingRouter(group *echo.Group, service *orm.AdminOnboardingService) (*OnboardingAdminRouter, error) {
	router := OnboardingAdminRouter{
		service: service,
	}
	r := group.Group("/vendors")
	r.GET("/reviews", router.getReviews)
	r.GET("/:id/documents", router.getDocument)
	r.PUT("/:id/documents", router.changeStatus)

	return &router, nil
}

func (api *OnboardingAdminRouter) changeStatus(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad id"))
	}

	request := new(ChangeStatusRequest)

	if err := ctx.Bind(request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(request); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	status, err := model.ReviewStatusFromString(request.Status)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad status"))
	}

	err = api.service.ChangeStatus(id, status, request.Message)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, "")
}

func (api *OnboardingAdminRouter) getDocument(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad id"))
	}

	doc, err := api.service.GetForVendor(id)
	if err != nil {
		return err
	}

	dto := DocumentsInfoResponseDTO{}
	err = mapper.Map(doc, &dto)
	if err != nil {
		return orm.NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Mapping dto error"))
	}
	dto.Status = doc.ReviewStatus.ToString()

	return ctx.JSON(http.StatusOK, dto)
}

func (api *OnboardingAdminRouter) getReviews(ctx echo.Context) error {
	offset := 0
	limit := 20
	status := model.ReviewUndefined

	if offsetParam := ctx.QueryParam("offset"); offsetParam != "" {
		if num, err := strconv.Atoi(offsetParam); err == nil {
			offset = num
		} else {
			return orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Bad limit"))
		}
	}

	if limitParam := ctx.QueryParam("limit"); limitParam != "" {
		if num, err := strconv.Atoi(limitParam); err == nil {
			limit = num
		} else {
			return orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Bad limit"))
		}
	}

	name := ctx.QueryParam("name")
	status, err := model.ReviewStatusFromString(ctx.QueryParam("status"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Bad status"))
	}

	sort := ctx.QueryParam("sort")
	requests, err := api.service.GetRequests(limit, offset, name, status, sort)

	if err != nil {
		return err
	}

	var dto []DocumentsInfoResponseDTO
	err = mapper.Map(requests, &dto)
	if err != nil {
		return orm.NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Dto mapping error"))
	}

	for i, doc := range requests {
		dto[i].Status = doc.ReviewStatus.ToString()
	}

	if dto == nil {
		dto = make([]DocumentsInfoResponseDTO, 0)
	}

	return ctx.JSON(http.StatusOK, dto)
}
