package issues

import (
	"testing"
	"time"
)

func TestIssueValidate(t *testing.T) {
	tests := []struct {
		name    string
		issue   *Issue
		wantErr error
	}{
		{
			name: "valid issue",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "Test Issue",
				Description: "Test description",
				Status:      StatusOpen,
				Priority:    1,
				IssueType:   TypeTask,
				SpecContext: "010-test-feature",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "invalid ID format",
			issue: &Issue{
				ID:          "invalid",
				Title:       "Test Issue",
				Status:      StatusOpen,
				Priority:    1,
				IssueType:   TypeTask,
				SpecContext: "010-test-feature",
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "empty title",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "",
				Status:      StatusOpen,
				Priority:    1,
				IssueType:   TypeTask,
				SpecContext: "010-test-feature",
			},
			wantErr: ErrInvalidTitle,
		},
		{
			name: "title too long",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       string(make([]byte, 201)),
				Status:      StatusOpen,
				Priority:    1,
				IssueType:   TypeTask,
				SpecContext: "010-test-feature",
			},
			wantErr: ErrInvalidTitle,
		},
		{
			name: "invalid status",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "Test Issue",
				Status:      IssueStatus("invalid"),
				Priority:    1,
				IssueType:   TypeTask,
				SpecContext: "010-test-feature",
			},
			wantErr: ErrInvalidStatus,
		},
		{
			name: "priority too low",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "Test Issue",
				Status:      StatusOpen,
				Priority:    -1,
				IssueType:   TypeTask,
				SpecContext: "010-test-feature",
			},
			wantErr: ErrInvalidPriority,
		},
		{
			name: "priority too high",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "Test Issue",
				Status:      StatusOpen,
				Priority:    6,
				IssueType:   TypeTask,
				SpecContext: "010-test-feature",
			},
			wantErr: ErrInvalidPriority,
		},
		{
			name: "invalid issue type",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "Test Issue",
				Status:      StatusOpen,
				Priority:    1,
				IssueType:   IssueType("invalid"),
				SpecContext: "010-test-feature",
			},
			wantErr: ErrInvalidIssueType,
		},
		{
			name: "invalid spec context",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "Test Issue",
				Status:      StatusOpen,
				Priority:    1,
				IssueType:   TypeTask,
				SpecContext: "invalid-context",
			},
			wantErr: ErrInvalidSpecContext,
		},
		{
			name: "empty spec context is valid",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "Test Issue",
				Status:      StatusOpen,
				Priority:    1,
				IssueType:   TypeTask,
				SpecContext: "",
			},
			wantErr: nil,
		},
		{
			name: "migrated spec context is valid",
			issue: &Issue{
				ID:          "SL-abc123",
				Title:       "Test Issue",
				Status:      StatusOpen,
				Priority:    1,
				IssueType:   TypeTask,
				SpecContext: "migrated",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.issue.Validate()
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status IssueStatus
		want   bool
	}{
		{StatusOpen, true},
		{StatusInProgress, true},
		{StatusClosed, true},
		{IssueStatus("invalid"), false},
		{IssueStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := IsValidStatus(tt.status); got != tt.want {
				t.Errorf("IsValidStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIsValidIssueType(t *testing.T) {
	tests := []struct {
		issueType IssueType
		want      bool
	}{
		{TypeEpic, true},
		{TypeFeature, true},
		{TypeTask, true},
		{TypeBug, true},
		{IssueType("invalid"), false},
		{IssueType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.issueType), func(t *testing.T) {
			if got := IsValidIssueType(tt.issueType); got != tt.want {
				t.Errorf("IsValidIssueType(%q) = %v, want %v", tt.issueType, got, tt.want)
			}
		})
	}
}

func TestDefinitionOfDone(t *testing.T) {
	t.Run("IsComplete nil", func(t *testing.T) {
		var dod *DefinitionOfDone
		if !dod.IsComplete() {
			t.Error("nil DefinitionOfDone should be complete")
		}
	})

	t.Run("IsComplete empty", func(t *testing.T) {
		dod := &DefinitionOfDone{Items: []ChecklistItem{}}
		if !dod.IsComplete() {
			t.Error("empty DefinitionOfDone should be complete")
		}
	})

	t.Run("IsComplete partial", func(t *testing.T) {
		dod := &DefinitionOfDone{
			Items: []ChecklistItem{
				{Item: "Write tests", Checked: true},
				{Item: "Review code", Checked: false},
			},
		}
		if dod.IsComplete() {
			t.Error("partial DefinitionOfDone should not be complete")
		}
	})

	t.Run("IsComplete all checked", func(t *testing.T) {
		dod := &DefinitionOfDone{
			Items: []ChecklistItem{
				{Item: "Write tests", Checked: true},
				{Item: "Review code", Checked: true},
			},
		}
		if !dod.IsComplete() {
			t.Error("all checked DefinitionOfDone should be complete")
		}
	})

	t.Run("CheckItem", func(t *testing.T) {
		dod := &DefinitionOfDone{
			Items: []ChecklistItem{
				{Item: "Write tests", Checked: false},
			},
		}
		if !dod.CheckItem("Write tests") {
			t.Error("CheckItem should return true for existing item")
		}
		if !dod.Items[0].Checked {
			t.Error("item should be checked")
		}
		if dod.CheckItem("nonexistent") {
			t.Error("CheckItem should return false for nonexistent item")
		}
	})

	t.Run("UncheckItem", func(t *testing.T) {
		dod := &DefinitionOfDone{
			Items: []ChecklistItem{
				{Item: "Write tests", Checked: true},
			},
		}
		if !dod.UncheckItem("Write tests") {
			t.Error("UncheckItem should return true for existing item")
		}
		if dod.Items[0].Checked {
			t.Error("item should be unchecked")
		}
		if dod.UncheckItem("nonexistent") {
			t.Error("UncheckItem should return false for nonexistent item")
		}
	})

	t.Run("GetUncheckedItems", func(t *testing.T) {
		dod := &DefinitionOfDone{
			Items: []ChecklistItem{
				{Item: "Write tests", Checked: true},
				{Item: "Review code", Checked: false},
				{Item: "Deploy", Checked: false},
			},
		}
		unchecked := dod.GetUncheckedItems()
		if len(unchecked) != 2 {
			t.Errorf("expected 2 unchecked items, got %d", len(unchecked))
		}
	})
}

func TestIssueIsReady(t *testing.T) {
	now := time.Now()

	t.Run("closed issue not ready", func(t *testing.T) {
		issue := &Issue{Status: StatusClosed}
		if issue.IsReady(nil) {
			t.Error("closed issue should not be ready")
		}
	})

	t.Run("open issue with no blockers is ready", func(t *testing.T) {
		issue := &Issue{Status: StatusOpen, BlockedBy: []string{}}
		if !issue.IsReady(nil) {
			t.Error("open issue with no blockers should be ready")
		}
	})

	t.Run("in_progress issue with no blockers is ready", func(t *testing.T) {
		issue := &Issue{Status: StatusInProgress, BlockedBy: []string{}}
		if !issue.IsReady(nil) {
			t.Error("in_progress issue with no blockers should be ready")
		}
	})

	t.Run("issue with closed blockers is ready", func(t *testing.T) {
		issue := &Issue{
			Status:    StatusOpen,
			BlockedBy: []string{"SL-block1"},
		}
		allIssues := map[string]*Issue{
			"SL-block1": {ID: "SL-block1", Status: StatusClosed},
		}
		if !issue.IsReady(allIssues) {
			t.Error("issue with closed blockers should be ready")
		}
	})

	t.Run("issue with open blocker is not ready", func(t *testing.T) {
		issue := &Issue{
			Status:    StatusOpen,
			BlockedBy: []string{"SL-block1"},
		}
		allIssues := map[string]*Issue{
			"SL-block1": {ID: "SL-block1", Status: StatusOpen},
		}
		if issue.IsReady(allIssues) {
			t.Error("issue with open blocker should not be ready")
		}
	})

	t.Run("issue with missing blocker is ready", func(t *testing.T) {
		issue := &Issue{
			Status:    StatusOpen,
			BlockedBy: []string{"SL-nonexistent"},
		}
		if !issue.IsReady(map[string]*Issue{}) {
			t.Error("issue with missing blocker should be ready")
		}
	})

	t.Run("issue with mixed blockers", func(t *testing.T) {
		issue := &Issue{
			ID:        "SL-test",
			Status:    StatusOpen,
			BlockedBy: []string{"SL-closed", "SL-open"},
		}
		allIssues := map[string]*Issue{
			"SL-closed": {ID: "SL-closed", Status: StatusClosed, UpdatedAt: now},
			"SL-open":   {ID: "SL-open", Status: StatusOpen, UpdatedAt: now},
		}
		if issue.IsReady(allIssues) {
			t.Error("issue with any open blocker should not be ready")
		}
	})
}

func TestIssueGetBlockers(t *testing.T) {
	t.Run("returns blocker details", func(t *testing.T) {
		issue := &Issue{
			BlockedBy: []string{"SL-block1", "SL-block2"},
		}
		allIssues := map[string]*Issue{
			"SL-block1": {ID: "SL-block1", Title: "Blocker 1", Status: StatusOpen},
			"SL-block2": {ID: "SL-block2", Title: "Blocker 2", Status: StatusClosed},
		}
		blockers := issue.GetBlockers(allIssues)
		if len(blockers) != 2 {
			t.Errorf("expected 2 blockers, got %d", len(blockers))
		}
	})

	t.Run("skips missing blockers", func(t *testing.T) {
		issue := &Issue{
			BlockedBy: []string{"SL-missing"},
		}
		blockers := issue.GetBlockers(map[string]*Issue{})
		if len(blockers) != 0 {
			t.Errorf("expected 0 blockers for missing reference, got %d", len(blockers))
		}
	})
}

func TestNewIssue(t *testing.T) {
	issue := NewIssue("Test Title", "Test Description", "010-test", TypeFeature, 2)
	if issue.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", issue.Title)
	}
	if issue.Status != StatusOpen {
		t.Errorf("expected status open, got %s", issue.Status)
	}
	if issue.IssueType != TypeFeature {
		t.Errorf("expected type feature, got %s", issue.IssueType)
	}
	if issue.Priority != 2 {
		t.Errorf("expected priority 2, got %d", issue.Priority)
	}
	if issue.SpecContext != "010-test" {
		t.Errorf("expected spec context '010-test', got %q", issue.SpecContext)
	}
}
