package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           int
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Accept", "Content-Type", "Authorization", "X-Request-ID"},
		MaxAge:       86400,
	}
}

func (c CORSConfig) Validate() error {
	for _, origin := range c.AllowOrigins {
		if origin == "*" && c.AllowCredentials {
			return errors.New("CORS wildcard origin is incompatible with AllowCredentials")
		}
	}
	return nil
}

func CORS(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			originAllowed := false
			for _, allowed := range config.AllowOrigins {
				if allowed == "*" || strings.EqualFold(allowed, origin) {
					originAllowed = true
					break
				}
			}
			if originAllowed {
				if config.AllowOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
