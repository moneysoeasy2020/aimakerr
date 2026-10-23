package config

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testExtension struct {
	Browser   string `yaml:"browser"`
	Telemetry bool   `yaml:"telemetry"`
}

func TestExtensions_Unmarshal(t *testing.T) {
	t.Parallel()

	// GIVEN
	cfg, err := NewFromString(heredoc.Doc(`
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
	require.NoError(t, err)

	// WHEN
	ext := NewExtension[testExtension]("my-app", cfg)
	require.NoError(t, err)

	data, err := ext.Unmarshal()
	require.NoError(t, err)

	// THEN
	assert.Equal(t, "firefox", data.Browser)
}

func TestExtensions_Marshal(t *testing.T) {
	t.Parallel()

	// GIVEN
	cfg, err := NewFromString(heredoc.Doc(`
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
	require.NoError(t, err)
	ext := NewExtension[testExtension]("my-app", cfg)
	require.NoError(t, err)

	data, err := ext.Unmarshal()
	require.NoError(t, err)

	// WHEN
	data.Browser = "chrome"
	err = ext.Marshal()
	require.NoError(t, err)

	// THEN
	assert.Equal(t, "chrome", cfg.Extensions()["my-app"].GetFields()["browser"].GetStringValue())
}

func TestExtensions_Marshal_NewExtension(t *testing.T) {
	t.Parallel()

	// GIVEN
	cfg, err := NewFromString(heredoc.Doc(`
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
	`))
	require.NoError(t, err)
	ext := NewExtension[testExtension]("my-app", cfg)
	require.NoError(t, err)

	data, err := ext.Unmarshal()
	require.NoError(t, err)

	// WHEN
	data.Browser = "chrome"
	err = ext.Marshal()
	require.NoError(t, err)

	// THEN
	assert.Equal(t, "chrome", cfg.Extensions()["my-app"].GetFields()["browser"].GetStringValue())
}

func TestExtensions_Unmarshal_ForContext(t *testing.T) {
	t.Parallel()

	// GIVEN
	cfg, err := NewFromString(heredoc.Doc(`
		version: gitlab.com/config/v1beta1

		current-context: gitlab-com

		contexts:
		  - name: gitlab-com
		    instance: gitlab-com
		    auth: token-env

		instances:
		  - name: gitlab-com
		    server: https://gitlab.com
		    extensions:
		      my-app:
		        browser: firefox

		auths:
		  - name: token-env
		    auth-info:
		      personal-access-token:
		        token-source:
		          env_var: GITLAB_TOKEN
	`))
	require.NoError(t, err)

	// WHEN
	ext := NewExtensionForContext[testExtension]("my-app", cfg, "gitlab-com")
	require.NoError(t, err)

	data, err := ext.Unmarshal()
	require.NoError(t, err)

	// THEN
	assert.Equal(t, "firefox", data.Browser)
}

func TestExtensions_Marshal_ForContext(t *testing.T) {
	t.Parallel()

	// GIVEN
	cfg, err := NewFromString(heredoc.Doc(`
		version: gitlab.com/config/v1beta1

		current-context: gitlab-com

		contexts:
		  - name: gitlab-com
		    instance: gitlab-com
		    auth: token-env

		instances:
		  - name: gitlab-com
		    server: https://gitlab.com
		    extensions:
		      my-app:
		        browser: firefox

		auths:
		  - name: token-env
		    auth-info:
		      personal-access-token:
		        token-source:
		          env_var: GITLAB_TOKEN
	`))
	require.NoError(t, err)
	ext := NewExtensionForContext[testExtension]("my-app", cfg, "gitlab-com")
	require.NoError(t, err)

	data, err := ext.Unmarshal()
	require.NoError(t, err)

	// WHEN
	data.Browser = "chrome"
	err = ext.Marshal()
	require.NoError(t, err)

	// THEN
	assert.Equal(t, "chrome", cfg.Instance("gitlab-com").Extensions["my-app"].GetFields()["browser"].GetStringValue())
}
