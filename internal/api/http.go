package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hal-proxy/internal/core"
	"hal-proxy/internal/domain"
	apperrors "hal-proxy/internal/errors"
	"hal-proxy/internal/metrics"
	"hal-proxy/internal/store"

	"hal-proxy/internal/api/dto"
	"hal-proxy/internal/api/middleware"
)

// MaxBodySize limits request body to 1MB for security
const MaxBodySize = 1 << 20 // 1MB

// RequestTimeout sets the maximum duration for request processing
const RequestTimeout = 30 * 1e9 // 30 seconds in nanoseconds

// Server handles HTTP API requests with security best practices
type Server struct {
	svc *core.Service
	db  *sql.DB
	// In production, this should be populated with actual authentication middleware
	authMiddleware func(http.Handler) http.Handler
	trustedProxies []string
}

// NewServer creates a new API server instance
func NewServer(svc *core.Service, db *sql.DB, jwtSecret string, jwtExpiryHours int, jwtIssuer string, trustedProxies []string) *Server {
	jwtConfig := middleware.JWTConfig{
		Secret:          jwtSecret,
		ExpirationHours: jwtExpiryHours,
		Issuer:          jwtIssuer,
	}

	var authMW func(http.Handler) http.Handler
	if jwtConfig.Secret != "" {
		authMW = middleware.JWTAuth(jwtConfig)
	} else {
		authMW = func(next http.Handler) http.Handler { return next }
	}

	return &Server{
		svc:            svc,
		db:             db,
		authMiddleware: authMW,
		trustedProxies: trustedProxies,
	}
}

// SetAuthMiddleware allows setting a custom authentication middleware
func (s *Server) SetAuthMiddleware(middleware func(http.Handler) http.Handler) {
	s.authMiddleware = middleware
}

// Handler returns the HTTP handler with all routes configured
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.route)
}

// normalizePath converts actual request paths to route templates to reduce Prometheus label cardinality
// e.g., /v1/cases/abc123 -> /v1/cases/{id}
func normalizePath(path string) string {
	switch {
	case path == "/v1/health":
		return "/v1/health"
	case path == "/v1/cases":
		return "/v1/cases"
	case strings.HasPrefix(path, "/v1/cases/") && strings.HasSuffix(path, "/run"):
		return "/v1/cases/{id}/run"
	case strings.HasPrefix(path, "/v1/cases/") && strings.HasSuffix(path, "/events"):
		return "/v1/cases/{id}/events"
	case strings.HasPrefix(path, "/v1/cases/"):
		return "/v1/cases/{id}"
	case path == "/v1/approvals":
		return "/v1/approvals"
	case strings.HasPrefix(path, "/v1/approvals/") && strings.HasSuffix(path, "/approve"):
		return "/v1/approvals/{id}/approve"
	case strings.HasPrefix(path, "/v1/approvals/") && strings.HasSuffix(path, "/reject"):
		return "/v1/approvals/{id}/reject"
	case strings.HasPrefix(path, "/v1/reports/") && strings.HasSuffix(path, "/build"):
		return "/v1/reports/{case_id}/build"
	case path == "/v1/sessions":
		return "/v1/sessions"
	case path == "/v1/devices":
		return "/v1/devices"
	default:
		return path
	}
}

// metricsResponseWriter wraps http.ResponseWriter to capture status code for metrics
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (m *metricsResponseWriter) WriteHeader(code int) {
	m.statusCode = code
	m.ResponseWriter.WriteHeader(code)
}

func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	path := normalizePath(r.URL.Path)

	// Increment in-flight requests gauge
	metrics.HTTPRequestsInFlight.Inc()
	defer metrics.HTTPRequestsInFlight.Dec()

	// Wrap ResponseWriter to capture status code for metrics
	metricsW := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	w = metricsW

	w.Header().Set("Content-Type", "application/json")

	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Add request timeout context
	ctx := r.Context()
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(r.Context(), RequestTimeout)
		defer cancel()
	}
	r = r.WithContext(ctx)

	// Apply auth middleware for protected routes (skip health check)
	if r.URL.Path != "/v1/health" {
		authHandler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.routeInternal(w, r)
		}))
		authHandler.ServeHTTP(w, r)
	} else {
		// Health check endpoint (no auth required)
		s.routeInternal(w, r)
	}

	// Record metrics with normalized path to reduce cardinality
	duration := time.Since(start).Seconds()
	metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, statusCodeToString(metricsW.statusCode)).Inc()
	metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
}

func statusCodeToString(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

func (s *Server) routeInternal(w http.ResponseWriter, r *http.Request) {
	// Route handling - improved RESTful design
	switch {
	// Health check endpoints (no auth required)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/health":
		s.handleHealthCheck(w)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/health/ready":
		s.handleHealthReady(w)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/health/live":
		s.handleHealthLive(w)

	// Case management
	case r.Method == http.MethodPost && r.URL.Path == "/v1/cases":
		s.handleCreateCase(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/cases":
		s.handleListCases(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/cases/"):
		s.handleGetCase(w, r)
	case strings.HasPrefix(r.URL.Path, "/v1/cases/") && strings.HasSuffix(r.URL.Path, "/run"):
		if r.Method == http.MethodPost {
			s.handleRunCase(w, r)
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		}
	case strings.HasPrefix(r.URL.Path, "/v1/cases/") && strings.HasSuffix(r.URL.Path, "/events"):
		if r.Method == http.MethodGet {
			s.handleListEvents(w, r)
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		}

	// Approval management
	case r.Method == http.MethodGet && r.URL.Path == "/v1/approvals":
		s.handleListApprovals(w, r)
	case strings.HasPrefix(r.URL.Path, "/v1/approvals/") && strings.HasSuffix(r.URL.Path, "/approve"):
		if r.Method == http.MethodPost {
			s.handleResolveApproval(w, r, "approve")
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		}
	case strings.HasPrefix(r.URL.Path, "/v1/approvals/") && strings.HasSuffix(r.URL.Path, "/reject"):
		if r.Method == http.MethodPost {
			s.handleResolveApproval(w, r, "reject")
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		}

	// Report management
	case strings.HasPrefix(r.URL.Path, "/v1/reports/") && strings.HasSuffix(r.URL.Path, "/build"):
		if r.Method == http.MethodPost {
			s.handleBuildReport(w, r)
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		}

	// Session management
	case r.Method == http.MethodGet && r.URL.Path == "/v1/sessions":
		s.handleListSessions(w, r)

	// Device management
	case r.Method == http.MethodGet && r.URL.Path == "/v1/devices":
		s.handleListDevices(w, r)

	default:
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
	}
}

// handleHealthCheck returns server health status
func (s *Server) handleHealthCheck(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// handleHealthReady returns server readiness with database connectivity check
func (s *Server) handleHealthReady(w http.ResponseWriter) {
	if s.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "reason": "no_database"})
		return
	}
	if err := s.db.PingContext(context.Background()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "reason": "database_unreachable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

// handleHealthLive returns server liveness - if server is responding, it's alive
func (s *Server) handleHealthLive(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "live"})
}

// handleCreateCase creates a new case with proper input validation
func (s *Server) handleCreateCase(w http.ResponseWriter, r *http.Request) {
	// FIX 1: Proper request body size limiting (security fix)
	// Use MaxBytesReader BEFORE reading the body to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)
	defer r.Body.Close()

	var spec domain.CaseSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		// Check if it's a body size error
		if strings.Contains(err.Error(), "http: request body too large") {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"error": "request_body_too_large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json_format"})
		return
	}

	// Validate input
	if spec.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "title_required"})
		return
	}
	if len(spec.Commands) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "commands_required"})
		return
	}

	c, err := s.svc.CreateCase(r.Context(), spec)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// handleListCases lists all cases with pagination
func (s *Server) handleListCases(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	cases, nextCursor, hasMore, err := s.svc.ListCasesPaginated(r.Context(), cursor, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]dto.CaseResponse, len(cases))
	for i, c := range cases {
		items[i] = dto.ToCaseResponse(c)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":       items,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

// extractID extracts and validates a resource ID from the URL path.
// Returns the ID and true if valid, or "" and false if invalid.
func extractID(r *http.Request, prefix, suffix string) (string, bool) {
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	return id, id != "" && !strings.Contains(id, "..")
}

func (s *Server) handleGetCase(w http.ResponseWriter, r *http.Request) {
	id, ok := extractID(r, "/v1/cases/", "")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_case_id"})
		return
	}
	c, err := s.svc.GetCase(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	id, ok := extractID(r, "/v1/cases/", "/events")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_case_id"})
		return
	}
	events, err := s.svc.ListEvents(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleRunCase(w http.ResponseWriter, r *http.Request) {
	id, ok := extractID(r, "/v1/cases/", "/run")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_case_id"})
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	result, err := s.svc.RunCase(r.Context(), id, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleListApprovals(w http.ResponseWriter, r *http.Request) {
	approvals, err := s.svc.ListApprovals(r.Context(), r.URL.Query().Get("case_id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, approvals)
}

func (s *Server) handleResolveApproval(w http.ResponseWriter, r *http.Request, decision string) {
	id, ok := extractID(r, "/v1/approvals/"+decision+"/", "")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_approval_id"})
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	approval, err := s.svc.ResolveApproval(r.Context(), id, userID, decision, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, approval)
}

func (s *Server) handleBuildReport(w http.ResponseWriter, r *http.Request) {
	caseID, ok := extractID(r, "/v1/reports/", "/build")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_case_id"})
		return
	}
	report, err := s.svc.BuildReport(r.Context(), caseID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.svc.ListSessions(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.svc.ListDevices(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, devices)
}

// FIX 2: Proper error handling - don't expose internal error details
func writeError(w http.ResponseWriter, err error) {
	// Check for AppError
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		status := http.StatusInternalServerError
		switch {
		case appErr.Code >= 1000 && appErr.Code < 2000:
			status = http.StatusBadRequest
		case appErr.Code >= 2000 && appErr.Code < 3000:
			if appErr.Code == 2001 || appErr.Code == 2002 { // case not found
				status = http.StatusNotFound
			} else {
				status = http.StatusBadRequest
			}
		case appErr.Code >= 3000 && appErr.Code < 4000:
			if appErr.Code == 3001 { // approval not found
				status = http.StatusNotFound
			} else {
				status = http.StatusBadRequest
			}
		case appErr.Code >= 4000 && appErr.Code < 5000:
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]any{
			"code":    appErr.Code,
			"error":   appErr.Error(),
			"message": appErr.Message,
		})
		return
	}

	// Fallback for non-AppError
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "resource_not_found"})
		return
	}
	log.Printf("internal server error: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]any{
		"error":   "internal_server_error",
		"message": "An unexpected error occurred",
	})
}

// Helper function for consistent JSON responses
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}
