package comment

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/specledger/specledger/pkg/cli/auth"
)

// mockAuthProvider is a test double for AuthProvider.
type mockAuthProvider struct {
	creds        *auth.Credentials
	loadErr      error
	refreshToken string
	refreshErr   error
	refreshCalls int
}

func (m *mockAuthProvider) LoadCredentials() (*auth.Credentials, error) {
	return m.creds, m.loadErr
}

func (m *mockAuthProvider) ForceRefreshAccessToken() (string, error) {
	m.refreshCalls++
	if m.refreshErr != nil {
		return "", m.refreshErr
	}
	return m.refreshToken, nil
}

// newTestClient creates a Client pointing at the given test server.
func newTestClient(serverURL, token string, ap AuthProvider) *Client {
	return &Client{
		BaseURL:      serverURL,
		AnonKey:      "test-key",
		accessToken:  token,
		HTTPClient:   &http.Client{Timeout: 5 * time.Second},
		AuthProvider: ap,
	}
}

func TestDoWithRetry_Success_NoRetry(t *testing.T) {
	// Server always returns 200.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{}
	client := newTestClient(srv.URL, "good-token", mock)

	resp, err := client.DoWithRetry(func(token string) (*http.Response, error) {
		return client.Get(token, "/test")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if mock.refreshCalls != 0 {
		t.Fatalf("expected 0 refresh calls, got %d", mock.refreshCalls)
	}
}

func TestDoWithRetry_401_ReloadsCredentialsFromDisk(t *testing.T) {
	// Server returns 401 for "stale-token", 200 for "fresh-token".
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer fresh-token" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"ok":true}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message":"invalid token"}`)
	}))
	defer srv.Close()

	// Simulate another process having refreshed credentials on disk.
	mock := &mockAuthProvider{
		creds: &auth.Credentials{
			AccessToken:  "fresh-token",
			RefreshToken: "new-refresh",
			ExpiresIn:    3600,
			CreatedAt:    time.Now().Unix(),
		},
	}

	client := newTestClient(srv.URL, "stale-token", mock)

	resp, err := client.DoWithRetry(func(token string) (*http.Response, error) {
		return client.Get(token, "/test")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	// Must NOT have called ForceRefreshAccessToken — the disk reload was enough.
	if mock.refreshCalls != 0 {
		t.Fatalf("expected 0 refresh calls (should use disk creds), got %d", mock.refreshCalls)
	}
	if client.accessToken != "fresh-token" {
		t.Fatalf("expected client token updated to fresh-token, got %s", client.accessToken)
	}
	if calls != 2 {
		t.Fatalf("expected 2 HTTP calls (1 stale + 1 retry), got %d", calls)
	}
}

func TestDoWithRetry_401_SameCredentials_ForceRefreshes(t *testing.T) {
	// Server returns 401 for "stale-token", 200 for "refreshed-token".
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer refreshed-token" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"ok":true}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message":"invalid token"}`)
	}))
	defer srv.Close()

	// Disk credentials are the SAME as what client already has (no other process refreshed).
	mock := &mockAuthProvider{
		creds: &auth.Credentials{
			AccessToken:  "stale-token", // same as client's token
			RefreshToken: "some-refresh",
			ExpiresIn:    3600,
			CreatedAt:    time.Now().Unix(),
		},
		refreshToken: "refreshed-token",
	}

	client := newTestClient(srv.URL, "stale-token", mock)

	resp, err := client.DoWithRetry(func(token string) (*http.Response, error) {
		return client.Get(token, "/test")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if mock.refreshCalls != 1 {
		t.Fatalf("expected 1 refresh call, got %d", mock.refreshCalls)
	}
	if client.accessToken != "refreshed-token" {
		t.Fatalf("expected client token updated to refreshed-token, got %s", client.accessToken)
	}
}

func TestDoWithRetry_401_ExpiredDiskCreds_ForceRefreshes(t *testing.T) {
	// Disk credentials exist but are expired — should force refresh, not use them.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer refreshed-token" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"ok":true}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{
		creds: &auth.Credentials{
			AccessToken:  "different-but-expired",
			RefreshToken: "some-refresh",
			ExpiresIn:    3600,
			CreatedAt:    time.Now().Add(-2 * time.Hour).Unix(), // expired 1h ago
		},
		refreshToken: "refreshed-token",
	}

	client := newTestClient(srv.URL, "stale-token", mock)

	resp, err := client.DoWithRetry(func(token string) (*http.Response, error) {
		return client.Get(token, "/test")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	// Even though disk creds are different, they're expired — must force refresh.
	if mock.refreshCalls != 1 {
		t.Fatalf("expected 1 refresh call (disk creds expired), got %d", mock.refreshCalls)
	}
}

func TestDoWithRetry_401_ForceRefreshFails_ReturnsError(t *testing.T) {
	// Server always returns 401.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	// Disk creds are the same, and force refresh fails (simulating "already used" error).
	mock := &mockAuthProvider{
		creds: &auth.Credentials{
			AccessToken:  "stale-token",
			RefreshToken: "used-refresh",
			ExpiresIn:    3600,
			CreatedAt:    time.Now().Unix(),
		},
		refreshErr: fmt.Errorf("failed to refresh token: refresh request failed (HTTP 400): refresh_token_already_used"),
	}

	client := newTestClient(srv.URL, "stale-token", mock)

	_, err := client.DoWithRetry(func(token string) (*http.Response, error) {
		return client.Get(token, "/test")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if mock.refreshCalls != 1 {
		t.Fatalf("expected 1 refresh call, got %d", mock.refreshCalls)
	}
}

func TestDoWithRetry_401_LoadCredentialsFails_ForceRefreshes(t *testing.T) {
	// LoadCredentials returns an error — should fall through to force refresh.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer refreshed-token" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"ok":true}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{
		loadErr:      fmt.Errorf("corrupted credentials file"),
		refreshToken: "refreshed-token",
	}

	client := newTestClient(srv.URL, "stale-token", mock)

	resp, err := client.DoWithRetry(func(token string) (*http.Response, error) {
		return client.Get(token, "/test")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if mock.refreshCalls != 1 {
		t.Fatalf("expected 1 refresh call, got %d", mock.refreshCalls)
	}
}

func TestDoWithRetry_401_NilDiskCreds_ForceRefreshes(t *testing.T) {
	// LoadCredentials returns nil (no credentials file) — should force refresh.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer refreshed-token" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"ok":true}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{
		creds:        nil,
		refreshToken: "refreshed-token",
	}

	client := newTestClient(srv.URL, "stale-token", mock)

	resp, err := client.DoWithRetry(func(token string) (*http.Response, error) {
		return client.Get(token, "/test")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if mock.refreshCalls != 1 {
		t.Fatalf("expected 1 refresh call, got %d", mock.refreshCalls)
	}
}

// --- Structured error tests ---

func TestReadJSON_ReturnsRawAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "row-level security violation"})
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	var dest []map[string]any
	err = ReadJSON(resp, &dest)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var rawErr *RawAPIError
	if !errors.As(err, &rawErr) {
		t.Fatalf("expected RawAPIError, got %T: %v", err, err)
	}
	if rawErr.StatusCode != 403 {
		t.Fatalf("expected status 403, got %d", rawErr.StatusCode)
	}
	if !strings.Contains(rawErr.Body, "row-level security") {
		t.Fatalf("expected body to contain 'row-level security', got: %s", rawErr.Body)
	}
}

func TestSummarizeAPIError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON with message field",
			input:    `{"code":"42501","message":"new row violates row-level security"}`,
			expected: "new row violates row-level security",
		},
		{
			name:     "plain text",
			input:    "something went wrong",
			expected: "something went wrong",
		},
		{
			name:     "long body truncated",
			input:    strings.Repeat("x", 250),
			expected: strings.Repeat("x", 197) + "...",
		},
		{
			name:     "JSON without message field",
			input:    `{"error":"oops"}`,
			expected: `{"error":"oops"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeAPIError(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestFormatAPIError_StatusGuidance(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		mustContain []string
	}{
		{
			name:       "401 suggests auth login",
			statusCode: 401,
			mustContain: []string{
				"TestOp failed (401)",
				"sl auth login",
				"--verbose",
			},
		},
		{
			name:       "403 suggests stale auth",
			statusCode: 403,
			mustContain: []string{
				"TestOp failed (403)",
				"stale auth",
				"sl auth login",
			},
		},
		{
			name:       "404 suggests verify ID",
			statusCode: 404,
			mustContain: []string{
				"not found",
				"Verify the ID",
			},
		},
		{
			name:       "400 suggests check input",
			statusCode: 400,
			mustContain: []string{
				"malformed",
				"Check the input",
			},
		},
		{
			name:       "500 suggests retry",
			statusCode: 500,
			mustContain: []string{
				"Server error",
				"Retry",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{Verbose: false}
			err := client.formatAPIError("TestOp", tt.statusCode, `{"message":"test error"}`)
			msg := err.Error()

			for _, s := range tt.mustContain {
				if !strings.Contains(msg, s) {
					t.Errorf("expected error to contain %q, got:\n%s", s, msg)
				}
			}
		})
	}
}

func TestFormatAPIError_VerboseShowsRawBody(t *testing.T) {
	rawBody := `{"code":"42501","message":"rls violation","details":"some detail"}`

	client := &Client{Verbose: true}
	err := client.formatAPIError("TestOp", 403, rawBody)
	msg := err.Error()

	if !strings.Contains(msg, "Raw response:") {
		t.Error("verbose mode should include 'Raw response:'")
	}
	if !strings.Contains(msg, rawBody) {
		t.Errorf("verbose mode should include raw body, got:\n%s", msg)
	}
	if strings.Contains(msg, "Run with --verbose") {
		t.Error("verbose mode should NOT suggest --verbose")
	}
}

func TestFormatAPIError_NonVerboseHidesRawBody(t *testing.T) {
	rawBody := `{"code":"42501","message":"rls violation","details":"secret"}`

	client := &Client{Verbose: false}
	err := client.formatAPIError("TestOp", 403, rawBody)
	msg := err.Error()

	if strings.Contains(msg, "secret") {
		t.Error("non-verbose mode should not include raw body details")
	}
	if !strings.Contains(msg, "Run with --verbose") {
		t.Error("non-verbose mode should suggest --verbose")
	}
}

func TestWrapErr_APIError(t *testing.T) {
	client := &Client{Verbose: false}
	rawErr := &RawAPIError{StatusCode: 403, Body: `{"message":"forbidden"}`}

	err := client.wrapErr("CreateReply", rawErr)
	msg := err.Error()

	if !strings.Contains(msg, "CreateReply failed (403)") {
		t.Errorf("expected structured format, got: %s", msg)
	}
	if !strings.Contains(msg, "forbidden") {
		t.Errorf("expected summary from JSON message, got: %s", msg)
	}
}

func TestWrapErr_NonAPIError(t *testing.T) {
	client := &Client{Verbose: false}
	err := client.wrapErr("FetchComments", fmt.Errorf("connection refused"))
	msg := err.Error()

	if msg != "FetchComments: connection refused" {
		t.Errorf("expected simple wrap, got: %s", msg)
	}
}

func TestGetProject_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"permission denied"}`)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{}
	client := newTestClient(srv.URL, "token", mock)

	_, err := client.GetProject("owner", "repo")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()

	if !strings.Contains(msg, "GetProject failed (403)") {
		t.Errorf("expected structured error, got: %s", msg)
	}
	if !strings.Contains(msg, "sl auth login") {
		t.Errorf("expected auth guidance, got: %s", msg)
	}
}

func TestGetProject_NotFound_Guidance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[]`)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{}
	client := newTestClient(srv.URL, "token", mock)

	_, err := client.GetProject("owner", "repo")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()

	if !strings.Contains(msg, "owner/repo") {
		t.Errorf("expected repo info in error, got: %s", msg)
	}
	if !strings.Contains(msg, "git remote -v") {
		t.Errorf("expected guidance, got: %s", msg)
	}
}

func TestFetchCommentByID_NotFound_Guidance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[]`)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{}
	client := newTestClient(srv.URL, "token", mock)

	_, err := client.FetchCommentByID("nonexistent-id")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()

	if !strings.Contains(msg, "nonexistent-id") {
		t.Errorf("expected ID in error, got: %s", msg)
	}
	if !strings.Contains(msg, "sl comment list") {
		t.Errorf("expected guidance to list comments, got: %s", msg)
	}
}

func TestGetSpec_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"permission denied"}`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	_, err := client.GetSpec("proj-id", "my-spec")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "GetSpec failed (403)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}

func TestGetSpec_NotFound_Guidance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[]`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	_, err := client.GetSpec("proj-id", "my-spec")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "my-spec") {
		t.Errorf("expected spec key in error, got: %s", err)
	}
}

func TestGetChange_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"denied"}`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	_, err := client.GetChange("spec-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "GetChange failed (403)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}

func TestGetChange_NotFound_Guidance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[]`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	_, err := client.GetChange("spec-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no change found") {
		t.Errorf("expected guidance, got: %s", err)
	}
}

func TestGetChange_ReturnsOpenChange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[{"id":"c1","spec_id":"s1","head_branch":"feat","base_branch":"main","state":"closed"},{"id":"c2","spec_id":"s1","head_branch":"feat","base_branch":"main","state":"open"}]`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	ch, err := client.GetChange("s1")
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != "c2" {
		t.Errorf("expected open change c2, got %s", ch.ID)
	}
}

func TestFetchComments_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"message":"internal error"}`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	_, err := client.FetchComments("change-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "FetchComments failed (500)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}

func TestFetchReplies_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"message":"internal error"}`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	_, err := client.FetchReplies("change-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "FetchReplies failed (500)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}

func TestFetchRepliesByParentID_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"denied"}`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	_, err := client.FetchRepliesByParentID("parent-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "FetchRepliesByParentID failed (403)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}

func TestFetchResolvedComments_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"denied"}`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	_, err := client.FetchResolvedComments("change-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "FetchResolvedComments failed (403)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}

func TestCreateReply_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `[{"id":"c1","change_id":"ch1","file_path":"main.go","content":"test","author_name":"a","author_email":"a@b.com","is_resolved":false,"created_at":"2024-01-01"}]`)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"rls violation"}`)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{
		creds: &auth.Credentials{UserID: "user-1", AccessToken: "token"},
	}
	client := newTestClient(srv.URL, "token", mock)
	_, err := client.CreateReply("c1", "hello")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "CreateReply failed (403)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}

func TestCreateReply_NoCreds_Guidance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[{"id":"c1","change_id":"ch1","file_path":"main.go","content":"test","author_name":"a","author_email":"a@b.com","is_resolved":false,"created_at":"2024-01-01"}]`)
	}))
	defer srv.Close()

	mock := &mockAuthProvider{loadErr: fmt.Errorf("no creds")}
	client := newTestClient(srv.URL, "token", mock)
	_, err := client.CreateReply("c1", "hello")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "sl auth login") {
		t.Errorf("expected auth guidance, got: %s", err)
	}
}

func TestResolveComment_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"denied"}`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	err := client.ResolveComment("comment-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "ResolveComment failed (403)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}

func TestResolveCommentWithReplies_StructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"denied"}`)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "token", &mockAuthProvider{})
	err := client.ResolveCommentWithReplies("comment-id", []string{"r1", "r2"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "ResolveCommentWithReplies failed (403)") {
		t.Errorf("expected structured error, got: %s", err)
	}
}
