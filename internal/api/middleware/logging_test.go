package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bridgeos/internal/logging"
)

func TestLoggingWithProxiesLogsMethodAndPath(t *testing.T) {
	var buf bytes.Buffer
	previous := defaultLogger
	defaultLogger = logging.New(&buf, "info")
	t.Cleanup(func() { defaultLogger = previous })

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()

	LoggingWithProxies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil).ServeHTTP(rr, req)

	logged := buf.String()
	if !strings.Contains(logged, `"method":"GET"`) {
		t.Errorf("log missing method field: %s", logged)
	}
	if !strings.Contains(logged, `"/v1/cases"`) {
		t.Errorf("log missing path field: %s", logged)
	}
	if !strings.Contains(logged, `"status":200`) {
		t.Errorf("log missing status field: %s", logged)
	}
}

func TestLoggingWithProxiesLogsErrorStatus(t *testing.T) {
	var buf bytes.Buffer
	previous := defaultLogger
	defaultLogger = logging.New(&buf, "info")
	t.Cleanup(func() { defaultLogger = previous })

	tests := []struct {
		name       string
		statusCode int
		statusStr  string
	}{
		{"400 Bad Request", http.StatusBadRequest, `"status":400`},
		{"404 Not Found", http.StatusNotFound, `"status":404`},
		{"500 Internal Server Error", http.StatusInternalServerError, `"status":500`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			req := httptest.NewRequest(http.MethodPost, "/v1/cases", nil)
			req.RemoteAddr = "127.0.0.1:1234"
			rr := httptest.NewRecorder()

			LoggingWithProxies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}), nil).ServeHTTP(rr, req)

			if !strings.Contains(buf.String(), tt.statusStr) {
				t.Errorf("log missing status %d: %s", tt.statusCode, buf.String())
			}
		})
	}
}

func TestLoggingWithProxiesUsesClientIP(t *testing.T) {
	var buf bytes.Buffer
	previous := defaultLogger
	defaultLogger = logging.New(&buf, "info")
	t.Cleanup(func() { defaultLogger = previous })

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	rr := httptest.NewRecorder()

	LoggingWithProxies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), []string{"10.0.0.1"}).ServeHTTP(rr, req)

	logged := buf.String()
	if !strings.Contains(logged, `"client_ip":"203.0.113.1"`) {
		t.Errorf("log missing expected client_ip: %s", logged)
	}
}

func TestLoggingWithProxiesIncludesRequestID(t *testing.T) {
	var buf bytes.Buffer
	previous := defaultLogger
	defaultLogger = logging.New(&buf, "info")
	t.Cleanup(func() { defaultLogger = previous })

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "test-req-789"))
	rr := httptest.NewRecorder()

	LoggingWithProxies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil).ServeHTTP(rr, req)

	logged := buf.String()
	if !strings.Contains(logged, `"request_id":"test-req-789"`) {
		t.Errorf("log missing request_id: %s", logged)
	}
}

func TestLoggingCallsNextHandler(t *testing.T) {
	var buf bytes.Buffer
	previous := defaultLogger
	defaultLogger = logging.New(&buf, "info")
	t.Cleanup(func() { defaultLogger = previous })

	called := false
	Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !called {
		t.Error("next handler was not called")
	}
}

func TestResponseWriterCapturesStatusCode(t *testing.T) {
	rr := httptest.NewRecorder()
	wrapper := &responseWriter{
		ResponseWriter: rr,
		statusCode:     http.StatusOK,
	}

	wrapper.WriteHeader(http.StatusNotFound)

	if wrapper.statusCode != http.StatusNotFound {
		t.Errorf("statusCode = %d, want %d", wrapper.statusCode, http.StatusNotFound)
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("underlying writer status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestResponseWriterDefaultStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	wrapper := &responseWriter{
		ResponseWriter: rr,
		statusCode:     http.StatusOK,
	}

	if wrapper.statusCode != http.StatusOK {
		t.Errorf("default statusCode = %d, want %d", wrapper.statusCode, http.StatusOK)
	}
}

func TestSanitizeLogValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no special chars", "hello", "hello"},
		{"newline replaced", "a\nb", "a_b"},
		{"carriage return replaced", "a\rb", "a_b"},
		{"tab replaced", "a\tb", "a_b"},
		{"multiple replaced", "a\nb\rc\td", "a_b_c_d"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLogValue(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeLogValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoggingWithProxiesLatency(t *testing.T) {
	var buf bytes.Buffer
	previous := defaultLogger
	defaultLogger = logging.New(&buf, "info")
	t.Cleanup(func() { defaultLogger = previous })

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()

	LoggingWithProxies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil).ServeHTTP(rr, req)

	logged := buf.String()
	if !strings.Contains(logged, `"latency"`) {
		t.Errorf("log missing latency field: %s", logged)
	}
}
