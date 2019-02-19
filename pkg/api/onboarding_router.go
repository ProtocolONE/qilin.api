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
	OnboardingClientRouter struct {
		service *orm.OnboardingService
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

func InitClientOnboardingRouter(group *echo.Group, service *orm.OnboardingService) (*OnboardingClientRouter, error) {
	router := OnboardingClientRouter{
		service: service,
	}
	r := group.Group("/vendors/:id")
	r.GET("/documents", router.getDocument)
	r.PUT("/documents", router.changeDocument)
	r.POST("/documents/reviews", router.sendToReview)

	return &router, nil
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
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Id")
	}

	document, err := api.service.GetForVendor(id)
	if err != nil {
		return err
	}

	result := DocumentsInfoResponseDTO{}

	if err := mapper.Map(document, &result); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
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
