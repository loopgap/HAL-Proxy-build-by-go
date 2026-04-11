package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"bridgeos/internal/core"
	"bridgeos/internal/domain"
	"bridgeos/internal/store"
)

// MaxBodySize limits request body to 1MB for security
const MaxBodySize = 1 << 20 // 1MB

// RequestTimeout sets the maximum duration for request processing
const RequestTimeout = 30 * 1e9 // 30 seconds in nanoseconds

// Server handles HTTP API requests with security best practices
type Server struct {
	svc *core.Service
	// In production, this should be populated with actual authentication middleware
	authMiddleware func(http.Handler) http.Handler
}

// NewServer creates a new API server instance
func NewServer(svc *core.Service) *Server {
	return &Server{
		svc: svc,
		// Default auth middleware (placeholder - should be replaced with actual auth)
		authMiddleware: func(next http.Handler) http.Handler { return next },
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

func (s *Server) route(w http.ResponseWriter, r *http.Request) {
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

	// Route handling - improved RESTful design
	switch {
	// Health check endpoint (no auth required)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/health":
		s.handleHealthCheck(w)

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
		"status": "healthy",
		"version": "1.0.0",
	})
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

// handleListCases lists all cases
func (s *Server) handleListCases(w http.ResponseWriter, r *http.Request) {
	cases, err := s.svc.ListCases(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cases)
}

func (s *Server) handleGetCase(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/cases/")
	// Validate ID format (prevent path traversal)
	if strings.Contains(id, "..") || id == "" {
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
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/cases/"), "/events")
	// Validate ID
	if strings.Contains(id, "..") || id == "" {
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
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/cases/"), "/run")
	// Validate ID
	if strings.Contains(id, "..") || id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_case_id"})
		return
	}
	result, err := s.svc.RunCase(r.Context(), id, "daemon")
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
	id := strings.TrimPrefix(r.URL.Path, "/v1/approvals/"+decision+"/")
	// Validate ID
	if strings.Contains(id, "..") || id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_approval_id"})
		return
	}
	approval, err := s.svc.ResolveApproval(r.Context(), id, "daemon", decision, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, approval)
}

func (s *Server) handleBuildReport(w http.ResponseWriter, r *http.Request) {
	caseID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/reports/"), "/build")
	// Validate ID
	if strings.Contains(caseID, "..") || caseID == "" {
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
	// Check for specific known errors and return user-friendly messages
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "resource_not_found"})
		return
	}

	// For unknown errors, return a generic message
	// In production, log the actual error internally
	writeJSON(w, http.StatusInternalServerError, map[string]any{
		"error": "internal_server_error",
		"message": "An unexpected error occurred",
	})
}

// Helper function for consistent JSON responses
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Log encoding error but don't expose details
		fmt.Printf("JSON encoding error: %v\n", err)
	}
}
