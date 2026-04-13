package policy

import "hal-proxy/internal/domain"

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

// NormalizeRisk ensures the risk class is valid
func NormalizeRisk(risk domain.RiskClass) domain.RiskClass {
	if _, ok := DefaultPolicyConfig[risk]; ok {
		return risk
	}
	return domain.RiskObserve
}

// RequiresApproval checks if a risk class requires approval
func RequiresApproval(risk domain.RiskClass) bool {
	risk = NormalizeRisk(risk)
	config, ok := DefaultPolicyConfig[risk]
	if !ok {
		return false
	}
	return config.RequiresApproval
}

// GetRiskPriority returns the priority level of a risk class (higher = more severe)
func GetRiskPriority(risk domain.RiskClass) int {
	risk = NormalizeRisk(risk)
	config, ok := DefaultPolicyConfig[risk]
	if !ok {
		return 0
	}
	return config.Priority
}

// GetRiskDescription returns a human-readable description of the risk class
func GetRiskDescription(risk domain.RiskClass) string {
	risk = NormalizeRisk(risk)
	config, ok := DefaultPolicyConfig[risk]
	if !ok {
		return "Unknown risk level"
	}
	return config.Description
}
