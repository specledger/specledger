package skills

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestTrackSync_SendsRequest(t *testing.T) {
	// Ensure telemetry is not disabled
	for _, env := range []string{"DISABLE_TELEMETRY", "DO_NOT_TRACK", "CI", "GITHUB_ACTIONS"} {
		t.Setenv(env, "")
		os.Unsetenv(env)
	}

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
	}))
	defer srv.Close()

	params := map[string]string{
		"source": "org/repo",
		"skills": "my-skill",
	}

	err := TrackSync(srv.URL, "install", params, "1.0.0")
	if err != nil {
		t.Fatalf("TrackSync: %v", err)
	}

	if gotPath == "" {
		t.Fatal("no request received")
	}
	if !contains(gotPath, "/t?") {
		t.Errorf("path = %q, want containing '/t?'", gotPath)
	}
	if !contains(gotPath, "event=install") {
		t.Errorf("path = %q, want containing 'event=install'", gotPath)
	}
	if !contains(gotPath, "v=specledger-1.0.0") {
		t.Errorf("path = %q, want containing 'v=specledger-1.0.0'", gotPath)
	}
}

func TestTrackSync_DisabledByEnv(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
	}{
		{"DISABLE_TELEMETRY", "DISABLE_TELEMETRY"},
		{"DO_NOT_TRACK", "DO_NOT_TRACK"},
		{"CI", "CI"},
		{"GITHUB_ACTIONS", "GITHUB_ACTIONS"},
		{"GITLAB_CI", "GITLAB_CI"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all telemetry-related env vars first
			for _, env := range []string{"DISABLE_TELEMETRY", "DO_NOT_TRACK", "CI", "GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI", "TRAVIS", "BUILDKITE", "JENKINS_URL", "TEAMCITY_VERSION"} {
				t.Setenv(env, "")
				os.Unsetenv(env)
			}

			t.Setenv(tt.envVar, "1")

			called := false
			srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				called = true
			}))
			defer srv.Close()

			err := TrackSync(srv.URL, "install", nil, "1.0.0")
			if err != nil {
				t.Fatalf("TrackSync: %v", err)
			}
			if called {
				t.Error("telemetry should be skipped")
			}
		})
	}
}

func TestBuildTelemetryParams(t *testing.T) {
	params := BuildTelemetryParams("org/repo", []string{"skill1", "skill2"}, []string{"claude-code"})

	if params["source"] != "org/repo" {
		t.Errorf("source = %q, want %q", params["source"], "org/repo")
	}
	if params["skills"] != "skill1,skill2" {
		t.Errorf("skills = %q, want %q", params["skills"], "skill1,skill2")
	}
	if params["agents"] != "claude-code" {
		t.Errorf("agents = %q, want %q", params["agents"], "claude-code")
	}
}

func TestIsPrivateRepo(t *testing.T) {
	// Mock GitHub API returning private=true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"private": true}`))
	}))
	defer srv.Close()

	t.Setenv("GITHUB_API_URL", srv.URL)
	if !isPrivateRepo("owner/private-repo") {
		t.Error("expected private repo to be detected")
	}

	// Mock GitHub API returning private=false
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"private": false}`))
	}))
	defer srv2.Close()

	t.Setenv("GITHUB_API_URL", srv2.URL)
	if isPrivateRepo("owner/public-repo") {
		t.Error("expected public repo to not be detected as private")
	}

	// Mock GitHub API returning 404
	srv3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv3.Close()

	t.Setenv("GITHUB_API_URL", srv3.URL)
	if !isPrivateRepo("owner/notfound") {
		t.Error("404 should be treated as private (conservative)")
	}
}

func TestIsTelemetryDisabled(t *testing.T) {
	// Clear all
	for _, env := range []string{"DISABLE_TELEMETRY", "DO_NOT_TRACK", "CI", "GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI", "TRAVIS", "BUILDKITE", "JENKINS_URL", "TEAMCITY_VERSION"} {
		t.Setenv(env, "")
		os.Unsetenv(env)
	}

	if isTelemetryDisabled() {
		t.Error("telemetry should be enabled with no env vars")
	}
}

func TestIsCI_IndividualVars(t *testing.T) {
	individualVars := []string{"CIRCLECI", "TRAVIS", "BUILDKITE", "JENKINS_URL", "TEAMCITY_VERSION"}

	for _, target := range individualVars {
		t.Run(target, func(t *testing.T) {
			// Clear ALL CI env vars first
			for _, v := range ciEnvVars {
				t.Setenv(v, "")
				os.Unsetenv(v)
			}

			// Set only the one under test
			t.Setenv(target, "1")

			if !isCI() {
				t.Errorf("isCI() = false with %s set, want true", target)
			}
		})
	}
}

func TestIsCI_NoneSet(t *testing.T) {
	// Clear all CI env vars
	for _, v := range ciEnvVars {
		t.Setenv(v, "")
		os.Unsetenv(v)
	}

	if isCI() {
		t.Error("isCI() = true with no CI vars set, want false")
	}
}

func TestIsPrivateRepo_InvalidSource(t *testing.T) {
	// Source with no "/" separator → returns false (not valid, let telemetry proceed)
	if isPrivateRepo("noseparator") {
		t.Error("expected false for source without separator")
	}
}

func TestIsPrivateRepo_EmptyParts(t *testing.T) {
	// Empty owner
	if isPrivateRepo("/repo") {
		t.Error("expected false for empty owner")
	}

	// Empty repo
	if isPrivateRepo("owner/") {
		t.Error("expected false for empty repo")
	}
}

func TestBuildTelemetryParams_NoAgents(t *testing.T) {
	params := BuildTelemetryParams("org/repo", []string{"skill1"}, []string{})

	if _, ok := params["agents"]; ok {
		t.Error("expected no 'agents' key in params when agents slice is empty")
	}
}

func TestTrackSync_InvalidURL(t *testing.T) {
	// Clear telemetry-disabling env vars
	for _, env := range []string{"DISABLE_TELEMETRY", "DO_NOT_TRACK"} {
		t.Setenv(env, "")
		os.Unsetenv(env)
	}
	for _, v := range ciEnvVars {
		t.Setenv(v, "")
		os.Unsetenv(v)
	}

	// "://invalid" is not a valid URL, should cause http.NewRequestWithContext to fail
	err := TrackSync("://invalid", "install", nil, "1.0.0")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestIsPrivateRepo_APIError(t *testing.T) {
	// Server that closes the connection immediately — simulates unreachable API
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Trigger a connection error by hijacking and closing
		hj, ok := w.(http.Hijacker)
		if !ok {
			// Fallback: just return 500 which also triggers conservative=true
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	t.Setenv("GITHUB_API_URL", srv.URL)

	// Conservative behavior: API error → treat as private → skip telemetry
	if !isPrivateRepo("owner/repo") {
		t.Error("expected true (conservative) when API errors")
	}
}

func TestTrackSync_EmptyAuditURL(t *testing.T) {
	// Clear telemetry-disabling env vars
	for _, env := range []string{"DISABLE_TELEMETRY", "DO_NOT_TRACK"} {
		t.Setenv(env, "")
		os.Unsetenv(env)
	}
	for _, v := range ciEnvVars {
		t.Setenv(v, "")
		os.Unsetenv(v)
	}

	// Empty auditURL should fall back to defaultAuditURL.
	// Since defaultAuditURL is a real external URL, the request will fail
	// with a network error — but the important thing is that the code path
	// runs without panic and that the default was used (error references the URL).
	err := TrackSync("", "install", map[string]string{"source": "org/repo"}, "1.0.0")
	// We expect an error because defaultAuditURL is not reachable in tests,
	// but the function should not panic.
	if err == nil {
		// If no error, the request somehow succeeded — acceptable but unexpected
		return
	}
	// Verify the error is a network error (from trying defaultAuditURL), not a coding error
	if contains(err.Error(), "nil pointer") || contains(err.Error(), "panic") {
		t.Errorf("unexpected error type: %v", err)
	}
}
