package gee

import "fmt"

type router struct {
	handlers map[string]Handler
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]Handler),
	}
}

func (r *router) addRoute(method, pattern string, handler Handler) {
	key := fmt.Sprintf("%s_%s", method, pattern)
	r.handlers[key] = handler
}

func (r *router) handle(c *Context) {
	key := fmt.Sprintf("%s_%s", c.method, c.path)
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		fmt.Fprintf(c.w, "404 NOT FOUND: %s\n", c.r.URL)
	}
}
