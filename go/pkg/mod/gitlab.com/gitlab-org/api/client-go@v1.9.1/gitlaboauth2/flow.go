package gitlaboauth2

import (
	"context"

	"golang.org/x/oauth2"
)

// AuthorizationFlow performs a complete OAuth2 authorization flow for GitLab authentication.
//
// This function provides a simplified interface for performing OAuth2 authentication
// by combining the OAuth2 configuration creation and callback server handling into
// a single convenient function call.
//
// The function:
// 1. Creates an OAuth2 configuration for the specified GitLab instance
// 2. Sets up a local callback server on port 7171
// 3. Opens the authorization URL in the user's browser
// 4. Waits for the user to complete authentication
// 5. Handles the callback and exchanges the code for an access token
// 6. Returns the access token
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - baseURL: The base URL of the GitLab instance. Use "" for GitLab.com,
//     or provide the full URL for self-managed instances (e.g., "https://gitlab.example.com")
//   - clientID: The OAuth2 client ID obtained from your GitLab application settings
//   - scopes: A slice of OAuth2 scopes to request (e.g., []string{"read_user", "api"})
//   - browser: A function that opens URLs in the user's browser
//
// Returns:
//   - *oauth2.Token: The OAuth2 access token on successful authentication
//   - error: An error if the authentication flow fails at any step
//
// This function is a convenience wrapper around NewOAuth2Config and NewCallbackServer.
// For more control over the authentication flow, use those functions directly.
//
// Example usage for GitLab.com:
//
//	browserFunc := func(url string) error {
//		return exec.Command("open", url).Start() // macOS
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	token, err := gitlaboauth2.AuthorizationFlow(ctx, "", "your-client-id",
//		"http://localhost:7171/auth/redirect",
//		[]string{"read_user", "api"}, ":7171", browserFunc)
//	if err != nil {
//		log.Fatalf("Authentication failed: %v", err)
//	}
//
//	fmt.Printf("Access token: %s\n", token.AccessToken)
//
// Example usage for self-managed GitLab:
//
//	browserFunc := func(url string) error {
//		return exec.Command("xdg-open", url).Start() // Linux
//	}
//
//	ctx := context.Background()
//	token, err := gitlaboauth2.AuthorizationFlow(ctx, "https://gitlab.company.com",
//		"your-client-id", "http://localhost:7171/auth/redirect", []string{"read_user"}, ":7171", browserFunc)
//	if err != nil {
//		log.Fatalf("Authentication failed: %v", err)
//	}
//
//	// Use the token to create an authenticated HTTP client
//	config := gitlaboauth2.NewOAuth2Config("https://gitlab.company.com", "your-client-id", []string{"read_user"})
//	client := config.Client(ctx, token)
//
//	// Now you can use the client to make authenticated requests
//	resp, err := client.Get("https://gitlab.company.com/api/v4/user")
//	if err != nil {
//		log.Fatalf("API request failed: %v", err)
//	}
//	defer resp.Body.Close()
func AuthorizationFlow(ctx context.Context, baseURL, clientID, redirectURL string, scopes []string, callbackServerListenAddr string, browser BrowserFunc) (*oauth2.Token, error) {
	config := NewOAuth2Config(baseURL, clientID, redirectURL, scopes)
	server := NewCallbackServer(config, callbackServerListenAddr, browser)

	return server.GetToken(ctx)
}
