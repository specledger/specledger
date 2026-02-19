package issues

import (
	"errors"
	"regexp"
	"time"
)

// IssueStatus represents the current state of an issue
type IssueStatus string

const (
	StatusOpen       IssueStatus = "open"
	StatusInProgress IssueStatus = "in_progress"
	StatusClosed     IssueStatus = "closed"
)

// IssueType represents the category of an issue
type IssueType string

const (
	TypeEpic    IssueType = "epic"
	TypeFeature IssueType = "feature"
	TypeTask    IssueType = "task"
	TypeBug     IssueType = "bug"
)

// Issue represents a tracking unit for work items
type Issue struct {
	// Required fields
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Status      IssueStatus `json:"status"`
	Priority    int         `json:"priority"` // 0=highest, 5=lowest
	IssueType   IssueType   `json:"issue_type"`
	SpecContext string      `json:"spec_context"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`

	// Optional fields
	ClosedAt           *time.Time        `json:"closed_at,omitempty"`
	DefinitionOfDone   *DefinitionOfDone `json:"definition_of_done,omitempty"`
	BlockedBy          []string          `json:"blocked_by,omitempty"` // Issue IDs
	Blocks             []string          `json:"blocks,omitempty"`     // Issue IDs
	Labels             []string          `json:"labels,omitempty"`
	Assignee           string            `json:"assignee,omitempty"`
	Notes              string            `json:"notes,omitempty"`
	Design             string            `json:"design,omitempty"`
	AcceptanceCriteria string            `json:"acceptance_criteria,omitempty"`

	// Migration metadata (optional, for Beads migration)
	BeadsMigration *BeadsMigration `json:"beads_migration,omitempty"`
}

// BeadsMigration contains metadata for issues migrated from Beads
type BeadsMigration struct {
	OriginalID string    `json:"original_id"`
	MigratedAt time.Time `json:"migrated_at"`
}

// DefinitionOfDone represents a checklist that must be completed before closing
type DefinitionOfDone struct {
	Items []ChecklistItem `json:"items"`
}

// ChecklistItem represents a single item in a definition of done checklist
type ChecklistItem struct {
	Item       string     `json:"item"`
	Checked    bool       `json:"checked"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

// IssueUpdate represents partial updates to an issue
type IssueUpdate struct {
	Title              *string
	Description        *string
	Status             *IssueStatus
	Priority           *int
	IssueType          *IssueType
	Assignee           *string
	Notes              *string
	Design             *string
	AcceptanceCriteria *string
	Labels             *[]string
	AddLabels          []string
	RemoveLabels       []string
	BlockedBy          *[]string
	Blocks             *[]string
	DefinitionOfDone   *DefinitionOfDone
	CheckDoDItem       string // Item to mark as checked
	UncheckDoDItem     string // Item to mark as unchecked
}

// ListFilter represents filtering options for listing issues
type ListFilter struct {
	Status      *IssueStatus
	IssueType   *IssueType
	Priority    *int
	Labels      []string
	SpecContext string // Empty = all specs
	All         bool   // Search across all specs
	Blocked     bool   // Only show blocked issues
}

// Validation errors
var (
	ErrInvalidID          = errors.New("issue ID must match format SL-xxxxxx")
	ErrInvalidTitle       = errors.New("title is required and must be 1-200 characters")
	ErrInvalidStatus      = errors.New("status must be one of: open, in_progress, closed")
	ErrInvalidPriority    = errors.New("priority must be between 0 and 5")
	ErrInvalidIssueType   = errors.New("issue type must be one of: epic, feature, task, bug")
	ErrInvalidSpecContext = errors.New("spec context must match pattern ###-name")
)

var (
	idPattern          = regexp.MustCompile(`^SL-[a-f0-9]{6}$`)
	specContextPattern = regexp.MustCompile(`^\d{3,}-[a-z0-9-]+$`)
)

// Validate validates all fields of an issue
func (i *Issue) Validate() error {
	if !idPattern.MatchString(i.ID) {
		return ErrInvalidID
	}
	if len(i.Title) == 0 || len(i.Title) > 200 {
		return ErrInvalidTitle
	}
	if !IsValidStatus(i.Status) {
		return ErrInvalidStatus
	}
	if i.Priority < 0 || i.Priority > 5 {
		return ErrInvalidPriority
	}
	if !IsValidIssueType(i.IssueType) {
		return ErrInvalidIssueType
	}
	if i.SpecContext != "" && !isValidSpecContext(i.SpecContext) {
		return ErrInvalidSpecContext
	}
	return nil
}

// isValidSpecContext checks if a spec context is valid (either matches pattern or is a special directory)
func isValidSpecContext(specContext string) bool {
	// Allow standard spec pattern (###-name)
	if specContextPattern.MatchString(specContext) {
		return true
	}
	// Allow special directories for migrated issues
	if specContext == "migrated" {
		return true
	}
	return false
}

// IsValidStatus checks if a status is valid
func IsValidStatus(s IssueStatus) bool {
	switch s {
	case StatusOpen, StatusInProgress, StatusClosed:
		return true
	default:
		return false
	}
}

// IsValidIssueType checks if an issue type is valid
func IsValidIssueType(t IssueType) bool {
	switch t {
	case TypeEpic, TypeFeature, TypeTask, TypeBug:
		return true
	default:
		return false
	}
}

// IsComplete returns true if all checklist items are checked
func (d *DefinitionOfDone) IsComplete() bool {
	if d == nil || len(d.Items) == 0 {
		return true // No DoD means it's complete
	}
	for _, item := range d.Items {
		if !item.Checked {
			return false
		}
	}
	return true
}

// CheckItem marks an item as checked
func (d *DefinitionOfDone) CheckItem(itemText string) bool {
	for i, item := range d.Items {
		if item.Item == itemText {
			now := time.Now()
			d.Items[i].Checked = true
			d.Items[i].VerifiedAt = &now
			return true
		}
	}
	return false
}

// UncheckItem marks an item as unchecked
func (d *DefinitionOfDone) UncheckItem(itemText string) bool {
	for i, item := range d.Items {
		if item.Item == itemText {
			d.Items[i].Checked = false
			d.Items[i].VerifiedAt = nil
			return true
		}
	}
	return false
}

// GetUncheckedItems returns all unchecked items
func (d *DefinitionOfDone) GetUncheckedItems() []string {
	var unchecked []string
	for _, item := range d.Items {
		if !item.Checked {
			unchecked = append(unchecked, item.Item)
		}
	}
	return unchecked
}

// NewIssue creates a new issue with defaults
func NewIssue(title, description, specContext string, issueType IssueType, priority int) *Issue {
	now := time.Now()
	return &Issue{
		ID:          GenerateIssueID(specContext, title, now),
		Title:       title,
		Description: description,
		Status:      StatusOpen,
		Priority:    priority,
		IssueType:   issueType,
		SpecContext: specContext,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
