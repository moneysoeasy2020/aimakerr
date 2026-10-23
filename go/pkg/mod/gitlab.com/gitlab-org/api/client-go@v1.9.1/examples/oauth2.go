package main

import (
	"context"
	"fmt"
	"os/exec"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gitlab.com/gitlab-org/api/client-go/gitlaboauth2"
)

func oauth2() {
	ctx := context.Background()
	// Authorize with GitLab.com and OAuth2
	clientID := "your-client-id-here"
	redirectURL := "http://localhost:9999/auth/redirect"
	scopes := []string{"read_api"}
	config := gitlaboauth2.NewOAuth2Config("", clientID, redirectURL, scopes)

	server := gitlaboauth2.NewCallbackServer(config, ":9999", func(url string) error {
		return exec.Command("open", url).Start()
	})

	token, err := server.GetToken(ctx)
	if err != nil {
		panic(err)
	}

	client, err := gitlab.NewAuthSourceClient(gitlab.OAuthTokenSource{TokenSource: config.TokenSource(ctx, token)})
	if err != nil {
		panic(err)
	}

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Current user: %s\n", user.Username)
}
