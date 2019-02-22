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
	"time"
)

type (
	OnboardingClientRouter struct {
		service             *orm.OnboardingService
		notificationService model.NotificationService
	}

	ContactDTO struct {
		Authorized AuthorizedDTO `json:"authorized" validate:"required,dive"`
		Technical  TechnicalDTO  `json:"technical" validate:"dive"`
	}

	TechnicalDTO struct {
		FullName string `json:"fullName"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}

	AuthorizedDTO struct {
		FullName string `json:"fullName" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Phone    string `json:"phone" validate:"required"`
		Position string `json:"position" validate:"required"`
	}

	BankingDTO struct {
		Currency      string `json:"currency" validate:"required,is_currency"`
		Name          string `json:"name" validate:"required"`
		Address       string `json:"address" validate:"required"`
		AccountNumber string `json:"accountNumber" validate:"required"`
		Swift         string `json:"swift" validate:"required"`
		Details       string `json:"details"`
	}

	CompanyDTO struct {
		Name               string `json:"name" validate:"required"`
		AlternativeName    string `json:"alternativeName"`
		Country            string `json:"country" validate:"required"`
		Region             string `json:"region" validate:"required"`
		Zip                string `json:"zip" validate:"required"`
		City               string `json:"city" validate:"required"`
		Address            string `json:"address" validate:"required"`
		AdditionalAddress  string `json:"additionalAddress"`
		RegistrationNumber string `json:"registrationNumber" validate:"required"`
		TaxId              string `json:"taxId" validate:"required"`
	}

	DocumentsInfoDTO struct {
		Company CompanyDTO `json:"company" validate:"required,dive"`
		Contact ContactDTO `json:"contact" validate:"required,dive"`
		Banking BankingDTO `json:"banking" validate:"required,dive"`
	}

	DocumentsInfoResponseDTO struct {
		Company CompanyDTO `json:"company" validate:"required,dive"`
		Contact ContactDTO `json:"contact" validate:"required,dive"`
		Banking BankingDTO `json:"banking" validate:"required,dive"`
		Status  string     `json:"status"`
	}
)

func InitClientOnboardingRouter(group *echo.Group, service *orm.OnboardingService, notificationService model.NotificationService) (*OnboardingClientRouter, error) {
	router := OnboardingClientRouter{
		service:             service,
		notificationService: notificationService,
	}
	r := group.Group("/vendors/:id")
	r.GET("/documents", router.getDocument)
	r.PUT("/documents", router.changeDocument)
	r.POST("/documents/reviews", router.sendToReview)
	r.GET("/messages", router.getNotifications)
	r.GET("/messages/:messageId", router.getNotification)
	r.PUT("/messages/:messageId/read", router.markAsRead)
	r.GET("/messages/short", router.getLastNotifications)

	return &router, nil
}

func (api *OnboardingClientRouter) getLastNotifications(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	notifications, err := api.notificationService.GetNotifications(id, 3, 0, "", "")
	if err != nil {
		return err
	}

	var result []ShortNotificationDTO
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

func (api *OnboardingClientRouter) getNotification(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("messageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}
	notification, err := api.notificationService.GetNotification(id)
	if err != nil {
		return err
	}

	dto := NotificationDTO{}
	err = mapper.Map(notification, &dto)
	if err != nil {
		return orm.NewServiceError(http.StatusInternalServerError, err)
	}
	dto.ID = notification.ID.String()
	dto.CreatedAt = notification.CreatedAt.Format(time.RFC3339)

	return ctx.JSON(http.StatusOK, dto)
}

func (api *OnboardingClientRouter) markAsRead(ctx echo.Context) error {
	_, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	id, err := uuid.FromString(ctx.Param("messageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	err = api.notificationService.MarkAsRead(id)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, "")
}

func (api *OnboardingClientRouter) getNotifications(ctx echo.Context) error {
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

func (api *OnboardingClientRouter) changeDocument(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	dto := new(DocumentsInfoDTO)

	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	document := model.DocumentsInfo{}
	if err := mapper.Map(dto, &document); err != nil {
		return orm.NewServiceError(http.StatusInternalServerError, err)
	}

	document.VendorID = id
	if err := api.service.ChangeDocument(&document); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, "")
}

func (api *OnboardingClientRouter) getDocument(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	document, err := api.service.GetForVendor(id)
	if err != nil {
		return err
	}

	result := DocumentsInfoResponseDTO{}

	if err := mapper.Map(document, &result); err != nil {
		return orm.NewServiceError(http.StatusInternalServerError, err)
	}

	result.Status = document.Status.ToString()

	return ctx.JSON(http.StatusOK, result)
}

func (api *OnboardingClientRouter) sendToReview(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	if err := api.service.SendToReview(id); err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, "")
}
