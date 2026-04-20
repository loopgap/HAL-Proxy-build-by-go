package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bridgeos/internal/core"
	"bridgeos/internal/domain"
	apperrors "bridgeos/internal/errors"
	"bridgeos/internal/metrics"
	"bridgeos/internal/store"
	"bridgeos/internal/version"

	"bridgeos/internal/api/middleware"
)

// MaxBodySize limits request body to 1MB for security
const MaxBodySize = 1 << 20 // 1MB

// RequestTimeout sets the maximum duration for request processing
const RequestTimeout = 30 * time.Second

// Server handles HTTP API requests with security best practices
type Server struct {
	svc            *core.Service
	healthChecker  healthChecker
	blacklist      tokenRevocationStore
	authMiddleware func(http.Handler) http.Handler
	trustedProxies []string
	localTrusted   bool
	localActor     string
	localRoles     []string
}

type healthChecker interface {
	PingContext(context.Context) error
}

type tokenRevocationStore interface {
	Add(context.Context, string, time.Time) error
	IsRevoked(context.Context, string) (bool, error)
}

// NewServer creates a new API server instance
func NewServer(svc *core.Service, healthChecker healthChecker, blacklist tokenRevocationStore, jwtSecret string, jwtExpiryHours int, jwtIssuer string, trustedProxies []string, localTrusted bool, localActor string, localRoles []string) *Server {
	jwtConfig := middleware.JWTConfig{
		Secret:          jwtSecret,
		ExpirationHours: jwtExpiryHours,
		Issuer:          jwtIssuer,
	}

	var authMW func(http.Handler) http.Handler
	if jwtConfig.Secret == "" {
		log.Fatal("FATAL: JWT secret is required. Set BRIDGEOS_JWT_SECRET or HAL_PROXY_JWT_SECRET.")
	}

	jwtAuth := middleware.NewJWTAuthenticator(jwtConfig)
	if blacklist != nil {
		jwtAuth.Blacklist = blacklist
	}
	authMW = jwtAuth.Middleware()

	return &Server{
		svc:            svc,
		healthChecker:  healthChecker,
		blacklist:      blacklist,
		authMiddleware: authMW,
		trustedProxies: trustedProxies,
		localTrusted:   localTrusted,
		localActor:     localActor,
		localRoles:     localRoles,
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
	case path == "/v1/health/ready":
		return "/v1/health/ready"
	case path == "/v1/health/live":
		return "/v1/health/live"
	case path == "/v1/cases":
		return "/v1/cases"
	case strings.HasPrefix(path, "/v1/cases/") && strings.HasSuffix(path, "/run"):
		return "/v1/cases/{id}/run"
	case strings.HasPrefix(path, "/v1/cases/") && strings.HasSuffix(path, ":run"):
		return "/v1/cases/{id}/run"
	case strings.HasPrefix(path, "/v1/cases/") && strings.HasSuffix(path, "/events"):
		return "/v1/cases/{id}/events"
	case strings.HasPrefix(path, "/v1/cases/"):
		return "/v1/cases/{id}"
	case path == "/v1/approvals":
		return "/v1/approvals"
	case strings.HasPrefix(path, "/v1/approvals/") && strings.HasSuffix(path, "/approve"):
		return "/v1/approvals/{id}/approve"
	case strings.HasPrefix(path, "/v1/approvals/") && strings.HasSuffix(path, ":approve"):
		return "/v1/approvals/{id}/approve"
	case strings.HasPrefix(path, "/v1/approvals/") && strings.HasSuffix(path, "/reject"):
		return "/v1/approvals/{id}/reject"
	case strings.HasPrefix(path, "/v1/approvals/") && strings.HasSuffix(path, ":reject"):
		return "/v1/approvals/{id}/reject"
	case path == "/v1/reports":
		return "/v1/reports"
	case strings.HasPrefix(path, "/v1/reports/") && strings.HasSuffix(path, "/content"):
		return "/v1/reports/{id}/content"
	case strings.HasPrefix(path, "/v1/reports/") && !strings.HasSuffix(path, "/build") && !strings.HasSuffix(path, ":build"):
		return "/v1/reports/{id}"
	case strings.HasPrefix(path, "/v1/reports/") && strings.HasSuffix(path, "/build"):
		return "/v1/reports/{case_id}/build"
	case strings.HasPrefix(path, "/v1/reports/") && strings.HasSuffix(path, ":build"):
		return "/v1/reports/{case_id}/build"
	case path == "/v1/sessions":
		return "/v1/sessions"
	case path == "/v1/devices":
		return "/v1/devices"
	case path == "/v1/auth/revoke":
		return "/v1/auth/revoke"
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
	if requiresAuth(r.URL.Path) {
		if s.shouldTrustLocalRequest(r) {
			r = s.withTrustedLocalClaims(r)
			s.routeInternal(w, r)
		} else {
			authHandler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				s.routeInternal(w, r)
			}))
			authHandler.ServeHTTP(w, r)
		}
	} else {
		// Health check endpoints (no auth required)
		s.routeInternal(w, r)
	}

	// Record metrics with normalized path to reduce cardinality
	duration := time.Since(start).Seconds()
	metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, statusCodeToString(metricsW.statusCode)).Inc()
	metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
}

func requiresAuth(path string) bool {
	switch path {
	case "/v1/health", "/v1/health/ready", "/v1/health/live":
		return false
	default:
		return true
	}
}

func (s *Server) shouldTrustLocalRequest(r *http.Request) bool {
	if !s.localTrusted {
		return false
	}
	if r.Header.Get("Authorization") != "" || r.Header.Get("X-API-Key") != "" {
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	host = strings.TrimSpace(host)
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (s *Server) withTrustedLocalClaims(r *http.Request) *http.Request {
	actor := strings.TrimSpace(r.Header.Get("X-BridgeOS-Actor"))
	if actor == "" {
		actor = s.localActor
	}
	if actor == "" {
		actor = "local-agent"
	}
	roles := append([]string(nil), s.localRoles...)
	if headerRoles := strings.TrimSpace(r.Header.Get("X-BridgeOS-Roles")); headerRoles != "" {
		roles = strings.Split(headerRoles, ",")
		for i := range roles {
			roles[i] = strings.TrimSpace(roles[i])
		}
	}
	claims := &middleware.Claims{
		UserID:   actor,
		Username: actor,
		Roles:    roles,
	}
	return r.WithContext(middleware.ContextWithClaims(r.Context(), claims))
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
	case strings.HasPrefix(r.URL.Path, "/v1/cases/") && (strings.HasSuffix(r.URL.Path, "/run") || strings.HasSuffix(r.URL.Path, ":run")):
		// Temporary compatibility path: keep `:run` during 0.2.x transition.
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
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/cases/"):
		s.handleGetCase(w, r)

	// Approval management
	case r.Method == http.MethodGet && r.URL.Path == "/v1/approvals":
		s.handleListApprovals(w, r)
	case strings.HasPrefix(r.URL.Path, "/v1/approvals/") && (strings.HasSuffix(r.URL.Path, "/approve") || strings.HasSuffix(r.URL.Path, ":approve")):
		// Temporary compatibility path: keep `:approve` during 0.2.x transition.
		if r.Method == http.MethodPost {
			s.handleResolveApproval(w, r, "approve")
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		}
	case strings.HasPrefix(r.URL.Path, "/v1/approvals/") && (strings.HasSuffix(r.URL.Path, "/reject") || strings.HasSuffix(r.URL.Path, ":reject")):
		// Temporary compatibility path: keep `:reject` during 0.2.x transition.
		if r.Method == http.MethodPost {
			s.handleResolveApproval(w, r, "reject")
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		}

	// Report management
	case r.Method == http.MethodGet && r.URL.Path == "/v1/reports":
		s.handleListReports(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/reports/") && strings.HasSuffix(r.URL.Path, "/content"):
		s.handleGetReportContent(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/reports/") && !strings.HasSuffix(r.URL.Path, "/build") && !strings.HasSuffix(r.URL.Path, ":build"):
		s.handleGetReport(w, r)
	case strings.HasPrefix(r.URL.Path, "/v1/reports/") && (strings.HasSuffix(r.URL.Path, "/build") || strings.HasSuffix(r.URL.Path, ":build")):
		// Temporary compatibility path: keep `:build` during 0.2.x transition.
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

	// Token revocation
	case r.Method == http.MethodPost && r.URL.Path == "/v1/auth/revoke":
		s.handleRevokeToken(w, r)

	default:
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
	}
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

// handleHealthCheck returns server health status
func (s *Server) handleHealthCheck(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "healthy",
		"name":    version.AppName,
		"version": version.Version,
	})
}

// handleHealthReady returns server readiness with database connectivity check
func (s *Server) handleHealthReady(w http.ResponseWriter) {
	if s.healthChecker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "reason": "no_database"})
		return
	}
	if err := s.healthChecker.PingContext(context.Background()); err != nil {
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

	userID := middleware.GetUserIDFromContext(r.Context())
	c, err := s.svc.CreateCase(r.Context(), spec, userID)
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

	userID := middleware.GetUserIDFromContext(r.Context())
	cases, nextCursor, hasMore, err := s.svc.ListCasesPaginated(r.Context(), cursor, limit, userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":       cases,
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

func extractActionID(r *http.Request, prefix string, suffixes ...string) (string, bool) {
	for _, suffix := range suffixes {
		if strings.HasSuffix(r.URL.Path, suffix) {
			return extractID(r, prefix, suffix)
		}
	}
	return "", false
}

func extractReportID(r *http.Request) (string, bool) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/reports/")
	if strings.HasSuffix(id, "/content") {
		id = strings.TrimSuffix(id, "/content")
	}
	return id, id != "" && !strings.Contains(id, "/") && !strings.Contains(id, "..")
}

func (s *Server) handleGetCase(w http.ResponseWriter, r *http.Request) {
	id, ok := extractID(r, "/v1/cases/", "")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_case_id"})
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	c, err := s.svc.GetCase(r.Context(), id, userID)
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

	// Parse pagination params
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	events, total, err := s.svc.ListEventsPaginated(r.Context(), id, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":  events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) handleRunCase(w http.ResponseWriter, r *http.Request) {
	id, ok := extractActionID(r, "/v1/cases/", "/run", ":run")
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
	id, ok := extractActionID(r, "/v1/approvals/", "/"+decision, ":"+decision)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_approval_id"})
		return
	}

	// Check if user has permission to approve/reject
	// Only admin or users with "approver" role can resolve approvals
	claims, ok := middleware.GetClaimsFromContext(r.Context())
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error":   "unauthorized",
			"message": "Missing authentication",
		})
		return
	}
	if !middleware.HasRole(claims, "admin") && !middleware.HasRole(claims, "approver") {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"error":   "forbidden",
			"message": "Only admins or approvers can resolve approvals",
		})
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
	caseID, ok := extractActionID(r, "/v1/reports/", "/build", ":build")
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

func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request) {
	reports, err := s.svc.ListReports(r.Context(), r.URL.Query().Get("case_id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reports)
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	reportID, ok := extractReportID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_report_id"})
		return
	}

	report, err := s.svc.GetReport(r.Context(), reportID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleGetReportContent(w http.ResponseWriter, r *http.Request) {
	reportID, ok := extractReportID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_report_id"})
		return
	}

	report, body, err := s.svc.GetReportContent(r.Context(), reportID)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(report.Path)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
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

// handleRevokeToken revokes the current user's token
func (s *Server) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaimsFromContext(r.Context())
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}

	// Reject tokens without jti - they cannot be properly revoked
	if claims.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "token_cannot_be_revoked"})
		return
	}

	if s.blacklist == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "blacklist_not_configured"})
		return
	}

	if err := s.blacklist.Add(r.Context(), claims.ID, claims.ExpiresAt.Time); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "revoked"})
}

// FIX 2: Proper error handling - don't expose internal error details
func writeError(w http.ResponseWriter, err error) {
	// Check for AppError
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		status := http.StatusBadRequest
		switch appErr.Code {
		case apperrors.ErrCodeUnauthorized:
			status = http.StatusUnauthorized
		case apperrors.ErrCodeForbidden:
			status = http.StatusForbidden
		case apperrors.ErrCodeConflict:
			status = http.StatusConflict
		case apperrors.ErrCodeTimeout:
			status = http.StatusGatewayTimeout
		}
		if appErr.Code == apperrors.ErrCodeCaseNotFound || appErr.Code == apperrors.ErrCodeApprovalNotFound || appErr.Code == apperrors.ErrCodeReportNotFound || appErr.Code == apperrors.ErrCodeReportContentMissing {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]any{
			"code":    appErr.Code,
			"error":   appErr.Message,
			"message": appErr.Message,
		})
		return
	}

	// Fallback for non-AppError
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "resource_not_found", "message": "Resource not found"})
		return
	}
	if errors.Is(err, store.ErrConcurrentModification) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "concurrent_modification", "message": "Concurrent modification detected"})
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		writeJSON(w, http.StatusGatewayTimeout, map[string]any{"error": "timeout", "message": "Request timed out"})
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
