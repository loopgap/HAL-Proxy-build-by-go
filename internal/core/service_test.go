package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

	report, err := svc.BuildReport(ctx, c.ID, "test-user")
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

func TestRunCaseConcurrencyRace(t *testing.T) {
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
		Title: "concurrent-test",
		Commands: []domain.CaseCommandSpec{
			{Name: "read", Action: "read_mem", RiskClass: domain.RiskObserve},
		},
	}, "test")
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.RunCase(ctx, c.ID, "test")
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	conflictCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else {
			conflictCount++
		}
	}

	if successCount != 1 {
		t.Errorf("expected exactly 1 success, got %d (conflicts: %d)", successCount, conflictCount)
	}
	if conflictCount != numGoroutines-1 {
		t.Errorf("expected %d conflicts, got %d", numGoroutines-1, conflictCount)
	}
}

func TestResolveApprovalConcurrencyRace(t *testing.T) {
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
		Title: "approval-concurrent-test",
		Commands: []domain.CaseCommandSpec{
			{Name: "reset", Action: "reset", RiskClass: domain.RiskDestructive},
		},
	}, "test")
	if err != nil {
		t.Fatal(err)
	}

	run1, err := svc.RunCase(ctx, c.ID, "test")
	if err != nil {
		t.Fatal(err)
	}
	if run1.PendingApproval == nil {
		t.Fatal("expected pending approval")
	}

	approvalID := run1.PendingApproval.ID

	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	type result struct {
		approval domain.Approval
		err      error
	}
	results := make(chan result, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(actor string) {
			defer wg.Done()
			approval, err := svc.ResolveApproval(ctx, approvalID, actor, "approve", "approved")
			results <- result{approval, err}
		}(fmt.Sprintf("approver-%d", i))
	}

	wg.Wait()
	close(results)

	approvedCount := 0
	for r := range results {
		if r.err == nil && r.approval.Status == domain.ApprovalApproved {
			approvedCount++
		}
	}

	if approvedCount != numGoroutines {
		t.Errorf("expected all %d to return approved (idempotent), got %d", numGoroutines, approvedCount)
	}
}
