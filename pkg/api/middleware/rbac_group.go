package middleware

import "github.com/labstack/echo"

type RbacGroup struct {
	group *echo.Group
	paths map[string][]string
}

// DELETE implements `Echo#DELETE()` for sub-routes within the Group.
func (g *RbacGroup) DELETE(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.DELETE(path, h, m...)
	g.paths[p.Path] = permissions
	return g
}

// GET implements `Echo#GET()` for sub-routes within the Group.
func (g *RbacGroup) GET(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.GET(path, h, m...)
	g.paths[p.Path] = permissions
	return g
}

// HEAD implements `Echo#HEAD()` for sub-routes within the Group.
func (g *RbacGroup) HEAD(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.HEAD(path, h, m...)
	g.paths[p.Path] = permissions
	return g
}

// OPTIONS implements `Echo#OPTIONS()` for sub-routes within the Group.
func (g *RbacGroup) OPTIONS(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.OPTIONS(path, h, m...)
	g.paths[p.Path] = permissions
	return g
}

// PATCH implements `Echo#PATCH()` for sub-routes within the Group.
func (g *RbacGroup) PATCH(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.PATCH(path, h, m...)
	g.paths[p.Path] = permissions
	return g
}

// POST implements `Echo#POST()` for sub-routes within the Group.
func (g *RbacGroup) POST(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.POST(path, h, m...)
	g.paths[p.Path] = permissions
	return g
}

// PUT implements `Echo#PUT()` for sub-routes within the Group.
func (g *RbacGroup) PUT(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.PUT(path, h, m...)
	g.paths[p.Path] = permissions
	return g
}

func (g *RbacGroup) Group(group *echo.Group, prefix string, router RbacRouter) *RbacGroup {
	g.paths = map[string][]string{}
	g.group = group.Group(prefix, CheckPermissions(router))
	return g
}
