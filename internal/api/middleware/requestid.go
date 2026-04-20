package middleware

import (
	"context"
	"net/http"
	"unicode"

	"github.com/google/uuid"
)

const requestIDKey = "requestID"

// sanitizeRequestID removes any characters that could be used for log injection
// or header injection attacks. Only alphanumeric characters, hyphens, and
// underscores are allowed. If the result is empty, a new UUID is generated.
func sanitizeRequestID(id string) string {
	if id == "" {
		return uuid.New().String()
	}
	var clean []rune
	for _, r := range id {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			clean = append(clean, r)
		}
	}
	if len(clean) == 0 {
		return uuid.New().String()
	}
	return string(clean)
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		requestID = sanitizeRequestID(requestID)
		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val := ctx.Value(requestIDKey); val != nil {
		if requestID, ok := val.(string); ok {
			return requestID
		}
	}
	return ""
}
