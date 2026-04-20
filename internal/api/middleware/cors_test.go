package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	t.Run("AllowCredentials header is set when configured", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:     []string{"http://localhost:3000"},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: true,
			MaxAge:           86400,
		}

		handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		allowCred := rr.Header().Get("Access-Control-Allow-Credentials")
		if allowCred != "true" {
			t.Errorf("expected Access-Control-Allow-Credentials 'true', got %q", allowCred)
		}

		allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "http://localhost:3000" {
			t.Errorf("expected Access-Control-Allow-Origin 'http://localhost:3000', got %q", allowOrigin)
		}
	})

	t.Run("AllowCredentials header is NOT set when not configured", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:     []string{"http://localhost:3000"},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: false,
			MaxAge:           86400,
		}

		handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		allowCred := rr.Header().Get("Access-Control-Allow-Credentials")
		if allowCred != "" {
			t.Errorf("expected Access-Control-Allow-Credentials to be empty, got %q", allowCred)
		}
	})

	t.Run("OPTIONS request sets AllowCredentials header", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:     []string{"http://localhost:3000"},
			AllowMethods:     []string{"GET", "POST", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: true,
			MaxAge:           86400,
		}

		handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status %d for OPTIONS, got %d", http.StatusNoContent, rr.Code)
		}

		allowCred := rr.Header().Get("Access-Control-Allow-Credentials")
		if allowCred != "true" {
			t.Errorf("expected Access-Control-Allow-Credentials 'true' for OPTIONS, got %q", allowCred)
		}
	})
}

func TestCORSConfigValidate(t *testing.T) {
	t.Run("wildcard origin with AllowCredentials is rejected", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: true,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for wildcard origin with AllowCredentials")
		}
	})

	t.Run("non-wildcard origin with AllowCredentials is valid", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:     []string{"http://localhost:3000"},
			AllowMethods:     []string{"GET"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: true,
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("wildcard origin without AllowCredentials is valid", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: false,
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("wildcard origin mixed with other origins with AllowCredentials is rejected", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:     []string{"https://a.com", "*"},
			AllowMethods:     []string{"GET"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: true,
		}

		err := config.Validate()
		if err == nil {
			t.Error("expected error for wildcard origin mixed with other origins when AllowCredentials is true")
		}
	})
}
