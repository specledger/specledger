package comment

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
