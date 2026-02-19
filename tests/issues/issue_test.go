package issues_test

import (
	"testing"
	"time"

	"github.com/specledger/specledger/pkg/issues"
)

func TestIssue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		issue   *issues.Issue
		wantErr error
	}{
		{
			name: "valid issue",
			issue: &issues.Issue{
				ID:          "SL-a3f5d8",
				Title:       "Add validation",
				Status:      issues.StatusOpen,
				Priority:    1,
				IssueType:   issues.TypeTask,
				SpecContext: "010-my-feature",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "invalid ID format",
			issue: &issues.Issue{
				ID:          "invalid",
				Title:       "Add validation",
				Status:      issues.StatusOpen,
				Priority:    1,
				IssueType:   issues.TypeTask,
				SpecContext: "010-my-feature",
			},
			wantErr: issues.ErrInvalidID,
		},
		{
			name: "empty title",
			issue: &issues.Issue{
				ID:          "SL-a3f5d8",
				Title:       "",
				Status:      issues.StatusOpen,
				Priority:    1,
				IssueType:   issues.TypeTask,
				SpecContext: "010-my-feature",
			},
			wantErr: issues.ErrInvalidTitle,
		},
		{
			name: "invalid status",
			issue: &issues.Issue{
				ID:          "SL-a3f5d8",
				Title:       "Add validation",
				Status:      issues.IssueStatus("invalid"),
				Priority:    1,
				IssueType:   issues.TypeTask,
				SpecContext: "010-my-feature",
			},
			wantErr: issues.ErrInvalidStatus,
		},
		{
			name: "priority too high",
			issue: &issues.Issue{
				ID:          "SL-a3f5d8",
				Title:       "Add validation",
				Status:      issues.StatusOpen,
				Priority:    6,
				IssueType:   issues.TypeTask,
				SpecContext: "010-my-feature",
			},
			wantErr: issues.ErrInvalidPriority,
		},
		{
			name: "priority negative",
			issue: &issues.Issue{
				ID:          "SL-a3f5d8",
				Title:       "Add validation",
				Status:      issues.StatusOpen,
				Priority:    -1,
				IssueType:   issues.TypeTask,
				SpecContext: "010-my-feature",
			},
			wantErr: issues.ErrInvalidPriority,
		},
		{
			name: "invalid issue type",
			issue: &issues.Issue{
				ID:          "SL-a3f5d8",
				Title:       "Add validation",
				Status:      issues.StatusOpen,
				Priority:    1,
				IssueType:   issues.IssueType("invalid"),
				SpecContext: "010-my-feature",
			},
			wantErr: issues.ErrInvalidIssueType,
		},
		{
			name: "invalid spec context",
			issue: &issues.Issue{
				ID:          "SL-a3f5d8",
				Title:       "Add validation",
				Status:      issues.StatusOpen,
				Priority:    1,
				IssueType:   issues.TypeTask,
				SpecContext: "invalid-context",
			},
			wantErr: issues.ErrInvalidSpecContext,
		},
		{
			name: "migrated spec context is valid",
			issue: &issues.Issue{
				ID:          "SL-a3f5d8",
				Title:       "Add validation",
				Status:      issues.StatusOpen,
				Priority:    1,
				IssueType:   issues.TypeTask,
				SpecContext: "migrated",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.issue.Validate()
			if err != tt.wantErr {
				t.Errorf("Issue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status   issues.IssueStatus
		expected bool
	}{
		{issues.StatusOpen, true},
		{issues.StatusInProgress, true},
		{issues.StatusClosed, true},
		{issues.IssueStatus("invalid"), false},
		{issues.IssueStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := issues.IsValidStatus(tt.status)
			if result != tt.expected {
				t.Errorf("IsValidStatus(%v) = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestIsValidIssueType(t *testing.T) {
	tests := []struct {
		issueType issues.IssueType
		expected  bool
	}{
		{issues.TypeEpic, true},
		{issues.TypeFeature, true},
		{issues.TypeTask, true},
		{issues.TypeBug, true},
		{issues.IssueType("invalid"), false},
		{issues.IssueType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.issueType), func(t *testing.T) {
			result := issues.IsValidIssueType(tt.issueType)
			if result != tt.expected {
				t.Errorf("IsValidIssueType(%v) = %v, want %v", tt.issueType, result, tt.expected)
			}
		})
	}
}

func TestNewIssue(t *testing.T) {
	title := "Add validation"
	description := "Implement input validation"
	specContext := "010-my-feature"
	issueType := issues.TypeTask
	priority := 1

	issue := issues.NewIssue(title, description, specContext, issueType, priority)

	if issue.Title != title {
		t.Errorf("NewIssue().Title = %v, want %v", issue.Title, title)
	}
	if issue.Description != description {
		t.Errorf("NewIssue().Description = %v, want %v", issue.Description, description)
	}
	if issue.SpecContext != specContext {
		t.Errorf("NewIssue().SpecContext = %v, want %v", issue.SpecContext, specContext)
	}
	if issue.IssueType != issueType {
		t.Errorf("NewIssue().IssueType = %v, want %v", issue.IssueType, issueType)
	}
	if issue.Priority != priority {
		t.Errorf("NewIssue().Priority = %v, want %v", issue.Priority, priority)
	}
	if issue.Status != issues.StatusOpen {
		t.Errorf("NewIssue().Status = %v, want %v", issue.Status, issues.StatusOpen)
	}
	if !isValidIssueIDFormat(issue.ID) {
		t.Errorf("NewIssue().ID = %v, want valid format", issue.ID)
	}
	if issue.CreatedAt.IsZero() {
		t.Error("NewIssue().CreatedAt should not be zero")
	}
	if issue.UpdatedAt.IsZero() {
		t.Error("NewIssue().UpdatedAt should not be zero")
	}
}

func TestDefinitionOfDone_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		dod      *issues.DefinitionOfDone
		expected bool
	}{
		{
			name:     "nil DoD",
			dod:      nil,
			expected: true,
		},
		{
			name:     "empty DoD",
			dod:      &issues.DefinitionOfDone{Items: []issues.ChecklistItem{}},
			expected: true,
		},
		{
			name: "all checked",
			dod: &issues.DefinitionOfDone{
				Items: []issues.ChecklistItem{
					{Item: "Test", Checked: true},
					{Item: "Review", Checked: true},
				},
			},
			expected: true,
		},
		{
			name: "some unchecked",
			dod: &issues.DefinitionOfDone{
				Items: []issues.ChecklistItem{
					{Item: "Test", Checked: true},
					{Item: "Review", Checked: false},
				},
			},
			expected: false,
		},
		{
			name: "all unchecked",
			dod: &issues.DefinitionOfDone{
				Items: []issues.ChecklistItem{
					{Item: "Test", Checked: false},
					{Item: "Review", Checked: false},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dod.IsComplete()
			if result != tt.expected {
				t.Errorf("DefinitionOfDone.IsComplete() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefinitionOfDone_CheckItem(t *testing.T) {
	dod := &issues.DefinitionOfDone{
		Items: []issues.ChecklistItem{
			{Item: "Write tests", Checked: false},
			{Item: "Code review", Checked: false},
		},
	}

	// Check first item
	result := dod.CheckItem("Write tests")
	if !result {
		t.Error("CheckItem() should return true when item exists")
	}
	if !dod.Items[0].Checked {
		t.Error("CheckItem() should set Checked to true")
	}
	if dod.Items[0].VerifiedAt == nil {
		t.Error("CheckItem() should set VerifiedAt")
	}

	// Check non-existent item
	result = dod.CheckItem("Non-existent")
	if result {
		t.Error("CheckItem() should return false when item doesn't exist")
	}
}

func TestDefinitionOfDone_GetUncheckedItems(t *testing.T) {
	dod := &issues.DefinitionOfDone{
		Items: []issues.ChecklistItem{
			{Item: "Write tests", Checked: true},
			{Item: "Code review", Checked: false},
			{Item: "Update docs", Checked: false},
		},
	}

	unchecked := dod.GetUncheckedItems()
	if len(unchecked) != 2 {
		t.Errorf("GetUncheckedItems() returned %d items, want 2", len(unchecked))
	}
	if unchecked[0] != "Code review" {
		t.Errorf("GetUncheckedItems()[0] = %v, want 'Code review'", unchecked[0])
	}
}
