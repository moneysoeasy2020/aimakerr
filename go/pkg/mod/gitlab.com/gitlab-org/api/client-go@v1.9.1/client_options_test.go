package gitlab

import (
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithBaseURL(t *testing.T) {
	t.Parallel()

	// GIVEN a new GitLab client
	client, err := NewClient("")
	require.NoError(t, err)
	customURL := "https://example.com/api/v4/"

	// WHEN the WithBaseURL option is applied
	opt := WithBaseURL(customURL)
	err = opt(client)
	require.NoError(t, err)

	// THEN the client's base URL is updated
	assert.Equal(t, customURL, client.BaseURL().String())
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	// GIVEN a new GitLab client
	client, err := NewClient("")
	require.NoError(t, err)
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// WHEN the WithHTTPClient option is applied
	opt := WithHTTPClient(customClient)
	err = opt(client)
	require.NoError(t, err)

	// THEN the client's internal HTTP client is replaced
	assert.Same(t, customClient, client.client.HTTPClient)
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	// GIVEN a new GitLab client
	client, err := NewClient("")
	require.NoError(t, err)
	customUA := "my-custom-agent/1.0"

	// WHEN the WithUserAgent option is applied
	opt := WithUserAgent(customUA)
	err = opt(client)
	require.NoError(t, err)

	// THEN the client's UserAgent string is updated
	assert.Equal(t, customUA, client.UserAgent)
}

func TestWithoutRetries(t *testing.T) {
	t.Parallel()

	// GIVEN a new GitLab client with retries enabled by default
	client, err := NewClient("")
	require.NoError(t, err)
	assert.False(t, client.disableRetries, "Retries should be enabled by default")

	// WHEN the WithoutRetries option is applied
	opt := WithoutRetries()
	err = opt(client)
	require.NoError(t, err)

	// THEN the client's retry mechanism is disabled
	assert.True(t, client.disableRetries)
}

func TestWithCustomRetryMax(t *testing.T) {
	t.Parallel()

	// GIVEN a new GitLab client
	client, err := NewClient("")
	require.NoError(t, err)

	// WHEN the WithCustomRetryMax option is applied
	opt := WithCustomRetryMax(10)
	err = opt(client)
	require.NoError(t, err)

	// THEN the client's maximum retry count is updated
	assert.Equal(t, 10, client.client.RetryMax)
}

func TestWithCustomRetryWaitMinMax(t *testing.T) {
	t.Parallel()

	// GIVEN a new GitLab client
	client, err := NewClient("")
	require.NoError(t, err)
	minWait := 1 * time.Second
	maxWait := 30 * time.Second

	// WHEN the WithCustomRetryWaitMinMax option is applied
	opt := WithCustomRetryWaitMinMax(minWait, maxWait)
	err = opt(client)
	require.NoError(t, err)

	// THEN the client's retry wait times are updated
	assert.Equal(t, minWait, client.client.RetryWaitMin)
	assert.Equal(t, maxWait, client.client.RetryWaitMax)
}

func TestWithRequestOptions(t *testing.T) {
	t.Parallel()

	// GIVEN a new GitLab client and a default request option
	mux, client := setup(t)
	var testHeader string
	ro := func(req *retryablehttp.Request) error {
		testHeader = "was-set"
		return nil
	}

	// WHEN the WithRequestOptions option is applied and a request is made
	opt := WithRequestOptions(ro)
	err := opt(client)
	require.NoError(t, err)
	require.Len(t, client.defaultRequestOptions, 1)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, err := client.NewRequest(http.MethodGet, "/", nil, nil)
	require.NoError(t, err)
	_, err = client.Do(req, nil)
	require.NoError(t, err)

	// THEN the default request option is executed
	assert.Equal(t, "was-set", testHeader)
}

func TestClientWithInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("should add valid interceptor", func(t *testing.T) {
		t.Parallel()

		// GIVEN a new GitLab client and a valid interceptor
		client, err := NewClient("")
		require.NoError(t, err)
		ic := func(req http.RoundTripper) http.RoundTripper {
			return nil
		}

		// WHEN the WithInterceptor option is applied
		opt := WithInterceptor(ic)
		err = opt(client)
		require.NoError(t, err)

		// THEN the interceptor is added to the client
		assert.Len(t, client.interceptors, 1)
	})

	t.Run("should return error for nil interceptor", func(t *testing.T) {
		t.Parallel()

		// GIVEN a new GitLab client
		client, err := NewClient("")
		require.NoError(t, err)

		// WHEN the WithInterceptor option is applied with a nil value
		optNil := WithInterceptor(nil)
		err = optNil(client)

		// THEN an error is returned and no interceptor is added
		require.Error(t, err)
		assert.Equal(t, "interceptor cannot be nil", err.Error())
		assert.Empty(t, client.interceptors, "Nil interceptor should not be added")
	})
}
