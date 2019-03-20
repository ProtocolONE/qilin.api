package api

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/middleware"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strconv"
	"testing"
	"time"
)

type (
	OnboardingClientRouter struct {
		service             *orm.OnboardingService
		notificationService model.NotificationService
		group               *middleware.RbacGroup
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

func testAccess(t *testing.T, path string, vendor string, roel string, owner string, res, id string, result bool) {
	//CREATE USER
	//Add ROLES
	//Echo call
	//
}

func InitClientOnboardingRouter(group *echo.Group, service *orm.OnboardingService, notificationService model.NotificationService) (*OnboardingClientRouter, error) {
	r := &middleware.RbacGroup{}

	router := OnboardingClientRouter{
		service:             service,
		notificationService: notificationService,
		group:               r,
	}

	r = r.Group(group, "/vendors/:id", &router)

	common := []string{"*", model.DocumentsType, model.VendorDomain}
	r.GET("/documents", router.getDocument, common)
	r.PUT("/documents", router.changeDocument, common)
	r.POST("/documents/reviews", router.sendToReview, common)
	r.DELETE("/documents/reviews", router.revokeReview, common)

	r.GET("/messages", router.getNotifications, common)
	r.GET("/messages/:messageId", router.getNotification, []string{"messageId", model.DocumentsType, model.VendorDomain})
	r.PUT("/messages/:messageId/read", router.markAsRead, []string{"messageId", model.DocumentsType, model.VendorDomain})
	r.GET("/messages/short", router.getLastNotifications, common)

	return &router, nil
}

func (r *OnboardingClientRouter) GetOwner(ctx middleware.QilinContext) (string, error) {
	return GetOwnerForVendor(ctx)
}

func (api *OnboardingClientRouter) getLastNotifications(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	notifications, _, err := api.notificationService.GetNotifications(id, 3, 0, "", "")
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
		result[i].HaveMsg = n.Message != ""
	}

	token := api.notificationService.GetUserToken(id)
	if token != "" {
		ctx.Response().Header().Add("X-Centrifugo-Token", token)
	}

	return ctx.JSON(http.StatusOK, result)
}

func (api *OnboardingClientRouter) getNotification(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	messageId, err := uuid.FromString(ctx.Param("messageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}
	notification, err := api.notificationService.GetNotification(vendorId, messageId)
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
	vendorId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	id, err := uuid.FromString(ctx.Param("messageId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	err = api.notificationService.MarkAsRead(vendorId, id)
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

	notifications, count, err := api.notificationService.GetNotifications(id, limit, offset, query, sort)
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

	ctx.Response().Header().Add("X-Items-Count", fmt.Sprintf("%d", count))

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

func (api *OnboardingClientRouter) revokeReview(ctx echo.Context) error {
	id, err := uuid.FromString(ctx.Param("id"))

	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, "Invalid Id")
	}

	if err := api.service.RevokeReviewRequest(id); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, "")
}
