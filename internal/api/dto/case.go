package dto

import (
	"hal-proxy/internal/domain"
	"time"
)

type CaseResponse struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Status      string       `json:"status"`
	Commands    []CommandDTO `json:"commands,omitempty"`
	NextCommand int          `json:"next_command"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type CommandDTO struct {
	Name       string         `json:"name"`
	Action     string         `json:"action"`
	RiskClass  string         `json:"risk_class"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type CreateCaseRequest struct {
	Title    string        `json:"title"`
	Commands []CommandSpec `json:"commands"`
}

type CommandSpec struct {
	Name       string         `json:"name"`
	Action     string         `json:"action"`
	RiskClass  string         `json:"risk_class"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

func ToCaseResponse(c domain.CaseRecord) CaseResponse {
	commands := make([]CommandDTO, len(c.Spec.Commands))
	for i, cmd := range c.Spec.Commands {
		commands[i] = CommandDTO{
			Name:       cmd.Name,
			Action:     cmd.Action,
			RiskClass:  string(cmd.RiskClass),
			Parameters: cmd.Parameters,
		}
	}
	return CaseResponse{
		ID:          c.ID,
		Title:       c.Title,
		Status:      string(c.Status),
		Commands:    commands,
		NextCommand: c.NextCommand,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
