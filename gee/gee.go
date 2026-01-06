package gee

import (
	"net/http"
	"strings"
)

type Handler func(c *Context)

type Engine struct {
	*RouteGroup
	router *router
	groups []*RouteGroup
}

func New() *Engine {
	e := &Engine{router: newRouter()}
	e.RouteGroup = &RouteGroup{
		prefix:   "",
		handlers: nil,
		engine:   e,
	}
	e.groups = append(e.groups, e.RouteGroup)

	return e
}

func Default() *Engine {
	e := &Engine{router: newRouter()}
	e.RouteGroup = &RouteGroup{
		prefix:   "",
		handlers: HandlerChain{Recovery()},
		engine:   e,
	}
	e.groups = append(e.groups, e.RouteGroup)

	return e
}

func (e *Engine) addRoute(method, pattern string, handler Handler) {
	e.router.addRoute(method, pattern, handler)
}

func (e *Engine) GET(pattern string, handler Handler) {
	e.addRoute("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler Handler) {
	e.addRoute("POST", pattern, handler)
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers HandlerChain
	for _, group := range e.groups {
		if group != nil && strings.HasPrefix(r.URL.Path, group.prefix) {
			handlers = append(handlers, group.handlers...)
		}
	}

	c := newContext(w, r)
	c.handlers = handlers
	e.router.handle(c)
}

type HandlerChain []Handler

type RouteGroup struct {
	prefix   string
	handlers HandlerChain
	engine   *Engine
}

func (g *RouteGroup) Group(prefix string) *RouteGroup {
	e := g.engine
	newGroup := &RouteGroup{
		prefix:   g.prefix + prefix,
		handlers: nil,
		engine:   e,
	}
	e.groups = append(e.groups, newGroup)

	return newGroup
}

func (g *RouteGroup) Use(handlers ...Handler) {
	g.handlers = append(g.handlers, handlers...)
}

func (g *RouteGroup) addRoute(method, pattern string, handler Handler) {
	pattern = g.prefix + pattern
	g.engine.router.addRoute(method, pattern, handler)
}

func (g *RouteGroup) GET(pattern string, handler Handler) {
	g.addRoute("GET", pattern, handler)
}

func (g *RouteGroup) POST(pattern string, handler Handler) {
	g.addRoute("POST", pattern, handler)
}
