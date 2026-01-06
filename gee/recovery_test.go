package gee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name           string
		handler        Handler
		expectedStatus int
		expectedBody   string
		shouldPanic    bool
	}{
		{
			name: "handler without panic",
			handler: func(c *Context) {
				c.String(http.StatusOK, "Success")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Success",
			shouldPanic:    false,
		},
		{
			name: "handler with panic - string",
			handler: func(c *Context) {
				panic("something went wrong")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
			shouldPanic:    true,
		},
		{
			name: "handler with panic - error",
			handler: func(c *Context) {
				panic("runtime error: index out of range")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
			shouldPanic:    true,
		},
		{
			name: "handler with panic - nil",
			handler: func(c *Context) {
				var nilValue interface{} = nil
				panic(nilValue) //nolint:staticcheck // intentionally testing nil panic recovery
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
			shouldPanic:    true,
		},
		{
			name: "handler with panic in nested handler",
			handler: func(c *Context) {
				c.Next()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
			shouldPanic:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Use(Recovery())

			if tt.shouldPanic && tt.name == "handler with panic in nested handler" {
				e.GET("/test", func(c *Context) {
					panic("nested panic")
				})
			} else {
				e.GET("/test", tt.handler)
			}

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			e.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestRecovery_WithMultipleHandlers(t *testing.T) {
	e := New()
	e.Use(Recovery())

	executionOrder := make([]string, 0)

	e.Use(func(c *Context) {
		executionOrder = append(executionOrder, "middleware-1")
		c.Next()
	})

	e.GET("/test", func(c *Context) {
		executionOrder = append(executionOrder, "before-panic")
		panic("test panic")
		// executionOrder = append(executionOrder, "after-panic") // 不会执行，已移除
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	e.ServeHTTP(rr, req)

	// Recovery 应该捕获 panic，返回 500 错误
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	if rr.Body.String() != "Internal server error" {
		t.Errorf("Expected body 'Internal server error', got %q", rr.Body.String())
	}

	// 验证中间件执行了
	if len(executionOrder) < 1 {
		t.Error("Expected at least one execution, got none")
	}
}

func TestRecovery_IndexOutOfRange(t *testing.T) {
	e := New()
	e.Use(Recovery())

	e.GET("/panic", func(c *Context) {
		s := []string{"1", "2", "3"}
		c.String(http.StatusOK, "get %s", s[3]) // 索引越界
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	rr := httptest.NewRecorder()

	e.ServeHTTP(rr, req)

	// Recovery 应该捕获 panic
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	if rr.Body.String() != "Internal server error" {
		t.Errorf("Expected body 'Internal server error', got %q", rr.Body.String())
	}
}

func TestRecovery_ContinueAfterRecovery(t *testing.T) {
	e := New()
	e.Use(Recovery())

	panicHandler := func(c *Context) {
		panic("panic in handler")
	}

	normalHandler := func(c *Context) {
		c.String(http.StatusOK, "Normal")
	}

	e.GET("/panic", panicHandler)
	e.GET("/normal", normalHandler)

	// 测试 panic 路由
	req1 := httptest.NewRequest("GET", "/panic", nil)
	rr1 := httptest.NewRecorder()
	e.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rr1.Code)
	}

	// 测试正常路由（验证 Recovery 不影响其他路由）
	req2 := httptest.NewRequest("GET", "/normal", nil)
	rr2 := httptest.NewRecorder()
	e.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr2.Code)
	}

	if rr2.Body.String() != "Normal" {
		t.Errorf("Expected body 'Normal', got %q", rr2.Body.String())
	}
}

func TestDefault(t *testing.T) {
	e := Default()

	if e == nil {
		t.Fatal("Default() returned nil")
	}

	if e.RouteGroup == nil {
		t.Fatal("Default() RouteGroup is nil")
	}

	// 验证 Default() 包含了 Recovery 中间件
	if len(e.RouteGroup.handlers) == 0 {
		t.Error("Expected Default() to have Recovery middleware")
	}

	// 测试 Default() 创建的 Engine 能正常工作
	e.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "Test")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	e.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "Test" {
		t.Errorf("Expected body 'Test', got %q", rr.Body.String())
	}
}

func TestDefault_WithPanic(t *testing.T) {
	e := Default()

	e.GET("/panic", func(c *Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	rr := httptest.NewRecorder()

	e.ServeHTTP(rr, req)

	// Default() 包含 Recovery，应该能捕获 panic
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	if rr.Body.String() != "Internal server error" {
		t.Errorf("Expected body 'Internal server error', got %q", rr.Body.String())
	}
}

