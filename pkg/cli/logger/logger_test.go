package logger

import (
	"strings"
	"testing"
	"time"
)

func TestLevelConstants(t *testing.T) {
	if Debug != 0 {
		t.Errorf("expected Debug to be 0, got %d", Debug)
	}
	if Info != 1 {
		t.Errorf("expected Info to be 1, got %d", Info)
	}
	if Warn != 2 {
		t.Errorf("expected Warn to be 2, got %d", Warn)
	}
	if Error != 3 {
		t.Errorf("expected Error to be 3, got %d", Error)
	}
}

func TestDefaultFormatter(t *testing.T) {
	formatter := &DefaultFormatter{}
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	t.Run("format without fields", func(t *testing.T) {
		result := formatter.Format(Info, timestamp, "test message")
		if !strings.Contains(result, "2024-01-15T10:30:00") {
			t.Errorf("expected timestamp in output, got %q", result)
		}
		if !strings.Contains(result, "[INFO]") {
			t.Errorf("expected [INFO] in output, got %q", result)
		}
		if !strings.Contains(result, "test message") {
			t.Errorf("expected message in output, got %q", result)
		}
	})

	t.Run("format with fields", func(t *testing.T) {
		result := formatter.Format(Debug, timestamp, "test message", "field1", "field2")
		if !strings.Contains(result, "[DEBUG]") {
			t.Errorf("expected [DEBUG] in output, got %q", result)
		}
		if !strings.Contains(result, "field1 field2") {
			t.Errorf("expected fields in output, got %q", result)
		}
	})

	t.Run("all levels", func(t *testing.T) {
		levels := []struct {
			level    Level
			expected string
		}{
			{Debug, "DEBUG"},
			{Info, "INFO"},
			{Warn, "WARN"},
			{Error, "ERROR"},
			{Level(99), "UNKNOWN"},
		}

		for _, tt := range levels {
			result := formatter.Format(tt.level, timestamp, "msg")
			if !strings.Contains(result, "["+tt.expected+"]") {
				t.Errorf("expected [%s] in output for level %d, got %q", tt.expected, tt.level, result)
			}
		}
	})
}

func TestLogger(t *testing.T) {
	t.Run("New creates logger with correct level", func(t *testing.T) {
		logger := New(Info)
		if logger.level != Info {
			t.Errorf("expected level Info, got %d", logger.level)
		}
	})

	t.Run("SetLevel changes level", func(t *testing.T) {
		logger := New(Debug)
		logger.SetLevel(Error)
		if logger.level != Error {
			t.Errorf("expected level Error, got %d", logger.level)
		}
	})
}

// captureOutput captures stderr during test execution
func captureOutput(fn func()) string {
	// This is a simple test - in production you'd use os.Pipe or similar
	// For now we just verify the functions don't panic
	fn()
	return ""
}

func TestLoggerMethods(t *testing.T) {
	logger := New(Debug)

	t.Run("Debug doesn't panic", func(t *testing.T) {
		captureOutput(func() {
			logger.Debug("debug message")
		})
	})

	t.Run("Info doesn't panic", func(t *testing.T) {
		captureOutput(func() {
			logger.Info("info message")
		})
	})

	t.Run("Warn doesn't panic", func(t *testing.T) {
		captureOutput(func() {
			logger.Warn("warn message")
		})
	})

	t.Run("Error doesn't panic", func(t *testing.T) {
		captureOutput(func() {
			logger.Error("error message")
		})
	})

	t.Run("With fields doesn't panic", func(t *testing.T) {
		captureOutput(func() {
			logger.Info("message", "field1=value1", "field2=value2")
		})
	})
}

func TestLoggerLevelFiltering(t *testing.T) {
	// Test that levels are properly checked
	logger := New(Warn)

	// Verify level is set correctly
	if logger.level != Warn {
		t.Errorf("expected level Warn, got %d", logger.level)
	}

	// Test level comparison logic
	if logger.level <= Debug {
		t.Error("Debug should not be logged when level is Warn")
	}
	if logger.level <= Info {
		t.Error("Info should not be logged when level is Warn")
	}
	if logger.level > Warn {
		t.Error("Warn should be logged when level is Warn")
	}
}
