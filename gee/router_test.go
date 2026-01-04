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

	if r.roots == nil {
		t.Fatal("newRouter() roots is nil")
	}

	if len(r.handlers) != 0 {
		t.Errorf("Expected empty handlers map, got %d handlers", len(r.handlers))
	}

	if len(r.roots) != 0 {
		t.Errorf("Expected empty roots map, got %d roots", len(r.roots))
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
			expectedStatus: http.StatusNotFound,
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
			expectedStatus: http.StatusNotFound,
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
			expectedStatus: http.StatusNotFound,
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

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "simple path",
			pattern:  "/hello",
			expected: []string{"hello"},
		},
		{
			name:     "nested path",
			pattern:  "/api/users",
			expected: []string{"api", "users"},
		},
		{
			name:     "path with param",
			pattern:  "/users/:id",
			expected: []string{"users", ":id"},
		},
		{
			name:     "path with wildcard",
			pattern:  "/static/*filepath",
			expected: []string{"static", "*filepath"},
		},
		{
			name:     "path with param and wildcard",
			pattern:  "/users/:id/files/*filepath",
			expected: []string{"users", ":id", "files", "*filepath"},
		},
		{
			name:     "empty path",
			pattern:  "/",
			expected: []string{},
		},
		{
			name:     "path with trailing slash",
			pattern:  "/hello/",
			expected: []string{"hello"},
		},
		{
			name:     "wildcard stops parsing",
			pattern:  "/static/*filepath/extra",
			expected: []string{"static", "*filepath"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePattern(tt.pattern)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d. Expected: %v, Got: %v", len(tt.expected), len(result), tt.expected, result)
				return
			}

			for i, v := range tt.expected {
				if result[i] != v {
					t.Errorf("Expected[%d] = %q, got %q", i, v, result[i])
				}
			}
		})
	}
}

func TestRouter_getRoute(t *testing.T) {
	tests := []struct {
		name           string
		setupRoutes    func(*router)
		method         string
		path           string
		expectedFound  bool
		expectedParams map[string]string
	}{
		{
			name: "exact match",
			setupRoutes: func(r *router) {
				r.addRoute("GET", "/hello", func(c *Context) {})
			},
			method:        "GET",
			path:          "/hello",
			expectedFound: true,
			expectedParams: map[string]string{},
		},
		{
			name: "not found",
			setupRoutes: func(r *router) {
				r.addRoute("GET", "/hello", func(c *Context) {})
			},
			method:        "GET",
			path:          "/world",
			expectedFound: false,
			expectedParams: nil,
		},
		{
			name: "param route",
			setupRoutes: func(r *router) {
				r.addRoute("GET", "/users/:id", func(c *Context) {})
			},
			method:        "GET",
			path:          "/users/123",
			expectedFound: true,
			expectedParams: map[string]string{"id": "123"},
		},
		{
			name: "multiple params",
			setupRoutes: func(r *router) {
				r.addRoute("GET", "/users/:id/posts/:postId", func(c *Context) {})
			},
			method:        "GET",
			path:          "/users/123/posts/456",
			expectedFound: true,
			expectedParams: map[string]string{"id": "123", "postId": "456"},
		},
		{
			name: "wildcard route",
			setupRoutes: func(r *router) {
				r.addRoute("GET", "/static/*filepath", func(c *Context) {})
			},
			method:        "GET",
			path:          "/static/css/style.css",
			expectedFound: true,
			expectedParams: map[string]string{"filepath": "css/style.css"},
		},
		{
			name: "wildcard route with multiple segments",
			setupRoutes: func(r *router) {
				r.addRoute("GET", "/static/*filepath", func(c *Context) {})
			},
			method:        "GET",
			path:          "/static/js/utils/helper.js",
			expectedFound: true,
			expectedParams: map[string]string{"filepath": "js/utils/helper.js"},
		},
		{
			name: "method mismatch",
			setupRoutes: func(r *router) {
				r.addRoute("POST", "/users", func(c *Context) {})
			},
			method:        "GET",
			path:          "/users",
			expectedFound: false,
			expectedParams: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter()
			tt.setupRoutes(r)

			node, params := r.getRoute(tt.method, tt.path)

			if tt.expectedFound {
				if node == nil {
					t.Errorf("Expected route to be found, but got nil")
					return
				}

				if len(params) != len(tt.expectedParams) {
					t.Errorf("Expected %d params, got %d", len(tt.expectedParams), len(params))
					return
				}

				for k, v := range tt.expectedParams {
					if params[k] != v {
						t.Errorf("Expected param[%s] = %q, got %q", k, v, params[k])
					}
				}
			} else {
				if node != nil {
					t.Errorf("Expected route not to be found, but got node with pattern %q", node.pattern)
				}
				if params != nil {
					t.Errorf("Expected params to be nil, got %v", params)
				}
			}
		})
	}
}

func TestRouter_handle_DynamicRoutes(t *testing.T) {
	tests := []struct {
		name           string
		setupHandlers  func(*router)
		requestMethod  string
		requestPath    string
		expectedStatus int
		expectedBody   string
		checkParams    map[string]string
	}{
		{
			name: "param route",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/users/:id", func(c *Context) {
					id := c.Param("id")
					c.String(http.StatusOK, "User ID: %s", id)
				})
			},
			requestMethod:  "GET",
			requestPath:    "/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   "User ID: 123",
			checkParams:    map[string]string{"id": "123"},
		},
		{
			name: "multiple params",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/users/:id/posts/:postId", func(c *Context) {
					id := c.Param("id")
					postId := c.Param("postId")
					c.String(http.StatusOK, "User: %s, Post: %s", id, postId)
				})
			},
			requestMethod:  "GET",
			requestPath:    "/users/123/posts/456",
			expectedStatus: http.StatusOK,
			expectedBody:   "User: 123, Post: 456",
			checkParams:    map[string]string{"id": "123", "postId": "456"},
		},
		{
			name: "wildcard route",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/static/*filepath", func(c *Context) {
					filepath := c.Param("filepath")
					c.String(http.StatusOK, "File: %s", filepath)
				})
			},
			requestMethod:  "GET",
			requestPath:    "/static/css/style.css",
			expectedStatus: http.StatusOK,
			expectedBody:   "File: css/style.css",
			checkParams:    map[string]string{"filepath": "css/style.css"},
		},
		{
			name: "wildcard route with multiple segments",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/files/*filepath", func(c *Context) {
					filepath := c.Param("filepath")
					c.String(http.StatusOK, "Path: %s", filepath)
				})
			},
			requestMethod:  "GET",
			requestPath:    "/files/js/utils/helper.js",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path: js/utils/helper.js",
			checkParams:    map[string]string{"filepath": "js/utils/helper.js"},
		},
		{
			name: "param route not matching",
			setupHandlers: func(r *router) {
				r.addRoute("GET", "/users/:id", func(c *Context) {
					c.String(http.StatusOK, "User")
				})
			},
			requestMethod:  "GET",
			requestPath:    "/users",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 NOT FOUND: /users\n",
			checkParams:    nil,
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

			if tt.checkParams != nil {
				for k, v := range tt.checkParams {
					if c.Param(k) != v {
						t.Errorf("Expected param[%s] = %q, got %q", k, v, c.Param(k))
					}
				}
			}
		})
	}
}

