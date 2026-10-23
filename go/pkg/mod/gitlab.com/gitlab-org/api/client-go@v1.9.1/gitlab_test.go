//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
)

var timeLayout = "2006-01-02T15:04:05Z07:00"

// Interface implementation checks.
var (
	_ AuthSource = OAuthTokenSource{}
	_ AuthSource = JobTokenAuthSource{}
	_ AuthSource = AccessTokenAuthSource{}
	_ AuthSource = (*PasswordCredentialsAuthSource)(nil)
)

// setup sets up a test HTTP server along with a gitlab.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup(t *testing.T) (*http.ServeMux, *Client) {
	// mux is the HTTP request multiplexer used with the test server.
	mux := http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// client is the GitLab client being tested.
	client, err := NewClient("",
		WithBaseURL(server.URL),
		// Disable backoff to speed up tests that expect errors.
		WithCustomBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
			return 0
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return mux, client
}

func testURL(t *testing.T, r *http.Request, want string) {
	if got := r.RequestURI; got != want {
		t.Errorf("Request url: %+v, want %s", got, want)
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %s, want %s", got, want)
	}
}

// Tests that a given form attribute has a value in a form request. Useful
// for testing file upload API requests.
func testFormBody(t *testing.T, r *http.Request, key string, want string) {
	if got := r.FormValue(key); got != want {
		t.Errorf("Request body for key %s got: %s, want %s", key, got, want)
	}
}

// testBodyJSON tests that the JSON request body is what we expect. The want
// argument is typically either a struct, a map[string]string, or a
// map[string]any, though other types are handled as well.
//
// Calls t.Fatal if decoding the request body fails, failing the test
// immediately.
//
// When the request body is not equal to "want", the error is reported but the
// test is allowed to continue. You can use the return value to end the test on
// error: returns true if the decoded body is identical to want, false
// otherwise.
func testBodyJSON[T any](t *testing.T, r *http.Request, want T) bool {
	var got T

	if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode JSON from request body: %v", err)
	}

	return assert.Equal(t, want, got)
}

// testParam checks whether the given request contains the expected parameter and whether the parameter has the expected value.
func testParam(t *testing.T, r *http.Request, key, value string) {
	require.True(t, r.URL.Query().Has(key), "Request does not contain the %q parameter", key)
	assert.Len(t, r.URL.Query()[key], 1, "Request contains multiple %q parameters when only one is expected", key)
	require.Equal(t, value, r.URL.Query().Get(key))
}

func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixture file: %v", err)
	}
	defer f.Close()

	if _, err = io.Copy(w, f); err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}

// mustWriteJSONResponse writes a JSON response to w.
// It uses t.Fatal to stop the test and report an error if encoding the response fails.
// This helper is useful when implementing handlers in unit tests.
func mustWriteJSONResponse(t *testing.T, w io.Writer, response any) {
	if err := json.NewEncoder(w).Encode(response); err != nil {
		t.Fatalf("Failed to write response: %v", err)
	}
}

// mustWriteErrorResponse writes an error response to w in a format that CheckResponse can parse.
// It uses t.Fatal to stop the test and report an error if encoding the response fails.
// This is useful when testing error conditions.
func mustWriteErrorResponse(t *testing.T, w io.Writer, err error) {
	mustWriteJSONResponse(t, w, map[string]any{
		"error": err.Error(),
	})
}

var errRequestOptionFunc = errors.New("RequestOptionFunc returns an error")

func errorOption(*retryablehttp.Request) error {
	return errRequestOptionFunc
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("Default Configuration", func(t *testing.T) {
		t.Parallel()
		c, err := NewClient("")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		expectedBaseURL := defaultBaseURL + apiVersionPath

		if c.BaseURL().String() != expectedBaseURL {
			t.Errorf("NewClient BaseURL is %s, want %s", c.BaseURL().String(), expectedBaseURL)
		}
		if c.UserAgent != userAgent {
			t.Errorf("NewClient UserAgent is %s, want %s", c.UserAgent, userAgent)
		}
	})

	t.Run("Custom UserAgent", func(t *testing.T) {
		t.Parallel()
		c, err := NewClient("", WithUserAgent("any-custom-user-agent"))
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		expectedBaseURL := defaultBaseURL + apiVersionPath

		if c.BaseURL().String() != expectedBaseURL {
			t.Errorf("NewClient BaseURL is %s, want %s", c.BaseURL().String(), expectedBaseURL)
		}
		if c.UserAgent != "any-custom-user-agent" {
			t.Errorf("NewClient UserAgent is %s, want any-custom-user-agent", c.UserAgent)
		}
	})

	t.Run("Custom Base URL", func(t *testing.T) {
		t.Parallel()
		customURL := "https://custom.gitlab.com/api/v4"
		c, err := NewClient("", WithBaseURL(customURL))
		require.NoError(t, err)
		require.NotNil(t, c, "Client is nil")

		// The client will append a trailing slash to the base URL
		expectedURL := customURL + "/"
		require.Equal(t, expectedURL, c.BaseURL().String(), "BaseURL mismatch")
	})

	t.Run("Invalid Base URL", func(t *testing.T) {
		t.Parallel()
		_, err := NewClient("", WithBaseURL(":invalid:"))
		require.Error(t, err)
	})
}

func TestSendingUserAgent_Default(t *testing.T) {
	t.Parallel()

	c, err := NewClient("")
	require.NoError(t, err)

	req, err := c.NewRequest(http.MethodGet, "test", nil, nil)
	require.NoError(t, err)

	assert.Equal(t, userAgent, req.Header.Get("User-Agent"))
}

func TestSendingUserAgent_Custom(t *testing.T) {
	t.Parallel()

	c, err := NewClient("", WithUserAgent("any-custom-user-agent"))
	require.NoError(t, err)

	req, err := c.NewRequest(http.MethodGet, "test", nil, nil)
	require.NoError(t, err)

	assert.Equal(t, "any-custom-user-agent", req.Header.Get("User-Agent"))
}

func TestCheckResponse(t *testing.T) {
	t.Parallel()
	c, err := NewClient("")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req, err := c.NewRequest(http.MethodGet, "test", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp := &http.Response{
		Request:    req.Request,
		StatusCode: http.StatusBadRequest,
		Body: io.NopCloser(strings.NewReader(`
		{
			"message": {
				"prop1": [
					"message 1",
					"message 2"
				],
				"prop2":[
					"message 3"
				],
				"embed1": {
					"prop3": [
						"msg 1",
						"msg2"
					]
				},
				"embed2": {
					"prop4": [
						"some msg"
					]
				}
			},
			"error": "message 1"
		}`)),
	}

	errResp := CheckResponse(resp)
	if errResp == nil {
		t.Fatal("Expected error response.")
	}

	want := "GET https://gitlab.com/api/v4/test: 400 {error: message 1}, {message: {embed1: {prop3: [msg 1, msg2]}}, {embed2: {prop4: [some msg]}}, {prop1: [message 1, message 2]}, {prop2: [message 3]}}"

	if errResp.Error() != want {
		t.Errorf("Expected error: %s, got %s", want, errResp.Error())
	}
}

func TestCheckResponseOnUnknownErrorFormat(t *testing.T) {
	t.Parallel()
	c, err := NewClient("")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req, err := c.NewRequest(http.MethodGet, "test", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp := &http.Response{
		Request:    req.Request,
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("some error message but not JSON")),
	}

	errResp := CheckResponse(resp)
	if errResp == nil {
		t.Fatal("Expected error response.")
	}

	want := "GET https://gitlab.com/api/v4/test: 400 failed to parse unknown error format: some error message but not JSON"

	if errResp.Error() != want {
		t.Errorf("Expected error: %s, got %s", want, errResp.Error())
	}
}

func TestCheckResponseOnHeadRequestError(t *testing.T) {
	t.Parallel()
	c, err := NewClient("")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req, err := c.NewRequest(http.MethodHead, "test", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp := &http.Response{
		Request:    req.Request,
		StatusCode: http.StatusNotFound,
		Body:       nil,
	}

	errResp := CheckResponse(resp)
	if errResp == nil {
		t.Fatal("Expected error response.")
	}

	want := "404 Not Found"

	if errResp.Error() != want {
		t.Errorf("Expected error: %s, got %s", want, errResp.Error())
	}
}

func TestRequestWithContext(t *testing.T) {
	t.Parallel()
	c, err := NewClient("")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	req, err := c.NewRequest(http.MethodGet, "test", nil, []RequestOptionFunc{WithContext(ctx)})
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	defer cancel()

	if req.Context() != ctx {
		t.Fatal("Context was not set correctly")
	}
}

func loadFixture(t *testing.T, filePath string) []byte {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	return content
}

func TestPathEscape(t *testing.T) {
	t.Parallel()
	want := "diaspora%2Fdiaspora"
	got := PathEscape("diaspora/diaspora")
	if want != got {
		t.Errorf("Expected: %s, got %s", want, got)
	}
}

func TestPaginationPopulatePageValuesEmpty(t *testing.T) {
	t.Parallel()
	wantPageHeaders := map[string]int64{
		xTotal:      0,
		xTotalPages: 0,
		xPerPage:    0,
		xPage:       0,
		xNextPage:   0,
		xPrevPage:   0,
	}
	wantLinkHeaders := map[string]string{
		linkPrev:  "",
		linkNext:  "",
		linkFirst: "",
		linkLast:  "",
	}

	r := newResponse(&http.Response{
		Header: http.Header{},
	})

	gotPageHeaders := map[string]int64{
		xTotal:      r.TotalItems,
		xTotalPages: r.TotalPages,
		xPerPage:    r.ItemsPerPage,
		xPage:       r.CurrentPage,
		xNextPage:   r.NextPage,
		xPrevPage:   r.PreviousPage,
	}
	for k, v := range wantPageHeaders {
		if v != gotPageHeaders[k] {
			t.Errorf("For %s, expected %d, got %d", k, v, gotPageHeaders[k])
		}
	}

	gotLinkHeaders := map[string]string{
		linkPrev:  r.PreviousLink,
		linkNext:  r.NextLink,
		linkFirst: r.FirstLink,
		linkLast:  r.LastLink,
	}
	for k, v := range wantLinkHeaders {
		if v != gotLinkHeaders[k] {
			t.Errorf("For %s, expected %s, got %s", k, v, gotLinkHeaders[k])
		}
	}
}

func TestPaginationPopulatePageValuesOffset(t *testing.T) {
	t.Parallel()
	wantPageHeaders := map[string]int64{
		xTotal:      100,
		xTotalPages: 5,
		xPerPage:    20,
		xPage:       2,
		xNextPage:   3,
		xPrevPage:   1,
	}
	wantLinkHeaders := map[string]string{
		linkPrev:  "https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=1&per_page=3",
		linkNext:  "https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=3&per_page=3",
		linkFirst: "https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=1&per_page=3",
		linkLast:  "https://gitlab.example.com/api/v4/projects/8/issues/8/notes?page=3&per_page=3",
	}

	h := http.Header{}
	for k, v := range wantPageHeaders {
		h.Add(k, fmt.Sprint(v))
	}
	var linkHeaderComponents []string
	for k, v := range wantLinkHeaders {
		if v != "" {
			linkHeaderComponents = append(linkHeaderComponents, fmt.Sprintf("<%s>; rel=\"%s\"", v, k))
		}
	}
	h.Add("Link", strings.Join(linkHeaderComponents, ", "))

	r := newResponse(&http.Response{
		Header: h,
	})

	gotPageHeaders := map[string]int64{
		xTotal:      r.TotalItems,
		xTotalPages: r.TotalPages,
		xPerPage:    r.ItemsPerPage,
		xPage:       r.CurrentPage,
		xNextPage:   r.NextPage,
		xPrevPage:   r.PreviousPage,
	}
	for k, v := range wantPageHeaders {
		if v != gotPageHeaders[k] {
			t.Errorf("For %s, expected %d, got %d", k, v, gotPageHeaders[k])
		}
	}

	gotLinkHeaders := map[string]string{
		linkPrev:  r.PreviousLink,
		linkNext:  r.NextLink,
		linkFirst: r.FirstLink,
		linkLast:  r.LastLink,
	}
	for k, v := range wantLinkHeaders {
		if v != gotLinkHeaders[k] {
			t.Errorf("For %s, expected %s, got %s", k, v, gotLinkHeaders[k])
		}
	}
}

func TestPaginationPopulatePageValuesKeyset(t *testing.T) {
	t.Parallel()
	wantPageHeaders := map[string]int64{
		xTotal:      0,
		xTotalPages: 0,
		xPerPage:    0,
		xPage:       0,
		xNextPage:   0,
		xPrevPage:   0,
	}
	wantLinkHeaders := map[string]string{
		linkPrev:  "",
		linkFirst: "",
		linkLast:  "",
	}

	h := http.Header{}
	for k, v := range wantPageHeaders {
		h.Add(k, fmt.Sprint(v))
	}
	var linkHeaderComponents []string
	for k, v := range wantLinkHeaders {
		if v != "" {
			linkHeaderComponents = append(linkHeaderComponents, fmt.Sprintf("<%s>; rel=\"%s\"", v, k))
		}
	}
	h.Add("Link", strings.Join(linkHeaderComponents, ", "))

	r := newResponse(&http.Response{
		Header: h,
	})

	gotPageHeaders := map[string]int64{
		xTotal:      r.TotalItems,
		xTotalPages: r.TotalPages,
		xPerPage:    r.ItemsPerPage,
		xPage:       r.CurrentPage,
		xNextPage:   r.NextPage,
		xPrevPage:   r.PreviousPage,
	}
	for k, v := range wantPageHeaders {
		if v != gotPageHeaders[k] {
			t.Errorf("For %s, expected %d, got %d", k, v, gotPageHeaders[k])
		}
	}
}

func TestNewRetryableHTTPClientWithRetryCheck(t *testing.T) {
	t.Parallel()

	_, client := setup(t)

	httpClient := &http.Client{}
	logger := struct{}{}
	retryWaitMin := 10 * time.Second
	retryWaitMax := 20 * time.Second
	retryMax := 30
	requestLogHook := retryablehttp.RequestLogHook(func(logger retryablehttp.Logger, request *http.Request, i int) {
	})
	checkRetry := retryablehttp.CheckRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		return false, nil
	})
	backoff := retryablehttp.Backoff(func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		return time.Second
	})
	errorHandler := retryablehttp.ErrorHandler(func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		return nil, nil
	})
	prepareRetry := retryablehttp.PrepareRetry(func(req *http.Request) error {
		return nil
	})

	newCheckRetry := retryablehttp.CheckRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		return false, nil
	})

	client.client.HTTPClient = httpClient
	client.client.Logger = logger
	client.client.RetryWaitMin = retryWaitMin
	client.client.RetryWaitMax = retryWaitMax
	client.client.RetryMax = retryMax
	client.client.RequestLogHook = requestLogHook
	client.client.CheckRetry = checkRetry
	client.client.Backoff = backoff
	client.client.ErrorHandler = errorHandler
	client.client.PrepareRetry = prepareRetry

	actual := client.newRetryableHTTPClientWithRetryCheck(newCheckRetry)

	assert.Equal(t, httpClient, actual.HTTPClient)
	assert.Equal(t, logger, actual.Logger)
	assert.Equal(t, retryWaitMin, actual.RetryWaitMin)
	assert.Equal(t, retryWaitMax, actual.RetryWaitMax)
	assert.Equal(t, retryMax, actual.RetryMax)
	assert.Equal(t, reflect.ValueOf(requestLogHook).Pointer(), reflect.ValueOf(actual.RequestLogHook).Pointer())
	assert.Equal(t, reflect.ValueOf(newCheckRetry).Pointer(), reflect.ValueOf(actual.CheckRetry).Pointer())
	assert.Equal(t, reflect.ValueOf(backoff).Pointer(), reflect.ValueOf(actual.Backoff).Pointer())
	assert.Equal(t, reflect.ValueOf(errorHandler).Pointer(), reflect.ValueOf(actual.ErrorHandler).Pointer())
	assert.Equal(t, reflect.ValueOf(prepareRetry).Pointer(), reflect.ValueOf(actual.PrepareRetry).Pointer())
}

func TestExponentialBackoffLogic(t *testing.T) {
	t.Parallel()
	// Can't use the default `setup` because it disabled the backoff
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client, err := NewClient("",
		WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a method that returns 429
	mux.HandleFunc("/api/v4/projects/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusTooManyRequests)
	})

	// Measure the time at the start of the test
	start := time.Now()

	// Send a request (which will get a bunch of 429s)
	// None of the responses matter, so ignore them all
	_, resp, _ := client.Projects.GetProject(1, nil)
	end := time.Now()

	// The test should run for _at least_ 3,200 milliseconds
	duration := float64(end.Sub(start))
	if duration < float64(3200*time.Millisecond) {
		t.Fatal("Wait was shorter than expected. Expected a minimum of 5 retries taking 3200 milliseconds, got:", duration)
	}
	if resp.StatusCode != 429 {
		t.Fatal("Expected to get a 429 code given the server is hard-coded to return this. Received instead:", resp.StatusCode)
	}
}

func TestErrorResponsePreservesURLEncoding(t *testing.T) {
	t.Parallel()

	projectID := "group/subgroup"
	fileName := "path/file.txt"

	expectedEscapedPath := "/api/v4/projects/group%2Fsubgroup/repository/files/path%2Ffile%2Etxt"

	escapedProjectID := PathEscape(projectID)
	escapedFileName := PathEscape(fileName)
	escapedPath := fmt.Sprintf("/api/v4/projects/%s/repository/files/%s",
		escapedProjectID, escapedFileName)

	require.Equal(t, expectedEscapedPath, escapedPath)

	fullURL := "https://gitlab.com" + expectedEscapedPath
	req, _ := http.NewRequest("GET", fullURL, nil)
	resp := &http.Response{
		Request:    req,
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"message":"Not Found"}`)),
	}

	errorResponse := &ErrorResponse{Response: resp, Message: "Not Found"}

	require.ErrorContains(t, errorResponse, expectedEscapedPath)

	unescapedPath := fmt.Sprintf("/api/v4/projects/%s/repository/files/%s", projectID, fileName)
	assert.NotContains(t, errorResponse.Error(), unescapedPath)
}

func TestNewClient_auth(t *testing.T) {
	t.Parallel()

	const token = "glpat-0123456789abcdefg"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Private-Token"), token; got != want {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Authorization = %q, want %q", got, want)
			return
		}

		fmt.Fprint(w, "[]")
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)

	client, err := NewClient(token,
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	projects, resp, err := client.Projects.ListProjects(&ListProjectsOptions{})
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []*Project{}, projects)
}

func TestNewJobClient_auth(t *testing.T) {
	t.Parallel()

	const token = "glcbt-0123456789abcdefg"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Job-Token"), token; got != want {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Authorization = %q, want %q", got, want)
			return
		}

		fmt.Fprint(w, "[]")
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)

	client, err := NewJobClient(token,
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	projects, resp, err := client.Projects.ListProjects(&ListProjectsOptions{})
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []*Project{}, projects)
}

func TestNewBasicAuthClient_auth(t *testing.T) {
	t.Parallel()

	const (
		username = "test-username"
		password = "test-p4ssw0rd"
		token    = "test-token"
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "r.ParseForm: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if got, want := r.Form.Get("grant_type"), "password"; got != want {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "grant_type = %q, want %q", got, want)
			return
		}

		if gotUsername, gotPassword := r.Form.Get("username"), r.Form.Get("password"); gotUsername != username || gotPassword != password {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "username is %q, want %q", gotUsername, username)
			fmt.Fprintf(w, "password is %q, want %q", gotPassword, password)
			return
		}

		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		fmt.Fprint(w, url.Values{
			"access_token": {token},
			"token_type":   {"bearer"},
			"expires_in":   {"1800"},
		}.Encode())
	})
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Authorization"), "Bearer "+token; got != want {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Authorization = %q, want %q", got, want)
			return
		}

		fmt.Fprint(w, "[]")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unexpected %s request to %s", r.Method, r.URL.String())
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewBasicAuthClient(username, password,
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	projects, resp, err := client.Projects.ListProjects(&ListProjectsOptions{})
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []*Project{}, projects)
}

func TestNewAuthSourceClient(t *testing.T) {
	t.Parallel()

	token := &oauth2.Token{
		AccessToken: "0123456789abcdefg",
	}
	ts := oauth2.StaticTokenSource(token)

	handler := func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Authorization"), "Bearer 0123456789abcdefg"; got != want {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Authorization = %q, want %q", got, want)
			return
		}

		fmt.Fprint(w, "[]")
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)

	client, err := NewAuthSourceClient(OAuthTokenSource{ts},
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	projects, resp, err := client.Projects.ListProjects(&ListProjectsOptions{})
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []*Project{}, projects)
}

func TestHasStatusCode(t *testing.T) {
	t.Parallel()

	// GIVEN
	tests := []struct {
		name          string
		err           error
		hasStatusCode int
		expect        bool
	}{
		{
			name:          "error is nil",
			err:           nil,
			hasStatusCode: http.StatusOK,
			expect:        false,
		},
		{
			name:          "error is not a ErrorResponse",
			err:           errors.New("dummy"),
			hasStatusCode: http.StatusOK,
			expect:        false,
		},
		{
			name:          "error is a ErrorResponse, but has no http.Response",
			err:           &ErrorResponse{},
			hasStatusCode: http.StatusOK,
			expect:        false,
		},
		{
			name:          "error has different status code",
			err:           &ErrorResponse{Response: &http.Response{StatusCode: http.StatusBadRequest}},
			hasStatusCode: http.StatusOK,
			expect:        false,
		},
		{
			name:          "error has expected status code",
			err:           &ErrorResponse{Response: &http.Response{StatusCode: http.StatusOK}},
			hasStatusCode: http.StatusOK,
			expect:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// WHEN
			actual := HasStatusCode(tt.err, tt.hasStatusCode)

			// THEN
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestNewRequestToURL_disallowedURL(t *testing.T) {
	t.Parallel()

	// GIVEN
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "wrong scheme",
			url:  "http://gitlab.example.com",
		},
		{
			name: "wrong hostname",
			url:  "https://gitlab2.example.com",
		},
		{
			name: "wrong port",
			url:  "https://gitlab.example.com:8080",
		},
	}
	c, err := NewClient("",
		WithBaseURL("https://gitlab.example.com"),
	)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u, err := url.Parse(tt.url)
			require.NoError(t, err)

			// WHEN
			_, err = c.NewRequestToURL(http.MethodGet, u, nil, nil)
			assert.Error(t, err)
		})
	}
}

func TestNewRequestToURL_allowedURL(t *testing.T) {
	t.Parallel()

	// GIVEN
	tests := []struct {
		url string
	}{
		{
			url: "https://gitlab.example.com",
		},
		{
			url: "https://gitlab.example.com/api/v4",
		},
	}
	c, err := NewClient("",
		WithBaseURL("https://gitlab.example.com"),
	)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()

			u, err := url.Parse(tt.url)
			require.NoError(t, err)

			// WHEN
			_, err = c.NewRequestToURL(http.MethodGet, u, nil, nil)
			assert.NoError(t, err)
		})
	}
}

func TestClient_CookieJar(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/user", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("test-cookie")
		assert.NoError(t, err)

		assert.Equal(t, "yummy", cookie.Value)

		http.SetCookie(w, &http.Cookie{
			Name:  "test-session-cookie",
			Value: "another-yummy",
			Path:  "/",
		})
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{}`)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	jar.SetCookies(u, []*http.Cookie{
		{
			Name:  "test-cookie",
			Value: "yummy",
			Path:  "/",
		},
	})

	client, err := NewClient("", WithBaseURL(server.URL), WithHTTPClient(server.Client()), WithCookieJar(jar))
	require.NoError(t, err)

	_, resp, err := client.Users.CurrentUser()
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cookiesInJar := client.HTTPClient().Jar.Cookies(u)
	cookieMap := make(map[string]string, len(cookiesInJar))
	for _, c := range cookiesInJar {
		cookieMap[c.Name] = c.Value
	}
	assert.Equal(t, "another-yummy", cookieMap["test-session-cookie"])
}

func TestWithInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("when nil interceptor has been passed, then it will result in an error", func(t *testing.T) {
		t.Parallel()

		_, err := NewClient("", WithInterceptor(nil))
		require.Error(t, err)
	})

	t.Run("when interceptor option is provided, then it is used in the client as part of the http round tripping of the transportation", func(t *testing.T) {
		t.Parallel()

		client, err := NewClient("",
			WithInterceptor(func(next http.RoundTripper) http.RoundTripper {
				assert.NotNil(t, next, "it was expected that the next middleware is not empty, most likely being the default transport in worse case scenario")
				return StubRoundTripper(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}"))}, nil
				})
			}),
		)
		require.NoError(t, err)

		_, resp, err := client.Users.CurrentUser()
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("enables request manipulation", func(t *testing.T) {
		t.Parallel()

		client, err := NewClient("",
			WithInterceptor(func(next http.RoundTripper) http.RoundTripper {
				assert.NotNil(t, next, "it was expected that the next middleware is not empty, most likely being the default transport in worse case scenario")
				return StubRoundTripper(func(r *http.Request) (*http.Response, error) {
					assert.Equal(t, "foo", r.Header.Get("X-Foo"))
					respHeaders := http.Header{}
					respHeaders.Set("X-Bar", "bar")
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader("{}")),
						Header:     respHeaders,
					}, nil
				})
			}),
		)
		require.NoError(t, err)

		_, resp, err := client.Users.CurrentUser(func(r *retryablehttp.Request) error {
			r.Request.Header.Set("X-Foo", "foo")
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "bar", resp.Header.Get("X-Bar"))
	})

	t.Run("ordering aligned to how interceptors are provided, as this makes it easier to read a option setup", func(t *testing.T) {
		t.Parallel()

		var ordering []int64
		client, err := NewClient("",
			WithInterceptor(func(next http.RoundTripper) http.RoundTripper {
				assert.NotNil(t, next)
				return StubRoundTripper(func(r *http.Request) (*http.Response, error) {
					ordering = append(ordering, 1)
					return next.RoundTrip(r)
				})
			}),
			WithInterceptor(func(next http.RoundTripper) http.RoundTripper {
				assert.NotNil(t, next)
				return StubRoundTripper(func(r *http.Request) (*http.Response, error) {
					ordering = append(ordering, 2)
					return next.RoundTrip(r)
				})
			}),
			WithInterceptor(func(next http.RoundTripper) http.RoundTripper {
				assert.NotNil(t, next)
				return StubRoundTripper(func(r *http.Request) (*http.Response, error) {
					ordering = append(ordering, 3)
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}"))}, nil
				})
			}),
		)
		require.NoError(t, err)

		_, _, _ = client.Users.CurrentUser()
		assert.Equal(t, []int64{1, 2, 3}, ordering)
	})

	t.Run("e2e", func(t *testing.T) {
		t.Parallel()

		const endpoint = "/api/v4/user"

		mux := http.NewServeMux()
		mux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Foo", "bar")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{}`)
		})

		server := httptest.NewServer(mux)
		t.Cleanup(server.Close)

		client, err := NewClient("",
			WithBaseURL(server.URL),
			WithHTTPClient(server.Client()),
			WithInterceptor(func(next http.RoundTripper) http.RoundTripper {
				return StubRoundTripper(func(r *http.Request) (*http.Response, error) {
					assert.Contains(t, r.URL.Path, endpoint)
					resp, err := next.RoundTrip(r)
					if err == nil {
						assert.Equal(t, "bar", resp.Header.Get("X-Foo"))
					}
					return resp, err
				})
			}))

		require.NoError(t, err)

		_, resp, err := client.Users.CurrentUser()
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

type StubRoundTripper func(r *http.Request) (*http.Response, error)

func (fn StubRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func TestClient_DefaultRetryPolicy_RetryOnStatusCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		statusCode    int
		expectedRetry bool
		expectedError bool
	}{
		// We can't write an invalid status code. See RetryOnZeroStatusCode test
		// {
		// 	statusCode:    0,
		// 	expectedRetry: true,
		// },
		{
			statusCode:    http.StatusOK,
			expectedRetry: false,
			expectedError: false,
		},
		{
			statusCode:    http.StatusNotImplemented,
			expectedRetry: false,
			expectedError: true,
		},
		{
			statusCode:    http.StatusBadRequest,
			expectedRetry: false,
			expectedError: true,
		},
		{
			statusCode:    http.StatusTooManyRequests,
			expectedRetry: true,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.statusCode), func(t *testing.T) {
			t.Parallel()

			// GIVEN
			retries := 0
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v4/user", func(w http.ResponseWriter, r *http.Request) {
				retries++

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{}`))
			})
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			client, err := NewClient(
				"",
				WithBaseURL(server.URL),
				WithHTTPClient(server.Client()),
				WithCustomRetryMax(1),
			)
			require.NoError(t, err)

			// WHEN
			_, resp, err := client.Users.CurrentUser()

			// THEN
			if tt.expectedError {
				require.Error(t, err)
				assert.True(t, HasStatusCode(err, tt.statusCode))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.statusCode, resp.StatusCode)
			}

			if tt.expectedRetry {
				assert.Equal(t, 2, retries, "Expected 2 retries, got %d", retries)
			} else {
				assert.Equal(t, 1, retries, "Didn't expect a retry to happen, but endpoint counter indicates that the request has been retried")
			}
		})
	}
}

func TestClient_DefaultRetryPolicy_RetryOnIdempotentRequests_ByMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method        string
		expectedRetry bool
	}{
		{
			method:        http.MethodGet,
			expectedRetry: true,
		},
		{
			method:        http.MethodPost,
			expectedRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			client, err := NewClient(
				"",
				WithCustomRetryMax(1),
			)
			require.NoError(t, err)

			// WHEN
			retry, err := client.retryHTTPCheck(context.Background(), &http.Response{Request: &http.Request{Method: tt.method}}, errors.New("dummy"))

			// THEN
			assert.Equal(t, tt.expectedRetry, retry)
			assert.NoError(t, err)
		})
	}
}

// TestClient_DefaultRetryPolicy_RetryOnZeroStatusCode tests retry for status code 0
//
// We test for this bogus status code because of AWS ALB, see:
// https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-troubleshooting.html#response-code-000
func TestClient_DefaultRetryPolicy_RetryOnZeroStatusCode(t *testing.T) {
	t.Parallel()

	// GIVEN
	client, err := NewClient(
		"",
	)
	require.NoError(t, err)

	// WHEN
	retry, err := client.retryHTTPCheck(context.Background(), &http.Response{StatusCode: 0}, nil)

	// THEN
	assert.True(t, retry)
	assert.NoError(t, err)
}

func TestClient_DefaultRetryPolicy_RetryOnNetworkErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		err           error
		expectedRetry bool
	}{
		{
			name:          "DNS error - NXDOMAIN should not retry",
			err:           &net.DNSError{Err: "no such host", IsNotFound: true},
			expectedRetry: false,
		},
		{
			name:          "DNS error - temporary DNS error should retry",
			err:           &net.DNSError{Err: "temporary failure", IsTimeout: true},
			expectedRetry: true,
		},
		{
			name:          "DNS error - other DNS error should retry",
			err:           &net.DNSError{Err: "server failure"},
			expectedRetry: true,
		},
		{
			name:          "OpError - temporary error should retry",
			err:           &net.OpError{Op: "read", Net: "tcp", Err: &mockTemporaryError{temporary: true}},
			expectedRetry: true,
		},
		{
			name:          "OpError - dial operation should retry",
			err:           &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection failed")},
			expectedRetry: true,
		},
		{
			name:          "OpError - non-temporary, non-dial should not retry",
			err:           &net.OpError{Op: "write", Net: "tcp", Err: errors.New("broken pipe")},
			expectedRetry: false,
		},
		{
			name:          "URL error - connection refused should retry",
			err:           &url.Error{Op: "Get", URL: "http://example.com", Err: errors.New("connection refused")},
			expectedRetry: true,
		},
		{
			name:          "URL error - other URL error should not retry",
			err:           &url.Error{Op: "Get", URL: "http://example.com", Err: errors.New("other error")},
			expectedRetry: false,
		},
		{
			name:          "TLS handshake timeout should retry",
			err:           &mockTLSHandshakeError{msg: "net/http: TLS handshake timeout"},
			expectedRetry: true,
		},
		{
			name:          "Other TLS error should not retry",
			err:           &mockTLSHandshakeError{msg: "tls: bad certificate"},
			expectedRetry: false,
		},
		{
			name:          "Unknown error should not retry (conservative)",
			err:           errors.New("unknown network error"),
			expectedRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			client, err := NewClient("")
			require.NoError(t, err)

			// Create a mock response with POST method (non-idempotent) to test error-specific logic
			// For idempotent methods like GET, the retry logic returns true immediately for any error
			resp := &http.Response{
				Request: &http.Request{Method: http.MethodPost},
			}

			// WHEN
			retry, err := client.retryHTTPCheck(context.Background(), resp, tt.err)

			// THEN
			assert.Equal(t, tt.expectedRetry, retry, "Expected retry=%v for error: %v", tt.expectedRetry, tt.err)
			if tt.expectedRetry {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_DefaultRetryPolicy_ContextCancellation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  func() context.Context
	}{
		{
			name: "context canceled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
		},
		{
			name: "context deadline exceeded",
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
				defer cancel()
				return ctx
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			client, err := NewClient("")
			require.NoError(t, err)

			ctx := tt.ctx()

			// WHEN
			retry, err := client.retryHTTPCheck(ctx, nil, errors.New("some error"))

			// THEN
			assert.False(t, retry)
			assert.Equal(t, ctx.Err(), err)
		})
	}
}

func TestClient_DefaultRetryPolicy_RetriesDisabled(t *testing.T) {
	t.Parallel()

	// GIVEN
	client, err := NewClient("", WithCustomRetryMax(0))
	require.NoError(t, err)

	// Manually disable retries to test this specific path
	client.disableRetries = true

	// WHEN
	retry, err := client.retryHTTPCheck(context.Background(), nil, errors.New("some error"))

	// THEN
	assert.False(t, retry)
	assert.NoError(t, err)
}

// mockTemporaryError implements a temporary error for testing
type mockTemporaryError struct {
	temporary bool
}

func (e *mockTemporaryError) Error() string {
	return "mock temporary error"
}

func (e *mockTemporaryError) Temporary() bool {
	return e.temporary
}

// mockTLSHandshakeError implements an error that looks like a TLS handshake error
type mockTLSHandshakeError struct {
	msg string
}

func (e *mockTLSHandshakeError) Error() string {
	return e.msg
}

func (e *mockTLSHandshakeError) Timeout() bool {
	return strings.Contains(e.msg, "timeout")
}

func (e *mockTLSHandshakeError) Temporary() bool {
	return strings.Contains(e.msg, "timeout")
}

func TestParseID(t *testing.T) {
	t.Parallel()

	// GIVEN a set of inputs
	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
	}{
		{
			name:     "int type",
			input:    42,
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "int64 type",
			input:    int64(15),
			expected: "15",
			wantErr:  false,
		},
		{
			name:     "string type",
			input:    "test-id",
			expected: "test-id",
			wantErr:  false,
		},
		{
			name:     "invalid type",
			input:    3.14,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// WHEN parseID is called with the input
			result, err := parseID(tt.input)

			// THEN it should return the expected result
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
