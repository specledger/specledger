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
	tree, err := c.FetchRepoTree("anthropics", "skills", "main")
	if err != nil {
		t.Fatalf("FetchRepoTree: %v", err)
	}
	if len(tree) == 0 {
		t.Fatal("expected tree entries, got none")
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
