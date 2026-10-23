// Package config provides functionality around an experimental GitLab config file
// to handle GitLab instance setups, including where to find them and how to
// authenticate with them.
//
// Attention: This package is experimental and the Go API and the config file API might change
// at any time.
package config

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"buf.build/go/protovalidate"
	"buf.build/go/protoyaml"
	"github.com/zalando/go-keyring"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gitlab.com/gitlab-org/api/client-go/config/v1beta1"
	"gitlab.com/gitlab-org/api/client-go/gitlaboauth2"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// SaaSHostname is the GitLab.com SaaS hostname
	SaaSHostname = "gitlab.com"

	// DefaultConfigFileName is the default name for the config file
	DefaultConfigFileName = "config.yaml"

	// DefaultConfigDirName is the default directory name for GitLab config
	DefaultConfigDirName = "gitlab"

	// ConfigVersion is the current API version for the config schema
	ConfigVersion = "gitlab.com/config/v1beta1"

	// DefaultAPIVersion is the default GitLab Rest API version
	DefaultAPIVersion = "v4"

	// defaultExecTimeout is the default timeout for command executions
	defaultExecTimeout = 60 * time.Second

	// EnvVarGitLabConfigFile is the environment variable name for specifying a custom config file path
	EnvVarGitLabConfigFile = "GITLAB_CONFIG"

	// EnvVarGitLabContext is the environment variable name for specifying the current context to use
	EnvVarGitLabContext = "GITLAB_CONTEXT"

	// CredentialSourceNotFoundExitCode is the command exit code used to determine if a credential source wasn't found
	CredentialSourceNotFoundExitCode = 2

	// execCredentialSourceSupported is a flag to enable support for the `exec` credential source. It's currently always disabled
	// We plan to harden it and enable it in a future iteration.
	execCredentialSourceSupported = false
)

var (
	errExecCredentialSourceNotSupported = errors.New("the exec credential source is not yet supported")

	secureCipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	}
)

type errCredentialSourceNotFound struct {
	inner error
}

func (e *errCredentialSourceNotFound) Error() string {
	return fmt.Sprintf("credential source not found: %s", e.inner)
}

// Config represents a GitLab configuration that manages instances, authentication,
// and contexts for creating GitLab API clients.
type Config struct {
	ctx                 context.Context
	path                string
	config              *v1beta1.Config
	oauth2Settings      *OAuth2Settings
	additionalValidator ConfigValidator
}

// OAuth2Settings configures OAuth2 authentication flow parameters.
type OAuth2Settings struct {
	AuthorizationFlowEnabled bool
	CallbackServerListenAddr string
	Browser                  gitlaboauth2.BrowserFunc
	ClientID                 string
	ClientSecret             string
	RedirectURL              string
	Scopes                   []string
}

// ConfigOption is a function that modifies a Config during initialization.
type ConfigOption func(c *Config)

// ConfigValidator is a function that validates a Config structure.
type ConfigValidator func(c *v1beta1.Config) error

// WithPath sets the config file path for the Config.
func WithPath(path string) ConfigOption {
	return func(c *Config) {
		c.path = path
	}
}

// WithContext sets the context for the Config.
func WithContext(ctx context.Context) ConfigOption {
	return func(c *Config) {
		c.ctx = ctx
	}
}

// WithOAuth2Settings configures OAuth2 authentication settings for the Config.
func WithOAuth2Settings(settings OAuth2Settings) ConfigOption {
	return func(c *Config) {
		c.oauth2Settings = &settings
	}
}

// WithAdditionalValidator adds a custom validator to the Config.
func WithAdditionalValidator(validator ConfigValidator) ConfigOption {
	return func(c *Config) {
		c.additionalValidator = validator
	}
}

// WithAutoCISupport automatically configures the Config for GitLab CI/CD environments
// by reading CI environment variables and setting up appropriate authentication.
func WithAutoCISupport() ConfigOption {
	return func(c *Config) {
		if isCi, found := os.LookupEnv("CI"); found && isCi == "true" {
			ciAPIURL, found := os.LookupEnv("CI_API_V4_URL")
			if !found || ciAPIURL == "" {
				return
			}

			c.path = ""
			c.config = &v1beta1.Config{
				Instances: []*v1beta1.Instance{
					{
						Name:   gitlab.Ptr("auto-ci-support"),
						Server: gitlab.Ptr(ciAPIURL),
					},
				},
				Auths: []*v1beta1.Auth{
					{
						Name: gitlab.Ptr("auto-ci-support-job-token"),
						AuthInfo: &v1beta1.AuthInfo{
							AuthProvider: &v1beta1.AuthInfo_JobToken{
								JobToken: &v1beta1.JobToken{
									JobToken: &v1beta1.JobToken_TokenSource{
										TokenSource: &v1beta1.CredentialSource{
											Source: &v1beta1.CredentialSource_EnvVar{
												EnvVar: "CI_JOB_TOKEN",
											},
										},
									},
								},
							},
						},
					},
				},
				Contexts: []*v1beta1.Context{
					{
						Name:     gitlab.Ptr("auto-ci-support"),
						Instance: gitlab.Ptr("auto-ci-support"),
						Auth:     gitlab.Ptr("auto-ci-support-job-token"),
					},
				},
				CurrentContext: gitlab.Ptr("auto-ci-support"),
			}

			if caFile, found := os.LookupEnv("CI_SERVER_TLS_CA_FILE"); found && caFile != "" {
				c.config.Instances[0].InstanceCa = &v1beta1.Instance_CertificateAuthoritySource{
					CertificateAuthoritySource: &v1beta1.CredentialSource{
						Source: &v1beta1.CredentialSource_EnvVar{
							EnvVar: "CI_SERVER_TLS_CA_FILE",
						},
					},
				}
			}

			if certFile, found := os.LookupEnv("CI_SERVER_TLS_CERT_FILE"); found && certFile != "" {
				c.config.Instances[0].InstanceClientCert = &v1beta1.Instance_ClientCertSource{
					ClientCertSource: &v1beta1.CredentialSource{
						Source: &v1beta1.CredentialSource_EnvVar{
							EnvVar: "CI_SERVER_TLS_CERT_FILE",
						},
					},
				}
			}

			if keyFile, found := os.LookupEnv("CI_SERVER_TLS_KEY_FILE"); found && keyFile != "" {
				c.config.Instances[0].InstanceClientKey = &v1beta1.Instance_ClientKeySource{
					ClientKeySource: &v1beta1.CredentialSource{
						Source: &v1beta1.CredentialSource_EnvVar{
							EnvVar: "CI_SERVER_TLS_KEY_FILE",
						},
					},
				}
			}
		}
	}
}

// New creates a new Config with default config path and applies the given options.
func New(options ...ConfigOption) *Config {
	return NewFromPath(DefaultConfigPath(), options...)
}

// NewFromPath creates a new Config from the specified path and applies the given options.
func NewFromPath(path string, options ...ConfigOption) *Config {
	c := &Config{
		path: path,
	}

	// apply option functions
	for _, f := range options {
		f(c)
	}

	// apply Config field defaults
	if c.ctx == nil {
		c.ctx = context.Background()
	}

	return c
}

// NewFromString creates a new Config from a YAML string and applies the given options.
func NewFromString(s string, options ...ConfigOption) (*Config, error) {
	c := &Config{}

	// apply option functions
	for _, f := range options {
		f(c)
	}

	if err := c.load([]byte(s)); err != nil {
		return nil, err
	}

	// apply Config field defaults
	if c.ctx == nil {
		c.ctx = context.Background()
	}

	return c, nil
}

// Empty creates a new empty Config with no configuration loaded.
func Empty(options ...ConfigOption) *Config {
	c, err := NewFromString(``, options...)
	if err != nil {
		panic(fmt.Errorf("creating an empty config should never result in an error, but it did: %w. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", err))
	}
	return c
}

// DefaultConfigPath returns the default configuration file path, checking environment
// variables and standard locations in order of precedence.
func DefaultConfigPath() string {
	if path, found := os.LookupEnv(EnvVarGitLabConfigFile); found {
		return path
	}

	configDir, err := os.UserConfigDir()
	if err == nil {
		path := filepath.Join(configDir, DefaultConfigDirName, DefaultConfigFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		path := filepath.Join(homeDir, "."+DefaultConfigDirName, DefaultConfigFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	path := filepath.Join(".", "."+DefaultConfigDirName, DefaultConfigFileName)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

// Load reads and parses the configuration file from the configured path.
// Returns an error if the file cannot be read or parsed.
func (c *Config) Load() error {
	var configData []byte
	if c.config == nil {
		if c.path == "" {
			return errors.New("unable to locate config file")
		}

		f, err := os.Open(c.path)
		if err != nil {
			return fmt.Errorf("unable to open config file: %w", err)
		}
		defer f.Close()

		configData, err = io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("unable to read from config file: %w", err)
		}
	}

	return c.load(configData)
}

func (c *Config) load(b []byte) error {
	if c.config == nil {
		options := protoyaml.UnmarshalOptions{
			Path: c.path,
		}
		var config v1beta1.Config
		if err := options.Unmarshal(b, &config); err != nil {
			return fmt.Errorf("unable to unmarshal config file: %w", err)
		}
		c.config = &config
	}

	c.defaulting()
	return c.validate()
}

// Save writes the configuration to the configured file path.
// Returns an error if the config is invalid or cannot be written.
func (c *Config) Save() error {
	if c.path == "" {
		return errors.New("unable to save config to empty path. Use New(), NewFromPath(p) to create config with a path")
	}

	if c.config == nil {
		return errors.New("unable to save empty config. Load config first or manually create a configuration")
	}

	if err := c.validate(); err != nil {
		return fmt.Errorf("unable to save config because validation failed: %w", err)
	}

	options := protoyaml.MarshalOptions{
		EmitUnpopulated: false,
	}
	configData, err := options.Marshal(c.config)
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}

	if err := os.WriteFile(c.path, configData, 0o600); err != nil {
		return fmt.Errorf("unable to write config to file: %w", err)
	}

	return nil
}

// Instances returns all configured GitLab instances.
func (c *Config) Instances() []*v1beta1.Instance {
	if c.config == nil {
		return nil
	}

	return c.config.Instances
}

// Auths returns all configured authentication methods.
func (c *Config) Auths() []*v1beta1.Auth {
	if c.config == nil {
		return nil
	}

	return c.config.Auths
}

// Contexts returns all configured contexts.
func (c *Config) Contexts() []*v1beta1.Context {
	if c.config == nil {
		return nil
	}

	return c.config.Contexts
}

// CurrentContext returns the currently active context.
func (c *Config) CurrentContext() *v1beta1.Context {
	if c.config == nil {
		return nil
	}

	if ctxFromEnv := os.Getenv(EnvVarGitLabContext); ctxFromEnv != "" {
		return c.Context(ctxFromEnv)
	}

	if c.config.CurrentContext != nil {
		return c.Context(*c.config.CurrentContext)
	}

	return nil
}

func (c *Config) Context(name string) *v1beta1.Context {
	if c.config.Contexts == nil {
		return nil
	}

	idx := slices.IndexFunc(c.config.Contexts, func(context *v1beta1.Context) bool {
		return *context.Name == name
	})
	if idx < 0 {
		return nil
	}

	return c.config.Contexts[idx]
}

// Auth returns the authentication configuration for the specified name.
func (c *Config) Auth(name string) *v1beta1.Auth {
	if c.config == nil {
		return nil
	}

	idx := slices.IndexFunc(c.config.Auths, func(auth *v1beta1.Auth) bool {
		return *auth.Name == name
	})
	if idx < 0 {
		return nil
	}

	return c.config.Auths[idx]
}

// Instance returns the instance configuration for the specified name.
func (c *Config) Instance(name string) *v1beta1.Instance {
	if c.config == nil {
		return nil
	}

	idx := slices.IndexFunc(c.config.Instances, func(instance *v1beta1.Instance) bool {
		return *instance.Name == name
	})
	if idx < 0 {
		return nil
	}

	return c.config.Instances[idx]
}

// Extensions returns the custom configuration data.
func (c *Config) Extensions() map[string]*structpb.Struct {
	return c.config.Extensions
}

// NewClient creates a new GitLab API client using the current context.
// Returns an error if the current context cannot be resolved.
func (c *Config) NewClient(options ...gitlab.ClientOptionFunc) (*gitlab.Client, error) {
	currentContext := c.CurrentContext()
	if currentContext == nil {
		return nil, errors.New("unable to resolve current context for new client")
	}
	return c.NewClientForContext(*currentContext.Name, options...)
}

// NewClientForContext creates a new GitLab API client using the specified context.
// Returns an error if the context cannot be resolved or the client cannot be created.
func (c *Config) NewClientForContext(name string, options ...gitlab.ClientOptionFunc) (*gitlab.Client, error) {
	configContext := c.Context(name)
	if configContext == nil {
		return nil, fmt.Errorf("unable to resolve context %s for new client", name)
	}

	auth := c.Auth(*configContext.Auth)
	instance := c.Instance(*configContext.Instance)

	baseURL, err := c.resolveBaseURL(instance)
	if err != nil {
		return nil, err
	}

	defaultRequestOptions := []gitlab.RequestOptionFunc{
		gitlab.WithContext(c.ctx),
	}

	// custom headers
	customHeaders, err := c.resolveCustomHeaders(instance.CustomHeaders)
	if err != nil {
		return nil, err
	}
	if len(customHeaders) > 0 {
		defaultRequestOptions = append(defaultRequestOptions, gitlab.WithHeaders(customHeaders))
	}

	configOptions := []gitlab.ClientOptionFunc{
		gitlab.WithBaseURL(baseURL.String()),
		gitlab.WithRequestOptions(defaultRequestOptions...),
	}

	if c.config.Preferences.RetryMax != nil {
		configOptions = append(configOptions, gitlab.WithCustomRetryMax(int(*c.config.Preferences.RetryMax)))
	}

	if c.config.Preferences.RetryWaitMin != nil && c.config.Preferences.RetryWaitMax != nil {
		configOptions = append(configOptions, gitlab.WithCustomRetryWaitMinMax(
			c.config.Preferences.RetryWaitMin.AsDuration(),
			c.config.Preferences.RetryWaitMax.AsDuration(),
		))
	}

	if instance.RateLimit != nil {
		configOptions = append(configOptions, gitlab.WithCustomLimiter(
			rate.NewLimiter(
				rate.Limit(*instance.RateLimit.RequestsPerSecond),
				int(*instance.RateLimit.Burst),
			),
		))
	}

	// Create TLS configuration based on client settings
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}

	if instance.InsecureSkipTlsVerify != nil {
		tlsConfig.InsecureSkipVerify = *instance.InsecureSkipTlsVerify
	}

	// Set secure cipher suites for gitlab.com
	if baseURL.Hostname() == SaaSHostname {
		tlsConfig.CipherSuites = secureCipherSuites
	}

	// Configure custom CA if provided
	ca, err := c.resolveCACertificate(instance)
	if err != nil {
		return nil, err
	}

	if ca != nil {
		// use system cert pool as a baseline
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		caCertPool.AppendCertsFromPEM(ca)
		tlsConfig.RootCAs = caCertPool
	}

	// Configure client certificates if provided
	clientCert, err := c.resolveClientCert(instance)
	if err != nil {
		return nil, err
	}
	clientKey, err := c.resolveClientKey(instance)
	if err != nil {
		return nil, err
	}
	if clientCert != nil && clientKey != nil {
		clientCert, err := tls.X509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	// Set appropriate timeouts based on whether custom CA is used
	dialTimeout := 5 * time.Second
	keepAlive := 5 * time.Second
	idleTimeout := 30 * time.Second
	if ca != nil {
		dialTimeout = 30 * time.Second
		keepAlive = 30 * time.Second
		idleTimeout = 90 * time.Second
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   dialTimeout,
				KeepAlive: keepAlive,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       idleTimeout,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       tlsConfig,
		},
	}

	// set HTTP client as context value for oauth2 package to use for token refreshes etc.
	if hc := c.ctx.Value(oauth2.HTTPClient); hc == nil {
		c.ctx = context.WithValue(c.ctx, oauth2.HTTPClient, httpClient)
	}

	configOptions = append(configOptions, gitlab.WithHTTPClient(httpClient))

	// append all user provided options after config options so that the former take precedence.
	configOptions = append(configOptions, options...)

	authSource, err := c.resolveAuthSource(auth, baseURL)
	if err != nil {
		return nil, err
	}

	client, err := gitlab.NewAuthSourceClient(authSource, configOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to create new client from config: %w", err)
	}

	return client, nil
}

func (c *Config) resolveAuthSource(auth *v1beta1.Auth, baseURL *url.URL) (gitlab.AuthSource, error) {
	if auth == nil || auth.AuthInfo == nil {
		return nil, errors.New("unable to resolve auth source for empty auth")
	}

	switch ap := auth.AuthInfo.AuthProvider.(type) {
	case *v1beta1.AuthInfo_PersonalAccessToken:
		return c.resolvePersonalAccessToken(ap.PersonalAccessToken)
	case *v1beta1.AuthInfo_JobToken:
		return c.resolveJobToken(ap.JobToken)
	case *v1beta1.AuthInfo_Oauth2:
		return c.resolveOAuth2(ap.Oauth2, baseURL)
	case *v1beta1.AuthInfo_BasicAuth:
		return c.resolveBasicAuth(ap.BasicAuth)
	default:
		panic(fmt.Sprintf("unexpected v1beta1.AuthInfo.AuthProvider type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", ap))
	}
}

func (c *Config) resolvePersonalAccessToken(pat *v1beta1.PersonalAccessToken) (gitlab.AuthSource, error) {
	switch t := pat.PersonalAccessToken.(type) {
	case *v1beta1.PersonalAccessToken_Token:
		return gitlab.AccessTokenAuthSource{Token: t.Token}, nil
	case *v1beta1.PersonalAccessToken_TokenSource:
		token, err := c.resolveCredentialSource(t.TokenSource)
		if err != nil {
			return nil, err
		}
		return gitlab.AccessTokenAuthSource{Token: token}, nil
	default:
		panic(fmt.Sprintf("unexpected v1beta1.PersonalAccessToken type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", t))
	}
}

func (c *Config) resolveJobToken(jt *v1beta1.JobToken) (gitlab.AuthSource, error) {
	switch j := jt.JobToken.(type) {
	case *v1beta1.JobToken_Token:
		return gitlab.JobTokenAuthSource{Token: j.Token}, nil
	case *v1beta1.JobToken_TokenSource:
		token, err := c.resolveCredentialSource(j.TokenSource)
		if err != nil {
			return nil, err
		}
		return gitlab.JobTokenAuthSource{Token: token}, nil
	default:
		panic(fmt.Sprintf("unexpected v1beta1.JobToken type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", j))
	}
}

func (c *Config) resolveOAuth2(o *v1beta1.OAuth2, baseURL *url.URL) (gitlab.AuthSource, error) {
	if c.oauth2Settings == nil {
		return nil, errors.New("unable to resolve OAuth2 token in config because config object was not created with OAuth2 settings, use WithOAuth2Settings")
	}

	clientSecret, err := c.resolveOAuth2ClientSecret(o)
	if err != nil {
		return nil, err
	}
	if c.oauth2Settings.ClientSecret != "" || clientSecret != "" {
		panic("support for OAuth2 client secrets is not yet implemented")
	}

	var u string
	if baseURL.Hostname() != SaaSHostname {
		u = baseURL.String()
	}

	clientID := c.oauth2Settings.ClientID
	if o.ClientId != nil {
		clientID = *o.ClientId
	}

	oauth2Config := gitlaboauth2.NewOAuth2Config(u, clientID, c.oauth2Settings.RedirectURL, c.oauth2Settings.Scopes)

	if c.oauth2Settings.AuthorizationFlowEnabled {
		// read token to figure out if authorization flow is required
		token, err := c.readOAuth2Token(o, true)
		if err != nil {
			return nil, fmt.Errorf("unable to read OAuth2 token: %w", err)
		}
		if token.RefreshToken == "" {
			server := gitlaboauth2.NewCallbackServer(oauth2Config, c.oauth2Settings.CallbackServerListenAddr, c.oauth2Settings.Browser)

			token, err = server.GetToken(c.ctx)
			if err != nil {
				return nil, err
			}

			if err := c.writeOAuth2Token(o, token); err != nil {
				return nil, err
			}
		}
	}

	tokenSource, err := NewConfigTokenSource(
		c.ctx,
		oauth2Config,
		func() (*oauth2.Token, error) {
			return c.readOAuth2Token(o, false)
		},
		func(t *oauth2.Token) error {
			return c.writeOAuth2Token(o, t)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create OAuth2 token source: %w", err)
	}

	return gitlab.OAuthTokenSource{TokenSource: tokenSource}, nil
}

func (c *Config) resolveBasicAuth(ba *v1beta1.BasicAuth) (gitlab.AuthSource, error) {
	var username, password string
	switch u := ba.BasicAuthUsername.(type) {
	case *v1beta1.BasicAuth_Username:
		username = u.Username
	case *v1beta1.BasicAuth_UsernameSource:
		us, err := c.resolveCredentialSource(u.UsernameSource)
		if err != nil {
			return nil, err
		}
		username = us
	default:
		panic(fmt.Sprintf("unexpected v1beta1.BasicAuthUsername: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", u))
	}
	switch p := ba.BasicAuthPassword.(type) {
	case *v1beta1.BasicAuth_Password:
		password = p.Password
	case *v1beta1.BasicAuth_PasswordSource:
		ps, err := c.resolveCredentialSource(p.PasswordSource)
		if err != nil {
			return nil, err
		}
		password = ps
	default:
		panic(fmt.Sprintf("unexpected v1beta1.BasicAuthPassword: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", p))
	}

	return &gitlab.PasswordCredentialsAuthSource{Username: username, Password: password}, nil
}

func (c *Config) readOAuth2Token(o *v1beta1.OAuth2, allowNotFound bool) (*oauth2.Token, error) {
	token := oauth2.Token{
		Expiry: o.ExpiresAt.AsTime(),
	}
	switch at := o.Oauth2AccessToken.(type) {
	case *v1beta1.OAuth2_AccessToken:
		token.AccessToken = at.AccessToken
	case *v1beta1.OAuth2_AccessTokenSource:
		x, err := c.resolveCredentialSource(at.AccessTokenSource)
		if err != nil {
			var e *errCredentialSourceNotFound
			if !errors.As(err, &e) || !allowNotFound {
				return nil, err
			}
		}
		token.AccessToken = x
	default:
		panic(fmt.Sprintf("unexpected v1beta1.Oauth2AccessToken type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", at))
	}

	switch rt := o.Oauth2RefreshToken.(type) {
	case *v1beta1.OAuth2_RefreshToken:
		token.RefreshToken = rt.RefreshToken
	case *v1beta1.OAuth2_RefreshTokenSource:
		x, err := c.resolveCredentialSource(rt.RefreshTokenSource)
		if err != nil {
			var e *errCredentialSourceNotFound
			if !errors.As(err, &e) || !allowNotFound {
				return nil, err
			}
		}
		token.RefreshToken = x
	default:
		panic(fmt.Sprintf("unexpected v1beta1.Oauth2RefreshToken type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", rt))
	}

	return &token, nil
}

func (c *Config) writeOAuth2Token(o *v1beta1.OAuth2, token *oauth2.Token) error {
	o.ExpiresAt = timestamppb.New(token.Expiry)

	switch at := o.Oauth2AccessToken.(type) {
	case *v1beta1.OAuth2_AccessToken:
		at.AccessToken = token.AccessToken
	case *v1beta1.OAuth2_AccessTokenSource:
		if err := c.writeToCredentialSource(at.AccessTokenSource, token.AccessToken); err != nil {
			return err
		}
	default:
		panic(fmt.Sprintf("unexpected v1beta1.Oauth2AccessToken type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", at))
	}

	switch rt := o.Oauth2RefreshToken.(type) {
	case *v1beta1.OAuth2_RefreshToken:
		rt.RefreshToken = token.RefreshToken
	case *v1beta1.OAuth2_RefreshTokenSource:
		if err := c.writeToCredentialSource(rt.RefreshTokenSource, token.RefreshToken); err != nil {
			return err
		}
	default:
		panic(fmt.Sprintf("unexpected v1beta1.Oauth2RefreshToken type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", rt))
	}

	return c.Save()
}

func (c *Config) resolveOAuth2ClientSecret(o *v1beta1.OAuth2) (string, error) {
	switch s := o.Oauth2ClientSecret.(type) {
	case nil:
		return "", nil
	case *v1beta1.OAuth2_ClientSecret:
		return s.ClientSecret, nil
	case *v1beta1.OAuth2_ClientSecretSource:
		v, err := c.resolveCredentialSource(s.ClientSecretSource)
		if err != nil {
			return "", err
		}
		return v, nil
	default:
		panic(fmt.Sprintf("unexpected v1beta1.Oauth2ClientSecret type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", s))
	}
}

func (c *Config) resolveCustomHeaders(headers []*v1beta1.Header) (map[string]string, error) {
	hs := make(map[string]string, len(headers))

	for _, header := range headers {
		switch v := header.HeaderValue.(type) {
		case *v1beta1.Header_Value:
			hs[*header.Name] = v.Value
		case *v1beta1.Header_ValueFrom:
			csv, err := c.resolveCredentialSource(v.ValueFrom)
			if err != nil {
				return nil, err
			}
			hs[*header.Name] = csv
		default:
			panic(fmt.Sprintf("unexpected v1beta.HeaderValue type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", v))
		}
	}

	return hs, nil
}

func (c *Config) resolveCredentialSource(cs *v1beta1.CredentialSource) (string, error) {
	switch s := cs.Source.(type) {
	case *v1beta1.CredentialSource_EnvVar:
		val, found := os.LookupEnv(s.EnvVar)
		if !found {
			return "", &errCredentialSourceNotFound{inner: fmt.Errorf("unable to lookup environment variable %q for credentials", s.EnvVar)}
		}
		return val, nil
	case *v1beta1.CredentialSource_Exec:
		if !execCredentialSourceSupported {
			return "", errExecCredentialSourceNotSupported
		}

		timeout := defaultExecTimeout
		if s.Exec.Timeout != nil {
			timeout = s.Exec.Timeout.AsDuration()
		}
		ctx, cancel := context.WithTimeout(c.ctx, timeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, *s.Exec.Command, s.Exec.Args...)
		cmd.Env = append(cmd.Env, envKeyValPair(s.Exec.Env)...)
		output, err := cmd.Output()
		if err != nil {
			e := fmt.Errorf("unable to read credential from command output: %w", err)
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				if ee.ExitCode() == CredentialSourceNotFoundExitCode {
					return "", &errCredentialSourceNotFound{inner: e}
				}
			}
			return "", e
		}
		return string(bytes.TrimSpace(output)), nil
	case *v1beta1.CredentialSource_File:
		expandedPath, err := expandPath(s.File)
		if err != nil {
			return "", err
		}

		content, err := os.ReadFile(expandedPath)
		if err != nil {
			e := fmt.Errorf("unable to read credentials from file %q: %w", expandedPath, err)
			if errors.Is(err, os.ErrNotExist) {
				return "", &errCredentialSourceNotFound{inner: e}
			}

			return "", e
		}
		return string(content), nil
	case *v1beta1.CredentialSource_Keyring:
		content, err := keyring.Get(*s.Keyring.Service, *s.Keyring.User)
		if err != nil {
			e := fmt.Errorf("unable to get credentials from keyring service %q and user %q: %w", *s.Keyring.Service, *s.Keyring.User, err)
			if errors.Is(err, keyring.ErrNotFound) {
				return "", &errCredentialSourceNotFound{inner: e}
			}
			return "", e
		}
		return content, nil
	case *v1beta1.CredentialSource_Value:
		return s.Value, nil
	default:
		panic(fmt.Sprintf("unexpected v1beta1.CredentialSource.Source type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", s))
	}
}

func (c *Config) writeToCredentialSource(cs *v1beta1.CredentialSource, value string) error {
	switch s := cs.Source.(type) {
	case *v1beta1.CredentialSource_EnvVar:
		if err := os.Setenv(s.EnvVar, value); err != nil {
			return fmt.Errorf("unable to persist value in environment variable %q for credential source: %w", s.EnvVar, err)
		}
		return nil
	case *v1beta1.CredentialSource_Exec:
		if !execCredentialSourceSupported {
			return errExecCredentialSourceNotSupported
		}

		timeout := defaultExecTimeout
		if s.Exec.Timeout != nil {
			timeout = s.Exec.Timeout.AsDuration()
		}
		ctx, cancel := context.WithTimeout(c.ctx, timeout)
		defer cancel()
		args := []string{"--write"}
		args = append(args, s.Exec.Args...)
		cmd := exec.CommandContext(ctx, *s.Exec.Command, args...)
		cmd.Env = append(cmd.Env, envKeyValPair(s.Exec.Env)...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("unable to write credential with command: %w", err)
		}
		return nil
	case *v1beta1.CredentialSource_File:
		expandedPath, err := expandPath(s.File)
		if err != nil {
			return err
		}

		if err := os.WriteFile(expandedPath, []byte(value), 0o600); err != nil {
			return fmt.Errorf("unable to write value to credential source file %q: %w", expandedPath, err)
		}
		return nil
	case *v1beta1.CredentialSource_Keyring:
		if err := keyring.Set(*s.Keyring.Service, *s.Keyring.User, value); err != nil {
			return fmt.Errorf("unable to write credentials to keyring service %q and user %q: %w", *s.Keyring.Service, *s.Keyring.User, err)
		}
		return nil
	case *v1beta1.CredentialSource_Value:
		s.Value = value
		return nil
	default:
		panic(fmt.Sprintf("unexpected v1beta1.CredentialSource.Source type: %#v. Please report at https://gitlab.com/gitlab-org/api/client-go/-/issues", s))
	}
}

func (c *Config) resolveCACertificate(instance *v1beta1.Instance) ([]byte, error) {
	switch ca := instance.InstanceCa.(type) {
	case *v1beta1.Instance_CertificateAuthority:
		return []byte(ca.CertificateAuthority), nil
	case *v1beta1.Instance_CertificateAuthoritySource:
		s, err := c.resolveCredentialSource(ca.CertificateAuthoritySource)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	default:
		return nil, nil
	}
}

func (c *Config) resolveClientCert(instance *v1beta1.Instance) ([]byte, error) {
	switch cert := instance.InstanceClientCert.(type) {
	case *v1beta1.Instance_ClientCert:
		return []byte(cert.ClientCert), nil
	case *v1beta1.Instance_ClientCertSource:
		s, err := c.resolveCredentialSource(cert.ClientCertSource)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	default:
		return nil, nil
	}
}

func (c *Config) resolveClientKey(instance *v1beta1.Instance) ([]byte, error) {
	switch key := instance.InstanceClientKey.(type) {
	case *v1beta1.Instance_ClientKey:
		return []byte(key.ClientKey), nil
	case *v1beta1.Instance_ClientKeySource:
		s, err := c.resolveCredentialSource(key.ClientKeySource)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	default:
		return nil, nil
	}
}

func (c *Config) resolveBaseURL(instance *v1beta1.Instance) (*url.URL, error) {
	u, err := url.Parse(*instance.Server)
	if err != nil {
		return nil, fmt.Errorf("unable to parse instance %q server URL %q: %w", *instance.Name, *instance.Server, err)
	}

	apiVersion := DefaultAPIVersion
	if instance.ApiVersion != nil {
		apiVersion = *instance.ApiVersion
	}

	apiSuffix := fmt.Sprintf("api/%s", apiVersion)
	if strings.HasSuffix(strings.TrimPrefix(u.Path, "/"), apiSuffix) {
		return u, nil
	}

	return u.JoinPath(apiSuffix), nil
}

func (c *Config) defaulting() {
	if c.config.Version == nil {
		c.config.Version = gitlab.Ptr(ConfigVersion)
	}

	if c.config.Preferences == nil {
		c.config.Preferences = &v1beta1.Preferences{}
	}

	for _, auth := range c.config.Auths {
		if auth.AuthInfo == nil {
			continue
		}

		switch p := auth.AuthInfo.AuthProvider.(type) {
		case *v1beta1.AuthInfo_Oauth2:
			if p.Oauth2.Oauth2AccessToken == nil {
				p.Oauth2.Oauth2AccessToken = &v1beta1.OAuth2_AccessToken{}
			}
			if p.Oauth2.Oauth2RefreshToken == nil {
				p.Oauth2.Oauth2RefreshToken = &v1beta1.OAuth2_RefreshToken{}
			}
		}
	}
}

func (c *Config) validate() error {
	if err := protovalidate.Validate(c.config); err != nil {
		return fmt.Errorf("failed to validate against schema: %w", err)
	}

	if c.additionalValidator == nil {
		return nil
	}

	return c.additionalValidator(c.config)
}

func envKeyValPair(m map[string]string) []string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return s
}

func expandPath(p string) (string, error) {
	p = os.ExpandEnv(p)

	if after, found := strings.CutPrefix(p, "~"); found {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to expand ~ in path: %w", err)
		}
		p = filepath.Join(h, after)
	}

	return p, nil
}
