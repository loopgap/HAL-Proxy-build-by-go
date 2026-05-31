package policy

import (
	"fmt"

	"bridgeos/internal/domain"
)

// RiskConfig defines configuration for a risk class
type RiskConfig struct {
	RequiresApproval bool
	Priority         int
	Description      string
}

// DefaultPolicyConfig provides default policy configurations
var DefaultPolicyConfig = map[domain.RiskClass]RiskConfig{
	domain.RiskObserve: {
		RequiresApproval: false,
		Priority:         0,
		Description:      "Read-only operations, no modification",
	},
	domain.RiskMutate: {
		RequiresApproval: true,
		Priority:         1,
		Description:      "Operations that modify state",
	},
	domain.RiskDestructive: {
		RequiresApproval: true,
		Priority:         2,
		Description:      "Operations that may cause data loss",
	},
	domain.RiskExclusive: {
		RequiresApproval: true,
		Priority:         3,
		Description:      "Operations that require exclusive access",
	},
}

// ValidateRisk checks if a risk class is known. Returns error for unknown risk classes.
func ValidateRisk(risk domain.RiskClass) error {
	if _, ok := DefaultPolicyConfig[risk]; ok {
		return nil
	}
	return fmt.Errorf("unknown risk class: %q; valid values are observe, mutate, destructive, exclusive", risk)
}

// NormalizeRisk ensures the risk class is valid. Returns error for unknown risk classes (fail-closed).
func NormalizeRisk(risk domain.RiskClass) (domain.RiskClass, error) {
	if _, ok := DefaultPolicyConfig[risk]; ok {
		return risk, nil
	}
	return "", fmt.Errorf("unknown risk class: %q", risk)
}

// RequiresApproval checks if a risk class requires approval.
// Unknown risk classes default to requiring approval (fail-closed).
func RequiresApproval(risk domain.RiskClass) bool {
	config, ok := DefaultPolicyConfig[risk]
	if !ok {
		return true // fail-closed: unknown risk requires approval
	}
	return config.RequiresApproval
}

// GetRiskPriority returns the priority level of a risk class (higher = more severe).
// Unknown risk classes return highest priority (fail-closed).
func GetRiskPriority(risk domain.RiskClass) int {
	config, ok := DefaultPolicyConfig[risk]
	if !ok {
		return 999 // fail-closed: unknown risk gets highest priority
	}
	return config.Priority
}

// GetRiskDescription returns a human-readable description of the risk class
func GetRiskDescription(risk domain.RiskClass) string {
	config, ok := DefaultPolicyConfig[risk]
	if !ok {
		return "Unknown risk level — treated as highest risk"
	}
	return config.Description
}
