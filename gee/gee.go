package gee

import (
	"log"
	"net/http"
)

type Handler func(c *Context)

type Engine struct {
	*RouteGroup
	router *router
}

func New() *Engine {
	e := &Engine{router: newRouter()}
	e.RouteGroup = &RouteGroup{
		prefix:   "",
		handlers: nil,
		engine:   e,
	}

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
	c := newContext(w, r)
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
		prefix: g.prefix + prefix,
		engine: e,
	}

	return newGroup
}

func (g *RouteGroup) addRoute(method, pattern string, handler Handler) {
	pattern = g.prefix + pattern
	log.Printf("Route %4s - %s", method, pattern)
	g.engine.router.addRoute(method, pattern, handler)
}

func (g *RouteGroup) GET(pattern string, handler Handler) {
	g.addRoute("GET", pattern, handler)
}

func (g *RouteGroup) POST(pattern string, handler Handler) {
	g.addRoute("POST", pattern, handler)
}
