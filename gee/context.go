package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]any

type Context struct {
	w          http.ResponseWriter
	r          *http.Request
	method     string
	path       string
	statusCode int
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		w:      w,
		r:      r,
		method: r.Method,
		path:   r.URL.Path,
	}
}

func (c *Context) PostForm(key string) string {
	return c.r.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.r.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.statusCode = code
	c.w.WriteHeader(code)
}

func (c *Context) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

func (c *Context) JSON(code int, obj any) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.w)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.w, err.Error(), code)
	}
}

func (c *Context) String(code int, format string, args ...any) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.w.Write([]byte(fmt.Sprintf(format, args...)))
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.w.Write(data)
}
