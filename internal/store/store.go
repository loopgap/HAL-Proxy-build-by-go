package store

import (
	"context"
	"errors"

	"hal-proxy/internal/domain"
)

var (
	ErrNotFound               = errors.New("not found")
	ErrConcurrentModification = errors.New("concurrent modification detected")
)

type Repository interface {
	Init(context.Context) error
	CreateCase(context.Context, domain.CaseRecord) error
	UpdateCase(context.Context, domain.CaseRecord) error
	GetCase(context.Context, string) (domain.CaseRecord, error)
	GetCaseWithRelations(context.Context, string) (domain.CaseWithRelations, error)
	ListCases(context.Context) ([]domain.CaseRecord, error)
	ListCasesPaginated(ctx context.Context, cursor string, limit int) ([]domain.CaseRecord, string, bool, error)
	AppendEvent(context.Context, domain.EventEnvelope) (domain.EventEnvelope, error)
	ListEvents(context.Context, string) ([]domain.EventEnvelope, error)
	ListEventsPaginated(ctx context.Context, caseID string, limit, offset int) ([]domain.EventEnvelope, int, error)
	CreateOrGetPendingApproval(context.Context, domain.Approval) (domain.Approval, error)
	GetApproval(context.Context, string) (domain.Approval, error)
	FindApprovalByCommand(context.Context, string, int) (domain.Approval, error)
	ListApprovals(context.Context, string) ([]domain.Approval, error)
	UpdateApproval(context.Context, domain.Approval) error
	CreateReport(context.Context, domain.ReportSummary) error
	GetLatestReport(context.Context, string) (domain.ReportSummary, error)
}
