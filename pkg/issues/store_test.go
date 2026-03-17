package issues

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	t.Run("with spec context", func(t *testing.T) {
		opts := StoreOptions{
			BasePath:    "specledger",
			SpecContext: "010-test-feature",
		}
		store, err := NewStore(opts)
		if err != nil {
			t.Fatalf("NewStore() error: %v", err)
		}
		if store.Path() != "specledger/010-test-feature/issues.jsonl" {
			t.Errorf("unexpected path: %s", store.Path())
		}
	})

	t.Run("without spec context", func(t *testing.T) {
		opts := StoreOptions{
			BasePath: "specledger",
		}
		store, err := NewStore(opts)
		if err != nil {
			t.Fatalf("NewStore() error: %v", err)
		}
		if store.Path() != "specledger" {
			t.Errorf("unexpected path: %s", store.Path())
		}
	})

	t.Run("default base path", func(t *testing.T) {
		opts := StoreOptions{
			SpecContext: "010-test",
		}
		store, err := NewStore(opts)
		if err != nil {
			t.Fatalf("NewStore() error: %v", err)
		}
		expected := "specledger/010-test/issues.jsonl"
		if store.Path() != expected {
			t.Errorf("expected path %s, got %s", expected, store.Path())
		}
	})
}

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "specledger")
	specContext := "010-test"

	// Create the spec directory
	specDir := filepath.Join(basePath, specContext)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	store, err := NewStore(StoreOptions{
		BasePath:    basePath,
		SpecContext: specContext,
	})
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	return store
}

func TestStoreCRUD(t *testing.T) {
	store := setupTestStore(t)

	// Freeze time for consistent testing
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	NowFunc = func() time.Time { return now }
	defer func() { NowFunc = time.Now }()

	t.Run("Create", func(t *testing.T) {
		issue := &Issue{
			ID:          "SL-abc123",
			Title:       "Test Issue",
			Description: "Test description",
			Status:      StatusOpen,
			Priority:    1,
			IssueType:   TypeTask,
			SpecContext: "010-test",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := store.Create(issue); err != nil {
			t.Fatalf("Create() error: %v", err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		issue, err := store.Get("SL-abc123")
		if err != nil {
			t.Fatalf("Get() error: %v", err)
		}
		if issue.Title != "Test Issue" {
			t.Errorf("expected title 'Test Issue', got %q", issue.Title)
		}
	})

	t.Run("Get not found", func(t *testing.T) {
		_, err := store.Get("SL-nonexistent")
		if err != ErrIssueNotFound {
			t.Errorf("expected ErrIssueNotFound, got %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		issues, err := store.List(ListFilter{})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(issues) != 1 {
			t.Errorf("expected 1 issue, got %d", len(issues))
		}
	})

	t.Run("Update", func(t *testing.T) {
		newTitle := "Updated Title"
		issue, err := store.Update("SL-abc123", IssueUpdate{
			Title: &newTitle,
		})
		if err != nil {
			t.Fatalf("Update() error: %v", err)
		}
		if issue.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got %q", issue.Title)
		}
	})

	t.Run("Update not found", func(t *testing.T) {
		_, err := store.Update("SL-nonexistent", IssueUpdate{})
		if err != ErrIssueNotFound {
			t.Errorf("expected ErrIssueNotFound, got %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := store.Delete("SL-abc123"); err != nil {
			t.Fatalf("Delete() error: %v", err)
		}

		issues, err := store.List(ListFilter{})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected 0 issues after delete, got %d", len(issues))
		}
	})

	t.Run("Delete not found", func(t *testing.T) {
		err := store.Delete("SL-nonexistent")
		if err != ErrIssueNotFound {
			t.Errorf("expected ErrIssueNotFound, got %v", err)
		}
	})
}

func TestStoreCreateDuplicate(t *testing.T) {
	store := setupTestStore(t)

	now := time.Now()
	issue := &Issue{
		ID:          "SL-abc123",
		Title:       "Test Issue",
		Status:      StatusOpen,
		Priority:    1,
		IssueType:   TypeTask,
		SpecContext: "010-test",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := store.Create(issue); err != nil {
		t.Fatalf("first Create() error: %v", err)
	}

	err := store.Create(issue)
	if err != ErrIssueAlreadyExists {
		t.Errorf("expected ErrIssueAlreadyExists, got %v", err)
	}
}

func TestStoreListFilter(t *testing.T) {
	store := setupTestStore(t)

	now := time.Now()

	// Create test issues
	issues := []*Issue{
		{ID: "SL-aaa111", Title: "Open Task", Status: StatusOpen, Priority: 1, IssueType: TypeTask, SpecContext: "010-test", CreatedAt: now, UpdatedAt: now},
		{ID: "SL-bbb222", Title: "Closed Bug", Status: StatusClosed, Priority: 2, IssueType: TypeBug, SpecContext: "010-test", CreatedAt: now, UpdatedAt: now},
		{ID: "SL-ccc333", Title: "In Progress Feature", Status: StatusInProgress, Priority: 3, IssueType: TypeFeature, SpecContext: "010-test", CreatedAt: now, UpdatedAt: now},
	}

	for _, issue := range issues {
		if err := store.Create(issue); err != nil {
			t.Fatalf("Create() error: %v", err)
		}
	}

	t.Run("filter by status", func(t *testing.T) {
		status := StatusOpen
		result, err := store.List(ListFilter{Status: &status})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 open issue, got %d", len(result))
		}
	})

	t.Run("filter by issue type", func(t *testing.T) {
		issueType := TypeBug
		result, err := store.List(ListFilter{IssueType: &issueType})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 bug issue, got %d", len(result))
		}
	})

	t.Run("filter by priority", func(t *testing.T) {
		priority := 2
		result, err := store.List(ListFilter{Priority: &priority})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 issue with priority 2, got %d", len(result))
		}
	})
}

func TestStoreParentChild(t *testing.T) {
	store := setupTestStore(t)

	now := time.Now()

	// Create parent issue
	parent := &Issue{
		ID:          "SL-aaaaaa",
		Title:       "Parent Issue",
		Status:      StatusOpen,
		Priority:    1,
		IssueType:   TypeEpic,
		SpecContext: "010-test",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := store.Create(parent); err != nil {
		t.Fatalf("Create parent error: %v", err)
	}

	// Create child issue
	child := &Issue{
		ID:          "SL-bbbbbb",
		Title:       "Child Issue",
		Status:      StatusOpen,
		Priority:    1,
		IssueType:   TypeTask,
		SpecContext: "010-test",
		ParentID:    strPtr("SL-aaaaaa"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := store.Create(child); err != nil {
		t.Fatalf("Create child error: %v", err)
	}

	t.Run("GetChildren", func(t *testing.T) {
		children, err := store.GetChildren("SL-aaaaaa")
		if err != nil {
			t.Fatalf("GetChildren() error: %v", err)
		}
		if len(children) != 1 {
			t.Errorf("expected 1 child, got %d", len(children))
		}
	})

	t.Run("self-parent prevention", func(t *testing.T) {
		child := &Issue{
			ID:          "SL-cccccc",
			Title:       "New Issue",
			Status:      StatusOpen,
			Priority:    1,
			IssueType:   TypeTask,
			SpecContext: "010-test",
			ParentID:    strPtr("SL-cccccc"),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := store.Create(child)
		if err == nil {
			t.Error("expected error for self-parent")
		}
	})

	t.Run("nonexistent parent prevention", func(t *testing.T) {
		child := &Issue{
			ID:          "SL-dddddd",
			Title:       "New Issue 2",
			Status:      StatusOpen,
			Priority:    1,
			IssueType:   TypeTask,
			SpecContext: "010-test",
			ParentID:    strPtr("SL-xxxxxx"),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := store.Create(child)
		if err == nil {
			t.Error("expected error for nonexistent parent")
		}
	})
}

func TestStoreUpdateLabels(t *testing.T) {
	store := setupTestStore(t)

	now := time.Now()
	issue := &Issue{
		ID:          "SL-eeeeee",
		Title:       "Test Issue",
		Status:      StatusOpen,
		Priority:    1,
		IssueType:   TypeTask,
		SpecContext: "010-test",
		Labels:      []string{"existing"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := store.Create(issue); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	t.Run("add labels", func(t *testing.T) {
		updated, err := store.Update("SL-eeeeee", IssueUpdate{
			AddLabels: []string{"new1", "new2"},
		})
		if err != nil {
			t.Fatalf("Update() error: %v", err)
		}
		if len(updated.Labels) != 3 {
			t.Errorf("expected 3 labels, got %d: %v", len(updated.Labels), updated.Labels)
		}
	})

	t.Run("remove labels", func(t *testing.T) {
		updated, err := store.Update("SL-eeeeee", IssueUpdate{
			RemoveLabels: []string{"existing"},
		})
		if err != nil {
			t.Fatalf("Update() error: %v", err)
		}
		if len(updated.Labels) != 2 {
			t.Errorf("expected 2 labels after removal, got %d: %v", len(updated.Labels), updated.Labels)
		}
	})

	t.Run("set labels", func(t *testing.T) {
		newLabels := []string{"a", "b"}
		updated, err := store.Update("SL-eeeeee", IssueUpdate{
			Labels: &newLabels,
		})
		if err != nil {
			t.Fatalf("Update() error: %v", err)
		}
		if len(updated.Labels) != 2 {
			t.Errorf("expected 2 labels, got %d", len(updated.Labels))
		}
	})
}

func TestStoreUpdateStatusWithClosedAt(t *testing.T) {
	store := setupTestStore(t)

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	NowFunc = func() time.Time { return now }
	defer func() { NowFunc = time.Now }()

	issue := &Issue{
		ID:          "SL-ffffff",
		Title:       "Test Issue",
		Status:      StatusOpen,
		Priority:    1,
		IssueType:   TypeTask,
		SpecContext: "010-test",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := store.Create(issue); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	closed := StatusClosed
	updated, err := store.Update("SL-ffffff", IssueUpdate{
		Status: &closed,
	})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	if updated.Status != StatusClosed {
		t.Errorf("expected status closed, got %s", updated.Status)
	}
	if updated.ClosedAt == nil {
		t.Error("expected ClosedAt to be set")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"a"}, "a", true},
	}

	for _, tt := range tests {
		if got := contains(tt.slice, tt.item); got != tt.expected {
			t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.expected)
		}
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
