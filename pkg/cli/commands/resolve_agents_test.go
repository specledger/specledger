package commands

import (
	"strings"
	"testing"
)

func TestResolveAgentFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		want    string
		wantErr string
	}{
		{
			name:  "single agent claude",
			flags: []string{"claude"},
			want:  "Claude Code",
		},
		{
			name:  "single agent opencode",
			flags: []string{"opencode"},
			want:  "OpenCode",
		},
		{
			name:  "single agent codex",
			flags: []string{"codex"},
			want:  "Codex",
		},
		{
			name:  "single agent copilot alias",
			flags: []string{"copilot"},
			want:  "Copilot CLI",
		},
		{
			name:  "single agent github-copilot command",
			flags: []string{"github-copilot"},
			want:  "Copilot CLI",
		},
		{
			name:  "multiple agents",
			flags: []string{"claude", "opencode"},
			want:  "Claude Code,OpenCode",
		},
		{
			name:  "all agents",
			flags: []string{"all"},
			want:  "Claude Code,OpenCode,Copilot CLI,Codex",
		},
		{
			name:    "all combined with other",
			flags:   []string{"all", "claude"},
			wantErr: "cannot be combined",
		},
		{
			name:    "invalid agent",
			flags:   []string{"vim"},
			wantErr: "unknown agent \"vim\"",
		},
		{
			name:    "invalid agent includes valid values",
			flags:   []string{"vim"},
			wantErr: "Valid values: claude, opencode, codex, copilot, all",
		},
		{
			name:    "empty string",
			flags:   []string{""},
			wantErr: "--agent value cannot be empty",
		},
		{
			name:    "whitespace only",
			flags:   []string{"  "},
			wantErr: "--agent value cannot be empty",
		},
		{
			name:  "case insensitive",
			flags: []string{"Claude"},
			want:  "Claude Code",
		},
		{
			name:  "dedup",
			flags: []string{"claude", "claude"},
			want:  "Claude Code",
		},
		{
			name:  "three agents",
			flags: []string{"claude", "opencode", "codex"},
			want:  "Claude Code,OpenCode,Codex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAgentFlags(tt.flags)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
