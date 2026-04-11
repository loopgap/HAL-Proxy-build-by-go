package api

import (
	"context"
	"encoding/json"
	"errors"
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

type Server struct {
	svc *core.Service
}

func NewServer(svc *core.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.route)
}

func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Add request timeout context
	ctx := r.Context()
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(r.Context(), RequestTimeout)
		defer cancel()
	}
	r = r.WithContext(ctx)

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/cases":
		s.handleCreateCase(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/cases/") && strings.HasSuffix(r.URL.Path, "/events"):
		s.handleListEvents(w, r)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/cases/") && strings.HasSuffix(r.URL.Path, ":run"):
		s.handleRunCase(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/cases/"):
		s.handleGetCase(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/approvals":
		s.handleListApprovals(w, r)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/approvals/") && strings.HasSuffix(r.URL.Path, ":approve"):
		s.handleResolveApproval(w, r, "approve")
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/approvals/") && strings.HasSuffix(r.URL.Path, ":reject"):
		s.handleResolveApproval(w, r, "reject")
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/reports/") && strings.HasSuffix(r.URL.Path, ":build"):
		s.handleBuildReport(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/sessions":
		s.handleListSessions(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/devices":
		s.handleListDevices(w, r)
	default:
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
	}
}

func (s *Server) handleCreateCase(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to prevent memory exhaustion
	if r.ContentLength > MaxBodySize {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"error": "request body too large"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)

	var spec domain.CaseSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON format"})
		return
	}
	c, err := s.svc.CreateCase(r.Context(), spec)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) handleGetCase(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/cases/")
	c, err := s.svc.GetCase(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/cases/"), "/events")
	events, err := s.svc.ListEvents(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleRunCase(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/cases/"), ":run")
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
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/approvals/"), ":"+decision)
	approval, err := s.svc.ResolveApproval(r.Context(), id, "daemon", decision, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, approval)
}

func (s *Server) handleBuildReport(w http.ResponseWriter, r *http.Request) {
	caseID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/reports/"), ":build")
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

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
		return
	}
	writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
