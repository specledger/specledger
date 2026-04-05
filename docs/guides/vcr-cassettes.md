# VCR Cassette Testing Guide

> **Purpose:** Reproducible HTTP tests with REAL captured API responses — no mocks, no network at test time.
>
> **Library:** [dnaeon/go-vcr v4](https://github.com/dnaeon/go-vcr) (`gopkg.in/dnaeon/go-vcr.v4`)
>
> **Reference implementation:** `pkg/cli/skills/` (skills.sh, GitHub APIs)

## Why VCR Cassettes?

Traditional approaches have fundamental problems:

| Approach | Problem |
|----------|---------|
| `httptest.NewServer` with hardcoded handlers | You're testing against **fake data you invented** — API format changes, edge cases, and real-world response structures are invisible |
| Live API calls in tests | Flaky in CI, rate-limited, slow, requires credentials |
| Mock interfaces | Requires maintaining mocks that drift from real APIs |

**VCR cassettes solve all three**: record once from real APIs, replay forever. Cassettes contain actual HTTP interactions (request + response), committed to git, replayed deterministically in <1ms per test.

## Architecture

```
Record (once, with network)          Replay (every test run, no network)
─────────────────────────────        ──────────────────────────────────
Client → Recorder → Real API         Client → Recorder → Cassette File
              ↓                                    ↓
         cassette.yaml                        cassette.yaml
         (saved to disk)                      (read from disk)
```

### File Layout

```
tests/testdata/cassettes/
├── skills/                     # One directory per feature
│   ├── search.yaml             # Real skills.sh search response
│   ├── search_empty.yaml       # Real empty search response
│   ├── audit.yaml              # Real audit API response (ATH/Socket/Snyk)
│   ├── github_trees.yaml       # Real GitHub Trees API response
│   └── github_raw.yaml         # Real raw.githubusercontent.com content
└── <other-feature>/
    └── *.yaml
```

### Test File Layout

```
pkg/cli/<feature>/
├── record_test.go              # Records cassettes from real APIs (gated)
├── vcr_test.go                 # Replays cassettes deterministically
├── client_test.go              # httptest-based unit tests (fine for simple validation)
└── ...
```

## Recording Cassettes

### Step 1: Write `record_test.go`

```go
package skills

import (
    "fmt"
    "os"
    "testing"

    "gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
    "gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func TestRecordCassettes(t *testing.T) {
    if os.Getenv("RECORD_CASSETTES") == "" {
        t.Skip("set RECORD_CASSETTES=1 to record from real APIs (requires network)")
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
                "../../../"+tt.cassette,       // path relative to test file
                recorder.WithMode(recorder.ModeRecordOnly),
                recorder.WithSkipRequestLatency(true),
                // CRITICAL: Strip auth headers before saving to git
                recorder.WithHook(func(i *cassette.Interaction) error {
                    delete(i.Request.Headers, "Authorization")
                    delete(i.Request.Headers, "Cookie")
                    return nil
                }, recorder.AfterCaptureHook),
            )
            if err != nil {
                t.Fatalf("recorder: %v", err)
            }

            // Use REAL API URLs — no httptest server
            client := &Client{
                SearchURL:  defaultSearchURL,   // https://skills.sh
                AuditURL:   defaultAuditURL,    // https://add-skill.vercel.sh
                GitHubURL:  defaultGitHubURL,   // https://api.github.com
                RawGHURL:   defaultRawGHURL,    // https://raw.githubusercontent.com
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

### Key Rules for Recording

1. **Use real default URLs** — not `httptest.NewServer()`. The recorder wraps `http.DefaultTransport` to intercept real network calls.
2. **Always strip auth headers** via `AfterCaptureHook` before saving. Cassettes are committed to git.
3. **Gate behind `RECORD_CASSETTES=1`** — recording requires network, but every other test run replays from disk.
4. **Use real, well-known data** — test against actual public repos/skills that exist. Example: `anthropics/skills`, `vercel-labs/agent-skills`. Verify the data exists before recording.
5. **Add assertions during recording** — if the real API returns empty, fail. This catches recording issues immediately.

### Step 2: Record

```bash
# Record all cassettes (requires network)
RECORD_CASSETTES=1 go test ./pkg/cli/skills/ -run TestRecordCassettes -v

# Output shows real data captured:
# search: got 10 results, first: web-design-guidelines (210639 installs)
# audit: got 1 results
# github_trees: got 457 entries
# github_raw: got 33168 bytes
```

### Step 3: Commit cassettes

```bash
git add tests/testdata/cassettes/skills/*.yaml
git commit -m "test: record VCR cassettes from real APIs"
```

## Replaying Cassettes

### Write `vcr_test.go`

```go
package skills

import (
    "testing"

    "gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

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
        SearchURL:  defaultSearchURL,   // Must match recording URLs
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
        t.Fatal("expected results")
    }
    // Assert on real data structure, not hardcoded values
    for _, r := range results {
        if r.Name == "" {
            t.Error("result has empty Name")
        }
    }
}
```

### Key Rules for Replay

1. **Use the same URLs as recording** — the default URL matcher compares full method + URL. Since cassettes contain real URLs (`https://skills.sh/api/search?...`), the client must use the same base URLs.
2. **No custom matcher needed** — unlike httptest cassettes (which have random ports), real-backend cassettes have stable URLs. The default matcher works.
3. **Assert on structure, not volatile values** — install counts change daily. Assert that fields are non-empty, that the right types are returned, not that `Installs == 210639`.
4. **Replay is instantaneous** — cassettes are read from disk, no network. Tests run in <1ms per cassette.

## When to Re-Record

Re-record cassettes when:

- **API response format changes** — tests will fail on replay (field missing, type mismatch)
- **Test scenarios change** — new query, different parameters
- **Periodically** — API data evolves (new skills, changed audit scores). Not urgent unless tests break.

```bash
# Delete old cassettes and re-record
rm tests/testdata/cassettes/skills/*.yaml
RECORD_CASSETTES=1 go test ./pkg/cli/skills/ -run TestRecordCassettes -v
git add tests/testdata/cassettes/skills/
git commit -m "test: refresh VCR cassettes"
```

## Anti-Patterns

### DO NOT: Record from httptest handlers

```go
// BAD — recording fake data defeats the purpose of VCR
srv := httptest.NewServer(handler)
rec, _ := recorder.New("cassette",
    recorder.WithRealTransport(srv.Client().Transport),  // routes to fake server
)
```

This creates cassettes with fabricated responses. The entire value of VCR is capturing **real API behavior** — response formats, headers, edge cases, error bodies that you'd never think to hardcode.

### DO NOT: Hardcode expected values from cassettes

```go
// BAD — brittle, breaks when cassette is refreshed
if results[0].Installs != 210639 {
    t.Error("wrong install count")
}

// GOOD — assert on structure
if results[0].Installs <= 0 {
    t.Error("expected positive install count")
}
```

### DO NOT: Set `RECORD_CASSETTES=1` in CI

CI must always replay from committed cassettes. Recording requires network and produces non-deterministic output (timestamps, counts). Add a CI guard:

```yaml
# .github/workflows/test.yml
- run: go test ./...
  env:
    RECORD_CASSETTES: ""  # never set in CI
```

## VCR vs httptest: When to Use Each

| Use Case | Tool | Why |
|----------|------|-----|
| Testing against external APIs (skills.sh, GitHub) | **VCR cassettes** | Real responses, deterministic replay, no network in CI |
| E2E integration tests (full CLI binary) | **httptest.Server** | Need dynamic behavior (different responses per scenario), full binary invocation |
| Simple request/response validation | **httptest.Server** | Fine for unit tests where you control the contract |
| Testing error handling (timeouts, 500s) | **httptest.Server** | Easier to simulate failures |

Both approaches coexist. VCR cassettes are the **primary** tool for API client testing. httptest is for E2E binary invocation and error simulation.

## Reference: go-vcr v4 Recorder Modes

| Mode | Description | Network Required |
|------|-------------|-----------------|
| `ModeRecordOnly` | Record all interactions, overwrite cassette | Yes |
| `ModeReplayOnly` | Replay from cassette, fail if no match | No |
| `ModeRecordOnce` | Record if cassette missing, replay if exists | First run only |
| `ModePassthrough` | No recording or replay, pass through to network | Yes |

For this project, we use `ModeRecordOnly` (gated behind env var) and `ModeReplayOnly` (default in tests).

## Cassette File Format

Cassettes are YAML files with HTTP interactions:

```yaml
---
version: 2
interactions:
  - id: 0
    request:
      body: ""
      form: {}
      headers:
        Accept:
          - application/json
      method: GET
      url: https://skills.sh/api/search?q=web+design&limit=10
    response:
      body: '{"skills":[{"id":"vercel-labs/agent-skills/web-design-guidelines",...}]}'
      headers:
        Content-Type:
          - application/json
      code: 200
```

These files are committed to git and reviewed like any other test fixture.
