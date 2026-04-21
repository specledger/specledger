package skills

import (
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// newVCRClient creates a Client backed by a go-vcr cassette for replay.
// Uses real API URLs since cassettes were recorded against real backends.
func newVCRClient(t *testing.T, cassetteName string) *Client {
	t.Helper()
	rec, err := recorder.New(
		"../../../tests/testdata/cassettes/skills/"+cassetteName,
		recorder.WithMode(recorder.ModeReplayOnly),
		recorder.WithSkipRequestLatency(true),
	)
	if err != nil {
		t.Fatalf("recorder.New(%s): %v", cassetteName, err)
	}
	t.Cleanup(func() { _ = rec.Stop() })

	return &Client{
		SearchURL:  defaultSearchURL,
		AuditURL:   defaultAuditURL,
		GitHubURL:  defaultGitHubURL,
		RawGHURL:   defaultRawGHURL,
		HTTPClient: rec.GetDefaultClient(),
	}
}

func TestVCR_Search(t *testing.T) {
	c := newVCRClient(t, "search")
	results, err := c.Search("web design", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results, got none")
	}
	// Verify structure from real API
	for _, r := range results {
		if r.Name == "" {
			t.Error("result has empty Name")
		}
		if r.Source == "" {
			t.Error("result has empty Source")
		}
	}
	t.Logf("got %d results, first: %s (%d installs)", len(results), results[0].Name, results[0].Installs)
}

func TestVCR_SearchEmpty(t *testing.T) {
	c := newVCRClient(t, "search_empty")
	results, err := c.Search("xyznonexistent99999", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestVCR_Audit(t *testing.T) {
	c := newVCRClient(t, "audit")
	results, err := c.FetchAudit("anthropics/skills", []string{"skill-creator"})
	if err != nil {
		t.Fatalf("FetchAudit: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected audit results, got none")
	}
	// Verify at least one partner has data
	for slug, r := range results {
		hasPartner := r.ATH != nil || r.Socket != nil || r.Snyk != nil
		if !hasPartner {
			t.Errorf("slug %q has no partner data", slug)
		}
	}
}

func TestVCR_GitHubTrees(t *testing.T) {
	c := newVCRClient(t, "github_trees")
	tree, resolvedRef, _, err := c.FetchRepoTree("anthropics", "skills", "main")
	if err != nil {
		t.Fatalf("FetchRepoTree: %v", err)
	}
	if len(tree) == 0 {
		t.Fatal("expected tree entries, got none")
	}
	if resolvedRef != "main" {
		t.Errorf("resolvedRef = %q, want %q", resolvedRef, "main")
	}

	// Verify SKILL.md entries exist in the anthropics/skills repo
	skillCount := 0
	for _, entry := range tree {
		if entry.Type == "blob" && contains(entry.Path, "SKILL.md") {
			skillCount++
		}
	}
	if skillCount == 0 {
		t.Error("no SKILL.md entries found in tree")
	}
	t.Logf("got %d tree entries, %d SKILL.md files", len(tree), skillCount)
}

func TestVCR_GitHubRaw(t *testing.T) {
	c := newVCRClient(t, "github_raw")
	data, err := c.FetchSkillContent("anthropics", "skills", "main", "skills/skill-creator/SKILL.md")
	if err != nil {
		t.Fatalf("FetchSkillContent: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected content, got empty")
	}
	// Verify it's a valid SKILL.md
	if !contains(string(data), "skill-creator") || !contains(string(data), "skill") {
		t.Error("content doesn't appear to be the skill-creator SKILL.md")
	}
	t.Logf("got %d bytes of SKILL.md content", len(data))
}

// Issue #172: repos with non-main default branches

func TestVCR_GitHubTreesNonMainDefault_404(t *testing.T) {
	// Reproduces #172: Trees API returns 404 when repo uses "dev" not "main"
	c := newVCRClient(t, "github_trees_nonmain_404")
	_, _, _, err := c.FetchRepoTree("different-ai", "openwork", "main")
	if err == nil {
		t.Fatal("expected error for 404 on main branch")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error = %q, want containing 'not found'", err.Error())
	}
}

func TestVCR_GitHubTreesNonMainDefault_HEAD(t *testing.T) {
	// Verifies HEAD ref resolves to the repo's actual default branch
	c := newVCRClient(t, "github_trees_nonmain")
	tree, resolvedRef, _, err := c.FetchRepoTree("different-ai", "openwork", "HEAD")
	if err != nil {
		t.Fatalf("FetchRepoTree: %v", err)
	}
	if resolvedRef != "HEAD" {
		t.Errorf("resolvedRef = %q, want %q", resolvedRef, "HEAD")
	}

	// Verify SKILL.md entries exist (different-ai/openwork has 13 skills)
	skillCount := 0
	for _, entry := range tree {
		if entry.Type == "blob" && contains(entry.Path, "SKILL.md") {
			skillCount++
		}
	}
	if skillCount == 0 {
		t.Error("no SKILL.md entries found in tree")
	}
	t.Logf("got %d tree entries, %d SKILL.md files", len(tree), skillCount)
}

func TestVCR_DiscoverNonMainDefault(t *testing.T) {
	// End-to-end: DiscoverSkills with empty ref triggers HEAD fallback
	c := newVCRClient(t, "discover_nonmain")
	source := &SkillSource{
		Owner:       "different-ai",
		Repo:        "openwork",
		SkillFilter: "opencode-primitives",
		Ref:         "", // auto-resolve
		Type:        SourceTypeGitHub,
	}

	skills, err := DiscoverSkills(c, source)
	if err != nil {
		t.Fatalf("DiscoverSkills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("len(skills) = %d, want 1", len(skills))
	}
	if skills[0].Name != "opencode-primitives" {
		t.Errorf("Name = %q, want %q", skills[0].Name, "opencode-primitives")
	}
	// Verify the ref was resolved (not still empty)
	if source.Ref == "" {
		t.Error("source.Ref should be resolved after discovery, got empty")
	}
	t.Logf("discovered %s, ref resolved to %q", skills[0].Name, source.Ref)
}

func TestVCR_GitHubRawNonMain(t *testing.T) {
	c := newVCRClient(t, "github_raw_nonmain")
	data, err := c.FetchSkillContent("different-ai", "openwork", "HEAD", ".opencode/skills/opencode-primitives/SKILL.md")
	if err != nil {
		t.Fatalf("FetchSkillContent: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected content, got empty")
	}
	// Verify it parses as a valid skill
	meta, err := ParseSkillFrontmatter(data)
	if err != nil {
		t.Fatalf("ParseSkillFrontmatter: %v", err)
	}
	if meta.Name != "opencode-primitives" {
		t.Errorf("Name = %q, want %q", meta.Name, "opencode-primitives")
	}
	t.Logf("got %d bytes, skill name: %s", len(data), meta.Name)
}

func TestVCR_FetchSkillFiles_MultiFile(t *testing.T) {
	c := newVCRClient(t, "github_multi_file")
	source := &SkillSource{
		Owner: "test-org",
		Repo:  "multi-repo",
		Ref:   "main",
		Type:  SourceTypeGitHub,
	}

	// Manually construct the SkillMetadata as discovery would produce it.
	// This tests FetchSkillFiles in isolation (discovery has its own tests).
	skill := SkillMetadata{
		Name:     "multi-skill",
		RepoPath: "skills/multi-skill/SKILL.md",
		Files: []string{
			"skills/multi-skill/GENERATION.md",
			"skills/multi-skill/SKILL.md",
			"skills/multi-skill/references/advanced.md",
			"skills/multi-skill/references/core.md",
		},
	}

	files, err := FetchSkillFiles(c, source, &skill)
	if err != nil {
		t.Fatalf("FetchSkillFiles: %v", err)
	}

	// Verify all 4 files returned with correct relative keys
	wantKeys := []string{"SKILL.md", "GENERATION.md", "references/advanced.md", "references/core.md"}
	if len(files) != len(wantKeys) {
		t.Fatalf("got %d files, want %d: %v", len(files), len(wantKeys), mapKeys(files))
	}
	for _, key := range wantKeys {
		if _, ok := files[key]; !ok {
			t.Errorf("missing key %q in files map", key)
		}
		if len(files[key]) == 0 {
			t.Errorf("file %q is empty", key)
		}
	}

	// Verify content fidelity
	if !contains(string(files["references/core.md"]), "Core concepts") {
		t.Errorf("references/core.md content wrong: %q", string(files["references/core.md"]))
	}

	t.Logf("FetchSkillFiles returned %d files", len(files))
}

func mapKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
