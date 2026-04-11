package middleware

import (
	"net/http"
)

type APIKeyConfig struct {
	HeaderName string
	QueryParam string
	ValidKeys  map[string]string
	SkipPaths  []string
}

func DefaultAPIKeyConfig() APIKeyConfig {
	return APIKeyConfig{
		HeaderName: "X-API-Key",
		QueryParam: "api_key",
		ValidKeys:  make(map[string]string),
		SkipPaths:  []string{"/v1/health"},
	}
}

func APIKeyAuth(config APIKeyConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range config.SkipPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}
			apiKey := r.Header.Get(config.HeaderName)
			if apiKey == "" && config.QueryParam != "" {
				apiKey = r.URL.Query().Get(config.QueryParam)
			}
			if apiKey == "" {
				http.Error(w, "missing_api_key", http.StatusUnauthorized)
				return
			}
			if _, valid := config.ValidKeys[apiKey]; !valid {
				http.Error(w, "invalid_api_key", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
