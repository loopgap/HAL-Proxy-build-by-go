package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRiskClassIsHighRisk(t *testing.T) {
	tests := []struct {
		name     string
		risk     RiskClass
		expected bool
	}{
		{"Observe is not high risk", RiskObserve, false},
		{"Mutate is high risk", RiskMutate, true},
		{"Destructive is high risk", RiskDestructive, true},
		{"Exclusive is high risk", RiskExclusive, true},
		{"Unknown risk class", RiskClass("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.risk.IsHighRisk(); got != tt.expected {
				t.Errorf("RiskClass.IsHighRisk() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCaseStatusCanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     CaseStatus
		to       CaseStatus
		expected bool
	}{
		// Draft transitions
		{"Draft can transition to Ready", CaseStatusDraft, CaseStatusReady, true},
		{"Draft cannot transition to Running", CaseStatusDraft, CaseStatusRunning, false},
		{"Draft cannot transition to Completed", CaseStatusDraft, CaseStatusCompleted, false},

		// Ready transitions
		{"Ready can transition to Running", CaseStatusReady, CaseStatusRunning, true},
		{"Ready can transition to Paused", CaseStatusReady, CaseStatusPaused, true},
		{"Ready cannot transition to Completed", CaseStatusReady, CaseStatusCompleted, false},

		// Running transitions
		{"Running can transition to Paused", CaseStatusRunning, CaseStatusPaused, true},
		{"Running can transition to Completed", CaseStatusRunning, CaseStatusCompleted, true},
		{"Running can transition to Rejected", CaseStatusRunning, CaseStatusRejected, true},
		{"Running cannot transition to Ready", CaseStatusRunning, CaseStatusReady, false},

		// Paused transitions
		{"Paused can transition to Ready", CaseStatusPaused, CaseStatusReady, true},
		{"Paused can transition to Rejected", CaseStatusPaused, CaseStatusRejected, true},
		{"Paused cannot transition to Completed", CaseStatusPaused, CaseStatusCompleted, false},
		{"Paused cannot transition to Running", CaseStatusPaused, CaseStatusRunning, false},

		// Terminal states cannot transition
		{"Completed cannot transition", CaseStatusCompleted, CaseStatusReady, false},
		{"Completed cannot transition", CaseStatusCompleted, CaseStatusRunning, false},
		{"Rejected cannot transition", CaseStatusRejected, CaseStatusReady, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("CaseStatus.CanTransitionTo() from %v to %v = %v, want %v", tt.from, tt.to, got, tt.expected)
			}
		})
	}
}

func TestCaseStatusIsTerminal(t *testing.T) {
	tests := []struct {
		name     string
		status   CaseStatus
		expected bool
	}{
		{"Draft is not terminal", CaseStatusDraft, false},
		{"Ready is not terminal", CaseStatusReady, false},
		{"Running is not terminal", CaseStatusRunning, false},
		{"Paused is not terminal", CaseStatusPaused, false},
		{"Completed is terminal", CaseStatusCompleted, true},
		{"Rejected is terminal", CaseStatusRejected, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.expected {
				t.Errorf("CaseStatus.IsTerminal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRiskClassConstants(t *testing.T) {
	// Test that constants are defined correctly
	tests := []struct {
		name     string
		risk     RiskClass
		expected string
	}{
		{"Observe constant", RiskObserve, "observe"},
		{"Mutate constant", RiskMutate, "mutate"},
		{"Destructive constant", RiskDestructive, "destructive"},
		{"Exclusive constant", RiskExclusive, "exclusive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.risk) != tt.expected {
				t.Errorf("RiskClass = %v, want %v", tt.risk, tt.expected)
			}
		})
	}
}

func TestCaseStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   CaseStatus
		expected string
	}{
		{"Draft constant", CaseStatusDraft, "draft"},
		{"Ready constant", CaseStatusReady, "ready"},
		{"Running constant", CaseStatusRunning, "running"},
		{"Paused constant", CaseStatusPaused, "paused"},
		{"Completed constant", CaseStatusCompleted, "completed"},
		{"Rejected constant", CaseStatusRejected, "rejected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("CaseStatus = %v, want %v", tt.status, tt.expected)
			}
		})
	}
}

func TestApprovalStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   ApprovalStatus
		expected string
	}{
		{"Pending constant", ApprovalPending, "pending"},
		{"Approved constant", ApprovalApproved, "approved"},
		{"Rejected constant", ApprovalRejected, "rejected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("ApprovalStatus = %v, want %v", tt.status, tt.expected)
			}
		})
	}
}

func TestCaseSpecJSONSerialization(t *testing.T) {
	spec := CaseSpec{
		Title: "Test Case",
		Commands: []CaseCommandSpec{
			{
				Name:       "read-memory",
				Action:     "read_mem",
				RiskClass:  RiskObserve,
				Parameters: map[string]any{"address": "0x20000000", "length": 16},
			},
			{
				Name:       "reset-device",
				Action:     "reset",
				RiskClass:  RiskDestructive,
				Parameters: map[string]any{"mode": "sysresetreq"},
			},
		},
	}

	// Test serialization
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal CaseSpec: %v", err)
	}

	// Test deserialization
	var decoded CaseSpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CaseSpec: %v", err)
	}

	if decoded.Title != spec.Title {
		t.Errorf("Title = %v, want %v", decoded.Title, spec.Title)
	}

	if len(decoded.Commands) != len(spec.Commands) {
		t.Errorf("Commands count = %v, want %v", len(decoded.Commands), len(spec.Commands))
	}

	if decoded.Commands[0].Name != spec.Commands[0].Name {
		t.Errorf("Command name = %v, want %v", decoded.Commands[0].Name, spec.Commands[0].Name)
	}
}

func TestCaseRecordJSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	record := CaseRecord{
		ID:     "test-case-001",
		Title:  "Test Case",
		Status: CaseStatusReady,
		Spec: CaseSpec{
			Title: "Test Case",
			Commands: []CaseCommandSpec{
				{
					Name:      "test-cmd",
					Action:    "test",
					RiskClass: RiskObserve,
				},
			},
		},
		NextCommand: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Test serialization
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Failed to marshal CaseRecord: %v", err)
	}

	// Test deserialization
	var decoded CaseRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CaseRecord: %v", err)
	}

	if decoded.ID != record.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, record.ID)
	}

	if decoded.Status != record.Status {
		t.Errorf("Status = %v, want %v", decoded.Status, record.Status)
	}
}

func TestEventEnvelopeJSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	payload := map[string]any{"key": "value"}
	payloadJSON, _ := json.Marshal(payload)

	event := EventEnvelope{
		Sequence:  1,
		CaseID:    "test-case-001",
		Type:      "test.event",
		Payload:   payloadJSON,
		CreatedAt: now,
	}

	// Test serialization
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal EventEnvelope: %v", err)
	}

	// Test deserialization
	var decoded EventEnvelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal EventEnvelope: %v", err)
	}

	if decoded.Sequence != event.Sequence {
		t.Errorf("Sequence = %v, want %v", decoded.Sequence, event.Sequence)
	}

	if decoded.Type != event.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, event.Type)
	}
}

func TestApprovalJSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	approval := Approval{
		ID:           "approval-001",
		CaseID:       "test-case-001",
		CommandIndex: 1,
		CommandName:  "reset-device",
		RiskClass:    RiskDestructive,
		Status:       ApprovalPending,
		CreatedAt:    now,
	}

	// Test serialization
	data, err := json.Marshal(approval)
	if err != nil {
		t.Fatalf("Failed to marshal Approval: %v", err)
	}

	// Test deserialization
	var decoded Approval
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Approval: %v", err)
	}

	if decoded.ID != approval.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, approval.ID)
	}

	if decoded.Status != approval.Status {
		t.Errorf("Status = %v, want %v", decoded.Status, approval.Status)
	}

	if decoded.RiskClass != approval.RiskClass {
		t.Errorf("RiskClass = %v, want %v", decoded.RiskClass, approval.RiskClass)
	}
}

func TestReportSummaryJSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	report := ReportSummary{
		ID:           "report-001",
		CaseID:       "test-case-001",
		Path:         "/artifacts/test-report.md",
		CommandCount: 5,
		EventCount:   20,
		CreatedAt:    now,
	}

	// Test serialization
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Failed to marshal ReportSummary: %v", err)
	}

	// Test deserialization
	var decoded ReportSummary
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ReportSummary: %v", err)
	}

	if decoded.ID != report.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, report.ID)
	}

	if decoded.CommandCount != report.CommandCount {
		t.Errorf("CommandCount = %v, want %v", decoded.CommandCount, report.CommandCount)
	}
}

func TestDeviceDescriptorJSONSerialization(t *testing.T) {
	device := DeviceDescriptor{
		ID:           "device-001",
		Name:         "Test Device",
		Capabilities: []string{"read", "write", "debug"},
		SupportLevel: "full",
	}

	// Test serialization
	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("Failed to marshal DeviceDescriptor: %v", err)
	}

	// Test deserialization
	var decoded DeviceDescriptor
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DeviceDescriptor: %v", err)
	}

	if decoded.ID != device.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, device.ID)
	}

	if len(decoded.Capabilities) != len(device.Capabilities) {
		t.Errorf("Capabilities count = %v, want %v", len(decoded.Capabilities), len(device.Capabilities))
	}
}

func TestSessionRecordJSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	session := SessionRecord{
		ID:        "session-001",
		DeviceID:  "device-001",
		Status:    "active",
		Owner:     "test-user",
		CreatedAt: now,
	}

	// Test serialization
	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal SessionRecord: %v", err)
	}

	// Test deserialization
	var decoded SessionRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SessionRecord: %v", err)
	}

	if decoded.ID != session.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, session.ID)
	}

	if decoded.Status != session.Status {
		t.Errorf("Status = %v, want %v", decoded.Status, session.Status)
	}
}

func TestErrInvalidTransition(t *testing.T) {
	// Test that error is properly defined
	if ErrInvalidTransition.Error() != "invalid status transition" {
		t.Errorf("ErrInvalidTransition = %v, want 'invalid status transition'", ErrInvalidTransition.Error())
	}
}

func TestAllStatusTransitions(t *testing.T) {
	// Comprehensive test of all possible status transitions
	allStatuses := []CaseStatus{
		CaseStatusDraft,
		CaseStatusReady,
		CaseStatusRunning,
		CaseStatusPaused,
		CaseStatusCompleted,
		CaseStatusRejected,
	}

	for _, from := range allStatuses {
		for _, to := range allStatuses {
			// Just ensure CanTransitionTo doesn't panic
			_ = from.CanTransitionTo(to)
		}
	}
}
