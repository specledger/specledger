package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestMetadataClient(server *httptest.Server) *MetadataClient {
	return &MetadataClient{
		baseURL: server.URL,
		anonKey: "test-anon-key",
		client:  server.Client(),
	}
}

func TestMetadataClientDelete(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"success", http.StatusNoContent, false},
		{"success 200", http.StatusOK, false},
		{"not found", http.StatusNotFound, true},
		{"unauthorized", http.StatusUnauthorized, true},
		{"server error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("missing or wrong Authorization header")
				}
				if r.Header.Get("apikey") != "test-anon-key" {
					t.Errorf("missing or wrong apikey header")
				}
				w.WriteHeader(tt.statusCode)
				if tt.statusCode >= 400 {
					_, _ = w.Write([]byte(`{"message":"error"}`))
				}
			}))
			defer server.Close()

			client := newTestMetadataClient(server)
			err := client.Delete("test-token", "test-session-id")

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetadataClientDeleteVerifiesURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/rest/v1/sessions"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}
		// Verify query parameter
		if r.URL.Query().Get("id") != "eq.abc-123" {
			t.Errorf("expected id=eq.abc-123, got %s", r.URL.Query().Get("id"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestMetadataClient(server)
	_ = client.Delete("test-token", "abc-123")
}

func TestMetadataClientQueryWithTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify tag filter is applied
		tags := r.URL.Query().Get("tags")
		if tags != "cs.{bugfix}" {
			t.Errorf("expected tags=cs.{bugfix}, got %s", tags)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := newTestMetadataClient(server)
	sessions, err := client.Query("test-token", &QueryOptions{
		ProjectID: "proj-1",
		Tag:       "bugfix",
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}
