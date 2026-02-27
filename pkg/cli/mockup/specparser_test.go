package mockup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpec_Basic(t *testing.T) {
	dir := t.TempDir()
	specContent := `# User Registration Feature

## Overview
A registration system for new users.

## User Stories

### US1: Basic Registration
**As a** new visitor, **I want to** create an account **so that** I can access the platform.

### US2: Email Verification
**As a** new user, **I want to** verify my email **so that** my account is secured.

## Functional Requirements

- FR-001: System shall display a registration form
- FR-002: System shall validate email format
- FR-003: System shall send verification email
`
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0600)

	sc, err := ParseSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}

	if sc.Title != "User Registration Feature" {
		t.Errorf("Title = %q, want %q", sc.Title, "User Registration Feature")
	}

	if len(sc.UserStories) == 0 {
		t.Error("expected at least one user story")
	}

	if len(sc.Requirements) == 0 {
		t.Error("expected at least one requirement")
	}

	if sc.FullContent == "" {
		t.Error("expected non-empty FullContent")
	}
}

func TestParseSpec_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(""), 0600)

	_, err := ParseSpec(specPath)
	if err == nil {
		t.Error("expected error for empty spec")
	}
}

func TestParseSpec_MissingFile(t *testing.T) {
	_, err := ParseSpec("/nonexistent/spec.md")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseSpec_NoUserStories(t *testing.T) {
	dir := t.TempDir()
	specContent := `# Simple Feature

## Overview
Just a simple feature with no user stories.
`
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0600)

	sc, err := ParseSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(sc.UserStories) != 0 {
		t.Errorf("expected 0 user stories, got %d", len(sc.UserStories))
	}
}
