package gee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter(t *testing.T) {
	r := newRouter()

	if r == nil {
		t.Fatal("newRouter() returned nil")
	}

	if r.handlers == nil {
		t.Fatal("newRouter() handlers is nil")
	}

	if len(r.handlers) != 0 {
		t.Errorf("Expected empty handlers map, got %d handlers", len(r.handlers))
	}
}

func TestRouter_addRoute(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		pattern      string
		expectedKey  string
	}{
		{
			name:        "add GET route",
			method:      "GET",
			pattern:     "/hello",
			expectedKey: "GET_/hello",
		},
		{
			name:        "add POST route",
			method:      "POST",
			pattern:     "/users",
			expectedKey: "POST_/users",
		},
		{
			name:        "add route with different path",
			method:      "GET",
			pattern:     "/api/users",
			expectedKey: "GET_/api/users",
		},
		{
			name:        "add route with empty path",
			method:      "GET",
			pattern:     "/",
			expectedKey: "GET_/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter()

			handler := func(c *Context) {
				c.Status(http.StatusOK)
				c.w.Write([]byte("test"))
			}

			r.addRoute(tt.method, tt.pattern, handler)

			if _, ok := r.handlers[tt.expectedKey]; !ok {
				t.Errorf("Route %s not found in handlers", tt.expectedKey)
			}

			if len(r.handlers) != 1 {
				t.Errorf("Expected 1 handler, got %d", len(r.handlers))
			}
		})
	}
}

func TestRouter_handle(t *testing.T) {
	tests := []struct {
		name           string
		setupHandlers  func(*router)
		requestMethod  string
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful handler execution",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/hello", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("Hello, World!"))
				})
			},
			requestMethod:  "GET",
			requestPath:    "/hello",
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
		},
		{
			name: "handler not found",
			setupHandlers: func(r *router) {
				// 不添加任何 handler
			},
			requestMethod:  "GET",
			requestPath:    "/notfound",
			expectedStatus: http.StatusOK, // 当前实现返回 200
			expectedBody:   "404 NOT FOUND: /notfound\n",
		},
		{
			name: "method mismatch",
			setupHandlers: func(r *router) {
				r.addRoute("POST", "/users", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("POST handler"))
				})
			},
			requestMethod:  "GET",
			requestPath:    "/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "404 NOT FOUND: /users\n",
		},
		{
			name: "path mismatch",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/hello", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("Hello"))
				})
			},
			requestMethod:  "GET",
			requestPath:    "/world",
			expectedStatus: http.StatusOK,
			expectedBody:   "404 NOT FOUND: /world\n",
		},
		{
			name: "multiple handlers same path different methods",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/api/users", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("GET response"))
				})
				r.addRoute("POST", "/api/users", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("POST response"))
				})
			},
			requestMethod:  "GET",
			requestPath:    "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "GET response",
		},
		{
			name: "multiple handlers same path different methods - POST",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/api/users", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("GET response"))
				})
				r.addRoute("POST", "/api/users", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("POST response"))
				})
			},
			requestMethod:  "POST",
			requestPath:    "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "POST response",
		},
		{
			name: "empty path handler",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("root"))
				})
			},
			requestMethod:  "GET",
			requestPath:    "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter()
			tt.setupHandlers(r)

			req := httptest.NewRequest(tt.requestMethod, tt.requestPath, nil)
			rr := httptest.NewRecorder()
			c := newContext(rr, req)

			r.handle(c)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected response body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

