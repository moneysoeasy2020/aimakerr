package gitlaboauth2

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// CallbackServer handles the OAuth2 callback flow for GitLab authentication.
//
// This server provides a complete OAuth2 flow implementation that:
// - Starts a local HTTP server to handle the OAuth2 callback
// - Opens the authorization URL in the user's browser
// - Handles the callback with the authorization code
// - Exchanges the code for an access token
// - Automatically shuts down after completion
//
// The server uses PKCE (Proof Key for Code Exchange) for enhanced security
// and handles state verification to prevent CSRF attacks.
type CallbackServer struct {
	config  *oauth2.Config
	server  *http.Server
	addr    string
	browser BrowserFunc
}

// BrowserFunc is a function type for opening URLs in a browser.
//
// This function should open the provided URL in the user's default browser
// or handle the URL opening in an appropriate way for the application.
//
// Parameters:
//   - url: The authorization URL to open in the browser
//
// Returns:
//   - error: An error if the browser could not be opened, nil otherwise
//
// Example implementation using the "xdg-open" command:
//
//	browserFunc := func(url string) error {
//		return exec.Command("xdg-open", url).Start()
//	}
type BrowserFunc func(url string) error

// NewCallbackServer creates a new callback server for handling OAuth2 authentication flow.
//
// This function initializes a CallbackServer that manages the complete OAuth2 flow,
// including starting a local HTTP server, opening the browser, and handling the callback.
//
// Parameters:
//   - config: The OAuth2 configuration created with NewOAuth2Config
//   - addr: The address for the local HTTP server (e.g., ":7171" or "localhost:8080")
//   - browser: A function that opens URLs in the user's browser
//
// Returns:
//   - *CallbackServer: A configured callback server ready to handle OAuth2 flow
//
// Example usage:
//
//	config := gitlaboauth2.NewOAuth2Config("", "client-id", []string{"read_user"})
//	browserFunc := func(url string) error {
//		return exec.Command("open", url).Start() // macOS
//	}
//	server := gitlaboauth2.NewCallbackServer(config, ":7171", browserFunc)
//
//	ctx := context.Background()
//	token, err := server.GetToken(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Access token: %s\n", token.AccessToken)
func NewCallbackServer(config *oauth2.Config, addr string, browser BrowserFunc) *CallbackServer {
	return &CallbackServer{
		config:  config,
		addr:    addr,
		browser: browser,
	}
}

// GetToken performs the complete OAuth2 flow and returns an access token.
//
// This method orchestrates the entire OAuth2 authentication flow by:
// 1. Generating a secure state and PKCE verifier for security
// 2. Starting a local HTTP server to handle the callback
// 3. Opening the authorization URL in the user's browser
// 4. Waiting for the user to complete authentication
// 5. Handling the callback with the authorization code
// 6. Exchanging the code for an access token
// 7. Automatically shutting down the server
//
// The method uses PKCE (Proof Key for Code Exchange) for enhanced security
// and implements proper state verification to prevent CSRF attacks.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - *oauth2.Token: The OAuth2 access token on successful authentication
//   - error: An error if the authentication flow fails at any step
//
// The method will block until:
// - The user completes authentication and a token is obtained
// - The context is canceled or times out
// - An error occurs during the authentication process
//
// Example usage:
//
//	config := gitlaboauth2.NewOAuth2Config("", "client-id", []string{"read_user"})
//	browserFunc := func(url string) error {
//		return exec.Command("open", url).Start()
//	}
//	server := gitlaboauth2.NewCallbackServer(config, ":7171", browserFunc)
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	token, err := server.GetToken(ctx)
//	if err != nil {
//		log.Fatalf("Authentication failed: %v", err)
//	}
//
//	fmt.Printf("Access token: %s\n", token.AccessToken)
//	fmt.Printf("Token type: %s\n", token.TokenType)
//	fmt.Printf("Expires at: %v\n", token.Expiry)
func (s *CallbackServer) GetToken(ctx context.Context) (*oauth2.Token, error) {
	// Channel to receive the result
	tokenChan := make(chan *oauth2.Token, 1)
	defer close(tokenChan)
	errorChan := make(chan error, 1)
	defer close(errorChan)

	state := oauth2.GenerateVerifier()
	verifier := oauth2.GenerateVerifier()
	authURL := s.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	u, err := url.Parse(s.config.RedirectURL)
	if err != nil {
		return nil, err
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/"+strings.TrimPrefix(u.Path, "/"), s.callbackHandler(ctx, tokenChan, errorChan, state, verifier))

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	// Start server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errorChan <- fmt.Errorf("server failed: %w", err)
		}
	}()

	if err := s.browser(authURL); err != nil {
		errorChan <- fmt.Errorf("failed to open browser: %w", err)
	}

	var token *oauth2.Token
	select {
	case <-ctx.Done():
		s.shutdown()
		return nil, ctx.Err()
	case err := <-errorChan:
		s.shutdown()
		return nil, err
	case token = <-tokenChan:
		s.shutdown()
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *CallbackServer) callbackHandler(ctx context.Context, tokenChan chan *oauth2.Token, errorChan chan error, expectedState, verifier string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for errors
		if errorParam := r.URL.Query().Get("error"); errorParam != "" {
			err := fmt.Errorf("authorization error: %s", errorParam)
			errorChan <- err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check for correct state
		state := r.URL.Query().Get("state")
		if state != expectedState {
			err := fmt.Errorf("invalid state")
			errorChan <- err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Extract authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			err := fmt.Errorf("no authorization code received")
			errorChan <- err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Exchange code for token
		token, err := s.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			errorChan <- err
			http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Send success response
		tokenChan <- token
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`
            <html>
            <body style="text-align: center; padding: 50px;">
                <h1 style="color: green;">Authentication Successful!</h1>
                <p>You can close this window and return to the application.</p>
            </body>
            </html>
        `))
	}
}

func (s *CallbackServer) shutdown() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.server.Shutdown(ctx)
	}
}
