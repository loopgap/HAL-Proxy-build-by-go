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

const timestampFormat = time.RFC3339Nano

type SQLiteRepository struct {
	db        *sql.DB
	Blacklist *TokenBlacklist
}

func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA cache_size=-64000",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply pragma: %w", err)
		}
	}

	blacklist := NewTokenBlacklist(db)

	return &SQLiteRepository{db: db, Blacklist: blacklist}, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// DB returns the underlying database connection for configuration purposes
func (r *SQLiteRepository) DB() *sql.DB {
	return r.db
}

func (r *SQLiteRepository) Init(ctx context.Context) error {
	// Initialize token blacklist
	if err := r.Blacklist.Init(ctx); err != nil {
		return fmt.Errorf("init token blacklist: %w", err)
	}

	// Create cases table with owner_id and version columns
	if _, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS cases (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			spec_json TEXT NOT NULL,
			next_command INTEGER NOT NULL,
			version INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`); err != nil {
		return fmt.Errorf("init cases table: %w", err)
	}

	// Migration: add version column if upgrading from older schema (pre-version field)
	if err := r.migrateAddVersionColumn(ctx); err != nil {
		return fmt.Errorf("migration add version: %w", err)
	}

	// Migration: add owner_id column if upgrading from older schema (pre-owner_id field)
	if err := r.migrateAddOwnerIDColumn(ctx); err != nil {
		return fmt.Errorf("migration add owner_id: %w", err)
	}

	stmts := []string{
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
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_approvals_case_command ON approvals(case_id, command_index);`,
		`CREATE INDEX IF NOT EXISTS idx_cases_status ON cases(status);`,
		`CREATE INDEX IF NOT EXISTS idx_cases_owner_id ON cases(owner_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reports_case_id ON reports(case_id);`,
		// Additional indexes for Wave 2 optimizations
		`CREATE INDEX IF NOT EXISTS idx_approvals_status ON approvals(status);`,
	}

	for _, stmt := range stmts {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
	}

	// Migration: add version column to approvals if upgrading from older schema
	if err := r.migrateAddApprovalVersionColumn(ctx); err != nil {
		return fmt.Errorf("migration add approval version: %w", err)
	}

	return nil
}

// migrateAddVersionColumn adds the version column to cases table if it doesn't exist
func (r *SQLiteRepository) migrateAddVersionColumn(ctx context.Context) error {
	// Check if version column already exists
	rows, err := r.db.QueryContext(ctx, `PRAGMA table_info(cases)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasVersion := false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt_value interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return err
		}
		if name == "version" {
			hasVersion = true
			break
		}
	}

	if !hasVersion {
		_, err = r.db.ExecContext(ctx, `ALTER TABLE cases ADD COLUMN version INTEGER NOT NULL DEFAULT 0`)
		if err != nil {
			return err
		}
	}
	return nil
}

// migrateAddOwnerIDColumn adds the owner_id column to cases table if it doesn't exist
func (r *SQLiteRepository) migrateAddOwnerIDColumn(ctx context.Context) error {
	// Check if owner_id column already exists
	rows, err := r.db.QueryContext(ctx, `PRAGMA table_info(cases)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasOwnerID := false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt_value interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return err
		}
		if name == "owner_id" {
			hasOwnerID = true
			break
		}
	}

	if !hasOwnerID {
		_, err = r.db.ExecContext(ctx, `ALTER TABLE cases ADD COLUMN owner_id TEXT NOT NULL DEFAULT ''`)
		if err != nil {
			return err
		}
	}
	return nil
}

// migrateAddApprovalVersionColumn adds the version column to approvals table if it doesn't exist
func (r *SQLiteRepository) migrateAddApprovalVersionColumn(ctx context.Context) error {
	// Check if version column already exists
	rows, err := r.db.QueryContext(ctx, `PRAGMA table_info(approvals)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasVersion := false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt_value interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return err
		}
		if name == "version" {
			hasVersion = true
			break
		}
	}

	if !hasVersion {
		_, err = r.db.ExecContext(ctx, `ALTER TABLE approvals ADD COLUMN version INTEGER NOT NULL DEFAULT 0`)
		if err != nil {
			return err
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
		`INSERT INTO cases (id, owner_id, title, status, spec_json, next_command, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.OwnerID, c.Title, c.Status, string(spec), c.NextCommand, 0, c.CreatedAt.Format(timestampFormat), c.UpdatedAt.Format(timestampFormat),
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

	// Optimistic locking: use version check in WHERE clause
	// Caller must increment c.Version before calling UpdateCase
	res, err := r.db.ExecContext(
		ctx,
		`UPDATE cases
		 SET title = ?, status = ?, spec_json = ?, next_command = ?, version = ?, updated_at = ?
		 WHERE id = ? AND version = ?`,
		c.Title, c.Status, string(spec), c.NextCommand, c.Version, c.UpdatedAt.Format(timestampFormat), c.ID, c.Version-1,
	)
	if err != nil {
		return fmt.Errorf("update case: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		// Check if case exists at all - if not, return ErrNotFound, otherwise ErrConcurrentModification
		existsQuery := `SELECT 1 FROM cases WHERE id = ?`
		var exists int
		if err := r.db.QueryRowContext(ctx, existsQuery, c.ID).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("UpdateCase: %w", err)
		}
		return ErrConcurrentModification
	}

	return nil
}

func (r *SQLiteRepository) GetCase(ctx context.Context, id string) (domain.CaseRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, owner_id, title, status, spec_json, next_command, version, created_at, updated_at FROM cases WHERE id = ?`, id)
	return scanCase(row)
}

func (r *SQLiteRepository) DeleteCase(ctx context.Context, id string) error {
	// Use transaction to ensure atomic deletion of case and related records
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete related events
	if _, err := tx.ExecContext(ctx, `DELETE FROM events WHERE case_id = ?`, id); err != nil {
		return fmt.Errorf("delete events: %w", err)
	}

	// Delete related approvals
	if _, err := tx.ExecContext(ctx, `DELETE FROM approvals WHERE case_id = ?`, id); err != nil {
		return fmt.Errorf("delete approvals: %w", err)
	}

	// Delete related reports
	if _, err := tx.ExecContext(ctx, `DELETE FROM reports WHERE case_id = ?`, id); err != nil {
		return fmt.Errorf("delete reports: %w", err)
	}

	// Delete the case
	result, err := tx.ExecContext(ctx, `DELETE FROM cases WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete case: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetCaseWithRelations(ctx context.Context, id string) (domain.CaseWithRelations, error) {
	c, err := r.GetCase(ctx, id)
	if err != nil {
		return domain.CaseWithRelations{}, err
	}

	events, err := r.ListEvents(ctx, id)
	if err != nil {
		return domain.CaseWithRelations{}, fmt.Errorf("list events: %w", err)
	}

	approvals, err := r.ListApprovals(ctx, id)
	if err != nil {
		return domain.CaseWithRelations{}, fmt.Errorf("list approvals: %w", err)
	}

	return domain.CaseWithRelations{
		Case:      c,
		Events:    events,
		Approvals: approvals,
	}, nil
}

func (r *SQLiteRepository) ListCases(ctx context.Context) ([]domain.CaseRecord, error) {
	// Limit to prevent unbounded memory growth - use ListCasesPaginated for large result sets
	rows, err := r.db.QueryContext(ctx, `SELECT id, owner_id, title, status, spec_json, next_command, version, created_at, updated_at FROM cases ORDER BY created_at ASC LIMIT 1000`)
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

func (r *SQLiteRepository) ListCasesPaginated(ctx context.Context, cursor string, limit int) ([]domain.CaseRecord, string, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `SELECT id, owner_id, title, status, spec_json, next_command, version, created_at, updated_at FROM cases ORDER BY created_at DESC LIMIT ?`
	args := []any{limit + 1} // Request one extra to check hasMore

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", false, err
	}
	defer rows.Close()

	var cases []domain.CaseRecord
	for rows.Next() {
		var c domain.CaseRecord
		var specJSON string
		err := rows.Scan(&c.ID, &c.OwnerID, &c.Title, &c.Status, &specJSON, &c.NextCommand, &c.Version, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, "", false, err
		}
		if err := json.Unmarshal([]byte(specJSON), &c.Spec); err != nil {
			return nil, "", false, err
		}
		cases = append(cases, c)
	}

	hasMore := len(cases) > limit
	if hasMore {
		cases = cases[:limit]
	}

	var nextCursor string
	if hasMore && len(cases) > 0 {
		lastCase := cases[len(cases)-1]
		nextCursor = lastCase.CreatedAt.Format(time.RFC3339Nano)
	}

	return cases, nextCursor, hasMore, nil
}

func (r *SQLiteRepository) AppendEvent(ctx context.Context, e domain.EventEnvelope) (domain.EventEnvelope, error) {
	res, err := r.db.ExecContext(
		ctx,
		`INSERT INTO events (case_id, type, payload_json, created_at) VALUES (?, ?, ?, ?)`,
		e.CaseID, e.Type, string(e.Payload), e.CreatedAt.Format(timestampFormat),
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

func (r *SQLiteRepository) ListEventsPaginated(ctx context.Context, caseID string, limit, offset int) ([]domain.EventEnvelope, int, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000 // Max limit to prevent abuse
	}

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM events WHERE case_id = ?", caseID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated events
	rows, err := r.db.QueryContext(ctx, `
		SELECT sequence, case_id, type, payload_json, created_at 
		FROM events 
		WHERE case_id = ? 
		ORDER BY sequence ASC 
		LIMIT ? OFFSET ?`, caseID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list events paginated: %w", err)
	}
	defer rows.Close()

	var out []domain.EventEnvelope
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}

	return out, total, rows.Err()
}

func (r *SQLiteRepository) CreateOrGetPendingApproval(ctx context.Context, a domain.Approval) (domain.Approval, error) {
	// Use INSERT ... ON CONFLICT to atomically handle race condition
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO approvals (id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at)
		VALUES (?, ?, ?, ?, ?, 'pending', '', '', NULL, ?)
		ON CONFLICT(case_id, command_index) DO UPDATE SET
			status = CASE 
				WHEN approvals.status = 'pending' THEN 'pending' 
				ELSE approvals.status 
			END
		RETURNING id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at`,
		a.ID, a.CaseID, a.CommandIndex, a.CommandName, string(a.RiskClass), now)

	if err != nil {
		return domain.Approval{}, fmt.Errorf("create approval: %w", err)
	}

	// Fetch the approval to return it
	return r.GetApproval(ctx, a.ID)
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
	// Optimistic locking: use version check in WHERE clause
	// Caller must increment a.Version before calling UpdateApproval
	res, err := r.db.ExecContext(
		ctx,
		`UPDATE approvals
		 SET status = ?, reason = ?, decided_by = ?, decided_at = ?, version = ?
		 WHERE id = ? AND version = ?`,
		a.Status, a.Reason, a.DecidedBy, nullableTime(a.DecidedAt), a.Version, a.ID, a.Version-1,
	)
	if err != nil {
		return fmt.Errorf("update approval: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		// Check if approval exists at all - if not, return ErrNotFound, otherwise ErrConcurrentModification
		existsQuery := `SELECT 1 FROM approvals WHERE id = ?`
		var exists int
		if err := r.db.QueryRowContext(ctx, existsQuery, a.ID).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("UpdateApproval: %w", err)
		}
		return ErrConcurrentModification
	}

	return nil
}

func (r *SQLiteRepository) CreateReport(ctx context.Context, rep domain.ReportSummary) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO reports (id, case_id, path, command_count, event_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		rep.ID, rep.CaseID, rep.Path, rep.CommandCount, rep.EventCount, rep.CreatedAt.Format(timestampFormat),
	)
	if err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) ListReports(ctx context.Context, caseID string) ([]domain.ReportSummary, error) {
	query := `SELECT id, case_id, path, command_count, event_count, created_at FROM reports`
	args := []any{}
	if caseID != "" {
		query += ` WHERE case_id = ?`
		args = append(args, caseID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	var reports []domain.ReportSummary
	for rows.Next() {
		rep, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}

	return reports, rows.Err()
}

func (r *SQLiteRepository) GetReport(ctx context.Context, id string) (domain.ReportSummary, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, case_id, path, command_count, event_count, created_at FROM reports WHERE id = ?`, id)
	return scanReport(row)
}

func (r *SQLiteRepository) GetLatestReport(ctx context.Context, caseID string) (domain.ReportSummary, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, case_id, path, command_count, event_count, created_at FROM reports WHERE case_id = ? ORDER BY created_at DESC LIMIT 1`, caseID)
	return scanReport(row)
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
	if err := s.Scan(&c.ID, &c.OwnerID, &c.Title, &status, &specJSON, &c.NextCommand, &c.Version, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.CaseRecord{}, ErrNotFound
		}
		return domain.CaseRecord{}, fmt.Errorf("scan case: %w", err)
	}
	c.Status = domain.CaseStatus(status)
	if err := json.Unmarshal([]byte(specJSON), &c.Spec); err != nil {
		return domain.CaseRecord{}, fmt.Errorf("unmarshal case spec: %w", err)
	}
	var err error
	c.CreatedAt, err = time.Parse(timestampFormat, createdAt)
	if err != nil {
		return domain.CaseRecord{}, fmt.Errorf("parse created_at: %w", err)
	}
	c.UpdatedAt, err = time.Parse(timestampFormat, updatedAt)
	if err != nil {
		return domain.CaseRecord{}, fmt.Errorf("parse updated_at: %w", err)
	}
	return c, nil
}

func scanEvent(s scanner) (domain.EventEnvelope, error) {
	var e domain.EventEnvelope
	var payloadJSON string
	var createdAt string
	var err error
	if err = s.Scan(&e.Sequence, &e.CaseID, &e.Type, &payloadJSON, &createdAt); err != nil {
		return domain.EventEnvelope{}, fmt.Errorf("scanEvent: %w", err)
	}
	e.Payload = json.RawMessage(payloadJSON)
	e.CreatedAt, err = time.Parse(timestampFormat, createdAt)
	if err != nil {
		return domain.EventEnvelope{}, fmt.Errorf("scanEvent: %w", err)
	}
	return e, nil
}

func scanReport(s scanner) (domain.ReportSummary, error) {
	var rep domain.ReportSummary
	var createdAt string
	if err := s.Scan(&rep.ID, &rep.CaseID, &rep.Path, &rep.CommandCount, &rep.EventCount, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ReportSummary{}, ErrNotFound
		}
		return domain.ReportSummary{}, fmt.Errorf("scan report: %w", err)
	}
	var err error
	rep.CreatedAt, err = time.Parse(timestampFormat, createdAt)
	if err != nil {
		return domain.ReportSummary{}, fmt.Errorf("parse report created_at: %w", err)
	}
	return rep, nil
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
		parsed, err := time.Parse(timestampFormat, decidedAt.String)
		if err == nil {
			a.DecidedAt = &parsed
		}
	}
	a.CreatedAt, _ = time.Parse(timestampFormat, createdAt)
	return a, nil
}

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(timestampFormat)
}

// txWrapper wraps sql.Tx to implement the store.Tx interface
type txWrapper struct {
	tx *sql.Tx
}

func (t *txWrapper) Commit() error {
	return t.tx.Commit()
}

func (t *txWrapper) Rollback() error {
	return t.tx.Rollback()
}

// BeginTx starts a new transaction
func (r *SQLiteRepository) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	return &txWrapper{tx: tx}, nil
}

// UpdateCaseInTx updates a case within a transaction
func (r *SQLiteRepository) UpdateCaseInTx(ctx context.Context, tx Tx, c domain.CaseRecord) error {
	wrapper, ok := tx.(*txWrapper)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}

	spec, err := json.Marshal(c.Spec)
	if err != nil {
		return fmt.Errorf("marshal case spec: %w", err)
	}

	res, err := wrapper.tx.ExecContext(
		ctx,
		`UPDATE cases
		 SET title = ?, status = ?, spec_json = ?, next_command = ?, version = ?, updated_at = ?
		 WHERE id = ? AND version = ?`,
		c.Title, c.Status, string(spec), c.NextCommand, c.Version, c.UpdatedAt.Format(timestampFormat), c.ID, c.Version-1,
	)
	if err != nil {
		return fmt.Errorf("update case: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		existsQuery := `SELECT 1 FROM cases WHERE id = ?`
		var exists int
		if err := wrapper.tx.QueryRowContext(ctx, existsQuery, c.ID).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("UpdateCaseInTx: %w", err)
		}
		return ErrConcurrentModification
	}

	return nil
}

// AppendEventInTx appends an event within a transaction
func (r *SQLiteRepository) AppendEventInTx(ctx context.Context, tx Tx, e domain.EventEnvelope) (domain.EventEnvelope, error) {
	wrapper, ok := tx.(*txWrapper)
	if !ok {
		return domain.EventEnvelope{}, fmt.Errorf("invalid transaction type")
	}

	res, err := wrapper.tx.ExecContext(
		ctx,
		`INSERT INTO events (case_id, type, payload_json, created_at) VALUES (?, ?, ?, ?)`,
		e.CaseID, e.Type, string(e.Payload), e.CreatedAt.Format(timestampFormat),
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

// FindApprovalByCommandInTx finds an approval by case ID and command index within a transaction
func (r *SQLiteRepository) FindApprovalByCommandInTx(ctx context.Context, tx Tx, caseID string, commandIndex int) (domain.Approval, error) {
	wrapper, ok := tx.(*txWrapper)
	if !ok {
		return domain.Approval{}, fmt.Errorf("invalid transaction type")
	}

	row := wrapper.tx.QueryRowContext(ctx, `SELECT id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at FROM approvals WHERE case_id = ? AND command_index = ? ORDER BY created_at DESC LIMIT 1`, caseID, commandIndex)
	return scanApproval(row)
}

// CreateOrGetPendingApprovalInTx creates or gets a pending approval within a transaction
func (r *SQLiteRepository) CreateOrGetPendingApprovalInTx(ctx context.Context, tx Tx, a domain.Approval) (domain.Approval, error) {
	wrapper, ok := tx.(*txWrapper)
	if !ok {
		return domain.Approval{}, fmt.Errorf("invalid transaction type")
	}

	now := time.Now().UTC()

	row := wrapper.tx.QueryRowContext(ctx, `
		INSERT INTO approvals (id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at)
		VALUES (?, ?, ?, ?, ?, 'pending', '', '', NULL, ?)
		ON CONFLICT(case_id, command_index) DO UPDATE SET
			status = CASE 
				WHEN approvals.status = 'pending' THEN 'pending' 
				ELSE approvals.status 
			END
		RETURNING id, case_id, command_index, command_name, risk_class, status, reason, decided_by, decided_at, created_at`,
		a.ID, a.CaseID, a.CommandIndex, a.CommandName, string(a.RiskClass), now)

	return scanApproval(row)
}

// CleanupExpiredTokens removes expired tokens from the blacklist
func (r *SQLiteRepository) CleanupExpiredTokens(ctx context.Context) error {
	return r.Blacklist.Cleanup(ctx)
}
