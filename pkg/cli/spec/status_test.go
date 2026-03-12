package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadStatus(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		want     string
		wantErr  bool
	}{
		{
			name:    "reads Draft status",
			content: "# Feature\n\n**Status**: Draft\n\nSome content",
			want:    "Draft",
		},
		{
			name:    "reads Approved status",
			content: "# Feature\n\n**Status**: Approved\n\nSome content",
			want:    "Approved",
		},
		{
			name:    "empty status field",
			content: "# Feature\n\n**Status**:\n\nSome content",
			wantErr: true,
		},
		{
			name:    "no status field",
			content: "# Feature\n\nSome content",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := ReadStatus(dir)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ReadStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadStatus_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadStatus(dir)
	if err == nil {
		t.Error("expected error for missing spec.md")
	}
}

func TestWriteStatus(t *testing.T) {
	dir := t.TempDir()
	original := "# Feature\n\n**Status**: Draft\n\nSome content"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteStatus(dir, "Approved"); err != nil {
		t.Fatalf("WriteStatus() error: %v", err)
	}

	got, err := ReadStatus(dir)
	if err != nil {
		t.Fatalf("ReadStatus() error: %v", err)
	}
	if got != "Approved" {
		t.Errorf("after WriteStatus, got %q, want %q", got, "Approved")
	}
}

func TestWriteStatus_NoField(t *testing.T) {
	dir := t.TempDir()
	content := "# Feature\n\nNo status here"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := WriteStatus(dir, "Approved")
	if err == nil {
		t.Error("expected error for missing status field")
	}
}
