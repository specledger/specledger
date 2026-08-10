package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestStorageClient(server *httptest.Server) *StorageClient {
	return &StorageClient{
		baseURL: server.URL,
		anonKey: "test-anon-key",
		client:  server.Client(),
	}
}

func TestStorageClientDelete(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"success", http.StatusOK, false},
		{"success no content", http.StatusNoContent, false},
		{"not found", http.StatusNotFound, true},
		{"unauthorized", http.StatusUnauthorized, true},
		{"server error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			client := newTestStorageClient(server)
			err := client.Delete("test-token", "proj/branch/file.json.gz")

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorageClientDeleteVerifiesURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/storage/v1/object/sessions/proj-1/main/abc123.json.gz"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestStorageClient(server)
	_ = client.Delete("test-token", "proj-1/main/abc123.json.gz")
}
