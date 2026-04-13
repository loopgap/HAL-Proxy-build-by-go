package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const requestIDKey = "requestID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
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
