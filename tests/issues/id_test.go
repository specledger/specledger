package issues_test

import (
	"testing"
	"time"

	"github.com/specledger/specledger/pkg/issues"
)

func TestGenerateIssueID(t *testing.T) {
	tests := []struct {
		name        string
		specContext string
		title       string
		createdAt   time.Time
		wantPrefix  string
	}{
		{
			name:        "generates valid ID",
			specContext: "010-my-feature",
			title:       "Add validation",
			createdAt:   time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC),
			wantPrefix:  "SL-",
		},
		{
			name:        "different specs produce different IDs",
			specContext: "020-other-feature",
			title:       "Add validation",
			createdAt:   time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC),
			wantPrefix:  "SL-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := issues.GenerateIssueID(tt.specContext, tt.title, tt.createdAt)

			if !isValidIssueIDFormat(id) {
				t.Errorf("GenerateIssueID() = %v, want format SL-xxxxxx", id)
			}

			if id[:3] != tt.wantPrefix {
				t.Errorf("GenerateIssueID() = %v, want prefix %v", id, tt.wantPrefix)
			}
		})
	}
}

func TestGenerateIssueID_Deterministic(t *testing.T) {
	specContext := "010-my-feature"
	title := "Add validation"
	createdAt := time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC)

	id1 := issues.GenerateIssueID(specContext, title, createdAt)
	id2 := issues.GenerateIssueID(specContext, title, createdAt)

	if id1 != id2 {
		t.Errorf("GenerateIssueID() should be deterministic, got %v and %v", id1, id2)
	}
}

func TestGenerateIssueID_DifferentTitles(t *testing.T) {
	specContext := "010-my-feature"
	createdAt := time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC)

	id1 := issues.GenerateIssueID(specContext, "Add validation", createdAt)
	id2 := issues.GenerateIssueID(specContext, "Add logging", createdAt)

	if id1 == id2 {
		t.Errorf("GenerateIssueID() should produce different IDs for different titles")
	}
}

func TestParseIssueID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid ID",
			id:      "SL-a3f5d8",
			wantErr: false,
		},
		{
			name:    "invalid - no prefix",
			id:      "a3f5d8",
			wantErr: true,
		},
		{
			name:    "invalid - wrong prefix",
			id:      "BD-a3f5d8",
			wantErr: true,
		},
		{
			name:    "invalid - too short",
			id:      "SL-a3f5",
			wantErr: true,
		},
		{
			name:    "invalid - too long",
			id:      "SL-a3f5d8e9",
			wantErr: true,
		},
		{
			name:    "invalid - non-hex characters",
			id:      "SL-ghijkl",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := issues.ParseIssueID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIssueID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func isValidIssueIDFormat(id string) bool {
	if len(id) != 9 {
		return false
	}
	if id[:3] != "SL-" {
		return false
	}
	for _, c := range id[3:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

