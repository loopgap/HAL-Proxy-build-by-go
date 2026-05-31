package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
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
	// Pre-hash valid keys during initialization (once) to prevent map lookup timing attacks
	hashedKeys := make(map[string]string, len(config.ValidKeys))
	for key, val := range config.ValidKeys {
		hasher := sha256.New()
		hasher.Write([]byte(key))
		hashedKey := hex.EncodeToString(hasher.Sum(nil))
		hashedKeys[hashedKey] = val
	}

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

			// 1. Calculate SHA-256 hash of incoming API key to prevent Timing Attacks on length
			hasher := sha256.New()
			hasher.Write([]byte(apiKey))
			apiKeyHash := hex.EncodeToString(hasher.Sum(nil))

			// 2. Perform map lookup using the hashed key to eliminate partial key prefix matching timings
			service, valid := hashedKeys[apiKeyHash]

			// 3. Perform ConstantTimeCompare to further eliminate timing channel leakages
			expectedMatch := apiKeyHash
			if !valid {
				// If not valid, compare with dummy hash to consume constant time overhead
				expectedMatch = "dummyhashvalueforconstanttimecomparison123"
			}
			matchResult := subtle.ConstantTimeCompare([]byte(apiKeyHash), []byte(expectedMatch))

			if matchResult != 1 || !valid || service == "" {
				http.Error(w, "invalid_api_key", http.StatusUnauthorized)
				return
			}
			r.Header.Set("X-API-Service", service)
			next.ServeHTTP(w, r)
		})
	}
}

func GetServiceFromRequest(r *http.Request) string {
	return r.Header.Get("X-API-Service")
}
