package policy

import (
	"testing"

	"bridgeos/internal/domain"
)

func TestNormalizeRisk(t *testing.T) {
	tests := []struct {
		name      string
		input     domain.RiskClass
		expected  domain.RiskClass
		wantError bool
	}{
		{"Observe stays as Observe", domain.RiskObserve, domain.RiskObserve, false},
		{"Mutate stays as Mutate", domain.RiskMutate, domain.RiskMutate, false},
		{"Destructive stays as Destructive", domain.RiskDestructive, domain.RiskDestructive, false},
		{"Exclusive stays as Exclusive", domain.RiskExclusive, domain.RiskExclusive, false},
		{"Unknown returns error", domain.RiskClass("unknown"), "", true},
		{"Empty returns error", domain.RiskClass(""), "", true},
		{"Random string returns error", domain.RiskClass("random"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeRisk(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("NormalizeRisk(%v) error = %v, wantError %v", tt.input, err, tt.wantError)
				return
			}
			if got != tt.expected {
				t.Errorf("NormalizeRisk(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateRisk(t *testing.T) {
	tests := []struct {
		name      string
		input     domain.RiskClass
		wantError bool
	}{
		{"Observe is valid", domain.RiskObserve, false},
		{"Mutate is valid", domain.RiskMutate, false},
		{"Destructive is valid", domain.RiskDestructive, false},
		{"Exclusive is valid", domain.RiskExclusive, false},
		{"Unknown is invalid", domain.RiskClass("unknown"), true},
		{"Empty is invalid", domain.RiskClass(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRisk(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRisk(%v) error = %v, wantError %v", tt.input, err, tt.wantError)
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
		{"Unknown requires approval (fail-closed)", domain.RiskClass("unknown"), true},
		{"Empty requires approval (fail-closed)", domain.RiskClass(""), true},
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
		{"Unknown has priority 999 (fail-closed)", domain.RiskClass("unknown"), 999},
		{"Empty has priority 999 (fail-closed)", domain.RiskClass(""), 999},
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
		{"Unknown description", domain.RiskClass("unknown"), "Unknown risk level — treated as highest risk"},
		{"Empty description", domain.RiskClass(""), "Unknown risk level — treated as highest risk"},
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
		normalized, err := NormalizeRisk(risk)
		if err != nil {
			t.Errorf("NormalizeRisk failed for known risk %v: %v", risk, err)
			continue
		}
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
		// Unknown risks should require approval (fail-closed)
		{domain.RiskClass("unknown"), true},
		{domain.RiskClass(""), true},
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
			_, _ = NormalizeRisk(risk)
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
