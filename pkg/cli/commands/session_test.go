package commands

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"30 days", "30d", 30 * 24 * time.Hour, false},
		{"7 days", "7d", 7 * 24 * time.Hour, false},
		{"1 day", "1d", 24 * time.Hour, false},
		{"90 days", "90d", 90 * 24 * time.Hour, false},
		{"invalid format", "30h", 0, true},
		{"no number", "d", 0, true},
		{"empty", "", 0, true},
		{"negative", "-5d", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 500, "500 B"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", 1048576, "1.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestSessionStatsComputation(t *testing.T) {
	// Test that SessionStats struct can be properly initialized
	stats := SessionStats{
		TotalSessions:    5,
		TotalSize:        1024,
		TotalRawSize:     4096,
		AvgMessages:      10.5,
		SessionsByBranch: map[string]int{"main": 3, "feat/test": 2},
		SizeByBranch:     map[string]int64{"main": 600, "feat/test": 424},
	}

	if stats.TotalSessions != 5 {
		t.Errorf("TotalSessions = %d, want 5", stats.TotalSessions)
	}
	if stats.SessionsByBranch["main"] != 3 {
		t.Errorf("SessionsByBranch[main] = %d, want 3", stats.SessionsByBranch["main"])
	}
}
