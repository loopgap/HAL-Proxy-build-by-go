package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bridgeos/internal/domain"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteRepository) Init(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS cases (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			spec_json TEXT NOT NULL,
			next_command INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS events (
			sequence INTEGER PRIMARY KEY AUTOINCREMENT,
			case_id TEXT NOT NULL,
			type TEXT NOT NULL,
			payload_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS approvals (
			id TEXT PRIMARY KEY,
			case_id TEXT NOT NULL,
			command_index INTEGER NOT NULL,
			command_name TEXT NOT NULL,
			risk_class TEXT NOT NULL,
			status TEXT NOT NULL,
			reason TEXT,
			decided_by TEXT,
			decided_at TEXT,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS reports (
			id TEXT PRIMARY KEY,
			case_id TEXT NOT NULL,
			path TEXT NOT NULL,
			command_count INTEGER NOT NULL,
			event_count INTEGER NOT NULL,
			created_at TEXT NOT NULL
		);`,
		// Performance indexes for faster queries
		`CREATE INDEX IF NOT EXISTS idx_events_case_id ON events(case_id);`,
		`CREATE INDEX IF NOT EXISTS idx_approvals_case_id ON approvals(case_id);`,
		`CREATE INDEX IF NOT EXISTS idx_approvals_case_command ON approvals(case_id, command_index);`,
		`CREATE INDEX IF NOT EXISTS idx_cases_status ON cases(status);`,
		`CREATE INDEX IF NOT EXISTS idx_reports_case_id ON reports(case_id);`,
	}

	for _, stmt := range stmts {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
	}

	return nil
}

func (r *SQLiteRepository) CreateCase(ctx context.Context, c domain.CaseRecord) error {
	spec, err := json.Marshal(c.Spec)
	if err != nil {
		return fmt.Errorf("marshal case spec: %w", err)
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO cases (id, title, status, spec_json, next_command, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Title, c.Status, string(spec), c.NextCommand, c.CreatedAt.Format(time.RFC3339Nano), c.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert case: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) UpdateCase(ctx context.Context, c domain.CaseRecord) error {
	spec, err := json.Marshal(c.Spec)
	if err != nil {
		return fmt.Errorf("marshal case spec: %w", err)
	}

	res, err := r.db.ExecContext(
		ctx,
		`UPDATE cases
		 SET title = ?, status = ?, spec_json = ?, next_command = ?, updated_at = ?
		 WHERE id = ?`,
		c.Title, c.Status, string(spec), c.NextCommand, c.UpdatedAt.Format(time.RFC3339Nano), c.ID,
	)
	if err != nil {
		return fmt.Errorf("update case: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SQLiteRepository) GetCase(ctx context.Context, id string) (domain.CaseRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, title, status, spec_json, next_command, created_at, updated_at FROM cases WHERE id = ?`, id)
	return scanCase(row)
}

func (r *SQLiteRepository) ListCases(ctx context.Context) ([]domain.CaseRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, title, status, spec_json, next_command, created_at, updated_at FROM cases ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("list cases: %w", err)
	}
	defer rows.Close()

	var out []domain.CaseRecord
	for rows.Next() {
		c, err := scanCase(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}

	return out, rows.Err()
}

func (r *SQLiteRepository) AppendEvent(ctx context.Context, e domain.EventEnvelope) (domain.EventEnvelope, error) {
	res, err := r.db.ExecContext(
		ctx,
		`INSERT INTO events (case_id, type, payload_json, created_at) VALUES (?, ?, ?, ?)`,
		e.CaseID, e.Type, string(e.Payload), e.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return domain.EventEnvelope{}, fmt.Errorf("append event: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return domain.EventEnvelope{}, fmt.Errorf("event id: %w", err)
	}
	e.Sequence = id
	return e, nil
}

func (r *SQLiteRepository) ListEvents(ctx context.Context, caseID string) ([]domain.EventEnvelope, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT sequence, case_id, type, payload_json, created_at FROM events WHERE case_id = ? ORDER BY sequence ASC`,
		caseID,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var out []domain.EventEnvelope
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}

	return out, rows.Err()
}

func (r *SQLiteRepository) CreateOrGetPendingApproval(ctx context.Context, a domain.Approval) (domain.Approval, error) {
	existing, err := r.FindApprovalByCommand(ctx, a.CaseID, a.CommandIndex)
	if err == nil && existing.Status == domain.ApprovalPending {
		return existing, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return domain.Approval{}, err
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO approvals (id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.CaseID, a.CommandIndex, a.CommandName, a.RiskClass, a.Status, a.Reason, a.DecidedBy, nullableTime(a.DecidedAt), a.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return domain.Approval{}, fmt.Errorf("insert approval: %w", err)
	}
	return a, nil
}

func (r *SQLiteRepository) GetApproval(ctx context.Context, id string) (domain.Approval, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at FROM approvals WHERE id = ?`, id)
	return scanApproval(row)
}

func (r *SQLiteRepository) FindApprovalByCommand(ctx context.Context, caseID string, commandIndex int) (domain.Approval, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at FROM approvals WHERE case_id = ? AND command_index = ? ORDER BY created_at DESC LIMIT 1`, caseID, commandIndex)
	return scanApproval(row)
}

func (r *SQLiteRepository) ListApprovals(ctx context.Context, caseID string) ([]domain.Approval, error) {
	query := `SELECT id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at FROM approvals`
	var args []any
	if caseID != "" {
		query += ` WHERE case_id = ?`
		args = append(args, caseID)
	}
	query += ` ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list approvals: %w", err)
	}
	defer rows.Close()

	var out []domain.Approval
	for rows.Next() {
		a, err := scanApproval(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}

	return out, rows.Err()
}

func (r *SQLiteRepository) UpdateApproval(ctx context.Context, a domain.Approval) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE approvals
		 SET status = ?, reason = ?, decided_by = ?, decided_at = ?
		 WHERE id = ?`,
		a.Status, a.Reason, a.DecidedBy, nullableTime(a.DecidedAt), a.ID,
	)
	if err != nil {
		return fmt.Errorf("update approval: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) CreateReport(ctx context.Context, rep domain.ReportSummary) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO reports (id, case_id, path, command_count, event_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		rep.ID, rep.CaseID, rep.Path, rep.CommandCount, rep.EventCount, rep.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetLatestReport(ctx context.Context, caseID string) (domain.ReportSummary, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, case_id, path, command_count, event_count, created_at FROM reports WHERE case_id = ? ORDER BY created_at DESC LIMIT 1`, caseID)
	var rep domain.ReportSummary
	var createdAt string
	if err := row.Scan(&rep.ID, &rep.CaseID, &rep.Path, &rep.CommandCount, &rep.EventCount, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ReportSummary{}, ErrNotFound
		}
		return domain.ReportSummary{}, fmt.Errorf("scan report: %w", err)
	}
	rep.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return rep, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCase(s scanner) (domain.CaseRecord, error) {
	var c domain.CaseRecord
	var status string
	var specJSON string
	var createdAt string
	var updatedAt string
	if err := s.Scan(&c.ID, &c.Title, &status, &specJSON, &c.NextCommand, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.CaseRecord{}, ErrNotFound
		}
		return domain.CaseRecord{}, fmt.Errorf("scan case: %w", err)
	}
	c.Status = domain.CaseStatus(status)
	if err := json.Unmarshal([]byte(specJSON), &c.Spec); err != nil {
		return domain.CaseRecord{}, fmt.Errorf("unmarshal case spec: %w", err)
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return c, nil
}

func scanEvent(s scanner) (domain.EventEnvelope, error) {
	var e domain.EventEnvelope
	var payloadJSON string
	var createdAt string
	if err := s.Scan(&e.Sequence, &e.CaseID, &e.Type, &payloadJSON, &createdAt); err != nil {
		return domain.EventEnvelope{}, fmt.Errorf("scan event: %w", err)
	}
	e.Payload = json.RawMessage(payloadJSON)
	e.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return e, nil
}

func scanApproval(s scanner) (domain.Approval, error) {
	var a domain.Approval
	var risk string
	var status string
	var decidedAt sql.NullString
	var createdAt string
	if err := s.Scan(&a.ID, &a.CaseID, &a.CommandIndex, &a.CommandName, &risk, &status, &a.Reason, &a.DecidedBy, &decidedAt, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Approval{}, ErrNotFound
		}
		return domain.Approval{}, fmt.Errorf("scan approval: %w", err)
	}
	a.RiskClass = domain.RiskClass(risk)
	a.Status = domain.ApprovalStatus(status)
	if decidedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, decidedAt.String)
		if err == nil {
			a.DecidedAt = &parsed
		}
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return a, nil
}

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339Nano)
}
