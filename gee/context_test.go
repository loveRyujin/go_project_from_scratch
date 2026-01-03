package gee

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	c := newContext(rr, req)

	if c == nil {
		t.Fatal("newContext() returned nil")
	}

	if c.w != rr {
		t.Error("Context writer not set correctly")
	}

	if c.r != req {
		t.Error("Context request not set correctly")
	}

	if c.method != "GET" {
		t.Errorf("Expected method GET, got %s", c.method)
	}

	if c.path != "/test" {
		t.Errorf("Expected path /test, got %s", c.path)
	}

	if c.statusCode != 0 {
		t.Errorf("Expected initial status code 0, got %d", c.statusCode)
	}
}

func TestContext_PostForm(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		body          string
		contentType   string
		key           string
		expectedValue string
	}{
		{
			name:          "get form value",
			method:        "POST",
			body:          "username=testuser&password=testpass",
			contentType:   "application/x-www-form-urlencoded",
			key:           "username",
			expectedValue: "testuser",
		},
		{
			name:          "get form value - password",
			method:        "POST",
			body:          "username=testuser&password=testpass",
			contentType:   "application/x-www-form-urlencoded",
			key:           "password",
			expectedValue: "testpass",
		},
		{
			name:          "form value not found",
			method:        "POST",
			body:          "username=testuser",
			contentType:   "application/x-www-form-urlencoded",
			key:           "email",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			value := c.PostForm(tt.key)

			if value != tt.expectedValue {
				t.Errorf("Expected value %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestContext_Query(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		key           string
		expectedValue string
	}{
		{
			name:          "get query parameter",
			url:           "/test?name=john&age=30",
			key:           "name",
			expectedValue: "john",
		},
		{
			name:          "get query parameter - age",
			url:           "/test?name=john&age=30",
			key:           "age",
			expectedValue: "30",
		},
		{
			name:          "query parameter not found",
			url:           "/test?name=john",
			key:           "email",
			expectedValue: "",
		},
		{
			name:          "no query parameters",
			url:           "/test",
			key:           "name",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			value := c.Query(tt.key)

			if value != tt.expectedValue {
				t.Errorf("Expected value %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestContext_Status(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		expectedCode int
	}{
		{
			name:         "set status OK",
			statusCode:   http.StatusOK,
			expectedCode: http.StatusOK,
		},
		{
			name:         "set status not found",
			statusCode:   http.StatusNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "set status internal server error",
			statusCode:   http.StatusInternalServerError,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "set status created",
			statusCode:   http.StatusCreated,
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			c.Status(tt.statusCode)

			if c.statusCode != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, c.statusCode)
			}

			if rr.Code != tt.expectedCode {
				t.Errorf("Expected response code %d, got %d", tt.expectedCode, rr.Code)
			}
		})
	}
}

func TestContext_SetHeader(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		expectedValue string
	}{
		{
			name:          "set content type",
			key:           "Content-Type",
			value:         "application/json",
			expectedValue: "application/json",
		},
		{
			name:          "set custom header",
			key:           "X-Custom-Header",
			value:         "custom-value",
			expectedValue: "custom-value",
		},
		{
			name:          "set authorization header",
			key:           "Authorization",
			value:         "Bearer token123",
			expectedValue: "Bearer token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			c.SetHeader(tt.key, tt.value)

			actualValue := rr.Header().Get(tt.key)
			if actualValue != tt.expectedValue {
				t.Errorf("Expected header value %q, got %q", tt.expectedValue, actualValue)
			}
		})
	}
}

func TestContext_JSON(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		obj            any
		expectedStatus int
		expectedBody   string
		checkHeader    bool
	}{
		{
			name:           "encode map",
			code:           http.StatusOK,
			obj:            H{"name": "john", "age": 30},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"age":30,"name":"john"}` + "\n",
			checkHeader:    true,
		},
		{
			name:           "encode struct",
			code:           http.StatusCreated,
			obj:            struct{ Message string }{"success"},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"Message":"success"}` + "\n",
			checkHeader:    true,
		},
		{
			name:           "encode string",
			code:           http.StatusOK,
			obj:            "test",
			expectedStatus: http.StatusOK,
			expectedBody:   `"test"` + "\n",
			checkHeader:    true,
		},
		{
			name:           "encode number",
			code:           http.StatusOK,
			obj:            42,
			expectedStatus: http.StatusOK,
			expectedBody:   "42\n",
			checkHeader:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			c.JSON(tt.code, tt.obj)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}

			if tt.checkHeader {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}

func TestContext_JSON_ErrorHandling(t *testing.T) {
	// 测试无法序列化的对象（使用 channel，无法序列化为 JSON）
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	c := newContext(rr, req)

	// 创建一个无法序列化的对象
	ch := make(chan int)
	c.JSON(http.StatusOK, ch)

	// 应该返回错误状态码
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// 响应体应该包含错误信息
	if rr.Body.Len() == 0 {
		t.Error("Expected error message in response body")
	}
}
