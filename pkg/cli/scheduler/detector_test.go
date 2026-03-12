package scheduler

import (
	"testing"
)

func TestIsFeatureBranchName(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   bool
	}{
		{"valid 3-digit", "127-feature-name", true},
		{"valid 4-digit", "1234-feature", true},
		{"main branch", "main", false},
		{"develop branch", "develop", false},
		{"no number prefix", "feature-branch", false},
		{"too short", "12-x", false},
		{"just numbers", "127", false},
		{"numbers with trailing dash", "127-", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFeatureBranchName(tt.branch); got != tt.want {
				t.Errorf("isFeatureBranchName(%q) = %v, want %v", tt.branch, got, tt.want)
			}
		})
	}
}
