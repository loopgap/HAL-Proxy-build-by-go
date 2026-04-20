package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"bridgeos/internal/domain"
	"bridgeos/internal/store"
)

func TestCaseRunApprovalAndReportFlow(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	repo, err := store.NewSQLiteRepository(filepath.Join(dir, "bridgeos.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	svc := NewService(repo, filepath.Join(dir, "artifacts"))
	ctx := context.Background()
	if err := svc.Init(ctx); err != nil {
		t.Fatal(err)
	}

	c, err := svc.CreateCase(ctx, domain.CaseSpec{
		Title: "test-case",
		Commands: []domain.CaseCommandSpec{
			{Name: "read", Action: "read_mem", RiskClass: domain.RiskObserve},
			{Name: "reset", Action: "reset", RiskClass: domain.RiskDestructive},
			{Name: "read-again", Action: "read_mem", RiskClass: domain.RiskObserve},
		},
	}, "test")
	if err != nil {
		t.Fatal(err)
	}

	run1, err := svc.RunCase(ctx, c.ID, "test")
	if err != nil {
		t.Fatal(err)
	}
	if run1.Status != "awaiting_approval" {
		t.Fatalf("expected awaiting_approval, got %s", run1.Status)
	}
	if run1.PendingApproval == nil {
		t.Fatal("expected pending approval")
	}

	approval, err := svc.ResolveApproval(ctx, run1.PendingApproval.ID, "tester", "approve", "safe to continue")
	if err != nil {
		t.Fatal(err)
	}
	if approval.Status != domain.ApprovalApproved {
		t.Fatalf("expected approved, got %s", approval.Status)
	}

	run2, err := svc.RunCase(ctx, c.ID, "test")
	if err != nil {
		t.Fatal(err)
	}
	if run2.Status != "completed" {
		t.Fatalf("expected completed, got %s", run2.Status)
	}

	events, err := svc.ListEvents(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected events")
	}

	report, err := svc.BuildReport(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if report.EventCount == 0 {
		t.Fatal("expected report event count")
	}
	if _, err := os.Stat(report.Path); err != nil {
		t.Fatalf("expected report file: %v", err)
	}
}
