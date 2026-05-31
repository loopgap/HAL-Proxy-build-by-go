package dto

import (
	"bridgeos/internal/domain"
	"testing"
	"time"
)

func TestToCaseResponse(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    domain.CaseRecord
		expected CaseResponse
	}{
		{
			name: "full case with multiple commands",
			input: domain.CaseRecord{
				ID:      "case-001",
				Title:   "Test Case",
				Status:  domain.CaseStatusRunning,
				Spec: domain.CaseSpec{
					Title: "Test Case",
					Commands: []domain.CaseCommandSpec{
						{
							Name:      "list-files",
							Action:    "ls -la",
							RiskClass: domain.RiskObserve,
						},
						{
							Name:      "delete-tmp",
							Action:    "rm -rf /tmp/*",
							RiskClass: domain.RiskDestructive,
							Parameters: map[string]any{
								"dry_run": true,
							},
						},
					},
				},
				NextCommand: 1,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			expected: CaseResponse{
				ID:     "case-001",
				Title:  "Test Case",
				Status: "running",
				Commands: []CommandDTO{
					{
						Name:      "list-files",
						Action:    "ls -la",
						RiskClass: "observe",
					},
					{
						Name:      "delete-tmp",
						Action:    "rm -rf /tmp/*",
						RiskClass: "destructive",
						Parameters: map[string]any{
							"dry_run": true,
						},
					},
				},
				NextCommand: 1,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			name: "case with zero commands",
			input: domain.CaseRecord{
				ID:      "case-002",
				Title:   "Empty Case",
				Status:  domain.CaseStatusDraft,
				Spec:    domain.CaseSpec{Title: "Empty Case"},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			expected: CaseResponse{
				ID:       "case-002",
				Title:    "Empty Case",
				Status:   "draft",
				Commands: []CommandDTO{},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
		},
		{
			name: "case with all risk classes",
			input: domain.CaseRecord{
				ID:     "case-003",
				Title:  "Risk Classes",
				Status: domain.CaseStatusReady,
				Spec: domain.CaseSpec{
					Title: "Risk Classes",
					Commands: []domain.CaseCommandSpec{
						{Name: "c1", Action: "a1", RiskClass: domain.RiskObserve},
						{Name: "c2", Action: "a2", RiskClass: domain.RiskMutate},
						{Name: "c3", Action: "a3", RiskClass: domain.RiskDestructive},
						{Name: "c4", Action: "a4", RiskClass: domain.RiskExclusive},
					},
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
			expected: CaseResponse{
				ID:     "case-003",
				Title:  "Risk Classes",
				Status: "ready",
				Commands: []CommandDTO{
					{Name: "c1", Action: "a1", RiskClass: "observe"},
					{Name: "c2", Action: "a2", RiskClass: "mutate"},
					{Name: "c3", Action: "a3", RiskClass: "destructive"},
					{Name: "c4", Action: "a4", RiskClass: "exclusive"},
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
		},
		{
			name: "case with parameters",
			input: domain.CaseRecord{
				ID:     "case-004",
				Title:  "With Params",
				Status: domain.CaseStatusCompleted,
				Spec: domain.CaseSpec{
					Title: "With Params",
					Commands: []domain.CaseCommandSpec{
						{
							Name:      "cmd1",
							Action:    "run",
							RiskClass: domain.RiskMutate,
							Parameters: map[string]any{
								"key1": "value1",
								"key2": 42,
								"key3": true,
							},
						},
					},
				},
				NextCommand: 0,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			expected: CaseResponse{
				ID:     "case-004",
				Title:  "With Params",
				Status: "completed",
				Commands: []CommandDTO{
					{
						Name:      "cmd1",
						Action:    "run",
						RiskClass: "mutate",
						Parameters: map[string]any{
							"key1": "value1",
							"key2": 42,
							"key3": true,
						},
					},
				},
				NextCommand: 0,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			name: "zero value case",
			input: domain.CaseRecord{},
			expected: CaseResponse{
				Commands: []CommandDTO{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCaseResponse(tt.input)
			if result.ID != tt.expected.ID {
				t.Errorf("ID: got %q, want %q", result.ID, tt.expected.ID)
			}
			if result.Title != tt.expected.Title {
				t.Errorf("Title: got %q, want %q", result.Title, tt.expected.Title)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Status: got %q, want %q", result.Status, tt.expected.Status)
			}
			if result.NextCommand != tt.expected.NextCommand {
				t.Errorf("NextCommand: got %d, want %d", result.NextCommand, tt.expected.NextCommand)
			}
			if len(result.Commands) != len(tt.expected.Commands) {
				t.Fatalf("Commands length: got %d, want %d", len(result.Commands), len(tt.expected.Commands))
			}
			for i, cmd := range result.Commands {
				exp := tt.expected.Commands[i]
				if cmd.Name != exp.Name {
					t.Errorf("Command[%d].Name: got %q, want %q", i, cmd.Name, exp.Name)
				}
				if cmd.Action != exp.Action {
					t.Errorf("Command[%d].Action: got %q, want %q", i, cmd.Action, exp.Action)
				}
				if cmd.RiskClass != exp.RiskClass {
					t.Errorf("Command[%d].RiskClass: got %q, want %q", i, cmd.RiskClass, exp.RiskClass)
				}
			}
			if !result.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("CreatedAt: got %v, want %v", result.CreatedAt, tt.expected.CreatedAt)
			}
			if !result.UpdatedAt.Equal(tt.expected.UpdatedAt) {
				t.Errorf("UpdatedAt: got %v, want %v", result.UpdatedAt, tt.expected.UpdatedAt)
			}
		})
	}
}

func TestToApprovalResponse(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	decidedAt := time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    domain.Approval
		expected ApprovalResponse
	}{
		{
			name: "pending approval",
			input: domain.Approval{
				ID:           "appr-001",
				CaseID:       "case-001",
				CommandIndex: 0,
				CommandName:  "delete-files",
				RiskClass:    domain.RiskDestructive,
				Status:       domain.ApprovalPending,
				CreatedAt:    fixedTime,
			},
			expected: ApprovalResponse{
				ID:           "appr-001",
				CaseID:       "case-001",
				CommandIndex: 0,
				CommandName:  "delete-files",
				RiskClass:    "destructive",
				Status:       "pending",
				CreatedAt:    fixedTime,
			},
		},
		{
			name: "approved approval with decision",
			input: domain.Approval{
				ID:           "appr-002",
				CaseID:       "case-001",
				CommandIndex: 1,
				CommandName:  "modify-config",
				RiskClass:    domain.RiskMutate,
				Status:       domain.ApprovalApproved,
				Reason:       "Looks safe enough",
				DecidedBy:    "admin",
				DecidedAt:    &decidedAt,
				CreatedAt:    fixedTime,
			},
			expected: ApprovalResponse{
				ID:           "appr-002",
				CaseID:       "case-001",
				CommandIndex: 1,
				CommandName:  "modify-config",
				RiskClass:    "mutate",
				Status:       "approved",
				Reason:       "Looks safe enough",
				DecidedBy:    "admin",
				DecidedAt:    &decidedAt,
				CreatedAt:    fixedTime,
			},
		},
		{
			name: "rejected approval",
			input: domain.Approval{
				ID:           "appr-003",
				CaseID:       "case-002",
				CommandIndex: 2,
				CommandName:  "drop-table",
				RiskClass:    domain.RiskExclusive,
				Status:       domain.ApprovalRejected,
				Reason:       "Too dangerous",
				DecidedBy:    "reviewer",
				DecidedAt:    &decidedAt,
				CreatedAt:    fixedTime,
			},
			expected: ApprovalResponse{
				ID:           "appr-003",
				CaseID:       "case-002",
				CommandIndex: 2,
				CommandName:  "drop-table",
				RiskClass:    "exclusive",
				Status:       "rejected",
				Reason:       "Too dangerous",
				DecidedBy:    "reviewer",
				DecidedAt:    &decidedAt,
				CreatedAt:    fixedTime,
			},
		},
		{
			name: "observe risk approval",
			input: domain.Approval{
				ID:           "appr-004",
				CaseID:       "case-003",
				CommandIndex: 0,
				CommandName:  "list-items",
				RiskClass:    domain.RiskObserve,
				Status:       domain.ApprovalPending,
				CreatedAt:    fixedTime,
			},
			expected: ApprovalResponse{
				ID:           "appr-004",
				CaseID:       "case-003",
				CommandIndex: 0,
				CommandName:  "list-items",
				RiskClass:    "observe",
				Status:       "pending",
				CreatedAt:    fixedTime,
			},
		},
		{
			name:     "zero value approval",
			input:    domain.Approval{},
			expected: ApprovalResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToApprovalResponse(tt.input)
			if result.ID != tt.expected.ID {
				t.Errorf("ID: got %q, want %q", result.ID, tt.expected.ID)
			}
			if result.CaseID != tt.expected.CaseID {
				t.Errorf("CaseID: got %q, want %q", result.CaseID, tt.expected.CaseID)
			}
			if result.CommandIndex != tt.expected.CommandIndex {
				t.Errorf("CommandIndex: got %d, want %d", result.CommandIndex, tt.expected.CommandIndex)
			}
			if result.CommandName != tt.expected.CommandName {
				t.Errorf("CommandName: got %q, want %q", result.CommandName, tt.expected.CommandName)
			}
			if result.RiskClass != tt.expected.RiskClass {
				t.Errorf("RiskClass: got %q, want %q", result.RiskClass, tt.expected.RiskClass)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Status: got %q, want %q", result.Status, tt.expected.Status)
			}
			if result.Reason != tt.expected.Reason {
				t.Errorf("Reason: got %q, want %q", result.Reason, tt.expected.Reason)
			}
			if result.DecidedBy != tt.expected.DecidedBy {
				t.Errorf("DecidedBy: got %q, want %q", result.DecidedBy, tt.expected.DecidedBy)
			}
			if !result.CreatedAt.Equal(tt.expected.CreatedAt) {
				t.Errorf("CreatedAt: got %v, want %v", result.CreatedAt, tt.expected.CreatedAt)
			}
			if (result.DecidedAt == nil) != (tt.expected.DecidedAt == nil) {
				t.Errorf("DecidedAt nil mismatch: got nil=%v, want nil=%v", result.DecidedAt == nil, tt.expected.DecidedAt == nil)
			} else if result.DecidedAt != nil && tt.expected.DecidedAt != nil && !result.DecidedAt.Equal(*tt.expected.DecidedAt) {
				t.Errorf("DecidedAt: got %v, want %v", *result.DecidedAt, *tt.expected.DecidedAt)
			}
		})
	}
}
