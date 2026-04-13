package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hal-proxy/internal/domain"
	"hal-proxy/internal/errors"
	"hal-proxy/internal/policy"
	"hal-proxy/internal/store"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	repo         store.Repository
	artifactsDir string
	tracer       trace.Tracer
}

type RunResult struct {
	Case            domain.CaseRecord `json:"case"`
	Status          string            `json:"status"`
	PendingApproval *domain.Approval  `json:"pending_approval,omitempty"`
}

func NewService(repo store.Repository, artifactsDir string) *Service {
	return &Service{
		repo:         repo,
		artifactsDir: artifactsDir,
		tracer:       otel.Tracer("hal-proxy/core"),
	}
}

func (s *Service) Init(ctx context.Context) error {
	return s.repo.Init(ctx)
}

// ListCases returns all cases from the repository
func (s *Service) ListCases(ctx context.Context) ([]domain.CaseRecord, error) {
	ctx, span := s.tracer.Start(ctx, "service.list_cases")
	defer span.End()
	return s.repo.ListCases(ctx)
}

func (s *Service) ListCasesPaginated(ctx context.Context, cursor string, limit int) ([]domain.CaseRecord, string, bool, error) {
	return s.repo.ListCasesPaginated(ctx, cursor, limit)
}

func (s *Service) CreateCase(ctx context.Context, spec domain.CaseSpec) (domain.CaseRecord, error) {
	ctx, span := s.tracer.Start(ctx, "service.create_case")
	defer span.End()

	now := time.Now().UTC()
	c := domain.CaseRecord{
		ID:          newID("case"),
		Title:       spec.Title,
		Status:      domain.CaseStatusReady,
		Spec:        spec,
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if c.Title == "" {
		c.Title = c.ID
	}

	if err := s.repo.CreateCase(ctx, c); err != nil {
		return domain.CaseRecord{}, err
	}
	if err := s.appendEvent(ctx, c.ID, "bridge.case.created", map[string]any{
		"title":         c.Title,
		"command_count": len(c.Spec.Commands),
	}); err != nil {
		return domain.CaseRecord{}, err
	}

	return c, nil
}

func (s *Service) GetCase(ctx context.Context, id string) (domain.CaseRecord, error) {
	ctx, span := s.tracer.Start(ctx, "service.get_case")
	defer span.End()
	return s.repo.GetCase(ctx, id)
}

func (s *Service) ListEvents(ctx context.Context, caseID string) ([]domain.EventEnvelope, error) {
	ctx, span := s.tracer.Start(ctx, "service.list_events")
	defer span.End()
	return s.repo.ListEvents(ctx, caseID)
}

func (s *Service) ListApprovals(ctx context.Context, caseID string) ([]domain.Approval, error) {
	ctx, span := s.tracer.Start(ctx, "service.list_approvals")
	defer span.End()
	return s.repo.ListApprovals(ctx, caseID)
}

func (s *Service) ResolveApproval(ctx context.Context, approvalID, actor, decision, reason string) (domain.Approval, error) {
	ctx, span := s.tracer.Start(ctx, "service.resolve_approval")
	defer span.End()

	approval, err := s.repo.GetApproval(ctx, approvalID)
	if err != nil {
		return domain.Approval{}, err
	}

	now := time.Now().UTC()
	switch decision {
	case "approve":
		approval.Status = domain.ApprovalApproved
	case "reject":
		approval.Status = domain.ApprovalRejected
	default:
		return domain.Approval{}, fmt.Errorf("unknown approval decision %q", decision)
	}
	approval.DecidedBy = actor
	approval.DecidedAt = &now
	approval.Reason = reason

	if err := s.repo.UpdateApproval(ctx, approval); err != nil {
		return domain.Approval{}, err
	}

	if err := s.appendEvent(ctx, approval.CaseID, "bridge.approval.resolved", map[string]any{
		"approval_id":     approval.ID,
		"command_index":   approval.CommandIndex,
		"command_name":    approval.CommandName,
		"status":          approval.Status,
		"decided_by":      approval.DecidedBy,
		"decision_reason": approval.Reason,
	}); err != nil {
		return domain.Approval{}, err
	}

	c, err := s.repo.GetCase(ctx, approval.CaseID)
	if err == nil && c.Status == domain.CaseStatusPaused {
		if approval.Status == domain.ApprovalApproved {
			c.Status = domain.CaseStatusReady
		} else {
			c.Status = domain.CaseStatusRejected
		}
		c.UpdatedAt = now
		if err := s.repo.UpdateCase(ctx, c); err != nil {
			return domain.Approval{}, err
		}
	}

	return approval, nil
}

func (s *Service) RunCase(ctx context.Context, caseID, actor string) (RunResult, error) {
	ctx, span := s.tracer.Start(ctx, "service.run_case")
	defer span.End()

	c, err := s.repo.GetCase(ctx, caseID)
	if err != nil {
		return RunResult{}, err
	}
	if !c.Status.CanTransitionTo(domain.CaseStatusRunning) {
		if c.Status.IsTerminal() {
			return RunResult{}, errors.ErrCaseNotRunnable(c.ID, fmt.Sprintf("case is %s", c.Status))
		}
		return RunResult{}, errors.ErrCaseInvalidStatus(string(c.Status), string(domain.CaseStatusRunning))
	}

	c.Status = domain.CaseStatusRunning
	c.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateCase(ctx, c); err != nil {
		return RunResult{}, err
	}
	if err := s.appendEvent(ctx, c.ID, "bridge.case.run_requested", map[string]any{"actor": actor}); err != nil {
		return RunResult{}, err
	}

	for idx := c.NextCommand; idx < len(c.Spec.Commands); idx++ {
		cmd := c.Spec.Commands[idx]
		risk := policy.NormalizeRisk(cmd.RiskClass)

		if err := s.appendEvent(ctx, c.ID, "bridge.step.started", map[string]any{
			"command_index": idx,
			"command_name":  cmd.Name,
			"action":        cmd.Action,
			"risk_class":    risk,
		}); err != nil {
			return RunResult{}, err
		}

		if err := s.appendEvent(ctx, c.ID, "bridge.command.dispatched", map[string]any{
			"command_index": idx,
			"command_name":  cmd.Name,
			"action":        cmd.Action,
			"parameters":    cmd.Parameters,
			"actor":         actor,
		}); err != nil {
			return RunResult{}, err
		}

		if policy.RequiresApproval(risk) {
			approval, err := s.repo.FindApprovalByCommand(ctx, c.ID, idx)
			if err == nil {
				switch approval.Status {
				case domain.ApprovalPending:
					c.Status = domain.CaseStatusPaused
					c.UpdatedAt = time.Now().UTC()
					if err := s.repo.UpdateCase(ctx, c); err != nil {
						return RunResult{}, err
					}
					return RunResult{Case: c, Status: "awaiting_approval", PendingApproval: &approval}, nil
				case domain.ApprovalRejected:
					c.Status = domain.CaseStatusRejected
					c.UpdatedAt = time.Now().UTC()
					if err := s.repo.UpdateCase(ctx, c); err != nil {
						return RunResult{}, err
					}
					return RunResult{Case: c, Status: "rejected", PendingApproval: &approval}, nil
				case domain.ApprovalApproved:
					if err := s.appendEvent(ctx, c.ID, "bridge.approval.accepted", map[string]any{
						"approval_id":   approval.ID,
						"command_index": idx,
						"command_name":  cmd.Name,
					}); err != nil {
						return RunResult{}, err
					}
					if err := s.appendObservation(ctx, c, idx, cmd); err != nil {
						return RunResult{}, err
					}
					c.NextCommand = idx + 1
					c.UpdatedAt = time.Now().UTC()
					if err := s.repo.UpdateCase(ctx, c); err != nil {
						return RunResult{}, err
					}
					continue
				}
			}

			if err == nil || err == store.ErrNotFound {
				if err == store.ErrNotFound {
					pending := domain.Approval{
						ID:           newID("approval"),
						CaseID:       c.ID,
						CommandIndex: idx,
						CommandName:  cmd.Name,
						RiskClass:    risk,
						Status:       domain.ApprovalPending,
						CreatedAt:    time.Now().UTC(),
					}
					approval, err = s.repo.CreateOrGetPendingApproval(ctx, pending)
					if err != nil {
						return RunResult{}, err
					}
					if err := s.appendEvent(ctx, c.ID, "bridge.approval.requested", map[string]any{
						"approval_id":   approval.ID,
						"command_index": idx,
						"command_name":  cmd.Name,
						"risk_class":    risk,
					}); err != nil {
						return RunResult{}, err
					}
				}

				c.Status = domain.CaseStatusPaused
				c.UpdatedAt = time.Now().UTC()
				if err := s.repo.UpdateCase(ctx, c); err != nil {
					return RunResult{}, err
				}
				return RunResult{Case: c, Status: "awaiting_approval", PendingApproval: &approval}, nil
			}

			return RunResult{}, err
		}

		if err := s.appendObservation(ctx, c, idx, cmd); err != nil {
			return RunResult{}, err
		}

		c.NextCommand = idx + 1
		c.UpdatedAt = time.Now().UTC()
		if err := s.repo.UpdateCase(ctx, c); err != nil {
			return RunResult{}, err
		}
	}

	c.Status = domain.CaseStatusCompleted
	c.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateCase(ctx, c); err != nil {
		return RunResult{}, err
	}
	if err := s.appendEvent(ctx, c.ID, "bridge.case.completed", map[string]any{
		"completed_commands": c.NextCommand,
	}); err != nil {
		return RunResult{}, err
	}

	return RunResult{Case: c, Status: "completed"}, nil
}

func (s *Service) BuildReport(ctx context.Context, caseID string) (domain.ReportSummary, error) {
	ctx, span := s.tracer.Start(ctx, "service.build_report")
	defer span.End()

	caseWithRelations, err := s.repo.GetCaseWithRelations(ctx, caseID)
	if err != nil {
		return domain.ReportSummary{}, err
	}
	c := caseWithRelations.Case
	events := caseWithRelations.Events
	approvals := caseWithRelations.Approvals

	if err := os.MkdirAll(s.artifactsDir, 0o755); err != nil {
		return domain.ReportSummary{}, fmt.Errorf("create artifacts dir: %w", err)
	}

	rep := domain.ReportSummary{
		ID:           newID("report"),
		CaseID:       caseID,
		Path:         filepath.Join(s.artifactsDir, fmt.Sprintf("%s-report.md", caseID)),
		CommandCount: len(c.Spec.Commands),
		EventCount:   len(events),
		CreatedAt:    time.Now().UTC(),
	}

	body := reportMarkdown(c, events, approvals)
	if err := os.WriteFile(rep.Path, []byte(body), 0o644); err != nil {
		return domain.ReportSummary{}, fmt.Errorf("write report: %w", err)
	}
	if err := s.repo.CreateReport(ctx, rep); err != nil {
		return domain.ReportSummary{}, err
	}
	if err := s.appendEvent(ctx, c.ID, "bridge.report.generated", map[string]any{
		"report_id": rep.ID,
		"path":      rep.Path,
	}); err != nil {
		return domain.ReportSummary{}, err
	}

	return rep, nil
}

func (s *Service) ListDevices(ctx context.Context) ([]domain.DeviceDescriptor, error) {
	return []domain.DeviceDescriptor{
		{
			ID:           "dev-001",
			Name:         "Primary Bridge Controller",
			Capabilities: []string{"execute", "monitor", "configure"},
			SupportLevel: "full",
		},
		{
			ID:           "dev-002",
			Name:         "Secondary Bridge Controller",
			Capabilities: []string{"execute", "monitor"},
			SupportLevel: "partial",
		},
	}, nil
}

func (s *Service) ListSessions(ctx context.Context) ([]domain.SessionRecord, error) {
	return []domain.SessionRecord{
		{
			ID:        "sess-001",
			DeviceID:  "dev-001",
			Status:    "active",
			Owner:     "system",
			CreatedAt: time.Now().Add(-time.Hour),
		},
	}, nil
}

func (s *Service) appendObservation(ctx context.Context, c domain.CaseRecord, idx int, cmd domain.CaseCommandSpec) error {
	observation := map[string]any{
		"command_index": idx,
		"command_name":  cmd.Name,
		"action":        cmd.Action,
		"summary":       fmt.Sprintf("simulated %s", cmd.Action),
		"parameters":    cmd.Parameters,
	}
	return s.appendEvent(ctx, c.ID, "bridge.observation.recorded", observation)
}

func (s *Service) appendEvent(ctx context.Context, caseID, eventType string, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	_, err = s.repo.AppendEvent(ctx, domain.EventEnvelope{
		CaseID:    caseID,
		Type:      eventType,
		Payload:   raw,
		CreatedAt: time.Now().UTC(),
	})
	return err
}

func reportMarkdown(c domain.CaseRecord, events []domain.EventEnvelope, approvals []domain.Approval) string {
	lines := []string{
		"# HAL-Proxy Report",
		"",
		fmt.Sprintf("- Case ID: `%s`", c.ID),
		fmt.Sprintf("- Title: `%s`", c.Title),
		fmt.Sprintf("- Status: `%s`", c.Status),
		fmt.Sprintf("- Commands: `%d`", len(c.Spec.Commands)),
		fmt.Sprintf("- Events: `%d`", len(events)),
		"",
		"## Commands",
		"",
	}

	for i, cmd := range c.Spec.Commands {
		lines = append(lines, fmt.Sprintf("%d. `%s` `%s` risk=`%s`", i+1, cmd.Name, cmd.Action, policy.NormalizeRisk(cmd.RiskClass)))
	}

	lines = append(lines, "", "## Approvals", "")
	if len(approvals) == 0 {
		lines = append(lines, "- None")
	} else {
		for _, approval := range approvals {
			lines = append(lines, fmt.Sprintf("- `%s` command=%d `%s` status=`%s`", approval.ID, approval.CommandIndex, approval.CommandName, approval.Status))
		}
	}

	lines = append(lines, "", "## Event Timeline", "")
	for _, event := range events {
		lines = append(lines, fmt.Sprintf("- %s `%s`", event.CreatedAt.Format(time.RFC3339), event.Type))
	}

	return strings.Join(lines, "\n") + "\n"
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(buf))
}
