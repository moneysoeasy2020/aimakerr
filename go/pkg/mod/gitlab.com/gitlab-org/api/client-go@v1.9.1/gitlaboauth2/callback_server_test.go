package gitlaboauth2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestBrowserFunc(t *testing.T) {
	t.Parallel()

	t.Run("successful browser function call", func(t *testing.T) {
		t.Parallel()

		called := false
		var capturedURL string

		browserFunc := BrowserFunc(func(url string) error {
			called = true
			capturedURL = url
			return nil
		})

		testURL := "https://gitlab.com/oauth/authorize?client_id=test"
		err := browserFunc(testURL)

		require.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, testURL, capturedURL)
	})

	t.Run("browser function returns error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("browser failed to open")

		browserFunc := BrowserFunc(func(url string) error {
			return expectedError
		})

		err := browserFunc("https://example.com")

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestNewCallbackServer(t *testing.T) { //nolint:paralleltest
	t.Run("creates callback server with valid parameters", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
			Scopes:      []string{"read_user"},
		}

		addr := ":8080"
		browserFunc := func(url string) error { return nil }

		server := NewCallbackServer(config, addr, browserFunc)

		assert.NotNil(t, server)
		assert.Equal(t, config, server.config)
		assert.Equal(t, addr, server.addr)
		assert.NotNil(t, server.browser)
	})

	t.Run("creates callback server with nil browser function", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":8080", nil)

		assert.NotNil(t, server)
		assert.Equal(t, config, server.config)
		assert.Nil(t, server.browser)
	})
}

func TestCallbackServer_GetToken(t *testing.T) { //nolint:paralleltest
	t.Run("successful token retrieval", func(t *testing.T) { //nolint:paralleltest
		// Create a test server to mock the OAuth2 provider
		mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mock token exchange endpoint
			if r.URL.Path == "/oauth/token" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{
					"access_token": "test-access-token",
					"token_type": "Bearer",
					"expires_in": 3600,
					"refresh_token": "test-refresh-token"
				}`)
			}
		}))
		defer mockProvider.Close()

		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:9999/auth/redirect",
			Endpoint: oauth2.Endpoint{
				TokenURL: mockProvider.URL + "/oauth/token",
			},
			Scopes: []string{"read_user"},
		}

		var capturedURL string
		browserFunc := func(authURL string) error {
			capturedURL = authURL
			// Simulate user completing OAuth flow by making a request to the callback
			go func() {
				time.Sleep(100 * time.Millisecond)

				// Parse the auth URL to extract state
				parsedURL, err := url.Parse(capturedURL)
				if err != nil {
					return
				}
				state := parsedURL.Query().Get("state")

				// Make callback request with timeout
				callbackURL := fmt.Sprintf("http://localhost:9999/auth/redirect?code=test-code&state=%s", state)
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(callbackURL)
				if err == nil {
					resp.Body.Close()
				}
			}()
			return nil
		}

		server := NewCallbackServer(config, ":9999", browserFunc)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token, err := server.GetToken(ctx)

		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "test-access-token", token.AccessToken)
		assert.Equal(t, "Bearer", token.TokenType)
		assert.NotEmpty(t, capturedURL)
		assert.Contains(t, capturedURL, "client_id=test-client-id")
		assert.Contains(t, capturedURL, "code_challenge")
		assert.Contains(t, capturedURL, "state")
	})

	t.Run("context timeout", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:9999/auth/redirect",
		}

		browserFunc := func(url string) error {
			// Don't make any callback request to simulate timeout
			return nil
		}

		server := NewCallbackServer(config, ":9999", browserFunc)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		token, err := server.GetToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("browser function fails", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:9999/auth/redirect",
		}

		expectedError := errors.New("failed to open browser")
		browserFunc := func(url string) error {
			return expectedError
		}

		server := NewCallbackServer(config, ":9999", browserFunc)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		token, err := server.GetToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "failed to open browser")
	})

	t.Run("invalid redirect URL", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "://invalid-url",
		}

		browserFunc := func(url string) error { return nil }
		server := NewCallbackServer(config, ":9999", browserFunc)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		token, err := server.GetToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "missing protocol scheme")
	})

	t.Run("server fails to start", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		browserFunc := func(url string) error { return nil }

		// Try to bind to an invalid address
		server := NewCallbackServer(config, "invalid-address", browserFunc)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		token, err := server.GetToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "server failed")
	})
}

func TestCallbackServer_CallbackHandler(t *testing.T) { //nolint:paralleltest
	t.Run("successful callback with valid code and state", func(t *testing.T) { //nolint:paralleltest
		// Create a test server to mock the OAuth2 provider
		mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth/token" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{
					"access_token": "test-access-token",
					"token_type": "Bearer",
					"expires_in": 3600
				}`)
			}
		}))
		defer mockProvider.Close()

		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
			Endpoint: oauth2.Endpoint{
				TokenURL: mockProvider.URL + "/oauth/token",
			},
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		// Create channels for testing
		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		expectedState := "test-state"
		verifier := "test-verifier"

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, expectedState, verifier)

		// Create test request
		req := httptest.NewRequest("GET", "/auth/redirect?code=test-code&state=test-state", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Check that token was received
		select {
		case token := <-tokenChan:
			assert.NotNil(t, token)
			assert.Equal(t, "test-access-token", token.AccessToken)
		case err := <-errorChan:
			t.Fatalf("Expected token but got error: %v", err)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for token")
		}

		// Check HTTP response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Authentication Successful!")
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
	})

	t.Run("callback with invalid state", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		expectedState := "expected-state"
		verifier := "test-verifier"

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, expectedState, verifier)

		// Create test request with wrong state
		req := httptest.NewRequest("GET", "/auth/redirect?code=test-code&state=wrong-state", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Check that error was received
		select {
		case <-tokenChan:
			t.Fatal("Expected error but got token")
		case err := <-errorChan:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid state")
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for error")
		}

		// Check HTTP response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid state")
	})

	t.Run("callback with missing authorization code", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		expectedState := "test-state"
		verifier := "test-verifier"

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, expectedState, verifier)

		// Create test request without code
		req := httptest.NewRequest("GET", "/auth/redirect?state=test-state", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Check that error was received
		select {
		case <-tokenChan:
			t.Fatal("Expected error but got token")
		case err := <-errorChan:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no authorization code received")
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for error")
		}

		// Check HTTP response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "no authorization code received")
	})

	t.Run("callback with OAuth error", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		expectedState := "test-state"
		verifier := "test-verifier"

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, expectedState, verifier)

		// Create test request with OAuth error
		req := httptest.NewRequest("GET", "/auth/redirect?error=access_denied&state=test-state", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Check that error was received
		select {
		case <-tokenChan:
			t.Fatal("Expected error but got token")
		case err := <-errorChan:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "authorization error: access_denied")
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for error")
		}

		// Check HTTP response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "authorization error: access_denied")
	})

	t.Run("token exchange fails", func(t *testing.T) { //nolint:paralleltest
		// Create a test server that returns an error
		mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth/token" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error": "invalid_grant"}`)
			}
		}))
		defer mockProvider.Close()

		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
			Endpoint: oauth2.Endpoint{
				TokenURL: mockProvider.URL + "/oauth/token",
			},
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		expectedState := "test-state"
		verifier := "test-verifier"

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, expectedState, verifier)

		// Create test request
		req := httptest.NewRequest("GET", "/auth/redirect?code=test-code&state=test-state", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Check that error was received
		select {
		case <-tokenChan:
			t.Fatal("Expected error but got token")
		case err := <-errorChan:
			assert.Error(t, err)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for error")
		}

		// Check HTTP response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Token exchange failed")
	})
}

func TestCallbackServer_Shutdown(t *testing.T) { //nolint:paralleltest
	t.Run("shutdown with no server", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		// Should not panic when server is nil
		assert.NotPanics(t, func() {
			server.shutdown()
		})
	})

	t.Run("shutdown with running server", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		// Start a server
		server.server = &http.Server{
			Addr:    ":9999",
			Handler: http.NewServeMux(),
		}

		// Should not panic
		assert.NotPanics(t, func() {
			server.shutdown()
		})
	})
}

func TestCallbackServer_EdgeCases(t *testing.T) { //nolint:paralleltest
	t.Run("handles malformed callback URLs", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, "test-state", "test-verifier")

		// Test with malformed query parameters
		req := httptest.NewRequest("GET", "/auth/redirect?code=test-code&state=test-state&malformed=%", nil)
		w := httptest.NewRecorder()

		// Should not panic
		assert.NotPanics(t, func() {
			handler(w, req)
		})
	})

	t.Run("handles empty query parameters", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, "test-state", "test-verifier")

		// Test with empty query parameters
		req := httptest.NewRequest("GET", "/auth/redirect?code=&state=", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Should receive error for missing code
		select {
		case <-tokenChan:
			t.Fatal("Expected error but got token")
		case err := <-errorChan:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid state")
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for error")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:9999/auth/redirect",
		}

		browserFunc := func(url string) error {
			// Don't make any callback request
			return nil
		}

		server := NewCallbackServer(config, ":9999", browserFunc)

		ctx, cancel := context.WithCancel(context.Background())

		// Cancel the context immediately
		cancel()

		token, err := server.GetToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestCallbackServer_ErrorHandling(t *testing.T) { //nolint:paralleltest
	t.Run("handles server startup failure gracefully", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		browserFunc := func(url string) error { return nil }

		// Use an address that's likely to fail
		server := NewCallbackServer(config, "256.256.256.256:80", browserFunc)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		token, err := server.GetToken(ctx)

		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("handles missing state parameter", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, "expected-state", "test-verifier")

		// Create test request without state parameter
		req := httptest.NewRequest("GET", "/auth/redirect?code=test-code", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Should receive error for invalid state (empty != expected)
		select {
		case <-tokenChan:
			t.Fatal("Expected error but got token")
		case err := <-errorChan:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid state")
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for error")
		}

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("handles multiple error parameters in callback", func(t *testing.T) { //nolint:paralleltest
		config := &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost:7171/auth/redirect",
		}

		server := NewCallbackServer(config, ":9999", func(url string) error { return nil })

		tokenChan := make(chan *oauth2.Token, 1)
		errorChan := make(chan error, 1)

		handler := server.callbackHandler(context.Background(), tokenChan, errorChan, "test-state", "test-verifier")

		// Create test request with multiple error scenarios
		req := httptest.NewRequest("GET", "/auth/redirect?error=access_denied&error_description=User+denied+access&state=test-state", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Should receive error for OAuth error
		select {
		case <-tokenChan:
			t.Fatal("Expected error but got token")
		case err := <-errorChan:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "authorization error: access_denied")
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for error")
		}

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
