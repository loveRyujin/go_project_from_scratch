package gee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEngine_New(t *testing.T) {
	newEngine := New()

	if newEngine == nil {
		t.Fatal("New() returned nil")
	}

	if newEngine.router == nil {
		t.Fatal("New() router is nil")
	}

	if len(newEngine.router.handlers) != 0 {
		t.Errorf("Expected empty router, got %d routes", len(newEngine.router.handlers))
	}
}

func TestEngine_addRoute(t *testing.T) {
	e := New()

	handler := func(c *Context) {
		c.Status(http.StatusOK)
		c.w.Write([]byte("test"))
	}

	e.addRoute("GET", "/test", handler)

	key := "GET_/test"
	if _, ok := e.router.handlers[key]; !ok {
		t.Errorf("Route %s not found in router", key)
	}

	if len(e.router.handlers) != 1 {
		t.Errorf("Expected 1 route, got %d", len(e.router.handlers))
	}
}

func TestEngine_RegisterRoute(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		registerFunc func(*Engine, string, Handler)
		expectedKey  string
	}{
		{
			name:         "register GET route",
			method:       "GET",
			path:         "/hello",
			registerFunc: func(e *Engine, path string, handler Handler) { e.Get(path, handler) },
			expectedKey:  "GET_/hello",
		},
		{
			name:         "register POST route",
			method:       "POST",
			path:         "/users",
			registerFunc: func(e *Engine, path string, handler Handler) { e.Post(path, handler) },
			expectedKey:  "POST_/users",
		},
		{
			name:         "register GET route with different path",
			method:       "GET",
			path:         "/api/users",
			registerFunc: func(e *Engine, path string, handler Handler) { e.Get(path, handler) },
			expectedKey:  "GET_/api/users",
		},
		{
			name:         "register POST route with different path",
			method:       "POST",
			path:         "/api/posts",
			registerFunc: func(e *Engine, path string, handler Handler) { e.Post(path, handler) },
			expectedKey:  "POST_/api/posts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()

			handler := func(c *Context) {
				c.Status(http.StatusOK)
				c.w.Write([]byte("handler"))
			}

			tt.registerFunc(e, tt.path, handler)

			if _, ok := e.router.handlers[tt.expectedKey]; !ok {
				t.Errorf("Route %s not found in router", tt.expectedKey)
			}
		})
	}
}

func TestEngine_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		setupRoutes    func(*Engine)
		requestMethod  string
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful GET request",
			setupRoutes: func(e *Engine) {
				e.Get("/hello", func(c *Context) {
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
			name: "successful POST request",
			setupRoutes: func(e *Engine) {
				e.Post("/users", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("POST handler"))
				})
			},
			requestMethod:  "POST",
			requestPath:    "/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "POST handler",
		},
		{
			name: "route not found",
			setupRoutes: func(e *Engine) {
				// 不注册任何路由
			},
			requestMethod:  "GET",
			requestPath:    "/notfound",
			expectedStatus: http.StatusOK, // 当前实现返回 200
			expectedBody:   "404 NOT FOUND: /notfound\n",
		},
		{
			name: "method mismatch - GET on POST route",
			setupRoutes: func(e *Engine) {
				e.Post("/users", func(c *Context) {
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
			name: "method mismatch - POST on GET route",
			setupRoutes: func(e *Engine) {
				e.Get("/hello", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("GET handler"))
				})
			},
			requestMethod:  "POST",
			requestPath:    "/hello",
			expectedStatus: http.StatusOK,
			expectedBody:   "404 NOT FOUND: /hello\n",
		},
		{
			name: "multiple routes same path different methods - GET",
			setupRoutes: func(e *Engine) {
				e.Get("/api/users", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("GET response"))
				})
				e.Post("/api/users", func(c *Context) {
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
			name: "multiple routes same path different methods - POST",
			setupRoutes: func(e *Engine) {
				e.Get("/api/users", func(c *Context) {
					c.Status(http.StatusOK)
					c.w.Write([]byte("GET response"))
				})
				e.Post("/api/users", func(c *Context) {
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
			name: "empty path",
			setupRoutes: func(e *Engine) {
				e.Get("/", func(c *Context) {
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
			e := New()

			tt.setupRoutes(e)

			req := httptest.NewRequest(tt.requestMethod, tt.requestPath, nil)
			rr := httptest.NewRecorder()

			e.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected response body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
