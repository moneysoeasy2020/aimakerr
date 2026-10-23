package config

import (
	"net/http"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gitlab.com/gitlab-org/api/client-go/config/v1beta1"
)

func TestConfig_EmptyConfig(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(``)

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestConfig_SingleInstance_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instances := c.Instances()
	assert.Len(t, instances, 1)
	assert.Equal(t, "example", *instances[0].Name)
	assert.Equal(t, "https://gitlab.example.com", *instances[0].Server)
}

func TestConfig_SingleInstance_Invalid_Name(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name:
		    server: https://gitlab.example.com
	`))

	// THEN
	require.ErrorContains(t, err, "value length must be at least 3 characters")
	require.Nil(t, c)
}

func TestConfig_MultipleInstances_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
		  - name: example-dev
		    server: https://gitlab-dev.example.com
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instances := c.Instances()
	assert.Len(t, instances, 2)
	assert.Equal(t, "example", *instances[0].Name)
	assert.Equal(t, "https://gitlab.example.com", *instances[0].Server)
	assert.Equal(t, "example-dev", *instances[1].Name)
	assert.Equal(t, "https://gitlab-dev.example.com", *instances[1].Server)
}

func TestConfig_MultipleInstances_Invalid_NotUniqueNames(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
		  - name: example
		    server: https://gitlab-dev.example.com
	`))

	// THEN
	require.ErrorContains(t, err, "instances: all names must be unique")
	require.Nil(t, c)
}

func TestConfig_Contexts_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
		  - name: example-dev
		    server: https://gitlab-dev.example.com

		auths:
		  - name: some-user
		    auth_info:
		      personal-access-token:
		        token: abc
		  - name: another-user
		    auth_info:
		      personal-access-token:
		        token: def

		contexts:
		  - name: example
		    instance: example
		    auth: some-user

		current-context: example
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestConfig_Auths_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: abc123
		  - name: job-token-user
		    auth_info:
		      job-token:
		        token: def456
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auths := c.Auths()
	assert.Len(t, auths, 2)
	assert.Equal(t, "pat-user", *auths[0].Name)
	assert.Equal(t, "job-token-user", *auths[1].Name)
}

func TestConfig_Auths_Invalid_Name(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name:
		    auth_info:
		      personal-access-token:
		        token: abc123
	`))

	// THEN
	require.ErrorContains(t, err, "value length must be at least 3 characters")
	require.Nil(t, c)
}

func TestConfig_Auths_Invalid_NotUniqueNames(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: duplicate-name
		    auth_info:
		      personal-access-token:
		        token: abc123
		  - name: duplicate-name
		    auth_info:
		      job-token:
		        token: def456
	`))

	// THEN
	require.ErrorContains(t, err, "auths: all names must be unique")
	require.Nil(t, c)
}

func TestConfig_Contexts_Invalid_Name(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: some-user
		    auth_info:
		      personal-access-token:
		        token: abc

		contexts:
		  - name:
		    instance: example
		    auth: some-user
	`))

	// THEN
	require.ErrorContains(t, err, "value length must be at least 3 characters")
	require.Nil(t, c)
}

func TestConfig_Contexts_Invalid_NotUniqueNames(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: some-user
		    auth_info:
		      personal-access-token:
		        token: abc

		contexts:
		  - name: duplicate-context
		    instance: example
		    auth: some-user
		  - name: duplicate-context
		    instance: example
		    auth: some-user
	`))

	// THEN
	require.ErrorContains(t, err, "contexts: all names must be unique")
	require.Nil(t, c)
}

func TestConfig_Contexts_Invalid_UnknownInstance(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: some-user
		    auth_info:
		      personal-access-token:
		        token: abc

		contexts:
		  - name: test-context
		    instance: unknown-instance
		    auth: some-user
	`))

	// THEN
	require.ErrorContains(t, err, "context.instance must reference an existing instance name")
	require.Nil(t, c)
}

func TestConfig_Contexts_Invalid_UnknownAuth(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: some-user
		    auth_info:
		      personal-access-token:
		        token: abc

		contexts:
		  - name: test-context
		    instance: example
		    auth: unknown-auth
	`))

	// THEN
	require.ErrorContains(t, err, "context.auth must reference an existing auth name")
	require.Nil(t, c)
}

func TestConfig_CurrentContext_Invalid_Unknown(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: some-user
		    auth_info:
		      personal-access-token:
		        token: abc

		contexts:
		  - name: test-context
		    instance: example
		    auth: some-user

		current-context: unknown-context
	`))

	// THEN
	require.ErrorContains(t, err, "current_context must reference an existing context name")
	require.Nil(t, c)
}

func TestConfig_Instance_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
		  - name: example-dev
		    server: https://gitlab-dev.example.com
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instance := c.Instance("example")
	assert.NotNil(t, instance)
	assert.Equal(t, "example", *instance.Name)
	assert.Equal(t, "https://gitlab.example.com", *instance.Server)

	instance = c.Instance("example-dev")
	assert.NotNil(t, instance)
	assert.Equal(t, "example-dev", *instance.Name)
	assert.Equal(t, "https://gitlab-dev.example.com", *instance.Server)
}

func TestConfig_Instance_NotFound(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instance := c.Instance("non-existent")
	assert.Nil(t, instance)
}

func TestConfig_Auth_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: abc123
		  - name: job-token-user
		    auth_info:
		      job-token:
		        token: def456
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("pat-user")
	assert.NotNil(t, auth)
	assert.Equal(t, "pat-user", *auth.Name)

	auth = c.Auth("job-token-user")
	assert.NotNil(t, auth)
	assert.Equal(t, "job-token-user", *auth.Name)
}

func TestConfig_Auth_NotFound(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: abc123
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("non-existent")
	assert.Nil(t, auth)
}

func TestConfig_CurrentContext_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: some-user
		    auth_info:
		      personal-access-token:
		        token: abc

		contexts:
		  - name: test-context
		    instance: example
		    auth: some-user

		current-context: test-context
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	currentContext := c.CurrentContext()
	assert.NotNil(t, currentContext)
	assert.Equal(t, "test-context", *currentContext.Name)
	assert.Equal(t, "example", *currentContext.Instance)
	assert.Equal(t, "some-user", *currentContext.Auth)
}

func TestConfig_CurrentContext_NotSet(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: some-user
		    auth_info:
		      personal-access-token:
		        token: abc

		contexts:
		  - name: test-context
		    instance: example
		    auth: some-user
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	currentContext := c.CurrentContext()
	assert.Nil(t, currentContext)
}

func TestConfig_Empty_ReturnsEmptyConfig(t *testing.T) {
	t.Parallel()

	// WHEN
	c := Empty()

	// THEN
	assert.NotNil(t, c)
	assert.Empty(t, c.Instances())
	assert.Empty(t, c.Auths())
	assert.Empty(t, c.Contexts())
	assert.Nil(t, c.CurrentContext())
}

func TestConfig_PersonalAccessToken_TokenSource_EnvVar(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          env-var: GITLAB_TOKEN
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("pat-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken())
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken().GetTokenSource())
	assert.Equal(t, "GITLAB_TOKEN", auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetEnvVar())
}

func TestConfig_JobToken_TokenSource_EnvVar(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: job-token-user
		    auth_info:
		      job-token:
		        token-source:
		          env-var: CI_JOB_TOKEN
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("job-token-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetJobToken())
	assert.NotNil(t, auth.AuthInfo.GetJobToken().GetTokenSource())
	assert.Equal(t, "CI_JOB_TOKEN", auth.AuthInfo.GetJobToken().GetTokenSource().GetEnvVar())
}

func TestConfig_BasicAuth_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: basic-auth-user
		    auth_info:
		      basic-auth:
		        username: testuser
		        password: testpass
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("basic-auth-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetBasicAuth())
	assert.Equal(t, "testuser", auth.AuthInfo.GetBasicAuth().GetUsername())
	assert.Equal(t, "testpass", auth.AuthInfo.GetBasicAuth().GetPassword())
}

func TestConfig_BasicAuth_WithSources(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: basic-auth-user
		    auth_info:
		      basic-auth:
		        username-source:
		          env-var: GITLAB_USERNAME
		        password-source:
		          env-var: GITLAB_PASSWORD
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("basic-auth-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetBasicAuth())
	assert.NotNil(t, auth.AuthInfo.GetBasicAuth().GetUsernameSource())
	assert.Equal(t, "GITLAB_USERNAME", auth.AuthInfo.GetBasicAuth().GetUsernameSource().GetEnvVar())
	assert.NotNil(t, auth.AuthInfo.GetBasicAuth().GetPasswordSource())
	assert.Equal(t, "GITLAB_PASSWORD", auth.AuthInfo.GetBasicAuth().GetPasswordSource().GetEnvVar())
}

func TestConfig_OAuth2_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: oauth2-user
		    auth_info:
		      oauth2:
		        access-token: access123
		        refresh-token: refresh456
		        client-id: client789
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("oauth2-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetOauth2())
	assert.Equal(t, "access123", auth.AuthInfo.GetOauth2().GetAccessToken())
	assert.Equal(t, "refresh456", auth.AuthInfo.GetOauth2().GetRefreshToken())
	assert.Equal(t, "client789", *auth.AuthInfo.GetOauth2().ClientId)
}

func TestConfig_Instance_WithRateLimit(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: rate-limited
		    server: https://gitlab.example.com
		    rate-limit:
		      requests-per-second: 100
		      burst: 10
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instance := c.Instance("rate-limited")
	assert.NotNil(t, instance)
	assert.NotNil(t, instance.RateLimit)
	assert.Equal(t, float64(100), *instance.RateLimit.RequestsPerSecond)
	assert.Equal(t, int32(10), *instance.RateLimit.Burst)
}

func TestConfig_Instance_WithTLSConfig(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: tls-configured
		    server: https://gitlab.example.com
		    insecure-skip-tls-verify: true
		    certificate-authority: |
		      -----BEGIN CERTIFICATE-----
		      MIIBkTCB+wIJANDY7JPmcK6/MA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNVBAMMCWxv
		      Y2FsaG9zdDAeFw0xNzEyMjgxNzExMTVaFw0xODEyMjgxNzExMTVaMBQxEjAQBgNV
		      BAMMCWxvY2FsaG9zdDBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQDYFCqLTAKG4GEa
		      QhrfFVdMh6LoN2fwyuv/NyvsQO2F3lJYO7ILN3Cq3K9KlcCrKtVhSJsLJ3KnJnvJ
		      CZp7d7wDAgMBAAEwDQYJKoZIhvcNAQELBQADQQBJlffJHybjDGxRMqaRmDhX0+6v
		      02q5S5OiPSUFhXRLWqk1C+aodO+QAiJWVrNKTN3+QOzqzDNDI9dwMWsYnNzC
		      -----END CERTIFICATE-----
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instance := c.Instance("tls-configured")
	assert.NotNil(t, instance)
	assert.NotNil(t, instance.InsecureSkipTlsVerify)
	assert.True(t, *instance.InsecureSkipTlsVerify)
	assert.NotNil(t, instance.GetCertificateAuthority())
}

func TestConfig_Instance_WithAPIVersion(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: custom-api-version
		    server: https://gitlab.example.com
		    api-version: v4
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instance := c.Instance("custom-api-version")
	assert.NotNil(t, instance)
	assert.NotNil(t, instance.ApiVersion)
	assert.Equal(t, "v4", *instance.ApiVersion)
}

func TestConfig_Preferences_Valid(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		preferences:
		  retry-max: 5
		  retry-wait-min: 100ms
		  retry-wait-max: 400ms
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	prefs := c.config.Preferences
	assert.NotNil(t, prefs)
	assert.Equal(t, int32(5), *prefs.RetryMax)
	assert.NotNil(t, prefs.RetryWaitMin)
	assert.NotNil(t, prefs.RetryWaitMax)
}

// Additional tests for increased coverage

func TestConfig_New_WithOptions(t *testing.T) {
	t.Parallel()

	// WHEN
	c := New(WithPath("/test/path"))

	// THEN
	assert.NotNil(t, c)
	assert.Equal(t, "/test/path", c.path)
}

func TestConfig_NewFromPath_WithOptions(t *testing.T) {
	t.Parallel()

	// WHEN
	c := NewFromPath("/custom/path", WithPath("/override/path"))

	// THEN
	assert.NotNil(t, c)
	assert.Equal(t, "/override/path", c.path)
}

func TestConfig_NewFromString_WithOptions(t *testing.T) {
	t.Parallel()

	// GIVEN
	testValidator := func(c *v1beta1.Config) error {
		return nil
	}

	// WHEN
	c, err := NewFromString(``, WithAdditionalValidator(testValidator))

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.additionalValidator)
}

func TestConfig_Load_FileNotFound(t *testing.T) {
	t.Parallel()

	// GIVEN
	c := NewFromPath("/non/existent/path")

	// WHEN
	err := c.Load()

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to open config file")
}

func TestConfig_Load_EmptyPath(t *testing.T) {
	t.Parallel()

	// GIVEN
	c := NewFromPath("")

	// WHEN
	err := c.Load()

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to locate config file")
}

func TestConfig_Save_EmptyPath(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(``)
	require.NoError(t, err)

	// WHEN
	err = c.Save()

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to save config to empty path")
}

func TestConfig_Save_NilConfig(t *testing.T) {
	t.Parallel()

	// GIVEN
	c := NewFromPath("/test/path")

	// WHEN
	err := c.Save()

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to save empty config")
}

func TestConfig_Instance_WithCertificateAuthorityFile(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: ca-file-instance
		    server: https://gitlab.example.com
		    certificate-authority-source:
		      file: /path/to/ca.crt
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instance := c.Instance("ca-file-instance")
	assert.NotNil(t, instance)
	assert.Equal(t, "/path/to/ca.crt", instance.GetCertificateAuthoritySource().GetFile())
}

func TestConfig_Instance_WithClientCertificates(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: client-cert-instance
		    server: https://gitlab.example.com
		    client-cert-source:
		      file: /path/to/client.crt
		    client-key-source:
		      file: /path/to/client.key
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instance := c.Instance("client-cert-instance")
	assert.NotNil(t, instance)
	assert.Equal(t, "/path/to/client.crt", instance.GetClientCertSource().GetFile())
	assert.Equal(t, "/path/to/client.key", instance.GetClientKeySource().GetFile())
}

func TestConfig_PersonalAccessToken_TokenSource_File(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: pat-file-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          file: /path/to/token
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("pat-file-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken())
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken().GetTokenSource())
	assert.Equal(t, "/path/to/token", auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetFile())
}

func TestConfig_PersonalAccessToken_TokenSource_Keyring(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: pat-keyring-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          keyring:
		            service: gitlab-sdk
		            user: personal-token
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("pat-keyring-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken())
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken().GetTokenSource())
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetKeyring())
	assert.Equal(t, "gitlab-sdk", *auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetKeyring().Service)
	assert.Equal(t, "personal-token", *auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetKeyring().User)
}

func TestConfig_PersonalAccessToken_TokenSource_Exec(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: pat-exec-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          exec:
		            command: echo
		            args: ["test-token"]
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("pat-exec-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken())
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken().GetTokenSource())
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetExec())
	assert.Equal(t, "echo", *auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetExec().Command)
	assert.Equal(t, []string{"test-token"}, auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetExec().Args)
}

func TestConfig_PersonalAccessToken_TokenSource_Value(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: pat-value-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          value: direct-token-value
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("pat-value-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken())
	assert.NotNil(t, auth.AuthInfo.GetPersonalAccessToken().GetTokenSource())
	assert.Equal(t, "direct-token-value", auth.AuthInfo.GetPersonalAccessToken().GetTokenSource().GetValue())
}

func TestConfig_OAuth2_WithTokenSources(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: oauth2-sources-user
		    auth_info:
		      oauth2:
		        access-token-source:
		          env-var: OAUTH_ACCESS_TOKEN
		        refresh-token-source:
		          env-var: OAUTH_REFRESH_TOKEN
		        client-secret-source:
		          env-var: OAUTH_CLIENT_SECRET
		        client-id: client123
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	auth := c.Auth("oauth2-sources-user")
	assert.NotNil(t, auth)
	assert.NotNil(t, auth.AuthInfo.GetOauth2())
	assert.NotNil(t, auth.AuthInfo.GetOauth2().GetAccessTokenSource())
	assert.Equal(t, "OAUTH_ACCESS_TOKEN", auth.AuthInfo.GetOauth2().GetAccessTokenSource().GetEnvVar())
	assert.NotNil(t, auth.AuthInfo.GetOauth2().GetRefreshTokenSource())
	assert.Equal(t, "OAUTH_REFRESH_TOKEN", auth.AuthInfo.GetOauth2().GetRefreshTokenSource().GetEnvVar())
	assert.NotNil(t, auth.AuthInfo.GetOauth2().GetClientSecretSource())
	assert.Equal(t, "OAUTH_CLIENT_SECRET", auth.AuthInfo.GetOauth2().GetClientSecretSource().GetEnvVar())
	assert.Equal(t, "client123", *auth.AuthInfo.GetOauth2().ClientId)
}

func TestConfig_Contexts_Valid_Multiple(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
		  - name: example-dev
		    server: https://gitlab-dev.example.com

		auths:
		  - name: user1
		    auth_info:
		      personal-access-token:
		        token: token1
		  - name: user2
		    auth_info:
		      personal-access-token:
		        token: token2

		contexts:
		  - name: prod-context
		    instance: example
		    auth: user1
		  - name: dev-context
		    instance: example-dev
		    auth: user2

		current-context: prod-context
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	contexts := c.Contexts()
	assert.Len(t, contexts, 2)
	assert.Equal(t, "prod-context", *contexts[0].Name)
	assert.Equal(t, "example", *contexts[0].Instance)
	assert.Equal(t, "user1", *contexts[0].Auth)
	assert.Equal(t, "dev-context", *contexts[1].Name)
	assert.Equal(t, "example-dev", *contexts[1].Instance)
	assert.Equal(t, "user2", *contexts[1].Auth)

	currentContext := c.CurrentContext()
	assert.NotNil(t, currentContext)
	assert.Equal(t, "prod-context", *currentContext.Name)
}

func TestConfig_Instance_Invalid_ServerURL(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: invalid-server
		    server:
	`))

	// THEN
	require.Error(t, err)
	require.Nil(t, c)
}

func TestConfig_Auths_Invalid_EmptyAuthInfo(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		auths:
		  - name: empty-auth
		    auth_info:
	`))

	// THEN
	require.Error(t, err)
	require.Nil(t, c)
}

func TestConfig_DefaultConfigPath_ReturnsPath(t *testing.T) {
	t.Parallel()

	// WHEN
	path := DefaultConfigPath()

	// THEN
	// Should return a string (may be empty if no config file exists)
	assert.IsType(t, "", path)
}

func TestConfig_Constants(t *testing.T) {
	t.Parallel()

	// THEN
	assert.Equal(t, "gitlab.com", SaaSHostname)
	assert.Equal(t, "config.yaml", DefaultConfigFileName)
	assert.Equal(t, "gitlab", DefaultConfigDirName)
	assert.Equal(t, "gitlab.com/config/v1beta1", ConfigVersion)
	assert.Equal(t, "v4", DefaultAPIVersion)
	assert.Equal(t, "GITLAB_CONFIG", EnvVarGitLabConfigFile)
}

func TestConfig_Instance_WithAllFields(t *testing.T) {
	t.Parallel()

	// WHEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: full-instance
		    server: https://gitlab.example.com
		    api-version: v4
		    insecure-skip-tls-verify: false
		    certificate-authority: |
		      -----BEGIN CERTIFICATE-----
		      MIIBkTCB+wIJANDY7JPmcK6/MA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNVBAMMCWxv
		      Y2FsaG9zdDAeFw0xNzEyMjgxNzExMTVaFw0xODEyMjgxNzExMTVaMBQxEjAQBgNV
		      BAMMCWxvY2FsaG9zdDBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQDYFCqLTAKG4GEa
		      QhrfFVdMh6LoN2fwyuv/NyvsQO2F3lJYO7ILN3Cq3K9KlcCrKtVhSJsLJ3KnJnvJ
		      CZp7d7wDAgMBAAEwDQYJKoZIhvcNAQELBQADQQBJlffJHybjDGxRMqaRmDhX0+6v
		      02q5S5OiPSUFhXRLWqk1C+aodO+QAiJWVrNKTN3+QOzqzDNDI9dwMWsYnNzC
		      -----END CERTIFICATE-----
		    client-cert: |
		      -----BEGIN CERTIFICATE-----
		      MIIBkTCB+wIJANDY7JPmcK6/MA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNVBAMMCWxv
		      Y2FsaG9zdDAeFw0xNzEyMjgxNzExMTVaFw0xODEyMjgxNzExMTVaMBQxEjAQBgNV
		      BAMMCWxvY2FsaG9zdDBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQDYFCqLTAKG4GEa
		      QhrfFVdMh6LoN2fwyuv/NyvsQO2F3lJYO7ILN3Cq3K9KlcCrKtVhSJsLJ3KnJnvJ
		      CZp7d7wDAgMBAAEwDQYJKoZIhvcNAQELBQADQQBJlffJHybjDGxRMqaRmDhX0+6v
		      02q5S5OiPSUFhXRLWqk1C+aodO+QAiJWVrNKTN3+QOzqzDNDI9dwMWsYnNzC
		      -----END CERTIFICATE-----
		    client-key: |
		      -----BEGIN PRIVATE KEY-----
		      MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDYFCqLTAKG4GEa
		      QhrfFVdMh6LoN2fwyuv/NyvsQO2F3lJYO7ILN3Cq3K9KlcCrKtVhSJsLJ3KnJnvJ
		      CZp7d7wDAgMBAAECggEAJR5bY5D6TcIZ6okQoJmf5qgGAFEcGsKJSzPBcCKBqOhP
		      -----END PRIVATE KEY-----
		    rate-limit:
		      requests-per-second: 50
		      burst: 5
	`))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, c)

	instance := c.Instance("full-instance")
	assert.NotNil(t, instance)
	assert.Equal(t, "full-instance", *instance.Name)
	assert.Equal(t, "https://gitlab.example.com", *instance.Server)
	assert.Equal(t, "v4", *instance.ApiVersion)
	assert.NotNil(t, instance.InsecureSkipTlsVerify)
	assert.False(t, *instance.InsecureSkipTlsVerify)
	assert.NotNil(t, instance.GetCertificateAuthority())
	assert.NotNil(t, instance.GetClientCert())
	assert.NotNil(t, instance.GetClientKey())
	assert.NotNil(t, instance.RateLimit)
	assert.Equal(t, float64(50), *instance.RateLimit.RequestsPerSecond)
	assert.Equal(t, int32(5), *instance.RateLimit.Burst)
}

func TestConfig_NilConfig_Methods(t *testing.T) {
	t.Parallel()

	// GIVEN
	c := &Config{} // nil config

	// THEN
	assert.Nil(t, c.Instances())
	assert.Nil(t, c.Auths())
	assert.Nil(t, c.Contexts())
	assert.Nil(t, c.CurrentContext())
	assert.Nil(t, c.Instance("test"))
	assert.Nil(t, c.Auth("test"))
}

func TestConfig_NewClient_Success_OverrideBaseURL(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: custom-api-path
		    server: https://gitlab.example.com/api/v4

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: custom-api-path
		    auth: pat-user

		current-context: test-context
	`))
	require.NoError(t, err)
	require.NotNil(t, c)

	// WHEN
	client, err := c.NewClient(gitlab.WithBaseURL("https://gitlab.com"))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "https://gitlab.com/api/v4/", client.BaseURL().String())
}

func TestConfig_NewClientForContext_Success_PersonalAccessToken(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_JobToken(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: job-token-user
		    auth_info:
		      job-token:
		        token: job-token-123

		contexts:
		  - name: test-context
		    instance: example
		    auth: job-token-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_BasicAuth(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: basic-auth-user
		    auth_info:
		      basic-auth:
		        username: testuser
		        password: testpass

		contexts:
		  - name: test-context
		    instance: example
		    auth: basic-auth-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Error_ContextNotFound(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("non-existent-context")

	// THEN
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "unable to resolve context non-existent-context for new client")
}

func TestConfig_NewClientForContext_Error_AuthNotFound(t *testing.T) {
	t.Parallel()

	// GIVEN - This test should use a valid config first, then test runtime error
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-user
	`))
	require.NoError(t, err)

	// Manually modify the context to point to non-existent auth to bypass validation
	c.config.Contexts[0].Auth = &[]string{"non-existent-auth"}[0]

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "unable to resolve auth source for empty auth")
}

func TestConfig_NewClientForContext_Success_WithRateLimit(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: rate-limited
		    server: https://gitlab.example.com
		    rate-limit:
		      requests-per-second: 100
		      burst: 10

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: rate-limited
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_WithRetryConfig(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		preferences:
		  retry-max: 5
		  retry-wait-min: 100ms
		  retry-wait-max: 400ms

		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_WithAPIVersion(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
		    api-version: v4

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "https://gitlab.example.com/api/v4/", client.BaseURL().String())
}

func TestConfig_NewClientForContext_Success_WithTLSConfig(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: tls-configured
		    server: https://gitlab.example.com
		    insecure-skip-tls-verify: true
		    certificate-authority: |
		      -----BEGIN CERTIFICATE-----
		      MIIBkTCB+wIJANDY7JPmcK6/MA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNVBAMMCWxv
		      Y2FsaG9zdDAeFw0xNzEyMjgxNzExMTVaFw0xODEyMjgxNzExMTVaMBQxEjAQBgNV
		      BAMMCWxvY2FsaG9zdDBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQDYFCqLTAKG4GEa
		      QhrfFVdMh6LoN2fwyuv/NyvsQO2F3lJYO7ILN3Cq3K9KlcCrKtVhSJsLJ3KnJnvJ
		      CZp7d7wDAgMBAAEwDQYJKoZIhvcNAQELBQADQQBJlffJHybjDGxRMqaRmDhX0+6v
		      02q5S5OiPSUFhXRLWqk1C+aodO+QAiJWVrNKTN3+QOzqzDNDI9dwMWsYnNzC
		      -----END CERTIFICATE-----

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: tls-configured
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_PersonalAccessTokenFromEnv(t *testing.T) {
	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: pat-env-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          env-var: GITLAB_TOKEN

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-env-user
	`))
	require.NoError(t, err)

	// Set environment variable
	t.Setenv("GITLAB_TOKEN", "env-token-123")

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_JobTokenFromEnv(t *testing.T) {
	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: job-token-env-user
		    auth_info:
		      job-token:
		        token-source:
		          env-var: CI_JOB_TOKEN

		contexts:
		  - name: test-context
		    instance: example
		    auth: job-token-env-user
	`))
	require.NoError(t, err)

	// Set environment variable
	t.Setenv("CI_JOB_TOKEN", "ci-job-token-123")

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_BasicAuthFromEnv(t *testing.T) {
	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: basic-auth-env-user
		    auth_info:
		      basic-auth:
		        username-source:
		          env-var: GITLAB_USERNAME
		        password-source:
		          env-var: GITLAB_PASSWORD

		contexts:
		  - name: test-context
		    instance: example
		    auth: basic-auth-env-user
	`))
	require.NoError(t, err)

	// Set environment variables
	t.Setenv("GITLAB_USERNAME", "testuser")
	t.Setenv("GITLAB_PASSWORD", "testpass")

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_GitLabSaaS(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: saas
		    server: https://gitlab.com

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: saas
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_WithClientCertificates(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: client-cert-instance
		    server: https://gitlab.example.com
		    client-cert: |
		        -----BEGIN CERTIFICATE-----
		        MIIDAzCCAeugAwIBAgIQYVyQi9/mh7pLS7BO/VMhZzANBgkqhkiG9w0BAQsFADAi
		        MQ8wDQYDVQQKEwZHaXRMYWIxDzANBgNVBAMTBkdpdExhYjAeFw0yNTA3MTcxMzQ5
		        MjZaFw0yNjA3MTcxMzQ5MjZaMCIxDzANBgNVBAoTBkdpdExhYjEPMA0GA1UEAxMG
		        R2l0TGFiMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1tLkU1s3X8ui
		        +61E850vCo/TP+g5EJSfFtnL30I7M6RQNCcc+CbhK6kKpPS7zEt5QyIC8i8HeInC
		        bRhPJ2wRGrSyzlkkD7GkMC1ktfhMVM3/DjsIU7BcbJWLlI6yCypd9zrNQR8lazac
		        XnrjPjGeXG1faXJoPQC9VnO0+iKWD34QQr8Hhlou9Jh+GGpZIjrfM1IYfgAl2SCG
		        WI2NynYx8GjSTEODBVFm3LApQxG9j3rR7Umop16uJOsm6Pr7q/IPed2cgtZ4Lm+s
		        cREoI/0ASU4eEUnUw55LcJDGuggEggJmlpqF+/Tn/K2uYB20ujpKLU2FWx6wuH2i
		        xS53B1b34QIDAQABozUwMzAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0lBAwwCgYIKwYB
		        BQUHAwIwDAYDVR0TAQH/BAIwADANBgkqhkiG9w0BAQsFAAOCAQEAItJBL74KO+hZ
		        Ormh9bAgrO2Rl5vR/+SYlO8Szy4FnwB0ctgaKYSdo1X5O27U6MERVa6nF+CZe3Ps
		        bFGbtBbecezlkCgtUypR4YgMNmlJX/gIDc5k+AShKUrszL3PEqYYOh+WKNf5DRaI
		        JbCbCQ97TJVBUhaYul95T2Yxlfwh5ro2pGXqDFZuLlJjDOQKM/i46mRT8LoGtCmP
		        AgIawKU8Ty3IvxwBuHCUyB95UM0qCicy09mpmeImrvC0AuqXZXS0BoIRNLpYhd4T
		        KGd4qrV/RHsXVIVcntLb9GhSTJxZlThl42E7mHNLHKzdvoxIR/7ibGH6aIdJmbhb
		        O8rMppIr1w==
		        -----END CERTIFICATE-----
		    client-key: |
		        -----BEGIN RSA PRIVATE KEY-----
		        MIIEowIBAAKCAQEA1tLkU1s3X8ui+61E850vCo/TP+g5EJSfFtnL30I7M6RQNCcc
		        +CbhK6kKpPS7zEt5QyIC8i8HeInCbRhPJ2wRGrSyzlkkD7GkMC1ktfhMVM3/DjsI
		        U7BcbJWLlI6yCypd9zrNQR8lazacXnrjPjGeXG1faXJoPQC9VnO0+iKWD34QQr8H
		        hlou9Jh+GGpZIjrfM1IYfgAl2SCGWI2NynYx8GjSTEODBVFm3LApQxG9j3rR7Umo
		        p16uJOsm6Pr7q/IPed2cgtZ4Lm+scREoI/0ASU4eEUnUw55LcJDGuggEggJmlpqF
		        +/Tn/K2uYB20ujpKLU2FWx6wuH2ixS53B1b34QIDAQABAoIBAHBYxnQZhjIhK1F3
		        4lGNaKabZR1M81sKftDSgl52IsP1MMS1l97nZmcQ9rIiE3zaE8baKLRDiCKv2PB5
		        ABxb1e4jhkeIMuXKP1W6x6qq+jB3suXcVZR+7TcUVnUQ02gndhDvvZxLD6SsYMbA
		        ecty46DuyjE5Ve5hTqPBy2ntYJEkpN6R2g/x+yVJgE8h3Fy/zP1nDLciLCXIpuEM
		        BlMBiHaF57dOdIuleyCrpO3MR2I8cUXUI3l0lCkgHPWscgMkybJtJ7u4X0sgcdrj
		        wVfHpvrgMY2qp5oVlckN/KiZ5yNRX+8jSOLupxQgHNCgpakHu/43GVmsd5YBU+yF
		        dnLHXqkCgYEA+PMLZu3sEzv6TYPbmqrqYceJjaYu4f2XcW8MU2N9plt2+TCbs95b
		        OZL1ZqDZzIyGm4NAl9ffNxJYDFYoRMbyRtx8stOKEDDZVIMwQ5/+cwjREBvz+4Gz
		        oZEiIXq0njIbXv2iW/+C+5+0T7AYbdpS/sAKmqHjialNUZ6lt2j6nosCgYEA3Oht
		        Rwp/6/QZE4zZHH+9DKvxk84C3bV3Y76YSGgjG9tYyX0CDas2H2ESkkg4aSCuyEk4
		        VgnNXzfUPBjTSND18qhKzYEbp46vr8mBnRGLJYua6cseJhC+0fpLEHTFdOIGw7ak
		        mSQzG2nXlkhtM4/HOqa8gQsUgyqMk8rByolrHMMCgYATShG6GflOzDjqxKrBYzjh
		        9qoL1bKQRCv12BrmYzEbML8ZM9D8sN/0qBRnrVLy7HiJmDPrEAj1pXA5FHvuSFQB
		        dZgb6xQpiP9t8vRMaRs4IpjAXMoc1MHsZOh2G6HfGBbS12g7JKMriAZanlRmPqJr
		        psmrjZup0Ppyto40lefFXQKBgQCTXX3goVF6zwiXcSM4jsJHnMB4MDrbOf4eDPw7
		        eTTKlYXiS8E96xQc1L311bXD86iFNcseIkXdmjm7qXfxIGyh5sCX3OPc4CO1KcCM
		        TjK75ih+hCBlllAldUnz/WHnugx3LPUar/pj9DR8LW6jsete5fHkR8b0RUMoKF8k
		        xI0uzwKBgAsvnxfQuOGs/uKIifKnVuPbmQfVCXbsoUQbpSl/nDij8HlG5FZ7X9Xd
		        QonwSBcCfXmL/46QpwgwrUF8fL1n5K+qDFgJ3m+IugKAtHfFmnzqh6IFukA0CllO
		        BghX2eJhsvKF7SGv3AKUgG0lPU0dK6eGLQKnFijAobESiDPlhXeH
		        -----END RSA PRIVATE KEY-----

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: client-cert-instance
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Error_MissingEnvVar(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: pat-env-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          env-var: MISSING_TOKEN

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-env-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "credential source not found")
}

func TestConfig_NewClientForContext_Success_WithCustomAPIPath(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: custom-api-path
		    server: https://gitlab.example.com/api/v4

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: custom-api-path
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Success_WithTokenFromValue(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com

		auths:
		  - name: pat-value-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          value: direct-token-value

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-value-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfig_NewClientForContext_Error_InvalidCertificateData(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: bad-cert-instance
		    server: https://gitlab.example.com
		    client-cert: "invalid-cert-data"
		    client-key: "invalid-key-data"

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: bad-cert-instance
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "tls: failed to find any PEM data")
}

func TestConfig_NewClientForContext_Error_MismatchedCertificateAndKey(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: mismatched-cert-instance
		    server: https://gitlab.example.com
		    client-cert: |
		        -----BEGIN CERTIFICATE-----
		        MIIDAzCCAeugAwIBAgIQYVyQi9/mh7pLS7BO/VMhZzANBgkqhkiG9w0BAQsFADAi
		        MQ8wDQYDVQQKEwZHaXRMYWIxDzANBgNVBAMTBkdpdExhYjAeFw0yNTA3MTcxMzQ5
		        MjZaFw0yNjA3MTcxMzQ5MjZaMCIxDzANBgNVBAoTBkdpdExhYjEPMA0GA1UEAxMG
		        R2l0TGFiMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1tLkU1s3X8ui
		        +61E850vCo/TP+g5EJSfFtnL30I7M6RQNCcc+CbhK6kKpPS7zEt5QyIC8i8HeInC
		        bRhPJ2wRGrSyzlkkD7GkMC1ktfhMVM3/DjsIU7BcbJWLlI6yCypd9zrNQR8lazac
		        XnrjPjGeXG1faXJoPQC9VnO0+iKWD34QQr8Hhlou9Jh+GGpZIjrfM1IYfgAl2SCG
		        WI2NynYx8GjSTEODBVFm3LApQxG9j3rR7Umop16uJOsm6Pr7q/IPed2cgtZ4Lm+s
		        cREoI/0ASU4eEUnUw55LcJDGuggEggJmlpqF+/Tn/K2uYB20ujpKLU2FWx6wuH2i
		        xS53B1b34QIDAQABozUwMzAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0lBAwwCgYIKwYB
		        BQUHAwIwDAYDVR0TAQH/BAIwADANBgkqhkiG9w0BAQsFAAOCAQEAItJBL74KO+hZ
		        Ormh9bAgrO2Rl5vR/+SYlO8Szy4FnwB0ctgaKYSdo1X5O27U6MERVa6nF+CZe3Ps
		        bFGbtBbecezlkCgtUypR4YgMNmlJX/gIDc5k+AShKUrszL3PEqYYOh+WKNf5DRaI
		        JbCbCQ97TJVBUhaYul95T2Yxlfwh5ro2pGXqDFZuLlJjDOQKM/i46mRT8LoGtCmP
		        AgIawKU8Ty3IvxwBuHCUyB95UM0qCicy09mpmeImrvC0AuqXZXS0BoIRNLpYhd4T
		        KGd4qrV/RHsXVIVcntLb9GhSTJxZlThl42E7mHNLHKzdvoxIR/7ibGH6aIdJmbhb
		        O8rMppIr1w==
		        -----END CERTIFICATE-----
		    # Using a different key that doesn't match the certificate
		    client-key: |
		        -----BEGIN RSA PRIVATE KEY-----
		        MIIEowIBAAKCAQEAuJuxWuv9fgOdfjJw81sugKOxcmvq6rCs69HCrAYhsQqe6pya
		        XGerHrY/hwmnrn/7Z/VeRNq5zwqRN6cThlIrvFaSVl3iV3++8WZr/SfT8KtoNoQF
		        dtgeq7UG0VJR2yFMuD3YW5Vcz4KWTBK3zZ3pPISXqGynNymBpG7m4chuG6YBB7T/
		        5B90VP+DO1u7c8NTfPC6WbxT0Qet8J3qy5eWCBcHMWUTLRTCDHit9bAY1Z+4bwj9
		        tQvzZoNDlJ/fi6LaRfWE8qDlVHcodyyGJirUVtFX9LSdww84zmiAIedGgZ7PzG4J
		        xj3T8StnnQqLghIHdd24yZgvScloEp7ahVYy3wIDAQABAoIBAQCLzOvsbNZU/avh
		        C5XZ1O1MYapZejw2aoEpPHpuB16wUdiy1tFWtPMzmNRXEQq47RaYTYqMHg+kKN58
		        BLyWddfFEtJTMVnc5VLWQLf6yJSJp2SOFECHFXd0lOyKzApNJdSRmdQk1uGoC76B
		        8ZLb1X/xYn/u/glLjtUsjwetaDlqGWxeCCXJf/SL2dnkU+v7S7mpd8pCO4FGHk7+
		        yUPEC72q5HZb3XIBOlYLPJY80n9cXGcIyxlxTHk1jV3lJ/PmNDJT9u4uZWKOzMf3
		        GKMhKMFzGFCZIfZF7jqn5+QY336nxNOblZQhjHe5eIUT+wr+ZcPZAYoI65FfypGR
		        DWA4fWNRAoGBAMkM8rJ5EsQx+LRymChH8GvVK60/vIXQe1KQgd2MduwXkKo9i7te
		        cXYc5osOKwWTYZC/q+Df0ZWyOF5siLq8lUYgIGD2Hpx9X1fn8DzuGTCmtbUejWix
		        MlH/HeqAI+f5c2gKw1AXrK1znvLBjR0QUrua1SgUos6wUuQguk4hnHt9AoGBAOsQ
		        TqLz6E/4CU494ENWSZRDIFeGBdpFLLqhgObFK5xmt+/dqNi3edbJIP/vXeSQBNcu
		        MS827zrAm3f3O6Xa5/EjcWlokyg6s2GTswlHhfvktlIlHI23+Vnd/+eMzW4dBtZL
		        Qpf4/GnA10SLGrre4+W9kBXuAooa/CLO5RauSJ6LAoGAIBBklHoiuA+QLpcoFSSD
		        /26b7KGBm4XIZT6Ot5qzTKvlcoEmS9egGMo7Kmo0Ckua/87RxqdrcYhe3RBKLh3t
		        YKW3BD+8WhDUp9xhwBXpBo1P5Xbd7ph0AgfB6ahOEa0C7tDonVlpPLB35RdhPgVg
		        bHMhE6dW38fXMHLXw6YworECgYAgac4+IB3/sPcvh8692k8pF5yFFSEHeRRy48RP
		        jg62cV+ZvtoCkEJHwNJBGHO9CbLxLRhxJ0UTt+14PGpIM4haMwX3gAkSug10Phap
		        B+jM1Dvj1eQ7EoxavQcFmd/V+ECyGgyjwhykRIgqlnfoHsYULvCIZZqKCrCL6DWk
		        zAGNgwKBgBQ3UqlQ7yd2HePmlAJZdRHLc5oXjA0beC6y4GurmK9lNIZQe+AKFXRS
		        kVzTChbyM4gPr9vO/OCtAhgt2B5KKtAq90qlz/vX98YLZPVTefF6VnaDRJzmKhX+
		        dvYKQUxuw8eaDzgEpSppBU61k30yEpUmm1VDEtiRmt40zZiMb0Pz
		        -----END RSA PRIVATE KEY-----

		auths:
		  - name: pat-user
		    auth_info:
		      personal-access-token:
		        token: test-token

		contexts:
		  - name: test-context
		    instance: mismatched-cert-instance
		    auth: pat-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")

	// THEN
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "tls: private key does not match public key")
}

func TestConfig_NewClientForContext_Success_CustomHeaderLiteral(t *testing.T) {
	t.Parallel()

	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
		    custom-headers:
		      - name: My-Custom-Header
		        value: my-custom-header-value

		auths:
		  - name: pat-value-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          value: direct-token-value

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-value-user
	`))
	require.NoError(t, err)

	// WHEN
	client, err := c.NewClientForContext("test-context")
	require.NoError(t, err)

	req, err := client.NewRequest(http.MethodGet, "any-path", nil, nil)
	require.NoError(t, err)

	// THEN
	assert.Equal(t, "my-custom-header-value", req.Header.Get("My-Custom-Header"))
}

func TestConfig_NewClientForContext_Success_CustomHeader_FromSource(t *testing.T) {
	// GIVEN
	c, err := NewFromString(heredoc.Doc(`
		instances:
		  - name: example
		    server: https://gitlab.example.com
		    custom-headers:
		      - name: My-Custom-Header
		        value-from:
		          env-var: MY_CUSTOM_HEADER_VALUE

		auths:
		  - name: pat-value-user
		    auth_info:
		      personal-access-token:
		        token-source:
		          value: direct-token-value

		contexts:
		  - name: test-context
		    instance: example
		    auth: pat-value-user
	`))
	require.NoError(t, err)

	t.Setenv("MY_CUSTOM_HEADER_VALUE", "my-custom-header-value")

	// WHEN
	client, err := c.NewClientForContext("test-context")
	require.NoError(t, err)

	req, err := client.NewRequest(http.MethodGet, "any-path", nil, nil)
	require.NoError(t, err)

	// THEN
	assert.Equal(t, "my-custom-header-value", req.Header.Get("My-Custom-Header"))
}
