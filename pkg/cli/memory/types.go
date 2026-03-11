package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// EntryStatus represents the promotion state of a knowledge entry
type EntryStatus string

const (
	StatusCandidate EntryStatus = "candidate"
	StatusPromoted  EntryStatus = "promoted"
	StatusArchived  EntryStatus = "archived"
)

// Validation constants
const (
	MaxTitleLength       = 200
	MaxDescriptionLength = 5000
	MaxTagCount          = 10
	MaxTagLength         = 50
	MinScore             = 0.0
	MaxScore             = 10.0
	PromotionThreshold   = 7.0
)

// Validation errors
var (
	ErrEmptyTitle           = errors.New("title is required")
	ErrTitleTooLong         = errors.New("title must be 200 characters or less")
	ErrEmptyDescription     = errors.New("description is required")
	ErrDescriptionTooLong   = errors.New("description must be 5000 characters or less")
	ErrNoTags               = errors.New("at least one tag is required")
	ErrTooManyTags          = errors.New("maximum 10 tags allowed")
	ErrTagTooLong           = errors.New("each tag must be 50 characters or less")
	ErrEmptyTag             = errors.New("tags must not be empty")
	ErrScoreOutOfRange      = errors.New("score must be between 0.0 and 10.0")
	ErrInvalidStatus        = errors.New("status must be candidate, promoted, or archived")
	ErrEntryNotFound        = errors.New("knowledge entry not found")
	ErrEntryAlreadyExists   = errors.New("knowledge entry already exists")
)

// Score holds the three-axis scoring dimensions for a knowledge entry.
type Score struct {
	Recurrence  float64 `json:"recurrence"`
	Impact      float64 `json:"impact"`
	Specificity float64 `json:"specificity"`
	Composite   float64 `json:"composite"`
}

// KnowledgeEntry represents a structured piece of organizational knowledge
// extracted from session transcripts.
type KnowledgeEntry struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	Tags            []string    `json:"tags"`
	SourceSessionID string      `json:"source_session_id,omitempty"`
	SourceBranch    string      `json:"source_branch,omitempty"`
	Scores          Score       `json:"scores"`
	Status          EntryStatus `json:"status"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	RecurrenceCount int         `json:"recurrence_count"`
}

// GenerateEntryID creates a deterministic ID for a knowledge entry
// using SHA-256 hash of title + created_at.
func GenerateEntryID(title string, createdAt time.Time) string {
	data := fmt.Sprintf("knowledge|%s|%d", title, createdAt.UnixNano())
	hash := sha256.Sum256([]byte(data))
	hexPart := hex.EncodeToString(hash[:3])
	return "KE-" + hexPart
}

// Validate validates all fields of a KnowledgeEntry.
func (e *KnowledgeEntry) Validate() error {
	if len(e.Title) == 0 {
		return ErrEmptyTitle
	}
	if len(e.Title) > MaxTitleLength {
		return ErrTitleTooLong
	}
	if len(e.Description) == 0 {
		return ErrEmptyDescription
	}
	if len(e.Description) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}
	if len(e.Tags) == 0 {
		return ErrNoTags
	}
	if len(e.Tags) > MaxTagCount {
		return ErrTooManyTags
	}
	for _, tag := range e.Tags {
		if len(tag) == 0 {
			return ErrEmptyTag
		}
		if len(tag) > MaxTagLength {
			return ErrTagTooLong
		}
	}
	if err := e.Scores.Validate(); err != nil {
		return err
	}
	if !IsValidStatus(e.Status) {
		return ErrInvalidStatus
	}
	return nil
}

// Validate checks that all score axes are within valid range.
func (s *Score) Validate() error {
	if s.Recurrence < MinScore || s.Recurrence > MaxScore {
		return fmt.Errorf("recurrence %w: %.1f", ErrScoreOutOfRange, s.Recurrence)
	}
	if s.Impact < MinScore || s.Impact > MaxScore {
		return fmt.Errorf("impact %w: %.1f", ErrScoreOutOfRange, s.Impact)
	}
	if s.Specificity < MinScore || s.Specificity > MaxScore {
		return fmt.Errorf("specificity %w: %.1f", ErrScoreOutOfRange, s.Specificity)
	}
	return nil
}

// IsValidStatus checks if a status value is valid.
func IsValidStatus(s EntryStatus) bool {
	switch s {
	case StatusCandidate, StatusPromoted, StatusArchived:
		return true
	default:
		return false
	}
}

// NewKnowledgeEntry creates a new KnowledgeEntry with defaults.
func NewKnowledgeEntry(title, description string, tags []string, scores Score) *KnowledgeEntry {
	now := time.Now()
	entry := &KnowledgeEntry{
		ID:              GenerateEntryID(title, now),
		Title:           title,
		Description:     description,
		Tags:            tags,
		Scores:          scores,
		Status:          StatusCandidate,
		CreatedAt:       now,
		UpdatedAt:       now,
		RecurrenceCount: 1,
	}
	return entry
}
