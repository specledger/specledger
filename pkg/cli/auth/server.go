package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// CallbackResult contains the authentication result from the browser
type CallbackResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserEmail    string `json:"user_email"`
	UserID       string `json:"user_id"`
	Error        string `json:"error"`
}

const (
	// DefaultCallbackPort is the default port for the CLI callback server
	DefaultCallbackPort = 2026
)

// CallbackServer handles the OAuth callback from the browser
type CallbackServer struct {
	port        int
	server      *http.Server
	listener    net.Listener
	result      chan CallbackResult
	frontendURL string
}

// NewCallbackServer creates a new callback server on port 2026
func NewCallbackServer(frontendURL string) (*CallbackServer, error) {
	return NewCallbackServerWithPort(DefaultCallbackPort, frontendURL)
}

// NewCallbackServerWithPort creates a new callback server on the specified port
func NewCallbackServerWithPort(port int, frontendURL string) (*CallbackServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	cs := &CallbackServer{
		port:        port,
		listener:    listener,
		result:      make(chan CallbackResult, 1),
		frontendURL: frontendURL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", cs.handleCallback)

	cs.server = &http.Server{
		Handler: mux,
	}

	return cs, nil
}

// Port returns the port the server is listening on
func (cs *CallbackServer) Port() int {
	return cs.port
}

// CallbackURL returns the URL that should be used for OAuth redirect
func (cs *CallbackServer) CallbackURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/callback", cs.port)
}

// Start starts the callback server in the background
func (cs *CallbackServer) Start() {
	go func() {
		if err := cs.server.Serve(cs.listener); err != nil && err != http.ErrServerClosed {
			cs.result <- CallbackResult{Error: err.Error()}
		}
	}()
}

// WaitForCallback waits for the authentication callback with a timeout
func (cs *CallbackServer) WaitForCallback(timeout time.Duration) (*CallbackResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case result := <-cs.result:
		if result.Error != "" {
			return nil, fmt.Errorf("authentication error: %s", result.Error)
		}
		return &result, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("authentication timed out after %s", timeout)
	}
}

// Shutdown gracefully shuts down the server
func (cs *CallbackServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return cs.server.Shutdown(ctx)
}

// handleCallback handles the OAuth callback request
func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Support both GET (query params) and POST (JSON body)
	var result CallbackResult

	if r.Method == http.MethodPost {
		// Parse JSON body
		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			result.Error = "invalid request body"
		}
	} else {
		// Parse query parameters
		q := r.URL.Query()
		result.AccessToken = q.Get("access_token")
		result.RefreshToken = q.Get("refresh_token")
		result.UserEmail = q.Get("email")
		result.UserID = q.Get("user_id")
		result.Error = q.Get("error")

		// Parse expires_in if present
		if expiresIn := q.Get("expires_in"); expiresIn != "" {
			fmt.Sscanf(expiresIn, "%d", &result.ExpiresIn)
		}
	}

	// Set CORS headers for browser requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Send result to channel
	cs.result <- result

	// Redirect browser back to frontend
	redirectURL, _ := url.Parse(cs.frontendURL)
	redirectURL.Path = "/cli/auth"

	q := redirectURL.Query()
	if result.Error != "" {
		q.Set("status", "error")
		q.Set("message", result.Error)
	} else {
		q.Set("status", "success")
	}
	redirectURL.RawQuery = q.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}
