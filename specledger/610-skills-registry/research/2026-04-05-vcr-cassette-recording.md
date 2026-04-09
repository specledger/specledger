# Research: VCR Cassette Recording Against Real Backends

**Date**: 2026-04-05
**Context**: Current `record_test.go` records cassettes from httptest handlers with hardcoded responses. We need cassettes with REAL responses from skills.sh, add-skill.vercel.sh, and GitHub APIs so tests are reliable without network.
**Time-box**: 30 minutes

## Question

Can we use `dnaeon/go-vcr` v4 to record cassettes from the actual backend APIs (skills.sh, GitHub) and replay them in unit tests without network access?

## Findings

### Finding 1: Current record_test.go records fake data

The existing `record_test.go` creates an `httptest.NewServer` with hardcoded JSON handlers, then wraps that with a VCR recorder. This means the cassettes contain **fabricated** data, not real API responses. The cassettes are valid YAML but the response bodies are hand-crafted test fixtures — the VCR layer is essentially recording from itself.

### Finding 2: go-vcr v4 supports direct real-backend recording

go-vcr v4's `ModeRecordOnly` can record against real backends without an httptest server. The key is:

```go
rec, err := recorder.New("cassette-path",
    recorder.WithMode(recorder.ModeRecordOnly),
    // No WithRealTransport needed — defaults to http.DefaultTransport
    recorder.WithSkipRequestLatency(true),
)
client := rec.GetDefaultClient()
// client now records ALL HTTP interactions to cassette file
```

When `WithRealTransport` is omitted (or set to `http.DefaultTransport`), the recorder uses real network transport. The current code passes `srv.Client().Transport` which routes to httptest — that's why it records fake data.

### Finding 3: Recording from real APIs — implementation pattern

To record against the actual backends:

```go
func TestRecordRealCassettes(t *testing.T) {
    if os.Getenv("RECORD_CASSETTES") == "" {
        t.Skip("set RECORD_CASSETTES=1 to record from real APIs")
    }

    tests := []struct {
        name     string
        cassette string
        call     func(c *Client) error
    }{
        {
            name:     "search_commit",
            cassette: "tests/testdata/cassettes/skills/search",
            call: func(c *Client) error {
                results, err := c.Search("commit", 10)
                if err != nil {
                    return err
                }
                if len(results) == 0 {
                    return fmt.Errorf("expected results, got none")
                }
                return nil
            },
        },
        // ... more scenarios
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            rec, err := recorder.New(
                "../../../"+tt.cassette,
                recorder.WithMode(recorder.ModeRecordOnly),
                recorder.WithSkipRequestLatency(true),
                // Sanitize auth headers before saving
                recorder.WithHook(func(i *cassette.Interaction) error {
                    delete(i.Request.Headers, "Authorization")
                    return nil
                }, recorder.AfterCaptureHook),
            )
            if err != nil {
                t.Fatalf("recorder: %v", err)
            }

            client := &Client{
                SearchURL:  defaultSearchURL,   // real URLs
                AuditURL:   defaultAuditURL,
                GitHubURL:  defaultGitHubURL,
                RawGHURL:   defaultRawGHURL,
                HTTPClient: rec.GetDefaultClient(),
            }

            if err := tt.call(client); err != nil {
                t.Fatalf("call: %v", err)
            }

            if err := rec.Stop(); err != nil {
                t.Fatalf("stop: %v", err)
            }
        })
    }
}
```

### Finding 4: Replay requires a path-based matcher

Cassettes recorded against real backends contain full URLs like `https://skills.sh/api/search?q=commit&limit=10`. For replay, the default matcher (method + URL) works directly since there's no httptest port randomization. This is actually **simpler** than the current approach which needs a `pathMethodMatcher` to strip the random httptest port.

### Finding 5: Security — sanitizing cassettes

go-vcr v4 provides hook kinds for sanitization:
- `AfterCaptureHook` — runs after each interaction is captured (best for stripping auth headers)
- `BeforeSaveHook` — runs before writing cassette to disk

Critical headers to strip:
- `Authorization` (GitHub token)
- Any cookies or session headers

### Finding 6: Quickstart scenarios mapped to API calls

| Quickstart Scenario | API Endpoint | Cassette File |
|---|---|---|
| Search "commit" | `GET skills.sh/api/search?q=commit&limit=10` | `search.yaml` |
| Search "testing" limit 3 | `GET skills.sh/api/search?q=testing&limit=3` | `search_testing.yaml` |
| Search no results | `GET skills.sh/api/search?q=xyznonexistent&limit=10` | `search_empty.yaml` |
| Install (fetch tree) | `GET api.github.com/repos/vercel-labs/agent-skills/git/trees/main?recursive=1` | `github_trees.yaml` |
| Install (fetch content) | `GET raw.githubusercontent.com/vercel-labs/agent-skills/main/skills/creating-pr/SKILL.md` | `github_raw.yaml` |
| Audit | `GET add-skill.vercel.sh/audit?source=vercel-labs/agent-skills&skills=creating-pr` | `audit.yaml` |

### Finding 7: Multiple interactions per cassette

go-vcr v4 supports multiple interactions in a single cassette. For the "install" flow (tree lookup + raw fetch + audit), we could either:
- **Option A**: One cassette per endpoint (current approach) — simpler replay matching
- **Option B**: One cassette per user scenario (e.g., `install_flow.yaml` with 3 interactions) — matches E2E flow

Option A is better for unit tests; Option B is better for integration tests. We can support both.

## Decisions

- **Decision 1**: Rewrite `record_test.go` to record from real backends instead of httptest handlers. The httptest approach defeats the purpose of VCR — we're recording fake data we already control. Real cassettes capture actual API response structure, headers, and edge cases.

- **Decision 2**: Keep cassettes at `tests/testdata/cassettes/skills/` (one per endpoint). Add scenario-specific cassettes (e.g., `search_empty.yaml`) for edge cases like no-results queries.

- **Decision 3**: Use `AfterCaptureHook` to strip `Authorization` headers before saving cassettes. Cassettes will be committed to git, so they must not contain secrets.

- **Decision 4**: Gate recording behind `RECORD_CASSETTES=1` (already done). Require network access only during recording; all other test runs replay from cassettes with `ModeReplayOnly`.

## Recommendations

1. **Rewrite `record_test.go`** to drop the httptest server and record against real backends. Use real default URLs, inject `rec.GetDefaultClient()` as `HTTPClient`. Add sanitization hook for auth headers.

2. **Expand cassette coverage** to match quickstart scenarios: add `search_testing.yaml` (limit=3), `search_empty.yaml` (no results), and a multi-skill audit cassette.

3. **Update replay tests** (`client_test.go`) to use `ModeReplayOnly` with the real-backend cassettes. The default URL matcher will work since cassettes contain real URLs — no custom `pathMethodMatcher` needed.

4. **Run the recording once** with `RECORD_CASSETTES=1 go test ./pkg/cli/skills/ -run TestRecordCassettes -v` to capture real API responses, then commit the cassettes.

5. **Add CI guard**: ensure `RECORD_CASSETTES` is never set in CI, so tests always replay from committed cassettes.

## References

- go-vcr v4 source: `gopkg.in/dnaeon/go-vcr.v4` (v4.0.6 in go.mod)
- Recorder modes: `pkg/recorder/recorder.go` — `ModeRecordOnly`, `ModeReplayOnly`, `ModeRecordOnce`
- Hook API: `recorder.WithHook(handler, kind)` with `AfterCaptureHook`, `BeforeSaveHook`
- Current cassettes: `tests/testdata/cassettes/skills/*.yaml`
- Quickstart scenarios: `specledger/610-skills-registry/quickstart.md`
