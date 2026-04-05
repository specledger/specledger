package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// skillsTestEnv sets up a mock API server and a project for skills testing.
type skillsTestEnv struct {
	slBinary    string
	projectPath string
	mockServer  *httptest.Server
	mux         *http.ServeMux
	t           *testing.T
}

func newSkillsTestEnv(t *testing.T) *skillsTestEnv {
	t.Helper()
	tempDir := t.TempDir()
	slBinary := buildSLBinary(t, tempDir)

	// Create a SpecLedger project
	projectPath := filepath.Join(tempDir, "test-skills-project")
	cmd := exec.Command(slBinary, "new", "--ci",
		"--project-name", "test-skills-project",
		"--short-code", "tsp",
		"--project-dir", tempDir)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create test project: %v\nOutput: %s", err, string(output))
	}

	// Create .claude/skills/ directory
	if err := os.MkdirAll(filepath.Join(projectPath, ".claude", "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	env := &skillsTestEnv{
		slBinary:    slBinary,
		projectPath: projectPath,
		mockServer:  srv,
		mux:         mux,
		t:           t,
	}

	// Register default mock handlers
	env.registerSearchHandler()
	env.registerTreesHandler()
	env.registerRawContentHandler()
	env.registerAuditHandler()
	env.registerTelemetryHandler()

	t.Cleanup(func() { srv.Close() })
	return env
}

func (e *skillsTestEnv) run(args ...string) (string, error) {
	cmd := exec.Command(e.slBinary, args...) // #nosec G204 -- test binary path is controlled
	cmd.Dir = e.projectPath
	cmd.Env = append(os.Environ(),
		"SKILLS_API_URL="+e.mockServer.URL,
		"SKILLS_AUDIT_URL="+e.mockServer.URL,
		"GITHUB_API_URL="+e.mockServer.URL,
		"GITHUB_RAW_URL="+e.mockServer.URL,
		"DISABLE_TELEMETRY=1",
	)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (e *skillsTestEnv) registerSearchHandler() {
	e.mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		type searchResp struct {
			Skills []map[string]interface{} `json:"skills"`
		}
		var resp searchResp
		if q == "xyznonexistent" {
			resp.Skills = []map[string]interface{}{}
		} else {
			resp.Skills = []map[string]interface{}{
				{"id": "test-org/test-repo/test-skill", "name": "test-skill", "source": "test-org/test-repo", "installs": 1234},
				{"id": "test-org/test-repo/other-skill", "name": "other-skill", "source": "test-org/test-repo", "installs": 567},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

func (e *skillsTestEnv) registerTreesHandler() {
	e.mux.HandleFunc("/repos/test-org/test-repo/git/trees/main", func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"sha": "abc123",
			"tree": []map[string]interface{}{
				{"path": "skills/test-skill/SKILL.md", "type": "blob"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

func (e *skillsTestEnv) registerRawContentHandler() {
	e.mux.HandleFunc("/test-org/test-repo/main/skills/test-skill/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("---\nname: test-skill\ndescription: A test skill for integration testing\n---\n# Test Skill\nThis is a test skill.\n"))
	})
}

func (e *skillsTestEnv) registerAuditHandler() {
	e.mux.HandleFunc("/audit", func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"test-skill": map[string]interface{}{
				"ath":    map[string]interface{}{"risk": "safe", "alerts": 0, "score": 100, "analyzedAt": "2026-01-01T00:00:00Z"},
				"socket": map[string]interface{}{"risk": "low", "alerts": 0, "score": 95, "analyzedAt": "2026-01-01T00:00:00Z"},
				"snyk":   map[string]interface{}{"risk": "safe", "alerts": 0, "score": 100, "analyzedAt": "2026-01-01T00:00:00Z"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

func (e *skillsTestEnv) registerTelemetryHandler() {
	e.mux.HandleFunc("/t", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// --- Search Tests ---

func TestSkillsSearch(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "search", "test")
	if err != nil {
		t.Fatalf("sl skill search failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "test-skill") {
		t.Errorf("output missing 'test-skill': %s", output)
	}
	if !strings.Contains(output, "1.2K installs") {
		t.Errorf("output missing install count: %s", output)
	}
}

func TestSkillsSearchJSON(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "search", "test", "--json")
	if err != nil {
		t.Fatalf("sl skill search --json failed: %v\nOutput: %s", err, output)
	}
	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("invalid JSON output: %v\nOutput: %s", err, output)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSkillsSearchNoResults(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "search", "xyznonexistent")
	if err != nil {
		t.Fatalf("sl skill search failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "No skills found") {
		t.Errorf("expected no-results message, got: %s", output)
	}
}

// --- Add Tests ---

func TestSkillsAdd(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "add", "test-org/test-repo@test-skill", "-y")
	if err != nil {
		t.Fatalf("sl skill add failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Installed test-skill") {
		t.Errorf("output missing install confirmation: %s", output)
	}
	if !strings.Contains(output, "Updated skills-lock.json") {
		t.Errorf("output missing lock file update: %s", output)
	}

	// Verify SKILL.md was written
	skillFile := filepath.Join(env.projectPath, ".claude", "skills", "test-skill", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Error("SKILL.md not created")
	}

	// Verify lock file updated
	lockPath := filepath.Join(env.projectPath, "skills-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("lock file not created: %v", err)
	}
	if !strings.Contains(string(data), "test-skill") {
		t.Error("lock file missing test-skill entry")
	}
}

func TestSkillsAddJSON(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "add", "test-org/test-repo@test-skill", "-y", "--json")
	if err != nil {
		t.Fatalf("sl skill add --json failed: %v\nOutput: %s", err, output)
	}
	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestSkillsAddInvalidSource(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "add", "invalid-source")
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
	if !strings.Contains(output, "invalid source") {
		t.Errorf("expected invalid source error, got: %s", output)
	}
}

// --- List Tests ---

func TestSkillsList(t *testing.T) {
	env := newSkillsTestEnv(t)

	// Install a skill first
	_, err := env.run("skill", "add", "test-org/test-repo@test-skill", "-y")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	output, err := env.run("skill", "list")
	if err != nil {
		t.Fatalf("sl skill list failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "test-skill") {
		t.Errorf("output missing 'test-skill': %s", output)
	}
	if !strings.Contains(output, "test-org/test-repo") {
		t.Errorf("output missing source: %s", output)
	}
}

func TestSkillsListEmpty(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "list")
	if err != nil {
		t.Fatalf("sl skill list failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected empty state message, got: %s", output)
	}
}

func TestSkillsListJSON(t *testing.T) {
	env := newSkillsTestEnv(t)
	_, _ = env.run("skill", "add", "test-org/test-repo@test-skill", "-y")

	output, err := env.run("skill", "list", "--json")
	if err != nil {
		t.Fatalf("sl skill list --json failed: %v\nOutput: %s", err, output)
	}
	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}
}

// --- Remove Tests ---

func TestSkillsRemove(t *testing.T) {
	env := newSkillsTestEnv(t)

	// Install first
	_, err := env.run("skill", "add", "test-org/test-repo@test-skill", "-y")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	output, err := env.run("skill", "remove", "test-skill")
	if err != nil {
		t.Fatalf("sl skill remove failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Removed test-skill") {
		t.Errorf("output missing removal confirmation: %s", output)
	}

	// Verify file removed
	skillFile := filepath.Join(env.projectPath, ".claude", "skills", "test-skill", "SKILL.md")
	if _, err := os.Stat(skillFile); !os.IsNotExist(err) {
		t.Error("SKILL.md still exists after remove")
	}
}

func TestSkillsRemoveNotInstalled(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "remove", "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-installed skill")
	}
	if !strings.Contains(output, "not installed") {
		t.Errorf("expected not-installed error, got: %s", output)
	}
}

func TestSkillsRemoveJSON(t *testing.T) {
	env := newSkillsTestEnv(t)
	_, _ = env.run("skill", "add", "test-org/test-repo@test-skill", "-y")

	output, err := env.run("skill", "remove", "test-skill", "--json")
	if err != nil {
		t.Fatalf("sl skill remove --json failed: %v\nOutput: %s", err, output)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}
	if result["status"] != "removed" {
		t.Errorf("expected status 'removed', got %v", result["status"])
	}
}

// --- Audit Tests ---

func TestSkillsAudit(t *testing.T) {
	env := newSkillsTestEnv(t)
	_, _ = env.run("skill", "add", "test-org/test-repo@test-skill", "-y")

	output, err := env.run("skill", "audit")
	if err != nil {
		t.Fatalf("sl skill audit failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Security Risk Assessments") {
		t.Errorf("output missing audit header: %s", output)
	}
	if !strings.Contains(output, "Safe") {
		t.Errorf("output missing risk level: %s", output)
	}
	if !strings.Contains(output, "No high or critical") {
		t.Errorf("output missing safety message: %s", output)
	}
}

func TestSkillsAuditJSON(t *testing.T) {
	env := newSkillsTestEnv(t)
	_, _ = env.run("skill", "add", "test-org/test-repo@test-skill", "-y")

	output, err := env.run("skill", "audit", "--json")
	if err != nil {
		t.Fatalf("sl skill audit --json failed: %v\nOutput: %s", err, output)
	}
	var results map[string]interface{}
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}
}

// --- Info Tests ---

func TestSkillsInfo(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "info", "test-org/test-repo@test-skill")
	if err != nil {
		t.Fatalf("sl skill info failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "test-skill") {
		t.Errorf("output missing skill name: %s", output)
	}
	if !strings.Contains(output, "test-org/test-repo") {
		t.Errorf("output missing source: %s", output)
	}
}

func TestSkillsInfoJSON(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "info", "test-org/test-repo@test-skill", "--json")
	if err != nil {
		t.Fatalf("sl skill info --json failed: %v\nOutput: %s", err, output)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}
	if result["name"] != "test-skill" {
		t.Errorf("name = %v, want test-skill", result["name"])
	}
}

// --- Additional Search Tests ---

func TestSkillsSearchLimit(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "search", "test", "--limit", "1", "--json")
	if err != nil {
		t.Fatalf("sl skill search --limit failed: %v\nOutput: %s", err, output)
	}
	// Server returns all results but --limit is passed as query param
	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, output)
	}
}

// --- Additional Add Tests ---

func TestSkillsAddOverwrite(t *testing.T) {
	env := newSkillsTestEnv(t)

	// Install first
	_, err := env.run("skill", "add", "test-org/test-repo@test-skill", "-y")
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Install again with -y (should overwrite without error)
	output, err := env.run("skill", "add", "test-org/test-repo@test-skill", "-y")
	if err != nil {
		t.Fatalf("overwrite install: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Installed test-skill") {
		t.Errorf("output missing install confirmation: %s", output)
	}
}

// --- Additional Audit Tests ---

func TestSkillsAuditSingle(t *testing.T) {
	env := newSkillsTestEnv(t)
	_, _ = env.run("skill", "add", "test-org/test-repo@test-skill", "-y")

	output, err := env.run("skill", "audit", "test-skill")
	if err != nil {
		t.Fatalf("sl skill audit test-skill failed: %v\nOutput: %s", err, output)
	}
	if !strings.Contains(output, "Security Risk Assessments") {
		t.Errorf("output missing audit header: %s", output)
	}
}

func TestSkillsAuditNotInstalled(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "audit", "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-installed skill audit")
	}
	if !strings.Contains(output, "not installed") {
		t.Errorf("expected not-installed error, got: %s", output)
	}
}

// --- Error Tests ---

func TestSkillsErrorInvalidSource(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "add", "no-slash-here")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(output, "invalid source") || !strings.Contains(output, "owner/repo") {
		t.Errorf("expected format guidance, got: %s", output)
	}
}

func TestSkillsErrorCorruptLock(t *testing.T) {
	env := newSkillsTestEnv(t)

	// Write corrupt lock file
	lockPath := filepath.Join(env.projectPath, "skills-lock.json")
	if err := os.WriteFile(lockPath, []byte("{corrupt json!!!}"), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := env.run("skill", "list")
	if err == nil {
		t.Fatal("expected error for corrupt lock file")
	}
	if !strings.Contains(output, "invalid") {
		t.Errorf("expected invalid JSON error, got: %s", output)
	}
}

func TestSkillsListJSONEmpty(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "list", "--json")
	if err != nil {
		t.Fatalf("sl skill list --json failed: %v\nOutput: %s", err, output)
	}
	// Must return [] not null
	trimmed := strings.TrimSpace(output)
	if trimmed != "[]" {
		t.Errorf("expected empty JSON array '[]', got: %q", trimmed)
	}
}

func TestSkillsSearchJSONEmpty(t *testing.T) {
	env := newSkillsTestEnv(t)
	output, err := env.run("skill", "search", "xyznonexistent", "--json")
	if err != nil {
		t.Fatalf("sl skill search --json failed: %v\nOutput: %s", err, output)
	}
	// Must return [] not null
	trimmed := strings.TrimSpace(output)
	if trimmed != "[]" {
		t.Errorf("expected empty JSON array '[]', got: %q", trimmed)
	}
}
