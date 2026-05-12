package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bridgeos/internal/logging"
)

func TestAPIKeyAuthSupportsSkipHeaderAndQuery(t *testing.T) {
	t.Parallel()

	handler := APIKeyAuth(APIKeyConfig{
		HeaderName: "X-API-Key",
		QueryParam: "api_key",
		ValidKeys:  map[string]string{"header-key": "header-service", "query-key": "query-service"},
		SkipPaths:  []string{"/v1/health"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(GetServiceFromRequest(r)))
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("skip path status=%d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-API-Key", "header-key")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "header-service" {
		t.Fatalf("header key status=%d body=%q", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/cases?api_key=query-key", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "query-service" {
		t.Fatalf("query key status=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyAuthRejectsMissingAndInvalidKeys(t *testing.T) {
	t.Parallel()

	handler := APIKeyAuth(APIKeyConfig{
		HeaderName: "X-API-Key",
		ValidKeys:  map[string]string{"good": "service"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/cases", nil))
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "missing_api_key") {
		t.Fatalf("missing key status=%d body=%q", rec.Code, rec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-API-Key", "bad")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "invalid_api_key") {
		t.Fatalf("invalid key status=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestRateLimitSecurityAndChain(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(1, time.Minute)
	defer limiter.Stop()

	var chainOrder []string
	mark := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				chainOrder = append(chainOrder, name+"-before")
				next.ServeHTTP(w, r)
				chainOrder = append(chainOrder, name+"-after")
			})
		}
	}

	handler := Chain(
		Security(RateLimit(limiter, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))),
		mark("outer"),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first request status=%d", rec.Code)
	}
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("security header=%q", got)
	}
	if strings.Join(chainOrder, ",") != "outer-before,outer-after" {
		t.Fatalf("unexpected middleware order %v", chainOrder)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status=%d", rec.Code)
	}
}

func TestLoggingWithProxiesSanitizesAndCapturesStatus(t *testing.T) {
	var buf bytes.Buffer
	previous := defaultLogger
	defaultLogger = logging.New(&buf, "info")
	t.Cleanup(func() { defaultLogger = previous })

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "req-1"))

	rec := httptest.NewRecorder()
	LoggingWithProxies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("logged handler status=%d", rec.Code)
	}
	if !strings.Contains(buf.String(), `"status":202`) || !strings.Contains(buf.String(), `"request_id":"req-1"`) {
		t.Fatalf("missing log fields: %s", buf.String())
	}
	if strings.Contains(buf.String(), "\n") && strings.Count(strings.TrimSpace(buf.String()), "\n") > 0 {
		t.Fatalf("expected single sanitized log line: %s", buf.String())
	}
	if got := sanitizeLogValue("a\nb\rc\td"); got != "a_b_c_d" {
		t.Fatalf("sanitizeLogValue=%q", got)
	}
}
