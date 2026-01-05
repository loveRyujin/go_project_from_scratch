package gee

import (
	"fmt"
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

	if c.params == nil {
		t.Error("Expected params map to be initialized")
	}

	if len(c.params) != 0 {
		t.Errorf("Expected empty params map, got %d params", len(c.params))
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

func TestContext_Param(t *testing.T) {
	tests := []struct {
		name          string
		params        map[string]string
		key           string
		expectedValue string
	}{
		{
			name:          "get existing param",
			params:        map[string]string{"id": "123", "name": "john"},
			key:           "id",
			expectedValue: "123",
		},
		{
			name:          "get another param",
			params:        map[string]string{"id": "123", "name": "john"},
			key:           "name",
			expectedValue: "john",
		},
		{
			name:          "get non-existent param",
			params:        map[string]string{"id": "123"},
			key:           "email",
			expectedValue: "",
		},
		{
			name:          "empty params map",
			params:        map[string]string{},
			key:           "id",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)
			c.params = tt.params

			value := c.Param(tt.key)

			if value != tt.expectedValue {
				t.Errorf("Expected value %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestContext_String(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		format         string
		args           []any
		expectedStatus int
		expectedBody   string
		checkHeader    bool
	}{
		{
			name:           "simple string",
			code:           http.StatusOK,
			format:         "Hello, World!",
			args:           nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
			checkHeader:    true,
		},
		{
			name:           "formatted string",
			code:           http.StatusOK,
			format:         "User ID: %d",
			args:           []any{123},
			expectedStatus: http.StatusOK,
			expectedBody:   "User ID: 123",
			checkHeader:    true,
		},
		{
			name:           "multiple args",
			code:           http.StatusOK,
			format:         "User: %s, Age: %d",
			args:           []any{"john", 30},
			expectedStatus: http.StatusOK,
			expectedBody:   "User: john, Age: 30",
			checkHeader:    true,
		},
		{
			name:           "not found status",
			code:           http.StatusNotFound,
			format:         "404 NOT FOUND: %s",
			args:           []any{"/test"},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 NOT FOUND: /test",
			checkHeader:    true,
		},
		{
			name:           "empty string",
			code:           http.StatusOK,
			format:         "",
			args:           nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
			checkHeader:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			// 构建预期的格式化字符串
			var formatted string
			if tt.args == nil {
				formatted = tt.format
			} else {
				formatted = fmt.Sprintf(tt.format, tt.args...)
			}

			// 使用常量格式字符串调用 String 方法
			c.String(tt.code, "%s", formatted)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}

			if tt.checkHeader {
				contentType := rr.Header().Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("Expected Content-Type text/plain, got %s", contentType)
				}
			}
		})
	}
}

func TestContext_Method(t *testing.T) {
	tests := []struct {
		name           string
		requestMethod  string
		expectedMethod string
	}{
		{
			name:           "GET method",
			requestMethod:  "GET",
			expectedMethod: "GET",
		},
		{
			name:           "POST method",
			requestMethod:  "POST",
			expectedMethod: "POST",
		},
		{
			name:           "PUT method",
			requestMethod:  "PUT",
			expectedMethod: "PUT",
		},
		{
			name:           "DELETE method",
			requestMethod:  "DELETE",
			expectedMethod: "DELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/test", nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			if c.Method() != tt.expectedMethod {
				t.Errorf("Expected method %s, got %s", tt.expectedMethod, c.Method())
			}
		})
	}
}

func TestContext_Path(t *testing.T) {
	tests := []struct {
		name         string
		requestPath  string
		expectedPath string
	}{
		{
			name:         "simple path",
			requestPath:  "/hello",
			expectedPath: "/hello",
		},
		{
			name:         "nested path",
			requestPath:  "/api/users",
			expectedPath: "/api/users",
		},
		{
			name:         "path with query",
			requestPath:  "/users?page=1",
			expectedPath: "/users",
		},
		{
			name:         "root path",
			requestPath:  "/",
			expectedPath: "/",
		},
		{
			name:         "path with params",
			requestPath:  "/users/123/posts/456",
			expectedPath: "/users/123/posts/456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			if c.Path() != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, c.Path())
			}
		})
	}
}

func TestContext_Data(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		data           []byte
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "simple byte array",
			code:           http.StatusOK,
			data:           []byte("Hello, World!"),
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
		},
		{
			name:           "empty byte array",
			code:           http.StatusOK,
			data:           []byte{},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "binary data",
			code:           http.StatusOK,
			data:           []byte{0x00, 0x01, 0x02, 0x03},
			expectedStatus: http.StatusOK,
			expectedBody:   string([]byte{0x00, 0x01, 0x02, 0x03}),
		},
		{
			name:           "not found status",
			code:           http.StatusNotFound,
			data:           []byte("Not Found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		},
		{
			name:           "JSON data",
			code:           http.StatusOK,
			data:           []byte(`{"name":"john","age":30}`),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"name":"john","age":30}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			c := newContext(rr, req)

			c.Data(tt.code, tt.data)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
