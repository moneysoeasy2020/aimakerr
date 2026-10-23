package main

import (
	"fmt"
	"log"

	"github.com/MakeNowJust/heredoc/v2"
	"gitlab.com/gitlab-org/api/client-go"
	"gitlab.com/gitlab-org/api/client-go/config"
)

func main() {
	// Example 0: empty config
	_ = config.Empty()

	// Example 1: Using the default config location
	basicConfigExample()

	// Example 2: auto CI support
	// autoCISupportExample()

	// Example 3: extensions
	extensions()

	// Example 4: custom headers
	customHeaders()
}

func basicConfigExample() {
	fmt.Println("=== Basic Config Example ===")

	// Create a config with default location (~/.config/gitlab/config.yaml)
	cfg := config.New(
		config.WithOAuth2Settings(config.OAuth2Settings{
			AuthorizationFlowEnabled: true,
			CallbackServerListenAddr: ":7171",
			Browser: func(url string) error {
				fmt.Printf("Open: %s\n", url)
				return nil
			},
			ClientID:    "41d48f9422ebd655dd9cf2947d6979681dfaddc6d0c56f7628f6ada59559af1e",
			RedirectURL: "http://localhost:7171/auth/redirect",
			Scopes:      []string{"openid", "profile", "read_user", "write_repository", "api"},
		}),
	)

	// Load the configuration
	if err := cfg.Load(); err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}

	client, err := cfg.NewClient(gitlab.WithUserAgent("my-app"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Use the client
	user, _, err := client.Users.CurrentUser()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	fmt.Printf("Authenticated as: %s (%s)\n", user.Name, user.Username)
}

func autoCISupportExample() {
	fmt.Println("=== Auto CI Support Example ===")

	// Create a config with precedence to CI support and falls back to the default paths for configs otherwise
	// cfg := config.New(config.WithAutoCISupport())
	cfg := config.Empty(config.WithAutoCISupport())

	// Load the configuration
	if err := cfg.Load(); err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}

	client, err := cfg.NewClient(gitlab.WithUserAgent("my-app"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Use the client
	user, _, err := client.Users.CurrentUser()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	fmt.Printf("Authenticated as: %s (%s)\n", user.Name, user.Username)
}

type demoExtension struct {
	Browser   string `yaml:"browser"`
	Telemetry bool   `yaml:"telemetry"`
}

func extensions() {
	fmt.Println("=== Extensions ===")

	cfg, err := config.NewFromString(heredoc.Doc(`
		version: gitlab.com/config/v1beta1

		current-context: gitlab-com

		contexts:
		  - name: gitlab-com
		    instance: gitlab-com
		    auth: token-env

		instances:
		  - name: gitlab-com
		    server: https://gitlab.com

		auths:
		  - name: token-env
		    auth-info:
		      personal-access-token:
		        token-source:
		          env_var: GITLAB_TOKEN

		extensions:
		  my-app:
		    browser: firefox
		    telemetry: true
	`))
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	client, err := cfg.NewClient(gitlab.WithUserAgent("my-app"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Use the client
	user, _, err := client.Users.CurrentUser()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	ext := config.NewExtension[demoExtension]("my-app", cfg)
	data, err := ext.Unmarshal()
	if err != nil {
		log.Fatalf("Failed to unmarshal extension: %v", err)
	}

	fmt.Printf("Browser from extension is: %q\n", data.Browser)

	fmt.Printf("Authenticated as: %s (%s)\n", user.Name, user.Username)
}

func customHeaders() {
	fmt.Println("=== Custom Headers ===")

	cfg, err := config.NewFromString(heredoc.Doc(`
		version: gitlab.com/config/v1beta1

		current-context: gitlab-com

		contexts:
		  - name: gitlab-com
		    instance: gitlab-com
		    auth: token-env

		instances:
		  - name: gitlab-com
		    server: https://gitlab.com
		    custom_headers:
		      - name: My-Custom-Header
		        value: my-header-value

		auths:
		  - name: token-env
		    auth-info:
		      personal-access-token:
		        token-source:
		          env_var: GITLAB_TOKEN
	`))
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	client, err := cfg.NewClient(gitlab.WithUserAgent("my-app"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Use the client
	user, _, err := client.Users.CurrentUser()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	fmt.Printf("Authenticated as: %s (%s)\n", user.Name, user.Username)
}
