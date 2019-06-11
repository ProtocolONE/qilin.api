package api

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/api/rbac_echo"
	"qilin-api/pkg/mapper"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
	"strings"
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

type InviteDTO struct {
	Email string          `json:"email" validate:"required"`
	Roles []RoleInviteDTO `json:"roles" validate:"required,dive"`
}

type RoleInviteDTO struct {
	Role     string            `json:"role" validate:"required,non_admin_role"`
	Resource InviteResourceDTO `json:"resource" validate:"required,dive"`
}

type InviteResourceDTO struct {
	Id     string `json:"id" validate:"required"`
	Domain string `json:"domain" validate:"required"`
}

type InviteCreatedDTO struct {
	Id  string `json:"id"`
	Url string `json:"url"`
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
	route.GET("/memberships/:userId/permissions", res.getUserPermissions, []string{"*", model.RoleUserType, model.VendorDomain})

	route.POST("/memberships/invites", res.sendInvite, []string{"*", model.InvitesType, model.VendorDomain})
	route.PUT("/memberships/invites/:inviteId", res.acceptInvite, []string{"*", model.InvitesType, model.VendorDomain})

	//TODO: Hack. Remove after needed functionality implemented
	group.POST("/to_delete/:userId/grantAdmin", res.addAdminRole)
	group.POST("/to_delete/:userId/dropAdmin", res.dropAdminRole)

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

func (api *MembershipRouter) dropAdminRole(ctx echo.Context) error {
	userId := ctx.Param("userId")
	if err := api.service.RemoveUserRole(userId, "*", model.SuperAdmin); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (api *MembershipRouter) GetOwner(ctx rbac_echo.AppContext) (string, error) {
	//HACK: we should skip checking rights for accepting invite and returning self as owner of resource for pass
	if strings.Contains(ctx.Path(), "/memberships/invites/:inviteId") && ctx.Request().Method == http.MethodPut {
		return context.GetAuthUserId(ctx)
	}

	return GetOwnerForVendor(ctx)
}

func (api *MembershipRouter) acceptInvite(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad vendor id"))
	}

	inviteId, err := uuid.FromString(ctx.Param("inviteId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad invite id"))
	}

	userId, err := context.GetAuthUserId(ctx)

	if err != nil {
		return err
	}

	err = api.service.AcceptInvite(vendorId, inviteId, userId)
	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (api *MembershipRouter) sendInvite(ctx echo.Context) error {
	vendorId, err := uuid.FromString(ctx.Param("vendorId"))
	if err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Bad vendor id"))
	}

	dto := &InviteDTO{}
	if err := ctx.Bind(dto); err != nil {
		return orm.NewServiceError(http.StatusBadRequest, errors.Wrap(err, "Binding to dto"))
	}

	if err := ctx.Validate(dto); err != nil {
		return orm.NewServiceError(http.StatusUnprocessableEntity, errors.Wrap(err, "Validation failed"))
	}

	invite := model.Invite{}
	if err := mapper.Map(dto, &invite); err != nil {
		return orm.NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Mapping from dto to model failed"))
	}

	result, err := api.service.SendInvite(vendorId, invite)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, &InviteCreatedDTO{Id: result.Id, Url: result.Url})
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

	currentUserId, err := context.GetAuthUserId(ctx)
	if currentUserId != userId {
		return orm.NewServiceErrorf(http.StatusForbidden, "User %s can't see permissions for user %s", currentUserId, userId)
	}

	permissions, err := api.service.GetUserPermissions(vendorId, userId)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, permissions)
}
