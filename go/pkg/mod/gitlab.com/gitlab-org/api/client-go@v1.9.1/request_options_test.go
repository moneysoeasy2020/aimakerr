// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitlab

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
)

func TestWithContext(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	mux.HandleFunc("/api/v4/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// WithContext is called once
	ctx1 := contextWithCheckRetry(context.Background(), func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if err != nil {
			return false, err
		}
		if resp.StatusCode == http.StatusMethodNotAllowed {
			return true, nil
		}
		return false, nil
	})

	req, err := client.NewRequest(
		http.MethodGet,
		"/ok",
		nil,
		[]RequestOptionFunc{WithContext(ctx1)},
	)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	// WithContext is called twice
	ctx1 = contextWithCheckRetry(context.Background(), func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if err != nil {
			return false, err
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusNotFound {
			return true, nil
		}
		return false, nil
	})
	ctx2 := contextWithCheckRetry(context.Background(), func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if err != nil {
			return false, err
		}
		if resp.StatusCode == http.StatusBadRequest {
			return true, nil
		}
		return false, nil
	})

	req, err = client.NewRequest(
		http.MethodGet,
		"/ok",
		nil,
		[]RequestOptionFunc{WithContext(ctx1), WithContext(ctx2)},
	)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)
}

func TestWithContextAndWithRequestRetry(t *testing.T) {
	t.Parallel()

	retryCount := 0
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/success-on-3rd", func(w http.ResponseWriter, r *http.Request) {
		retryCount += 1

		if retryCount < 3 {
			w.WriteHeader(http.StatusMethodNotAllowed)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	// retryableStatusCodes in context is restored when WithContext is called
	newCtx := context.Background()
	req, err := client.NewRequest(
		http.MethodGet,
		"/success-on-3rd",
		nil,
		[]RequestOptionFunc{
			WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
				if ctx.Err() != nil {
					return false, ctx.Err()
				}
				if err != nil {
					return false, err
				}
				if resp.StatusCode == http.StatusMethodNotAllowed {
					return true, nil
				}
				return false, nil
			}),
			WithContext(newCtx),
		},
	)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)
}

func TestWithHeader(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/without-header", func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("X-Custom-Header"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"X-Custom-Header": %s`, r.Header.Get("X-Custom-Header"))
	})
	mux.HandleFunc("/api/v4/with-header", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "randomtokenstring", r.Header.Get("X-Custom-Header"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"X-Custom-Header": %s`, r.Header.Get("X-Custom-Header"))
	})

	// ensure that X-Custom-Header hasn't been set at all
	req, err := client.NewRequest(http.MethodGet, "/without-header", nil, nil)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	// ensure that X-Custom-Header is set for only one request
	req, err = client.NewRequest(
		http.MethodGet,
		"/with-header",
		nil,
		[]RequestOptionFunc{WithHeader("X-Custom-Header", "randomtokenstring")},
	)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	req, err = client.NewRequest(http.MethodGet, "/without-header", nil, nil)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	// ensure that X-Custom-Header is set for all client requests
	addr := client.BaseURL().String()
	client, err = NewClient("",
		// same base options as setup
		WithBaseURL(addr),
		// Disable backoff to speed up tests that expect errors.
		WithCustomBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
			return 0
		}),
		// add client headers
		WithRequestOptions(WithHeader("X-Custom-Header", "randomtokenstring")))
	assert.NoError(t, err)
	assert.NotNil(t, client)

	req, err = client.NewRequest(http.MethodGet, "/with-header", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "randomtokenstring", req.Header.Get("X-Custom-Header"))

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	req, err = client.NewRequest(http.MethodGet, "/with-header", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "randomtokenstring", req.Header.Get("X-Custom-Header"))

	_, err = client.Do(req, nil)
	assert.NoError(t, err)
}

func TestWithHeaders(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/without-headers", func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("X-Custom-Header-1"))
		assert.Empty(t, r.Header.Get("X-Custom-Header-2"))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/v4/with-headers", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "randomtokenstring", r.Header.Get("X-Custom-Header-1"))
		assert.Equal(t, "randomtokenstring2", r.Header.Get("X-Custom-Header-2"))
		w.WriteHeader(http.StatusOK)
	})

	headers := map[string]string{
		"X-Custom-Header-1": "randomtokenstring",
		"X-Custom-Header-2": "randomtokenstring2",
	}

	// ensure that X-Custom-Header hasn't been set at all
	req, err := client.NewRequest(http.MethodGet, "/without-headers", nil, nil)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	// ensure that X-Custom-Header is set for only one request
	req, err = client.NewRequest(
		http.MethodGet,
		"/with-headers",
		nil,
		[]RequestOptionFunc{WithHeaders(headers)},
	)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	req, err = client.NewRequest(http.MethodGet, "/without-headers", nil, nil)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	// ensure that X-Custom-Header is set for all client requests
	addr := client.BaseURL().String()
	client, err = NewClient("",
		// same base options as setup
		WithBaseURL(addr),
		// Disable backoff to speed up tests that expect errors.
		WithCustomBackoff(func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
			return 0
		}),
		// add client headers
		WithRequestOptions(WithHeaders(headers)),
	)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	req, err = client.NewRequest(http.MethodGet, "/with-headers", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "randomtokenstring", req.Header.Get("X-Custom-Header-1"))

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	req, err = client.NewRequest(http.MethodGet, "/with-headers", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "randomtokenstring", req.Header.Get("X-Custom-Header-1"))

	_, err = client.Do(req, nil)
	assert.NoError(t, err)
}

func TestWithKeysetPaginationParameters(t *testing.T) {
	t.Parallel()
	req, err := retryablehttp.NewRequest("GET", "https://gitlab.example.com/api/v4/groups?pagination=keyset&per_page=50&order_by=name&sort=asc", nil)
	assert.NoError(t, err)

	linkNext := "https://gitlab.example.com/api/v4/groups?pagination=keyset&per_page=50&order_by=name&sort=asc&cursor=eyJuYW1lIjoiRmxpZ2h0anMiLCJpZCI6IjI2IiwiX2tkIjoibiJ9"

	err = WithKeysetPaginationParameters(linkNext)(req)
	assert.NoError(t, err)

	values := req.URL.Query()
	// Ensure all original parameters remain
	assert.Equal(t, "keyset", values.Get("pagination"))
	assert.Equal(t, "50", values.Get("per_page"))
	assert.Equal(t, "name", values.Get("order_by"))
	assert.Equal(t, "asc", values.Get("sort"))

	// Ensure cursor gets properly pulled from "next link" header
	assert.Equal(t, "eyJuYW1lIjoiRmxpZ2h0anMiLCJpZCI6IjI2IiwiX2tkIjoibiJ9", values.Get("cursor"))
}

func TestWithRequestRetry(t *testing.T) {
	t.Parallel()

	retryCount := 0

	mux, client := setup(t)
	mux.HandleFunc("/api/v4/success-on-3rd", func(w http.ResponseWriter, r *http.Request) {
		retryCount += 1
		switch retryCount {
		case 1:
			w.WriteHeader(http.StatusMethodNotAllowed)
		case 2:
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusOK)
		}
	})

	req, err := client.NewRequest(
		http.MethodGet,
		"/success-on-3rd",
		nil,
		[]RequestOptionFunc{
			WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
				if ctx.Err() != nil {
					return false, ctx.Err()
				}
				if err != nil {
					return false, err
				}
				if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusUnprocessableEntity {
					return true, nil
				}
				return false, nil
			}),
		},
	)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.NoError(t, err)

	// fails because it is retried only in case of StatusUnprocessableEntity error. (WithRetryForStatusCodes(StatusMethodNotAllowed) is overwritten)
	retryCount = 0
	req, err = client.NewRequest(
		http.MethodGet,
		"/success-on-3rd",
		nil,
		[]RequestOptionFunc{
			WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
				if ctx.Err() != nil {
					return false, ctx.Err()
				}
				if err != nil {
					return false, err
				}
				if resp.StatusCode == http.StatusMethodNotAllowed {
					return true, nil
				}
				return false, nil
			}),
			WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
				if ctx.Err() != nil {
					return false, ctx.Err()
				}
				if err != nil {
					return false, err
				}
				if resp.StatusCode == http.StatusUnprocessableEntity {
					return true, nil
				}
				return false, nil
			}),
		},
	)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.ErrorContains(t, err, ": 405", "expect to returns StatusMethodNotAllowed error")

	// fails because 422 error is not allow to retryable
	retryCount = 0
	req, err = client.NewRequest(
		http.MethodGet,
		"/success-on-3rd",
		nil,
		[]RequestOptionFunc{
			WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
				if ctx.Err() != nil {
					return false, ctx.Err()
				}
				if err != nil {
					return false, err
				}
				if resp.StatusCode == http.StatusMethodNotAllowed {
					return true, nil
				}
				return false, nil
			}),
		},
	)
	assert.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.ErrorContains(t, err, ": 422", "expect to returns StatusUnprocessableEntity error")
}

func ExampleWithRequestRetry_createMergeRequestAndSetAutoMerge() {
	git, err := NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	projectName := "example/example"

	// Create a new Merge Request
	mr, _, err := git.MergeRequests.CreateMergeRequest(projectName, &CreateMergeRequestOptions{
		SourceBranch:       Ptr("my-topic-branch"),
		TargetBranch:       Ptr("main"),
		Title:              Ptr("New MergeRequest"),
		Description:        Ptr("New MergeRequest"),
		RemoveSourceBranch: Ptr(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Set auto-merge to created Merge Request
	// c.f. https://docs.gitlab.com/user/project/merge_requests/auto_merge/
	_, _, err = git.MergeRequests.AcceptMergeRequest(
		projectName, mr.IID, &AcceptMergeRequestOptions{MergeWhenPipelineSucceeds: Ptr(true)},

		// client-go provides retries on rate limit (429) and server (>= 500) errors by default.
		//
		// But Method Not Allowed (405) and Unprocessable Content (422) errors will be returned
		// when AcceptMergeRequest is called immediately after CreateMergeRequest.
		//
		// c.f. https://docs.gitlab.com/api/merge_requests/#merge-a-merge-request
		//
		// Therefore, add a retryable status code only for AcceptMergeRequest calls
		WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			if err != nil {
				return false, err
			}
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusUnprocessableEntity {
				return true, nil
			}
			return false, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
