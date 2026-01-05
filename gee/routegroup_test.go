package gee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteGroup_Group(t *testing.T) {
	e := New()

	// 创建根路由组
	rootGroup := e.RouteGroup

	// 创建子路由组
	apiGroup := rootGroup.Group("/api")
	if apiGroup == nil {
		t.Fatal("Group() returned nil")
	}

	if apiGroup.prefix != "/api" {
		t.Errorf("Expected prefix /api, got %s", apiGroup.prefix)
	}

	if apiGroup.engine != e {
		t.Error("Group engine not set correctly")
	}

	// 创建嵌套路由组
	v1Group := apiGroup.Group("/v1")
	if v1Group.prefix != "/api/v1" {
		t.Errorf("Expected prefix /api/v1, got %s", v1Group.prefix)
	}

	// 创建另一个嵌套路由组
	usersGroup := v1Group.Group("/users")
	if usersGroup.prefix != "/api/v1/users" {
		t.Errorf("Expected prefix /api/v1/users, got %s", usersGroup.prefix)
	}
}

func TestRouteGroup_GET(t *testing.T) {
	tests := []struct {
		name           string
		setupGroup     func(*Engine) *RouteGroup
		pattern        string
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "GET route without prefix",
			setupGroup: func(e *Engine) *RouteGroup {
				return e.RouteGroup
			},
			pattern:        "/hello",
			requestPath:    "/hello",
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello",
		},
		{
			name: "GET route with prefix",
			setupGroup: func(e *Engine) *RouteGroup {
				return e.Group("/api")
			},
			pattern:        "/users",
			requestPath:    "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "Users",
		},
		{
			name: "GET route with nested prefix",
			setupGroup: func(e *Engine) *RouteGroup {
				return e.Group("/api").Group("/v1")
			},
			pattern:        "/users",
			requestPath:    "/api/v1/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "V1 Users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			group := tt.setupGroup(e)

			group.GET(tt.pattern, func(c *Context) {
				c.Status(http.StatusOK)
				c.w.Write([]byte(tt.expectedBody))
			})

			req := httptest.NewRequest("GET", tt.requestPath, nil)
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

func TestRouteGroup_POST(t *testing.T) {
	tests := []struct {
		name           string
		setupGroup     func(*Engine) *RouteGroup
		pattern        string
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "POST route without prefix",
			setupGroup: func(e *Engine) *RouteGroup {
				return e.RouteGroup
			},
			pattern:        "/users",
			requestPath:    "/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "Create User",
		},
		{
			name: "POST route with prefix",
			setupGroup: func(e *Engine) *RouteGroup {
				return e.Group("/api")
			},
			pattern:        "/users",
			requestPath:    "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "API Create User",
		},
		{
			name: "POST route with nested prefix",
			setupGroup: func(e *Engine) *RouteGroup {
				return e.Group("/api").Group("/v1")
			},
			pattern:        "/posts",
			requestPath:    "/api/v1/posts",
			expectedStatus: http.StatusOK,
			expectedBody:   "V1 Create Post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			group := tt.setupGroup(e)

			group.POST(tt.pattern, func(c *Context) {
				c.Status(http.StatusOK)
				c.w.Write([]byte(tt.expectedBody))
			})

			req := httptest.NewRequest("POST", tt.requestPath, nil)
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

func TestRouteGroup_PrefixConcatenation(t *testing.T) {
	e := New()

	// 测试前缀拼接
	apiGroup := e.Group("/api")
	v1Group := apiGroup.Group("/v1")
	usersGroup := v1Group.Group("/users")

	usersGroup.GET("/:id", func(c *Context) {
		id := c.Param("id")
		c.String(http.StatusOK, "User ID: %s", id)
	})

	// 测试路由是否正确注册
	expectedPattern := "/api/v1/users/:id"
	key := "GET_" + expectedPattern
	if _, ok := e.router.handlers[key]; !ok {
		t.Errorf("Route %s not found in handlers", key)
	}

	// 测试路由是否能正确匹配
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	rr := httptest.NewRecorder()
	e.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	expectedBody := "User ID: 123"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected response body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestRouteGroup_MultipleGroups(t *testing.T) {
	e := New()

	// 创建多个独立的路由组
	apiGroup := e.Group("/api")
	adminGroup := e.Group("/admin")

	apiGroup.GET("/users", func(c *Context) {
		c.String(http.StatusOK, "API Users")
	})

	adminGroup.GET("/users", func(c *Context) {
		c.String(http.StatusOK, "Admin Users")
	})

	// 测试 /api/users
	req1 := httptest.NewRequest("GET", "/api/users", nil)
	rr1 := httptest.NewRecorder()
	e.ServeHTTP(rr1, req1)

	if rr1.Body.String() != "API Users" {
		t.Errorf("Expected 'API Users', got %q", rr1.Body.String())
	}

	// 测试 /admin/users
	req2 := httptest.NewRequest("GET", "/admin/users", nil)
	rr2 := httptest.NewRecorder()
	e.ServeHTTP(rr2, req2)

	if rr2.Body.String() != "Admin Users" {
		t.Errorf("Expected 'Admin Users', got %q", rr2.Body.String())
	}
}

func TestRouteGroup_EmptyPrefix(t *testing.T) {
	e := New()

	// 测试空前缀的路由组
	emptyGroup := e.Group("")
	emptyGroup.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "Empty Prefix")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	e.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "Empty Prefix" {
		t.Errorf("Expected 'Empty Prefix', got %q", rr.Body.String())
	}
}

func TestRouteGroup_RootGroup(t *testing.T) {
	e := New()

	// 测试根路由组（prefix 为空）
	rootGroup := e.RouteGroup

	if rootGroup.prefix != "" {
		t.Errorf("Expected empty prefix for root group, got %q", rootGroup.prefix)
	}

	rootGroup.GET("/root", func(c *Context) {
		c.String(http.StatusOK, "Root")
	})

	req := httptest.NewRequest("GET", "/root", nil)
	rr := httptest.NewRecorder()
	e.ServeHTTP(rr, req)

	if rr.Body.String() != "Root" {
		t.Errorf("Expected 'Root', got %q", rr.Body.String())
	}
}

