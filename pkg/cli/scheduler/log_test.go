package scheduler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteHookLog_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, "nested", "logs")

	if err := WriteHookLog(logDir, "127-feature", "127-feature", "triggered", "spawned"); err != nil {
		t.Fatalf("WriteHookLog() error: %v", err)
	}

	logPath := filepath.Join(logDir, hookLogFileName)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("log file not created")
	}
}

func TestWriteHookLog_StructuredFormat(t *testing.T) {
	logDir := t.TempDir()

	if err := WriteHookLog(logDir, "127-feat", "127-feat", "skip", "not approved"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(logDir, hookLogFileName))
	content := string(data)
	if !strings.Contains(content, "branch=127-feat") {
		t.Error("missing branch field")
	}
	if !strings.Contains(content, "action=skip") {
		t.Error("missing action field")
	}
	if !strings.Contains(content, "detail=not approved") {
		t.Error("missing detail field")
	}
}

func TestWriteHookLog_AppendsEntries(t *testing.T) {
	logDir := t.TempDir()

	_ = WriteHookLog(logDir, "b1", "f1", "a1", "d1")
	_ = WriteHookLog(logDir, "b2", "f2", "a2", "d2")

	data, _ := os.ReadFile(filepath.Join(logDir, hookLogFileName))
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 entries, got %d", len(lines))
	}
}

func TestWriteHookLog_Rotation(t *testing.T) {
	logDir := t.TempDir()

	// Write more than maxLogEntries
	for i := 0; i < maxLogEntries+10; i++ {
		if err := WriteHookLog(logDir, "b", "f", "a", "d"); err != nil {
			t.Fatal(err)
		}
	}

	data, _ := os.ReadFile(filepath.Join(logDir, hookLogFileName))
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != maxLogEntries {
		t.Errorf("expected %d entries after rotation, got %d", maxLogEntries, len(lines))
	}
}
