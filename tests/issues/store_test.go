package issues_test

import (
	"os"
	"testing"
	"time"

	"github.com/specledger/specledger/pkg/issues"
)

func TestStore_ParentValidation(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create spec context directory
	specDir := tmpDir + "/010-test-feature"
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec dir: %v", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    tmpDir,
		SpecContext: "010-test-feature",
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create parent issue
	parent := &issues.Issue{
		ID:          "SL-aaaaaa",
		Title:       "Parent Issue",
		Status:      issues.StatusOpen,
		Priority:    1,
		IssueType:   issues.TypeFeature,
		SpecContext: "010-test-feature",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = store.Create(parent)
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	// Create child issue with parent
	child := &issues.Issue{
		ID:          "SL-bbbbbb",
		Title:       "Child Issue",
		Status:      issues.StatusOpen,
		Priority:    2,
		IssueType:   issues.TypeTask,
		SpecContext: "010-test-feature",
		ParentID:    strPtr("SL-aaaaaa"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = store.Create(child)
	if err != nil {
		t.Fatalf("Failed to create child with parent: %v", err)
	}

	// Verify child was created with parent
	retrieved, err := store.Get("SL-bbbbbb")
	if err != nil {
		t.Fatalf("Failed to get child: %v", err)
	}
	if retrieved.ParentID == nil || *retrieved.ParentID != "SL-aaaaaa" {
		t.Error("Child should have parent set")
	}
}

func TestStore_SelfParentValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create spec context directory
	specDir := tmpDir + "/010-test-feature"
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec dir: %v", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    tmpDir,
		SpecContext: "010-test-feature",
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Try to create issue with self as parent
	issue := &issues.Issue{
		ID:          "SL-cccccc",
		Title:       "Self Parent Issue",
		Status:      issues.StatusOpen,
		Priority:    1,
		IssueType:   issues.TypeTask,
		SpecContext: "010-test-feature",
		ParentID:    strPtr("SL-cccccc"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = store.Create(issue)
	if err == nil {
		t.Error("Should not allow self as parent")
	}
}

func TestStore_CycleDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create spec context directory
	specDir := tmpDir + "/010-test-feature"
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec dir: %v", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    tmpDir,
		SpecContext: "010-test-feature",
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create grandparent
	grandparent := &issues.Issue{
		ID:          "SL-dddddd",
		Title:       "Grandparent",
		Status:      issues.StatusOpen,
		Priority:    0,
		IssueType:   issues.TypeEpic,
		SpecContext: "010-test-feature",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = store.Create(grandparent)
	if err != nil {
		t.Fatalf("Failed to create grandparent: %v", err)
	}

	// Create parent with grandparent
	parent := &issues.Issue{
		ID:          "SL-eeeeee",
		Title:       "Parent",
		Status:      issues.StatusOpen,
		Priority:    1,
		IssueType:   issues.TypeFeature,
		SpecContext: "010-test-feature",
		ParentID:    strPtr("SL-dddddd"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = store.Create(parent)
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	// Try to update grandparent to have parent as child (would create cycle)
	update := &issues.IssueUpdate{
		ParentID: strPtr("SL-eeeeee"),
	}
	_, err = store.Update("SL-dddddd", *update)
	if err == nil {
		t.Error("Should detect circular parent-child relationship")
	}
}

func TestStore_GetChildren(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create spec context directory
	specDir := tmpDir + "/010-test-feature"
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec dir: %v", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    tmpDir,
		SpecContext: "010-test-feature",
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create parent
	parent := &issues.Issue{
		ID:          "SL-ffffff",
		Title:       "Parent",
		Status:      issues.StatusOpen,
		Priority:    1,
		IssueType:   issues.TypeFeature,
		SpecContext: "010-test-feature",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = store.Create(parent)
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	// Create children with different priorities
	child1 := &issues.Issue{
		ID:          "SL-111111",
		Title:       "Child 1",
		Status:      issues.StatusOpen,
		Priority:    2,
		IssueType:   issues.TypeTask,
		SpecContext: "010-test-feature",
		ParentID:    strPtr("SL-ffffff"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	child2 := &issues.Issue{
		ID:          "SL-222222",
		Title:       "Child 2",
		Status:      issues.StatusOpen,
		Priority:    1, // Higher priority (lower number)
		IssueType:   issues.TypeTask,
		SpecContext: "010-test-feature",
		ParentID:    strPtr("SL-ffffff"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = store.Create(child1)
	if err != nil {
		t.Fatalf("Failed to create child1: %v", err)
	}
	err = store.Create(child2)
	if err != nil {
		t.Fatalf("Failed to create child2: %v", err)
	}

	// Get children
	children, err := store.GetChildren("SL-ffffff")
	if err != nil {
		t.Fatalf("Failed to get children: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}

	// Children should be ordered by priority (descending), then ID (ascending)
	// child2 has priority 1 (higher), child1 has priority 2 (lower)
	if len(children) >= 2 {
		if children[0].ID != "SL-222222" {
			t.Errorf("First child should be SL-222222 (higher priority), got %s", children[0].ID)
		}
		if children[1].ID != "SL-111111" {
			t.Errorf("Second child should be SL-111111 (lower priority), got %s", children[1].ID)
		}
	}
}

func TestStore_DoDCheckUncheck(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create spec context directory
	specDir := tmpDir + "/010-test-feature"
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec dir: %v", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    tmpDir,
		SpecContext: "010-test-feature",
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create issue with DoD
	issue := &issues.Issue{
		ID:          "SL-333333",
		Title:       "DoD Test",
		Status:      issues.StatusOpen,
		Priority:    1,
		IssueType:   issues.TypeTask,
		SpecContext: "010-test-feature",
		DefinitionOfDone: &issues.DefinitionOfDone{
			Items: []issues.ChecklistItem{
				{Item: "Write tests", Checked: false},
				{Item: "Code review", Checked: false},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = store.Create(issue)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	// Check DoD item
	update := &issues.IssueUpdate{
		CheckDoDItem: "Write tests",
	}
	updated, err := store.Update("SL-333333", *update)
	if err != nil {
		t.Fatalf("Failed to check DoD item: %v", err)
	}

	// Verify item was checked
	if updated.DefinitionOfDone == nil || len(updated.DefinitionOfDone.Items) == 0 {
		t.Fatal("DefinitionOfDone should not be nil/empty")
	}
	if !updated.DefinitionOfDone.Items[0].Checked {
		t.Error("First DoD item should be checked")
	}
	if updated.DefinitionOfDone.Items[0].VerifiedAt == nil {
		t.Error("Checked item should have VerifiedAt set")
	}

	// Try to check non-existent item
	update2 := &issues.IssueUpdate{
		CheckDoDItem: "Non-existent item",
	}
	_, err = store.Update("SL-333333", *update2)
	if err == nil {
		t.Error("Should return error when checking non-existent DoD item")
	}

	// Uncheck DoD item
	update3 := &issues.IssueUpdate{
		UncheckDoDItem: "Write tests",
	}
	updated, err = store.Update("SL-333333", *update3)
	if err != nil {
		t.Fatalf("Failed to uncheck DoD item: %v", err)
	}

	// Verify item was unchecked
	if updated.DefinitionOfDone.Items[0].Checked {
		t.Error("First DoD item should be unchecked")
	}
	if updated.DefinitionOfDone.Items[0].VerifiedAt != nil {
		t.Error("Unchecked item should have VerifiedAt cleared")
	}
}

func TestStore_DoDReplacement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create spec context directory
	specDir := tmpDir + "/010-test-feature"
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec dir: %v", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    tmpDir,
		SpecContext: "010-test-feature",
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create issue with DoD
	issue := &issues.Issue{
		ID:          "SL-444444",
		Title:       "DoD Replace Test",
		Status:      issues.StatusOpen,
		Priority:    1,
		IssueType:   issues.TypeTask,
		SpecContext: "010-test-feature",
		DefinitionOfDone: &issues.DefinitionOfDone{
			Items: []issues.ChecklistItem{
				{Item: "Old item 1", Checked: true},
				{Item: "Old item 2", Checked: false},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = store.Create(issue)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	// Replace DoD
	newDoD := &issues.DefinitionOfDone{
		Items: []issues.ChecklistItem{
			{Item: "New item 1", Checked: false},
			{Item: "New item 2", Checked: false},
			{Item: "New item 3", Checked: false},
		},
	}
	update := &issues.IssueUpdate{
		DefinitionOfDone: newDoD,
	}
	updated, err := store.Update("SL-444444", *update)
	if err != nil {
		t.Fatalf("Failed to replace DoD: %v", err)
	}

	// Verify DoD was replaced
	if len(updated.DefinitionOfDone.Items) != 3 {
		t.Errorf("Expected 3 DoD items, got %d", len(updated.DefinitionOfDone.Items))
	}
	if updated.DefinitionOfDone.Items[0].Item != "New item 1" {
		t.Errorf("First item should be 'New item 1', got '%s'", updated.DefinitionOfDone.Items[0].Item)
	}
}

func TestStore_ParentNotExistValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "specledger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create spec context directory
	specDir := tmpDir + "/010-test-feature"
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec dir: %v", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    tmpDir,
		SpecContext: "010-test-feature",
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Try to create issue with non-existent parent
	issue := &issues.Issue{
		ID:          "SL-555555",
		Title:       "Orphan Issue",
		Status:      issues.StatusOpen,
		Priority:    1,
		IssueType:   issues.TypeTask,
		SpecContext: "010-test-feature",
		ParentID:    strPtr("SL-999999"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = store.Create(issue)
	if err == nil {
		t.Error("Should return error when parent does not exist")
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
