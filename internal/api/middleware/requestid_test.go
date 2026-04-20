package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode"
)

func TestSanitizeRequestID(t *testing.T) {
	isValid := func(s string) bool {
		for _, r := range s {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
				return false
			}
		}
		return true
	}

	isUUID := func(s string) bool {
		if len(s) != 36 {
			return false
		}
		if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
			return false
		}
		return true
	}

	tests := []struct {
		name     string
		input    string
		wantUUID bool // true if we expect a UUID to be generated (empty or all invalid chars)
	}{
		{
			name:     "empty string returns UUID",
			input:    "",
			wantUUID: true,
		},
		{
			name:     "valid alphanumeric ID is preserved",
			input:    "abc123",
			wantUUID: false,
		},
		{
			name:     "valid ID with hyphens is preserved",
			input:    "req-123-abc",
			wantUUID: false,
		},
		{
			name:     "valid ID with underscores is preserved",
			input:    "req_123_abc",
			wantUUID: false,
		},
		{
			name:     "newline injection is sanitized",
			input:    "test\nmalicious",
			wantUUID: false,
		},
		{
			name:     "carriage return injection is sanitized",
			input:    "test\rmalicious",
			wantUUID: false,
		},
		{
			name:     "tab injection is sanitized",
			input:    "test\tmalicious",
			wantUUID: false,
		},
		{
			name:     "ANSI escape codes are sanitized",
			input:    "\x1b[31mmalicious",
			wantUUID: false,
		},
		{
			name:     "HTML script injection is sanitized",
			input:    "<script>alert(1)</script>",
			wantUUID: false,
		},
		{
			name:     "newline only returns UUID",
			input:    "\n",
			wantUUID: true,
		},
		{
			name:     "only invalid characters returns UUID",
			input:    "!!!@@@",
			wantUUID: true,
		},
		{
			name:     "mixed valid and invalid returns only valid",
			input:    "req-123\nabc",
			wantUUID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeRequestID(tt.input)

			if tt.wantUUID {
				if !isUUID(got) {
					t.Errorf("sanitizeRequestID(%q) = %q, want UUID", tt.input, got)
				}
			} else {
				if !isValid(got) {
					t.Errorf("sanitizeRequestID(%q) = %q, contains invalid characters", tt.input, got)
				}
				for _, c := range got {
					if c == '\n' || c == '\r' || c == '\t' || c == '<' || c == '>' {
						t.Errorf("sanitizeRequestID(%q) = %q, contains dangerous character %c", tt.input, got, c)
					}
				}
			}
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("valid X-Request-ID header is used", func(t *testing.T) {
		handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "my-request-123")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		got := rr.Header().Get("X-Request-ID")
		if got != "my-request-123" {
			t.Errorf("X-Request-ID = %q, want %q", got, "my-request-123")
		}
	})

	t.Run("invalid X-Request-ID header is sanitized", func(t *testing.T) {
		handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "test\nmalicious")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		got := rr.Header().Get("X-Request-ID")
		// Should not contain newline
		for _, c := range got {
			if c == '\n' || c == '\r' || c == '\t' {
				t.Errorf("X-Request-ID %q contains control character", got)
			}
		}
		// Should only have valid chars
		for _, c := range got {
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' && c != '_' {
				t.Errorf("X-Request-ID %q contains invalid character %c", got, c)
			}
		}
	})

	t.Run("missing X-Request-ID generates UUID", func(t *testing.T) {
		handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		got := rr.Header().Get("X-Request-ID")
		// Should be a valid UUID
		if len(got) != 36 {
			t.Errorf("X-Request-ID = %q, want valid UUID", got)
		}
	})

	t.Run("X-Request-ID is set in context", func(t *testing.T) {
		var ctxRequestID string
		handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxRequestID = GetRequestIDFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "test-req-456")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if ctxRequestID != "test-req-456" {
			t.Errorf("context X-Request-ID = %q, want %q", ctxRequestID, "test-req-456")
		}
	})
}
