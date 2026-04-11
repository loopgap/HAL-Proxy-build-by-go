package domain

import (
	"encoding/json"
	"errors"
	"time"
)

// Validation errors for domain operations
var (
	ErrInvalidTransition = errors.New("invalid status transition")
)

// RiskClass represents the risk level of an operation
type RiskClass string

const (
	RiskObserve     RiskClass = "observe"
	RiskMutate      RiskClass = "mutate"
	RiskDestructive RiskClass = "destructive"
	RiskExclusive   RiskClass = "exclusive"
)

// IsHighRisk returns true if the risk class requires approval
func (r RiskClass) IsHighRisk() bool {
	return r == RiskMutate || r == RiskDestructive || r == RiskExclusive
}

// CaseStatus represents the current state of a case
type CaseStatus string

const (
	CaseStatusDraft     CaseStatus = "draft"
	CaseStatusReady     CaseStatus = "ready"
	CaseStatusRunning   CaseStatus = "running"
	CaseStatusPaused    CaseStatus = "paused"
	CaseStatusCompleted CaseStatus = "completed"
	CaseStatusRejected  CaseStatus = "rejected"
)

// validTransitions defines allowed state transitions for cases
var validCaseTransitions = map[CaseStatus][]CaseStatus{
	CaseStatusDraft:     {CaseStatusReady},
	CaseStatusReady:     {CaseStatusRunning, CaseStatusPaused},
	CaseStatusRunning:   {CaseStatusPaused, CaseStatusCompleted, CaseStatusRejected},
	CaseStatusPaused:    {CaseStatusReady, CaseStatusRejected},
	CaseStatusCompleted: {},
	CaseStatusRejected:  {},
}

// CanTransitionTo checks if the case can transition to the target status
func (s CaseStatus) CanTransitionTo(target CaseStatus) bool {
	allowed, ok := validCaseTransitions[s]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// IsTerminal returns true if this status is a terminal state
func (s CaseStatus) IsTerminal() bool {
	return s == CaseStatusCompleted || s == CaseStatusRejected
}

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type CaseSpec struct {
	Title    string            `json:"title"`
	Commands []CaseCommandSpec `json:"commands"`
}

type CaseCommandSpec struct {
	Name       string         `json:"name"`
	Action     string         `json:"action"`
	RiskClass  RiskClass      `json:"risk_class"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type CaseRecord struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Status      CaseStatus `json:"status"`
	Spec        CaseSpec   `json:"spec"`
	NextCommand int        `json:"next_command"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type EventEnvelope struct {
	Sequence  int64           `json:"sequence"`
	CaseID    string          `json:"case_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

type Approval struct {
	ID           string         `json:"id"`
	CaseID       string         `json:"case_id"`
	CommandIndex int            `json:"command_index"`
	CommandName  string         `json:"command_name"`
	RiskClass    RiskClass      `json:"risk_class"`
	Status       ApprovalStatus `json:"status"`
	Reason       string         `json:"reason,omitempty"`
	DecidedBy    string         `json:"decided_by,omitempty"`
	DecidedAt    *time.Time     `json:"decided_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type ReportSummary struct {
	ID           string    `json:"id"`
	CaseID       string    `json:"case_id"`
	Path         string    `json:"path"`
	CommandCount int       `json:"command_count"`
	EventCount   int       `json:"event_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type DeviceDescriptor struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
	SupportLevel string   `json:"support_level"`
}

type SessionRecord struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Status    string    `json:"status"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}
