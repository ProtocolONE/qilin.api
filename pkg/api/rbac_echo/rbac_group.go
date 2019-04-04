package rbac_echo

import "github.com/labstack/echo/v4"

type RbacGroup struct {
	group       *echo.Group
	router      Router
	paths       map[string][]string
	permissions []string
}

// DELETE implements `Echo#DELETE()` for sub-routes within the Group.
func (g *RbacGroup) DELETE(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.DELETE(path, h, m...)
	if permissions != nil {
		g.paths[p.Path] = permissions
	} else if g.permissions != nil {
		g.paths[p.Path] = g.permissions
	} else {
		panic("Permissions not set")
	}
	return g
}

// GET implements `Echo#GET()` for sub-routes within the Group.
func (g *RbacGroup) GET(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.GET(path, h, m...)
	if permissions != nil {
		g.paths[p.Path] = permissions
	} else if g.permissions != nil {
		g.paths[p.Path] = g.permissions
	} else {
		panic("Permissions not set")
	}
	return g
}

// HEAD implements `Echo#HEAD()` for sub-routes within the Group.
func (g *RbacGroup) HEAD(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.HEAD(path, h, m...)
	if permissions != nil {
		g.paths[p.Path] = permissions
	} else if g.permissions != nil {
		g.paths[p.Path] = g.permissions
	} else {
		panic("Permissions not set")
	}
	return g
}

// OPTIONS implements `Echo#OPTIONS()` for sub-routes within the Group.
func (g *RbacGroup) OPTIONS(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.OPTIONS(path, h, m...)
	if permissions != nil {
		g.paths[p.Path] = permissions
	} else if g.permissions != nil {
		g.paths[p.Path] = g.permissions
	} else {
		panic("Permissions not set")
	}
	return g
}

// PATCH implements `Echo#PATCH()` for sub-routes within the Group.
func (g *RbacGroup) PATCH(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.PATCH(path, h, m...)
	if permissions != nil {
		g.paths[p.Path] = permissions
	} else if g.permissions != nil {
		g.paths[p.Path] = g.permissions
	} else {
		panic("Permissions not set")
	}
	return g
}

// POST implements `Echo#POST()` for sub-routes within the Group.
func (g *RbacGroup) POST(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.POST(path, h, m...)
	if permissions != nil {
		g.paths[p.Path] = permissions
	} else if g.permissions != nil {
		g.paths[p.Path] = g.permissions
	} else {
		panic("Permissions not set")
	}
	return g
}

// PUT implements `Echo#PUT()` for sub-routes within the Group.
func (g *RbacGroup) PUT(path string, h echo.HandlerFunc, permissions []string, m ...echo.MiddlewareFunc) *RbacGroup {
	p := g.group.PUT(path, h, m...)
	if permissions != nil {
		g.paths[p.Path] = permissions
	} else if g.permissions != nil {
		g.paths[p.Path] = g.permissions
	} else {
		panic("Permissions not set")
	}
	return g
}

func Group(group *echo.Group, prefix string, router Router, permissions []string, middleware ...echo.MiddlewareFunc) *RbacGroup {
	g := &RbacGroup{}
	g.paths = map[string][]string{}
	m := make([]echo.MiddlewareFunc, 0)
	m = append(m, CheckPermissions(g, router))
	m = append(m, middleware...)
	g.group = group.Group(prefix, m...)
	g.permissions = permissions
	return g
}
