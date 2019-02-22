package api

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"net/http"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strconv"
	"time"
)

type OnboardingAdminRouter struct {
	service             *orm.AdminOnboardingService
	notificationService model.NotificationService
}

type ChangeStatusRequest struct {
	Message string `json:"message"`
	Status  string `json:"status" validate:"required"`
}

type NotificationRequest struct {
	Message string `json:"message"`
	Title   string `json:"title" validate:"required"`
}

type NotificationDTO struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	IsRead    bool   `json:"isRead"`
}

type ShortNotificationDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	IsRead    bool   `json:"isRead"`
}

func InitAdminOnboardingRouter(group *echo.Group, service *orm.AdminOnboardingService, notificationService model.NotificationService) (*OnboardingAdminRouter, error) {
	router := OnboardingAdminRouter{
		service:             service,
		notificationService: notificationService,
	}
	r := group.Group("/vendors")
	r.GET("/reviews", router.getReviews)
	r.GET("/:id/documents", router.getDocument)
	r.PUT("/:id/documents", router.changeStatus)
	r.POST("/:id/messages", router.sendNotification)
	r.GET("/:id/messages", router.getNotifications)

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

	err = api.service.ChangeStatus(id, status)
	if err != nil {
		return err
	}

	if request.Message != "" {
		_, err := api.notificationService.SendNotification(&model.Notification{Title: request.Message, Message: request.Message, VendorID: id})
		if err != nil {
			zap.L().Error(err.Error())
		}
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

func (api *OnboardingAdminRouter) getNotifications(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	offset := 0
	limit := 20

	if offsetParam := ctx.QueryParam("offset"); offsetParam != "" {
		if num, err := strconv.Atoi(offsetParam); err == nil {
			offset = num
		} else {
			return orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Bad offset"))
		}
	}

	if limitParam := ctx.QueryParam("limit"); limitParam != "" {
		if num, err := strconv.Atoi(limitParam); err == nil {
			limit = num
		} else {
			return orm.NewServiceError(http.StatusBadRequest, errors.Wrapf(err, "Bad limit"))
		}
	}

	query := ctx.QueryParam("query")
	sort := ctx.QueryParam("sort")

	notifications, err := api.notificationService.GetNotifications(id, limit, offset, query, sort)
	if err != nil {
		return err
	}

	var result []NotificationDTO
	err = mapper.Map(notifications, &result)
	if err != nil {
		return orm.NewServiceErrorf(http.StatusInternalServerError, "Can't map to dto %#v", notifications)
	}

	for i, n := range notifications {
		result[i].ID = n.ID.String()
		result[i].CreatedAt = n.CreatedAt.Format(time.RFC3339)
	}

	return ctx.JSON(http.StatusOK, result)
}

func (api *OnboardingAdminRouter) sendNotification(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	request := new(NotificationRequest)
	if err := ctx.Bind(request); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(request); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	notification, err := api.notificationService.SendNotification(&model.Notification{Message: request.Message, Title: request.Title, VendorID: id})
	if err != nil {
		return err
	}

	result := NotificationDTO{}
	err = mapper.Map(notification, &result)
	if err != nil {
		return orm.NewServiceErrorf(http.StatusInternalServerError, "Can't map to DTO `%#v`", notification)
	}

	result.CreatedAt = notification.CreatedAt.Format(time.RFC3339)
	result.ID = notification.ID.String()

	return ctx.JSON(http.StatusOK, result)
}
