package middleware

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/ProtocolONE/rbac"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
)

type QilinContext struct {
	echo.Context
	enf *rbac.Enforcer
	db  *orm.Database
}

type RbacPathPermission struct {
	Type            model.ResourceType
	Domain          model.Domain
	ResourceIDQuery string
}

func (c *QilinContext) CheckPermissions(userId, domain, resource, resourceId, owner, action string) error {
	ctx := rbac.Context{
		Domain:        domain,
		User:          userId,
		ResourceId:    resourceId,
		Resource:      resource,
		ResourceOwner: owner,
		Action:        action,
	}
	if c.enf.Enforce(ctx) == false {
		return orm.NewServiceErrorf(http.StatusForbidden, "Enforce failed for user: `%s`, resource `%s` with id `%s` and action `%s` in domain `%s`", userId, resource, resourceId, action, domain)
	}
	return nil
}

func (c *QilinContext) GetOwnerForVendor(uuid uuid.UUID) (uuid.UUID, error) {
	return orm.GetOwnerForVendor(c.db.DB(), uuid)
}

func (c *QilinContext) GetUserIdByExternal(id string) (uuid.UUID, error) {
	return orm.GetUserId(c.db.DB(), id)
}

func CheckPermissions(router RbacRouter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			qilinCtx := c.(QilinContext)

			owner, err := router.GetOwner(qilinCtx)
			if err != nil {
				return err
			}

			id, err := context.GetAuthExternalUserId(c)
			if err != nil {
				return err
			}

			paths := router.GetPermissionsMap()
			perm, ok := paths[c.Path()]
			if !ok {
				return orm.NewServiceError(http.StatusForbidden, "")
			}

			resourceId := "*"
			if perm[0] != "*" {
				resourceId = c.Param(perm[0])
			}

			userId, err := qilinCtx.GetUserIdByExternal(id)

			action := "any"
			switch c.Request().Method {
			case echo.GET:
				action = "read"
			case echo.PUT:
			case echo.POST:
			case echo.PATCH:
			case echo.DELETE:
				action = "write"
			}

			err = qilinCtx.CheckPermissions(userId.String(), perm[2], perm[1], resourceId, owner, action)
			if err != nil {
				return err
			}

			return next(c)
		}
	}
}

func QilinContextMiddleware(db *orm.Database, enf *rbac.Enforcer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			context := QilinContext{
				enf:     enf,
				db:      db,
				Context: c,
			}
			return next(context)
		}
	}
}
