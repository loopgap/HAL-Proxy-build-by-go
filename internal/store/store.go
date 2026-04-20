package store

import (
	"context"
	"errors"

	"bridgeos/internal/domain"
)

var (
	ErrNotFound               = errors.New("not found")
	ErrConcurrentModification = errors.New("concurrent modification detected")
)

// Tx represents a database transaction
type Tx interface {
	Commit() error
	Rollback() error
}

// Repository defines the interface for data persistence
type Repository interface {
	Init(context.Context) error
	Close() error
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
	ListReports(context.Context, string) ([]domain.ReportSummary, error)
	GetReport(context.Context, string) (domain.ReportSummary, error)
	GetLatestReport(context.Context, string) (domain.ReportSummary, error)

	// Transaction support
	BeginTx(context.Context) (Tx, error)

	// Transaction-aware variants for atomic operations
	UpdateCaseInTx(ctx context.Context, tx Tx, c domain.CaseRecord) error
	AppendEventInTx(ctx context.Context, tx Tx, e domain.EventEnvelope) (domain.EventEnvelope, error)
	FindApprovalByCommandInTx(ctx context.Context, tx Tx, caseID string, commandIndex int) (domain.Approval, error)
	CreateOrGetPendingApprovalInTx(ctx context.Context, tx Tx, a domain.Approval) (domain.Approval, error)

	// DeleteCase removes a case and its associated events and approvals
	DeleteCase(ctx context.Context, id string) error
}
