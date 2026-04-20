package policy

import (
	"testing"

	"bridgeos/internal/domain"
)

func TestNormalizeRisk(t *testing.T) {
	tests := []struct {
		name     string
		input    domain.RiskClass
		expected domain.RiskClass
	}{
		{"Observe stays as Observe", domain.RiskObserve, domain.RiskObserve},
		{"Mutate stays as Mutate", domain.RiskMutate, domain.RiskMutate},
		{"Destructive stays as Destructive", domain.RiskDestructive, domain.RiskDestructive},
		{"Exclusive stays as Exclusive", domain.RiskExclusive, domain.RiskExclusive},
		{"Unknown becomes Observe", domain.RiskClass("unknown"), domain.RiskObserve},
		{"Empty becomes Observe", domain.RiskClass(""), domain.RiskObserve},
		{"Random string becomes Observe", domain.RiskClass("random"), domain.RiskObserve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeRisk(tt.input); got != tt.expected {
				t.Errorf("NormalizeRisk(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRequiresApproval(t *testing.T) {
	tests := []struct {
		name     string
		risk     domain.RiskClass
		expected bool
	}{
		{"Observe does not require approval", domain.RiskObserve, false},
		{"Mutate requires approval", domain.RiskMutate, true},
		{"Destructive requires approval", domain.RiskDestructive, true},
		{"Exclusive requires approval", domain.RiskExclusive, true},
		{"Unknown does not require approval", domain.RiskClass("unknown"), false},
		{"Empty does not require approval", domain.RiskClass(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RequiresApproval(tt.risk); got != tt.expected {
				t.Errorf("RequiresApproval(%v) = %v, want %v", tt.risk, got, tt.expected)
			}
		})
	}
}

func TestGetRiskPriority(t *testing.T) {
	tests := []struct {
		name     string
		risk     domain.RiskClass
		expected int
	}{
		{"Observe has priority 0", domain.RiskObserve, 0},
		{"Mutate has priority 1", domain.RiskMutate, 1},
		{"Destructive has priority 2", domain.RiskDestructive, 2},
		{"Exclusive has priority 3", domain.RiskExclusive, 3},
		{"Unknown has priority 0", domain.RiskClass("unknown"), 0},
		{"Empty has priority 0", domain.RiskClass(""), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRiskPriority(tt.risk); got != tt.expected {
				t.Errorf("GetRiskPriority(%v) = %v, want %v", tt.risk, got, tt.expected)
			}
		})
	}
}

func TestGetRiskDescription(t *testing.T) {
	tests := []struct {
		name     string
		risk     domain.RiskClass
		expected string
	}{
		{"Observe description", domain.RiskObserve, "Read-only operations, no modification"},
		{"Mutate description", domain.RiskMutate, "Operations that modify state"},
		{"Destructive description", domain.RiskDestructive, "Operations that may cause data loss"},
		{"Exclusive description", domain.RiskExclusive, "Operations that require exclusive access"},
		// Unknown risks are normalized to Observe, so they get the Observe description
		{"Unknown description", domain.RiskClass("unknown"), "Read-only operations, no modification"},
		{"Empty description", domain.RiskClass(""), "Read-only operations, no modification"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRiskDescription(tt.risk); got != tt.expected {
				t.Errorf("GetRiskDescription(%v) = %v, want %v", tt.risk, got, tt.expected)
			}
		})
	}
}

func TestRiskPriorityOrdering(t *testing.T) {
	// Verify that risk priorities are in correct order
	risks := []domain.RiskClass{
		domain.RiskObserve,
		domain.RiskMutate,
		domain.RiskDestructive,
		domain.RiskExclusive,
	}

	for i := 0; i < len(risks)-1; i++ {
		currentPriority := GetRiskPriority(risks[i])
		nextPriority := GetRiskPriority(risks[i+1])
		if currentPriority >= nextPriority {
			t.Errorf("Risk priority ordering violated: %v (%d) >= %v (%d)",
				risks[i], currentPriority, risks[i+1], nextPriority)
		}
	}
}

func TestApprovalRequirementMatchesPriority(t *testing.T) {
	// High priority risks should require approval
	for _, risk := range []domain.RiskClass{
		domain.RiskMutate,
		domain.RiskDestructive,
		domain.RiskExclusive,
	} {
		priority := GetRiskPriority(risk)
		requiresApproval := RequiresApproval(risk)
		if priority > 0 && !requiresApproval {
			t.Errorf("Risk %v has priority %d but does not require approval", risk, priority)
		}
	}

	// Low priority risks should not require approval
	observePriority := GetRiskPriority(domain.RiskObserve)
	observeRequiresApproval := RequiresApproval(domain.RiskObserve)
	if observePriority == 0 && observeRequiresApproval {
		t.Error("Risk observe has priority 0 but requires approval")
	}
}

func TestDefaultPolicyConfig完整性(t *testing.T) {
	// Verify all defined risk classes have configurations
	expectedRisks := []domain.RiskClass{
		domain.RiskObserve,
		domain.RiskMutate,
		domain.RiskDestructive,
		domain.RiskExclusive,
	}

	for _, risk := range expectedRisks {
		if _, ok := DefaultPolicyConfig[risk]; !ok {
			t.Errorf("Missing configuration for risk class: %v", risk)
		}
	}

	// Verify all configurations have required fields
	for risk, config := range DefaultPolicyConfig {
		if config.Description == "" {
			t.Errorf("Risk %v has empty description", risk)
		}
	}
}

func TestNormalizeRiskPreservesKnownRisks(t *testing.T) {
	// Test that all known risks are preserved after normalization
	knownRisks := []domain.RiskClass{
		domain.RiskObserve,
		domain.RiskMutate,
		domain.RiskDestructive,
		domain.RiskExclusive,
	}

	for _, risk := range knownRisks {
		normalized := NormalizeRisk(risk)
		if normalized != risk {
			t.Errorf("NormalizeRisk should preserve known risk %v, but got %v", risk, normalized)
		}
	}
}

func TestRequiresApprovalWithNormalizedRisk(t *testing.T) {
	// Test that RequiresApproval works correctly with both normalized and non-normalized risks
	testCases := []struct {
		risk          domain.RiskClass
		shouldApprove bool
	}{
		{domain.RiskObserve, false},
		{domain.RiskMutate, true},
		{domain.RiskDestructive, true},
		{domain.RiskExclusive, true},
		// Unknown risks should be treated as observe (no approval needed)
		{domain.RiskClass("unknown"), false},
		{domain.RiskClass(""), false},
	}

	for _, tc := range testCases {
		result := RequiresApproval(tc.risk)
		if result != tc.shouldApprove {
			t.Errorf("RequiresApproval(%v) = %v, want %v", tc.risk, result, tc.shouldApprove)
		}
	}
}

func BenchmarkNormalizeRisk(b *testing.B) {
	risks := []domain.RiskClass{
		domain.RiskObserve,
		domain.RiskMutate,
		domain.RiskDestructive,
		domain.RiskExclusive,
		domain.RiskClass("unknown"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, risk := range risks {
			_ = NormalizeRisk(risk)
		}
	}
}

func BenchmarkRequiresApproval(b *testing.B) {
	risks := []domain.RiskClass{
		domain.RiskObserve,
		domain.RiskMutate,
		domain.RiskDestructive,
		domain.RiskExclusive,
		domain.RiskClass("unknown"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, risk := range risks {
			_ = RequiresApproval(risk)
		}
	}
}
