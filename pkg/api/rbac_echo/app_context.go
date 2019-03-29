package rbac_echo

import (
	"github.com/ProtocolONE/rbac"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"
	"qilin-api/pkg/api/context"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm"
)

type AppContext struct {
	echo.Context
	enf           *rbac.Enforcer
	ownerProvider model.OwnerProvider
}

func (c *AppContext) CheckPermissions(userId, domain, resource, resourceId, owner, action string) error {
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

func (c *AppContext) GetOwnerForGame(uuid uuid.UUID) (string, error) {
	return c.ownerProvider.GetOwnerForGame(uuid)
}

func (c *AppContext) GetOwnerForVendor(uuid uuid.UUID) (string, error) {
	return c.ownerProvider.GetOwnerForVendor(uuid)
}

func CheckPermissions(group *RbacGroup, router Router) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			qilinCtx := c.(AppContext)

			paths := group.paths
			path := c.Path()
			perm, ok := paths[path]
			if !ok {
				return orm.NewServiceErrorf(http.StatusForbidden, "Could not map `%s` in paths", path)
			}

			owner, err := router.GetOwner(qilinCtx)
			if err != nil {
				return err
			}

			userId, err := context.GetAuthUserId(c)
			if err != nil {
				return err
			}

			resourceId := "*"
			if perm[0] != "*" {
				resourceId = c.Param(perm[0])
			}

			action := "any"
			method := c.Request().Method
			switch method {
			case echo.GET:
				action = "read"
			case echo.PUT:
			case echo.POST:
			case echo.PATCH:
			case echo.DELETE:
				action = "write"
			}

			err = qilinCtx.CheckPermissions(userId, perm[2], perm[1], resourceId, owner, action)
			if err != nil {
				return err
			}

			return next(c)
		}
	}
}

func NewAppContextMiddleware(ownerProvider model.OwnerProvider, enf *rbac.Enforcer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			context := AppContext{
				enf:           enf,
				ownerProvider: ownerProvider,
				Context:       c,
			}
			return next(context)
		}
	}
}
