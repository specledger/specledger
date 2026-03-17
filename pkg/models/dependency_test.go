package models

import (
	"testing"
)

func TestDependencyString(t *testing.T) {
	t.Run("with alias", func(t *testing.T) {
		d := &Dependency{
			RepositoryURL: "https://github.com/test/repo",
			Alias:         "my-dep",
		}
		if got := d.String(); got != "my-dep -> https://github.com/test/repo" {
			t.Errorf("String() = %q, want %q", got, "my-dep -> https://github.com/test/repo")
		}
	})

	t.Run("without alias", func(t *testing.T) {
		d := &Dependency{
			RepositoryURL: "https://github.com/test/repo",
		}
		if got := d.String(); got != "https://github.com/test/repo" {
			t.Errorf("String() = %q, want %q", got, "https://github.com/test/repo")
		}
	})
}

func TestDependencyValidate(t *testing.T) {
	tests := []struct {
		name    string
		dep     Dependency
		wantErr bool
	}{
		{
			name: "valid dependency",
			dep: Dependency{
				RepositoryURL: "https://github.com/test/repo",
				Version:       "main",
				SpecPath:      "specledger/spec.md",
			},
			wantErr: false,
		},
		{
			name: "empty repository URL",
			dep: Dependency{
				Version:  "main",
				SpecPath: "specledger/spec.md",
			},
			wantErr: true,
		},
		{
			name: "empty version",
			dep: Dependency{
				RepositoryURL: "https://github.com/test/repo",
				SpecPath:      "specledger/spec.md",
			},
			wantErr: true,
		},
		{
			name: "empty spec path",
			dep: Dependency{
				RepositoryURL: "https://github.com/test/repo",
				Version:       "main",
			},
			wantErr: true,
		},
		{
			name: "valid alias",
			dep: Dependency{
				RepositoryURL: "https://github.com/test/repo",
				Version:       "main",
				SpecPath:      "specledger/spec.md",
				Alias:         "my-dep",
			},
			wantErr: false,
		},
		{
			name: "alias with underscore",
			dep: Dependency{
				RepositoryURL: "https://github.com/test/repo",
				Version:       "main",
				SpecPath:      "specledger/spec.md",
				Alias:         "my_dep",
			},
			wantErr: false,
		},
		{
			name: "alias too long",
			dep: Dependency{
				RepositoryURL: "https://github.com/test/repo",
				Version:       "main",
				SpecPath:      "specledger/spec.md",
				Alias:         string(make([]byte, 51)),
			},
			wantErr: true,
		},
		{
			name: "alias with invalid char",
			dep: Dependency{
				RepositoryURL: "https://github.com/test/repo",
				Version:       "main",
				SpecPath:      "specledger/spec.md",
				Alias:         "my@dep",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dep.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDependencyManifest(t *testing.T) {
	t.Run("NewDependencyManifest", func(t *testing.T) {
		m := NewDependencyManifest("1.0", "test-id", "specledger/spec.mod")
		if m.Version != "1.0" {
			t.Errorf("expected version 1.0, got %s", m.Version)
		}
		if m.ID != "test-id" {
			t.Errorf("expected id test-id, got %s", m.ID)
		}
		if len(m.Dependencies) != 0 {
			t.Errorf("expected empty dependencies, got %d", len(m.Dependencies))
		}
	})

	t.Run("AddDependency", func(t *testing.T) {
		m := NewDependencyManifest("1.0", "test-id", "specledger/spec.mod")
		dep := Dependency{
			RepositoryURL: "https://github.com/test/repo",
			Version:       "main",
			SpecPath:      "specledger/spec.md",
		}
		if err := m.AddDependency(dep); err != nil {
			t.Fatalf("AddDependency() error: %v", err)
		}
		if len(m.Dependencies) != 1 {
			t.Errorf("expected 1 dependency, got %d", len(m.Dependencies))
		}
	})

	t.Run("AddDependency duplicate", func(t *testing.T) {
		m := NewDependencyManifest("1.0", "test-id", "specledger/spec.mod")
		dep := Dependency{
			RepositoryURL: "https://github.com/test/repo",
			Version:       "main",
			SpecPath:      "specledger/spec.md",
		}
		_ = m.AddDependency(dep)
		err := m.AddDependency(dep)
		if err == nil {
			t.Error("expected error for duplicate dependency")
		}
	})

	t.Run("AddDependency invalid", func(t *testing.T) {
		m := NewDependencyManifest("1.0", "test-id", "specledger/spec.mod")
		dep := Dependency{} // Missing required fields
		err := m.AddDependency(dep)
		if err == nil {
			t.Error("expected error for invalid dependency")
		}
	})

	t.Run("RemoveDependency", func(t *testing.T) {
		m := NewDependencyManifest("1.0", "test-id", "specledger/spec.mod")
		dep := Dependency{
			RepositoryURL: "https://github.com/test/repo",
			Version:       "main",
			SpecPath:      "specledger/spec.md",
		}
		_ = m.AddDependency(dep)

		removed := m.RemoveDependency("https://github.com/test/repo", "specledger/spec.md")
		if !removed {
			t.Error("expected dependency to be removed")
		}
		if len(m.Dependencies) != 0 {
			t.Errorf("expected 0 dependencies, got %d", len(m.Dependencies))
		}
	})

	t.Run("RemoveDependency not found", func(t *testing.T) {
		m := NewDependencyManifest("1.0", "test-id", "specledger/spec.mod")
		removed := m.RemoveDependency("https://github.com/nonexistent", "spec.md")
		if removed {
			t.Error("expected dependency not to be removed")
		}
	})
}

func TestLockfileModels(t *testing.T) {
	t.Run("NewLockfile", func(t *testing.T) {
		lf := NewLockfile("1.0")
		if lf.Version != "1.0" {
			t.Errorf("expected version 1.0, got %s", lf.Version)
		}
		if len(lf.Entries) != 0 {
			t.Errorf("expected empty entries, got %d", len(lf.Entries))
		}
	})

	t.Run("AddEntry", func(t *testing.T) {
		lf := NewLockfile("1.0")
		entry := LockfileEntry{
			RepositoryURL: "https://github.com/test/repo",
			CommitHash:    "abc123",
			Size:          1024,
		}
		lf.AddEntry(entry)
		if len(lf.Entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(lf.Entries))
		}
		if lf.TotalSize != 1024 {
			t.Errorf("expected total size 1024, got %d", lf.TotalSize)
		}
	})
}
