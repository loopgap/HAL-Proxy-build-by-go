package dto

import (
	"bridgeos/internal/domain"
	"time"
)

type ApprovalResponse struct {
	ID           string     `json:"id"`
	CaseID       string     `json:"case_id"`
	CommandIndex int        `json:"command_index"`
	CommandName  string     `json:"command_name"`
	RiskClass    string     `json:"risk_class"`
	Status       string     `json:"status"`
	Reason       string     `json:"reason,omitempty"`
	DecidedBy    string     `json:"decided_by,omitempty"`
	DecidedAt    *time.Time `json:"decided_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func ToApprovalResponse(a domain.Approval) ApprovalResponse {
	return ApprovalResponse{
		ID:           a.ID,
		CaseID:       a.CaseID,
		CommandIndex: a.CommandIndex,
		CommandName:  a.CommandName,
		RiskClass:    string(a.RiskClass),
		Status:       string(a.Status),
		Reason:       a.Reason,
		DecidedBy:    a.DecidedBy,
		DecidedAt:    a.DecidedAt,
		CreatedAt:    a.CreatedAt,
	}
}
