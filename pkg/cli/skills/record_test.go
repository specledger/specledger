package skills

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// TestRecordCassettes records go-vcr v4 cassettes from REAL API backends.
// Requires network access. Skipped by default.
// Run with: RECORD_CASSETTES=1 go test ./pkg/cli/skills/ -run TestRecordCassettes -v
//
// This captures actual responses from skills.sh, add-skill.vercel.sh, and GitHub APIs.
// Cassettes are committed to git for deterministic replay in CI (no network needed).
func TestRecordCassettes(t *testing.T) {
	if os.Getenv("RECORD_CASSETTES") == "" {
		t.Skip("skipped: set RECORD_CASSETTES=1 to record from real APIs (requires network)")
	}

	tests := []struct {
		name     string
		cassette string
		call     func(c *Client) error
	}{
		{
			name:     "search",
			cassette: "tests/testdata/cassettes/skills/search",
			call: func(c *Client) error {
				results, err := c.Search("web design", 10)
				if err != nil {
					return err
				}
				if len(results) == 0 {
					return fmt.Errorf("expected search results, got none")
				}
				t.Logf("search: got %d results, first: %s (%d installs)", len(results), results[0].Name, results[0].Installs)
				return nil
			},
		},
		{
			name:     "search_empty",
			cassette: "tests/testdata/cassettes/skills/search_empty",
			call: func(c *Client) error {
				results, err := c.Search("xyznonexistent99999", 10)
				if err != nil {
					return err
				}
				t.Logf("search_empty: got %d results", len(results))
				return nil
			},
		},
		{
			name:     "audit",
			cassette: "tests/testdata/cassettes/skills/audit",
			call: func(c *Client) error {
				results, err := c.FetchAudit("anthropics/skills", []string{"skill-creator"})
				if err != nil {
					return err
				}
				if len(results) == 0 {
					return fmt.Errorf("expected audit results, got none")
				}
				t.Logf("audit: got %d results", len(results))
				return nil
			},
		},
		{
			name:     "github_trees",
			cassette: "tests/testdata/cassettes/skills/github_trees",
			call: func(c *Client) error {
				tree, _, _, err := c.FetchRepoTree("anthropics", "skills", "main")
				if err != nil {
					return err
				}
				if len(tree) == 0 {
					return fmt.Errorf("expected tree entries, got none")
				}
				t.Logf("github_trees: got %d entries", len(tree))
				return nil
			},
		},
		{
			name:     "github_raw",
			cassette: "tests/testdata/cassettes/skills/github_raw",
			call: func(c *Client) error {
				data, err := c.FetchSkillContent("anthropics", "skills", "main", "skills/skill-creator/SKILL.md")
				if err != nil {
					return err
				}
				if len(data) == 0 {
					return fmt.Errorf("expected content, got empty")
				}
				t.Logf("github_raw: got %d bytes", len(data))
				return nil
			},
		},
		// Issue #172: repos with non-main default branches
		{
			name:     "github_trees_nonmain_404",
			cassette: "tests/testdata/cassettes/skills/github_trees_nonmain_404",
			call: func(c *Client) error {
				// different-ai/openwork uses "dev" as default branch; "main" should 404
				_, _, _, err := c.FetchRepoTree("different-ai", "openwork", "main")
				if err == nil {
					return fmt.Errorf("expected 404 for main branch, but got success")
				}
				t.Logf("github_trees_nonmain_404: got expected error: %v", err)
				return nil
			},
		},
		{
			name:     "github_trees_nonmain",
			cassette: "tests/testdata/cassettes/skills/github_trees_nonmain",
			call: func(c *Client) error {
				// HEAD resolves to the repo's default branch (dev)
				tree, _, _, err := c.FetchRepoTree("different-ai", "openwork", "HEAD")
				if err != nil {
					return err
				}
				skillCount := 0
				for _, e := range tree {
					if e.Type == "blob" && strings.HasSuffix(e.Path, "/SKILL.md") {
						skillCount++
					}
				}
				if skillCount == 0 {
					return fmt.Errorf("expected SKILL.md files, found none")
				}
				t.Logf("github_trees_nonmain: got %d entries, %d SKILL.md files", len(tree), skillCount)
				return nil
			},
		},
		{
			name:     "github_raw_nonmain",
			cassette: "tests/testdata/cassettes/skills/github_raw_nonmain",
			call: func(c *Client) error {
				// Fetch a skill from the dev branch via HEAD
				data, err := c.FetchSkillContent("different-ai", "openwork", "HEAD", ".opencode/skills/opencode-primitives/SKILL.md")
				if err != nil {
					return err
				}
				if len(data) == 0 {
					return fmt.Errorf("expected content, got empty")
				}
				t.Logf("github_raw_nonmain: got %d bytes", len(data))
				return nil
			},
		},
		{
			name:     "discover_nonmain",
			cassette: "tests/testdata/cassettes/skills/discover_nonmain",
			call: func(c *Client) error {
				// Full discovery flow: ref="" triggers HEAD/main/master fallback
				source := &SkillSource{
					Owner:       "different-ai",
					Repo:        "openwork",
					SkillFilter: "opencode-primitives",
					Ref:         "",
					Type:        SourceTypeGitHub,
				}
				skills, err := DiscoverSkills(c, source)
				if err != nil {
					return err
				}
				if len(skills) == 0 {
					return fmt.Errorf("expected skills, got none")
				}
				t.Logf("discover_nonmain: found %d skills, ref resolved to %q", len(skills), source.Ref)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := recorder.New(
				"../../../"+tt.cassette,
				recorder.WithMode(recorder.ModeRecordOnly),
				recorder.WithSkipRequestLatency(true),
				// Strip auth headers before saving cassettes to git
				recorder.WithHook(func(i *cassette.Interaction) error {
					delete(i.Request.Headers, "Authorization")
					delete(i.Request.Headers, "Cookie")
					return nil
				}, recorder.AfterCaptureHook),
			)
			if err != nil {
				t.Fatalf("recorder.New() error: %v", err)
			}

			// Use real API URLs — recorder intercepts via http.DefaultTransport
			client := &Client{
				SearchURL:  defaultSearchURL,
				AuditURL:   defaultAuditURL,
				GitHubURL:  defaultGitHubURL,
				RawGHURL:   defaultRawGHURL,
				HTTPClient: rec.GetDefaultClient(),
			}

			if err := tt.call(client); err != nil {
				t.Fatalf("call() error: %v", err)
			}

			if err := rec.Stop(); err != nil {
				t.Fatalf("recorder.Stop() error: %v", err)
			}
		})
	}
}
