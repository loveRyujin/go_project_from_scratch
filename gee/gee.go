package gee

import (
	"net/http"
)

type Handler func(c *Context)

type Engine struct {
	router *router
}

func New() *Engine {
	return &Engine{
		router: newRouter(),
	}
}

func (e *Engine) addRoute(method, pattern string, handler Handler) {
	e.router.addRoute(method, pattern, handler)
}

func (e *Engine) Get(pattern string, handler Handler) {
	e.addRoute("GET", pattern, handler)
}

func (e *Engine) Post(pattern string, handler Handler) {
	e.addRoute("POST", pattern, handler)
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	e.router.handle(c)
}
