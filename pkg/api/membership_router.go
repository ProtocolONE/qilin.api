package api

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/rbac_echo"
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

	permissions := []string{"*", model.RolesType, model.VendorDomain}
	route := rbac_echo.Group(group, "/vendors/:vendorId", res, permissions)

	route.GET("/memberships", res.getUsers, nil)
	route.GET("/memberships/:userId", res.getUser, nil)
	route.PUT("/memberships/:userId", res.changeUserRoles, nil)
	route.GET("/memberships/:userId/permissions", res.getUserPermissions, nil)

	//TODO: Hack. Remove after needed functionality implemented
	group.POST("/to_delete/:userId/grantAdmin", res.addAdminRole)

	return res, nil
}

//TODO: Hack. Remove after needed functionality implemented
func (api *MembershipRouter) addAdminRole(ctx echo.Context) error {
	userId := ctx.Param("userId")
	if err := api.service.AddRoleToUser(userId, "*", model.SuperAdmin); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (api *MembershipRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	return GetOwnerForVendor(ctx)
}

func (api *MembershipRouter) getUsers(ctx echo.Context) error {
	vendorIdParam := ctx.Param("vendorId")
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
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
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
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad vendor id"))
	}

	userId := ctx.Param("userId")
	if userId == "" {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad user id"))
	}

	userRole, err := api.service.GetUser(vendorId, userId)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, userRole)
}

func (api *MembershipRouter) getUserPermissions(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad vendor id"))
	}

	userId := ctx.Param("userId")
	if userId == "" {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad user id"))
	}

	permissions, err := api.service.GetUserPermissions(vendorId, userId)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, permissions)
}
