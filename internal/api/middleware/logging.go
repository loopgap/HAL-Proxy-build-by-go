package middleware

import (
	"net/http"
	"strings"
	"time"

	"bridgeos/internal/logging"
)

var defaultLogger = logging.Default()

// Logging returns an http.Handler that logs requests using the default logger
func Logging(next http.Handler) http.Handler {
	return LoggingWithProxies(next, nil)
}

// LoggingWithProxies returns an http.Handler that logs requests with client IP
func LoggingWithProxies(next http.Handler, trustedProxies []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Extract request_id from context if available
		requestID := GetRequestIDFromContext(r.Context())
		var reqLogger *logging.RequestLogger
		if requestID != "" {
			reqLogger = defaultLogger.WithRequestID(requestID)
		}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		fields := map[string]interface{}{
			"method":    sanitizeLogValue(r.Method),
			"path":      sanitizeLogValue(r.URL.Path),
			"status":    wrapper.statusCode,
			"latency":   duration.String(),
			"client_ip": sanitizeLogValue(getClientIP(r, trustedProxies)),
		}
		if reqLogger != nil {
			reqLogger.InfoWithFields("HTTP request", fields)
		} else {
			defaultLogger.InfoWithFields("HTTP request", fields)
		}
	})
}

func sanitizeLogValue(s string) string {
	s = strings.ReplaceAll(s, "\n", "_")
	s = strings.ReplaceAll(s, "\r", "_")
	s = strings.ReplaceAll(s, "\t", "_")
	return s
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
