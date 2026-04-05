package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const telemetryTimeout = 3 * time.Second

// ciEnvVars are environment variables that indicate a CI environment.
var ciEnvVars = []string{
	"CI",
	"GITHUB_ACTIONS",
	"GITLAB_CI",
	"CIRCLECI",
	"TRAVIS",
	"BUILDKITE",
	"JENKINS_URL",
	"TEAMCITY_VERSION",
}

// Track sends a fire-and-forget telemetry ping to skills.sh.
// It runs in a goroutine and never blocks the caller.
// Telemetry is skipped if disabled via env vars, running in CI, or source repo is private.
func Track(auditURL, event string, params map[string]string, version string) {
	if isTelemetryDisabled() {
		return
	}

	// Skip for private repos (check source param)
	if source, ok := params["source"]; ok && isPrivateRepo(source) {
		return
	}

	go func() {
		_ = trackSync(auditURL, event, params, version)
	}()
}

// TrackSync sends a telemetry ping synchronously (for testing).
func TrackSync(auditURL, event string, params map[string]string, version string) error {
	if isTelemetryDisabled() {
		return nil
	}
	return trackSync(auditURL, event, params, version)
}

func trackSync(auditURL, event string, params map[string]string, version string) error {
	if auditURL == "" {
		auditURL = defaultAuditURL
	}

	q := url.Values{}
	q.Set("v", fmt.Sprintf("specledger-%s", version))
	q.Set("event", event)
	for k, v := range params {
		q.Set(k, v)
	}

	reqURL := auditURL + "/t?" + q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), telemetryTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err // fire-and-forget, swallowed in goroutine
	}
	resp.Body.Close()
	return nil
}

func isTelemetryDisabled() bool {
	if os.Getenv("DISABLE_TELEMETRY") != "" {
		return true
	}
	if os.Getenv("DO_NOT_TRACK") != "" {
		return true
	}
	return isCI()
}

func isCI() bool {
	for _, env := range ciEnvVars {
		if os.Getenv(env) != "" {
			return true
		}
	}
	return false
}

// isPrivateRepo checks if a GitHub repo is private by hitting the API.
// Returns true if the repo is private or the check fails (conservative: skip telemetry on error).
func isPrivateRepo(source string) bool {
	parts := strings.SplitN(source, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false // not a valid source, let telemetry proceed
	}

	ghURL := envOrDefault("GITHUB_API_URL", defaultGitHubURL)
	reqURL := fmt.Sprintf("%s/repos/%s/%s", ghURL, parts[0], parts[1])

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return true // skip telemetry on error (conservative)
	}
	req.Header.Set("Accept", "application/json")
	if token := githubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return true // skip telemetry on error (conservative)
	}
	defer resp.Body.Close()

	// If we get 404 without auth or the repo metadata says private, skip
	if resp.StatusCode == http.StatusNotFound {
		return true
	}

	// Parse response to check "private" field
	var repoInfo struct {
		Private bool `json:"private"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return false
	}
	return repoInfo.Private
}

// BuildTelemetryParams creates the standard telemetry params for an event.
func BuildTelemetryParams(source string, skillNames []string, agents []string) map[string]string {
	params := map[string]string{
		"source": source,
		"skills": strings.Join(skillNames, ","),
	}
	if len(agents) > 0 {
		params["agents"] = strings.Join(agents, ",")
	}
	return params
}
