package metadata

import (
	"testing"
	"time"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple name", "my-project", false},
		{"valid with numbers", "project123", false},
		{"valid with hyphens", "my-awesome-project", false},
		{"empty name", "", true},
		{"with spaces", "my project", true},
		{"with special chars", "my_project!", true},
		{"with underscores", "my_project", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateShortCode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid 2 chars", "ab", false},
		{"valid 4 chars", "abcd", false},
		{"valid 10 chars", "abcdefghij", false},
		{"valid with numbers", "ab12", false},
		{"too short", "a", true},
		{"too long", "abcdefghijk", true},
		{"with hyphens", "ab-cd", true},
		{"with special chars", "ab!", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShortCode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateShortCode(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid ssh with .git", "git@github.com:org/repo.git", false},
		{"valid ssh without .git", "git@github.com:org/repo", false},
		{"valid https", "https://github.com/org/repo", false},
		{"valid https with path", "https://github.com/org/repo/path", false},
		{"invalid no protocol", "github.com/org/repo", true},
		{"invalid http", "http://github.com/org/repo", true},
		{"invalid format", "not-a-url", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCommitSHA(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid sha", "abc123def456789012345678901234567890abcd", false},
		{"valid all lowercase", "abcdef1234567890abcdef1234567890abcdef12", false},
		{"too short", "abc123", true},
		{"too long", "abc123def456789012345678901234567890abcdef", true},
		{"with uppercase", "ABC123DEF456789012345678901234567890ABCD", true},
		{"with special chars", "abc123def456789012345678901234567890abc!", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommitSHA(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommitSHA(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestProjectMetadataValidate(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)

	validMetadata := &ProjectMetadata{
		Version: "1.0.0",
		Project: ProjectInfo{
			Name:      "test-project",
			ShortCode: "tp",
			Created:   past,
			Modified:  now,
			Version:   "0.1.0",
		},
		Framework: FrameworkInfo{
			Choice: FrameworkNone,
		},
		Dependencies: []Dependency{},
	}

	t.Run("valid metadata", func(t *testing.T) {
		if err := validMetadata.Validate(); err != nil {
			t.Errorf("expected valid metadata to pass validation, got error: %v", err)
		}
	})

	t.Run("invalid version", func(t *testing.T) {
		m := *validMetadata
		m.Version = "2.0.0"
		if err := m.Validate(); err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("invalid project name", func(t *testing.T) {
		m := *validMetadata
		m.Project.Name = "invalid name!"
		if err := m.Validate(); err == nil {
			t.Error("expected error for invalid project name")
		}
	})

	t.Run("invalid short code", func(t *testing.T) {
		m := *validMetadata
		m.Project.ShortCode = "x"
		if err := m.Validate(); err == nil {
			t.Error("expected error for invalid short code")
		}
	})

	t.Run("modified before created", func(t *testing.T) {
		m := *validMetadata
		m.Project.Modified = past
		m.Project.Created = now
		if err := m.Validate(); err == nil {
			t.Error("expected error when modified is before created")
		}
	})

	t.Run("invalid framework choice", func(t *testing.T) {
		m := *validMetadata
		m.Framework.Choice = "invalid"
		if err := m.Validate(); err == nil {
			t.Error("expected error for invalid framework choice")
		}
	})

	t.Run("invalid dependency url", func(t *testing.T) {
		m := *validMetadata
		m.Dependencies = []Dependency{
			{URL: "not-a-valid-url"},
		}
		if err := m.Validate(); err == nil {
			t.Error("expected error for invalid dependency URL")
		}
	})

	t.Run("invalid dependency commit sha", func(t *testing.T) {
		m := *validMetadata
		m.Dependencies = []Dependency{
			{
				URL:            "git@github.com:org/repo.git",
				ResolvedCommit: "invalid-sha",
			},
		}
		if err := m.Validate(); err == nil {
			t.Error("expected error for invalid commit SHA")
		}
	})
}

func TestFrameworkChoice(t *testing.T) {
	validChoices := []FrameworkChoice{
		FrameworkSpecKit,
		FrameworkOpenSpec,
		FrameworkBoth,
		FrameworkNone,
	}

	for _, choice := range validChoices {
		t.Run(string(choice), func(t *testing.T) {
			metadata := &ProjectMetadata{
				Version: "1.0.0",
				Project: ProjectInfo{
					Name:      "test",
					ShortCode: "ts",
					Created:   time.Now(),
					Modified:  time.Now(),
					Version:   "0.1.0",
				},
				Framework: FrameworkInfo{
					Choice: choice,
				},
				Dependencies: []Dependency{},
			}

			if err := metadata.Validate(); err != nil {
				t.Errorf("expected %s to be valid, got error: %v", choice, err)
			}
		})
	}
}
