package gee

import (
	"fmt"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request)

type Engine struct {
	router map[string]Handler
}

func New() *Engine {
	return &Engine{
		router: make(map[string]Handler),
	}
}

func (e *Engine) addRoute(method, pattern string, handler Handler) {
	key := fmt.Sprintf("%s_%s", method, pattern)
	e.router[key] = handler
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
	key := fmt.Sprintf("%s_%s", r.Method, r.URL.Path)
	if handler, ok := e.router[key]; ok {
		handler(w, r)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", r.URL)
	}
}
