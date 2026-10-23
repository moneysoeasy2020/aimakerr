// Package gitlaboauth2 provides OAuth2 configuration utilities for GitLab API clients.
//
// This package simplifies the creation of OAuth2 configurations for both
// GitLab.com and self-managed GitLab instances. It handles the proper
// endpoint configuration automatically based on the provided base URL.
//
// Example usage for GitLab.com:
//
//	config := gitlaboauth2.NewOAuth2Config("", "your-client-id", []string{"read_user", "read_repository"})
//
// Example usage for self-managed GitLab:
//
//	config := gitlaboauth2.NewOAuth2Config("https://gitlab.example.com", "your-client-id", []string{"read_user"})
//
// The package automatically configures the appropriate OAuth2 endpoints
// and uses a default redirect URI suitable for local development.
package gitlaboauth2

import (
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

// NewOAuth2Config creates a new OAuth2 configuration for GitLab authentication.
//
// This function configures OAuth2 settings for both GitLab.com and self-managed
// GitLab instances. It automatically determines the appropriate OAuth2 endpoints
// based on the provided base URL.
//
// Parameters:
//   - baseURL: The base URL of the GitLab instance. Use an empty string "" for GitLab.com.
//     For self-managed instances, provide the full URL (e.g., "https://gitlab.example.com").
//     The function automatically handles URL normalization by removing trailing slashes
//     and "/api/v4" suffixes.
//   - clientID: The OAuth2 client ID obtained from your GitLab application settings.
//   - scopes: A slice of OAuth2 scopes to request. Common scopes include:
//     "read_user", "read_repository", "write_repository", "api", "read_api", etc.
//
// Returns:
//   - *oauth2.Config: A configured OAuth2 configuration ready for use with the
//     golang.org/x/oauth2 package.
//
// The function automatically sets a default redirect URI suitable for local
// development (http://localhost:7171/auth/redirect).
//
// Example usage for GitLab.com:
//
//	config := gitlaboauth2.NewOAuth2Config("", "your-client-id", []string{"read_user", "api"})
//	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
//
// Example usage for self-managed GitLab:
//
//	config := gitlaboauth2.NewOAuth2Config("https://gitlab.company.com", "your-client-id", []string{"read_user"})
//	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
//
//	// Later, exchange the authorization code for a token
//	token, err := config.Exchange(context.Background(), "authorization-code")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create an HTTP client with the token
//	client := config.Client(context.Background(), token)
func NewOAuth2Config(baseURL, clientID, redirectURL string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURL,
		Endpoint:    endpoint(baseURL),
		Scopes:      scopes,
	}
}

func endpoint(baseURL string) oauth2.Endpoint {
	if baseURL == "" {
		return endpoints.GitLab
	}

	// Self-Managed
	baseURL = strings.TrimSuffix(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/api/v4")
	return oauth2.Endpoint{
		AuthURL:       baseURL + "/oauth/authorize",
		TokenURL:      baseURL + "/oauth/token",
		DeviceAuthURL: baseURL + "/oauth/authorize_device",
	}
}
