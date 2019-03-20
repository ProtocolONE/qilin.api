package api

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/middleware"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
)

type MembershipRouter struct {
	service model.MembershipService
}

type ChangeUserRolesDTO struct {
	Added   []UserRoleDTO `json:"added"`
	Removed []UserRoleDTO `json:"removed"`
}

type UserRoleDTO struct {
	Id    string   `json:"id"`
	Roles []string `json:"roles"`
}

func InitClientMembershipRouter(group *echo.Group, service model.MembershipService) (*MembershipRouter, error) {
	res := &MembershipRouter{
		service: service,
	}

	route := &middleware.RbacGroup{}
	route = route.Group(group, "/vendors/:id", res)
	route.GET("/memberships", res.getUsers, []string{"*", model.RolesType, model.VendorDomain})
	route.PUT("/memberships/:userId", res.changeUserRoles, []string{"userId", model.RolesType, model.VendorDomain})

	return res, nil
}

func (api *MembershipRouter) GetOwner(ctx middleware.QilinContext) (string, error) {
	return GetOwnerForVendor(ctx)
}

func (api *MembershipRouter) getUsers(ctx echo.Context) error {
	vendorIdParam := ctx.Param("id")
	vendorId, err := uuid.FromString(vendorIdParam)
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad vendor id"))
	}

	users, err := api.service.GetUsers(vendorId)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, users)
}

func (api *MembershipRouter) changeUserRoles(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad vendor id"))
	}

	userId := ctx.Param("userId")
	if userId == "" {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad user id"))
	}

	dto := new(ChangeUserRolesDTO)

	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, err)
	}

	if errs := ctx.Validate(dto); errs != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errs)
	}

	for _, remove := range dto.Removed {
		for _, role := range remove.Roles {
			err = api.service.RemoveRoleToUserInGame(vendorId, userId, remove.Id, role)
			if err != nil {
				return err
			}
		}
	}

	for _, remove := range dto.Added {
		for _, role := range remove.Roles {
			err = api.service.AddRoleToUserInGame(vendorId, userId, remove.Id, role)
			if err != nil {
				return err
			}
		}
	}

	user, err := api.service.GetUser(vendorId, userId)

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, user)
}

func (api *MembershipRouter) getUser(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad vendor id"))
	}

	userId := ctx.Param("id")
	if userId == "" {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad user id"))
	}

	userRole, err := api.service.GetUser(vendorId, userId)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, userRole)
}
