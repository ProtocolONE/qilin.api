package api

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
)

type MembershipRouter struct {
	service model.MembershipService
}


func InitClientMembershipRouter(group *echo.Group) (*MembershipRouter, error) {
	res := &MembershipRouter{}

	route := group.Group("/vendors/:id")
	route.GET("/membership", res.getUsers)

	return res, nil
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

func (api *MembershipRouter) getUser(ctx echo.Context) error {
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

