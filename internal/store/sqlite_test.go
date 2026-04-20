package store

import (
	"context"
	"os"
	"testing"
	"time"

	"bridgeos/internal/domain"
)

func setupTestDB(t *testing.T) (*SQLiteRepository, func()) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "bridgeos-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	repo, err := NewSQLiteRepository(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Initialize the schema
	ctx := context.Background()
	if err := repo.Init(ctx); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to init repository: %v", err)
	}

	cleanup := func() {
		repo.Close()
		os.Remove(tmpFile.Name())
	}

	return repo, cleanup
}

func TestSQLiteRepositoryInit(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Test that Init doesn't return an error (already initialized in setupTestDB)
	ctx := context.Background()
	err := repo.Init(ctx)
	if err != nil {
		t.Errorf("Init should not return error on already initialized db: %v", err)
	}
}

func TestCreateAndGetCase(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	caseRecord := domain.CaseRecord{
		ID:     "test-case-001",
		Title:  "Test Case",
		Status: domain.CaseStatusReady,
		Spec: domain.CaseSpec{
			Title: "Test Case",
			Commands: []domain.CaseCommandSpec{
				{
					Name:      "test-cmd",
					Action:    "test",
					RiskClass: domain.RiskObserve,
				},
			},
		},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create case
	err := repo.CreateCase(ctx, caseRecord)
	if err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Get case
	retrieved, err := repo.GetCase(ctx, "test-case-001")
	if err != nil {
		t.Fatalf("Failed to get case: %v", err)
	}

	if retrieved.ID != caseRecord.ID {
		t.Errorf("Case ID mismatch: got %v, want %v", retrieved.ID, caseRecord.ID)
	}

	if retrieved.Title != caseRecord.Title {
		t.Errorf("Case Title mismatch: got %v, want %v", retrieved.Title, caseRecord.Title)
	}

	if retrieved.Status != caseRecord.Status {
		t.Errorf("Case Status mismatch: got %v, want %v", retrieved.Status, caseRecord.Status)
	}

	if retrieved.NextCommand != caseRecord.NextCommand {
		t.Errorf("Case NextCommand mismatch: got %v, want %v", retrieved.NextCommand, caseRecord.NextCommand)
	}
}

func TestUpdateCase(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	caseRecord := domain.CaseRecord{
		ID:     "test-case-002",
		Title:  "Original Title",
		Status: domain.CaseStatusReady,
		Spec: domain.CaseSpec{
			Title:    "Original Title",
			Commands: []domain.CaseCommandSpec{},
		},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create case
	err := repo.CreateCase(ctx, caseRecord)
	if err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Update case
	caseRecord.Title = "Updated Title"
	caseRecord.Status = domain.CaseStatusRunning
	caseRecord.NextCommand = 5
	caseRecord.UpdatedAt = time.Now().UTC()
	caseRecord.Version++

	err = repo.UpdateCase(ctx, caseRecord)
	if err != nil {
		t.Fatalf("Failed to update case: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetCase(ctx, "test-case-002")
	if err != nil {
		t.Fatalf("Failed to get case: %v", err)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("Title not updated: got %v, want %v", retrieved.Title, "Updated Title")
	}

	if retrieved.Status != domain.CaseStatusRunning {
		t.Errorf("Status not updated: got %v, want %v", retrieved.Status, domain.CaseStatusRunning)
	}

	if retrieved.NextCommand != 5 {
		t.Errorf("NextCommand not updated: got %v, want %v", retrieved.NextCommand, 5)
	}
}

func TestGetCaseNotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, err := repo.GetCase(ctx, "non-existent-id")

	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestUpdateCaseNotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	caseRecord := domain.CaseRecord{
		ID:          "non-existent-id",
		Title:       "Test",
		Status:      domain.CaseStatusReady,
		Spec:        domain.CaseSpec{},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := repo.UpdateCase(ctx, caseRecord)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestListCases(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// Create multiple cases
	for i := 0; i < 3; i++ {
		caseRecord := domain.CaseRecord{
			ID:          "test-case-list-" + string(rune('a'+i)),
			Title:       "Test Case",
			Status:      domain.CaseStatusReady,
			Spec:        domain.CaseSpec{},
			NextCommand: 0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := repo.CreateCase(ctx, caseRecord); err != nil {
			t.Fatalf("Failed to create case: %v", err)
		}
	}

	// List cases
	cases, err := repo.ListCases(ctx)
	if err != nil {
		t.Fatalf("Failed to list cases: %v", err)
	}

	if len(cases) != 3 {
		t.Errorf("Expected 3 cases, got %d", len(cases))
	}
}

func TestAppendAndListEvents(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// Create a case first
	caseRecord := domain.CaseRecord{
		ID:          "test-case-events",
		Title:       "Test Case",
		Status:      domain.CaseStatusReady,
		Spec:        domain.CaseSpec{},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateCase(ctx, caseRecord); err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Append events
	for i := 0; i < 3; i++ {
		event := domain.EventEnvelope{
			CaseID:    "test-case-events",
			Type:      "test.event",
			Payload:   []byte(`{"index": ` + string(rune('0'+i)) + `}`),
			CreatedAt: now,
		}
		_, err := repo.AppendEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to append event: %v", err)
		}
	}

	// List events
	events, err := repo.ListEvents(ctx, "test-case-events")
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Verify event order
	for i, event := range events {
		if event.Type != "test.event" {
			t.Errorf("Event type mismatch at index %d", i)
		}
	}
}

func TestCreateAndGetApproval(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// Create a case first
	caseRecord := domain.CaseRecord{
		ID:          "test-case-approval",
		Title:       "Test Case",
		Status:      domain.CaseStatusReady,
		Spec:        domain.CaseSpec{},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateCase(ctx, caseRecord); err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Create approval
	approval := domain.Approval{
		ID:           "approval-001",
		CaseID:       "test-case-approval",
		CommandIndex: 0,
		CommandName:  "test-cmd",
		RiskClass:    domain.RiskDestructive,
		Status:       domain.ApprovalPending,
		CreatedAt:    now,
	}

	createdApproval, err := repo.CreateOrGetPendingApproval(ctx, approval)
	if err != nil {
		t.Fatalf("Failed to create approval: %v", err)
	}

	if createdApproval.ID != approval.ID {
		t.Errorf("Approval ID mismatch: got %v, want %v", createdApproval.ID, approval.ID)
	}

	// Get approval
	retrieved, err := repo.GetApproval(ctx, "approval-001")
	if err != nil {
		t.Fatalf("Failed to get approval: %v", err)
	}

	if retrieved.Status != domain.ApprovalPending {
		t.Errorf("Approval status mismatch: got %v, want %v", retrieved.Status, domain.ApprovalPending)
	}
}

func TestUpdateApproval(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// Create a case first
	caseRecord := domain.CaseRecord{
		ID:          "test-case-update-approval",
		Title:       "Test Case",
		Status:      domain.CaseStatusReady,
		Spec:        domain.CaseSpec{},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateCase(ctx, caseRecord); err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Create approval
	approval := domain.Approval{
		ID:           "approval-update-001",
		CaseID:       "test-case-update-approval",
		CommandIndex: 0,
		CommandName:  "test-cmd",
		RiskClass:    domain.RiskDestructive,
		Status:       domain.ApprovalPending,
		CreatedAt:    now,
	}
	_, err := repo.CreateOrGetPendingApproval(ctx, approval)
	if err != nil {
		t.Fatalf("Failed to create approval: %v", err)
	}

	// Update approval
	decidedAt := time.Now().UTC()
	approval.Status = domain.ApprovalApproved
	approval.DecidedBy = "test-user"
	approval.DecidedAt = &decidedAt
	approval.Reason = "Test approval"
	approval.Version++

	err = repo.UpdateApproval(ctx, approval)
	if err != nil {
		t.Fatalf("Failed to update approval: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetApproval(ctx, "approval-update-001")
	if err != nil {
		t.Fatalf("Failed to get approval: %v", err)
	}

	if retrieved.Status != domain.ApprovalApproved {
		t.Errorf("Approval status not updated: got %v, want %v", retrieved.Status, domain.ApprovalApproved)
	}

	if retrieved.DecidedBy != "test-user" {
		t.Errorf("Approval decided_by not updated: got %v, want %v", retrieved.DecidedBy, "test-user")
	}
}

func TestListApprovals(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// Create a case first
	caseRecord := domain.CaseRecord{
		ID:          "test-case-list-approvals",
		Title:       "Test Case",
		Status:      domain.CaseStatusReady,
		Spec:        domain.CaseSpec{},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateCase(ctx, caseRecord); err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Create multiple approvals
	for i := 0; i < 3; i++ {
		approval := domain.Approval{
			ID:           "approval-list-" + string(rune('a'+i)),
			CaseID:       "test-case-list-approvals",
			CommandIndex: i,
			CommandName:  "test-cmd",
			RiskClass:    domain.RiskDestructive,
			Status:       domain.ApprovalPending,
			CreatedAt:    now,
		}
		_, err := repo.CreateOrGetPendingApproval(ctx, approval)
		if err != nil {
			t.Fatalf("Failed to create approval: %v", err)
		}
	}

	// List approvals for case
	approvals, err := repo.ListApprovals(ctx, "test-case-list-approvals")
	if err != nil {
		t.Fatalf("Failed to list approvals: %v", err)
	}

	if len(approvals) != 3 {
		t.Errorf("Expected 3 approvals, got %d", len(approvals))
	}

	// List all approvals
	allApprovals, err := repo.ListApprovals(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list all approvals: %v", err)
	}

	if len(allApprovals) < 3 {
		t.Errorf("Expected at least 3 approvals, got %d", len(allApprovals))
	}
}

func TestCreateAndGetReport(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// Create a case first
	caseRecord := domain.CaseRecord{
		ID:          "test-case-report",
		Title:       "Test Case",
		Status:      domain.CaseStatusReady,
		Spec:        domain.CaseSpec{},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateCase(ctx, caseRecord); err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Create report
	report := domain.ReportSummary{
		ID:           "report-001",
		CaseID:       "test-case-report",
		Path:         "/artifacts/test-report.md",
		CommandCount: 5,
		EventCount:   10,
		CreatedAt:    now,
	}

	err := repo.CreateReport(ctx, report)
	if err != nil {
		t.Fatalf("Failed to create report: %v", err)
	}

	// Get latest report
	retrieved, err := repo.GetLatestReport(ctx, "test-case-report")
	if err != nil {
		t.Fatalf("Failed to get latest report: %v", err)
	}

	if retrieved.ID != report.ID {
		t.Errorf("Report ID mismatch: got %v, want %v", retrieved.ID, report.ID)
	}

	if retrieved.CommandCount != report.CommandCount {
		t.Errorf("Report CommandCount mismatch: got %v, want %v", retrieved.CommandCount, report.CommandCount)
	}
}

func TestGetLatestReportNotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, err := repo.GetLatestReport(ctx, "non-existent-case")

	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestListReports(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	caseA := domain.CaseRecord{
		ID:          "case-reports-a",
		OwnerID:     "owner-a",
		Title:       "Case A",
		Status:      domain.CaseStatusCompleted,
		Spec:        domain.CaseSpec{Title: "Case A"},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	caseB := domain.CaseRecord{
		ID:          "case-reports-b",
		OwnerID:     "owner-b",
		Title:       "Case B",
		Status:      domain.CaseStatusCompleted,
		Spec:        domain.CaseSpec{Title: "Case B"},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateCase(ctx, caseA); err != nil {
		t.Fatalf("Failed to create case A: %v", err)
	}
	if err := repo.CreateCase(ctx, caseB); err != nil {
		t.Fatalf("Failed to create case B: %v", err)
	}

	reports := []domain.ReportSummary{
		{
			ID:           "report-old",
			CaseID:       caseA.ID,
			Path:         "/artifacts/old.md",
			CommandCount: 1,
			EventCount:   2,
			CreatedAt:    now.Add(-2 * time.Hour),
		},
		{
			ID:           "report-new",
			CaseID:       caseA.ID,
			Path:         "/artifacts/new.md",
			CommandCount: 3,
			EventCount:   4,
			CreatedAt:    now.Add(-1 * time.Hour),
		},
		{
			ID:           "report-other",
			CaseID:       caseB.ID,
			Path:         "/artifacts/other.md",
			CommandCount: 5,
			EventCount:   6,
			CreatedAt:    now,
		},
	}
	for _, report := range reports {
		if err := repo.CreateReport(ctx, report); err != nil {
			t.Fatalf("Failed to create report %s: %v", report.ID, err)
		}
	}

	allReports, err := repo.ListReports(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list all reports: %v", err)
	}
	if len(allReports) != 3 {
		t.Fatalf("Expected 3 reports, got %d", len(allReports))
	}
	if allReports[0].ID != "report-other" || allReports[1].ID != "report-new" || allReports[2].ID != "report-old" {
		t.Fatalf("Reports not sorted by created_at desc: %+v", allReports)
	}

	caseReports, err := repo.ListReports(ctx, caseA.ID)
	if err != nil {
		t.Fatalf("Failed to list case reports: %v", err)
	}
	if len(caseReports) != 2 {
		t.Fatalf("Expected 2 reports for case A, got %d", len(caseReports))
	}
	if caseReports[0].ID != "report-new" || caseReports[1].ID != "report-old" {
		t.Fatalf("Filtered reports not sorted desc: %+v", caseReports)
	}
}

func TestGetReport(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	caseRecord := domain.CaseRecord{
		ID:          "case-report-by-id",
		OwnerID:     "owner-a",
		Title:       "Case Report ID",
		Status:      domain.CaseStatusCompleted,
		Spec:        domain.CaseSpec{Title: "Case Report ID"},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateCase(ctx, caseRecord); err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	report := domain.ReportSummary{
		ID:           "report-by-id",
		CaseID:       caseRecord.ID,
		Path:         "/artifacts/by-id.md",
		CommandCount: 2,
		EventCount:   3,
		CreatedAt:    now,
	}
	if err := repo.CreateReport(ctx, report); err != nil {
		t.Fatalf("Failed to create report: %v", err)
	}

	retrieved, err := repo.GetReport(ctx, report.ID)
	if err != nil {
		t.Fatalf("Failed to get report by id: %v", err)
	}
	if retrieved.ID != report.ID {
		t.Fatalf("Expected report id %s, got %s", report.ID, retrieved.ID)
	}

	if _, err := repo.GetReport(ctx, "missing-report"); err != ErrNotFound {
		t.Fatalf("Expected ErrNotFound for missing report, got %v", err)
	}
}

func TestClose(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "bridgeos-close-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	repo, err := NewSQLiteRepository(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Close should not panic
	err = repo.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

func TestFindApprovalByCommand(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// Create a case first
	caseRecord := domain.CaseRecord{
		ID:          "test-case-find-approval",
		Title:       "Test Case",
		Status:      domain.CaseStatusReady,
		Spec:        domain.CaseSpec{},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.CreateCase(ctx, caseRecord); err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Create approval
	approval := domain.Approval{
		ID:           "approval-find-001",
		CaseID:       "test-case-find-approval",
		CommandIndex: 5,
		CommandName:  "specific-cmd",
		RiskClass:    domain.RiskDestructive,
		Status:       domain.ApprovalPending,
		CreatedAt:    now,
	}
	_, err := repo.CreateOrGetPendingApproval(ctx, approval)
	if err != nil {
		t.Fatalf("Failed to create approval: %v", err)
	}

	// Find approval by command
	found, err := repo.FindApprovalByCommand(ctx, "test-case-find-approval", 5)
	if err != nil {
		t.Fatalf("Failed to find approval: %v", err)
	}

	if found.ID != "approval-find-001" {
		t.Errorf("Found wrong approval: got %v, want %v", found.ID, "approval-find-001")
	}

	// Find non-existent approval
	_, err = repo.FindApprovalByCommand(ctx, "test-case-find-approval", 999)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for non-existent approval, got %v", err)
	}
}
