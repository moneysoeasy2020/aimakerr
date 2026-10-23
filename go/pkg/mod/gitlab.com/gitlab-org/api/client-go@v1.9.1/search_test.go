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
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchService_Users(t *testing.T) {
	t.Parallel()

	t.Run("successful search", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/search", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			mustWriteHTTPResponse(t, w, "testdata/search_users.json")
		})

		opts := &SearchOptions{ListOptions: ListOptions{PerPage: 2}}
		users, resp, err := client.Search.Users("doe", opts)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		want := []*User{{
			ID:        1,
			Username:  "user1",
			Name:      "John Doe1",
			State:     "active",
			AvatarURL: "http://www.gravatar.com/avatar/c922747a93b40d1ea88262bf1aebee62?s=80&d=identicon",
			WebURL:    "http://localhost/user1",
		}}
		assert.Equal(t, want, users)
	})

	t.Run("error handling", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/search", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"message": "Internal Server Error"}`)
		})

		opts := &SearchOptions{ListOptions: ListOptions{PerPage: 20}}
		users, resp, err := client.Search.Users("doe", opts)

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("empty search term", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/search", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `[]`)
		})

		opts := &SearchOptions{ListOptions: ListOptions{PerPage: 20}}
		users, resp, err := client.Search.Users("", opts)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, users)
	})
	t.Run("pagination - second page", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/search", func(w http.ResponseWriter, r *http.Request) {
			testParam(t, r, "page", "2")
			testParam(t, r, "per_page", "1")
			testMethod(t, r, http.MethodGet)
			mustWriteHTTPResponse(t, w, "testdata/search_users_pagination.json")
		})

		opts := &SearchOptions{ListOptions: ListOptions{Page: 2, PerPage: 1}}
		users, _, err := client.Search.Users("doe", opts)

		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, int64(2), users[0].ID)
	})
}

func TestSearchService_UsersByGroup(t *testing.T) {
	t.Parallel()

	t.Run("valid group ID - returns expected users", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/groups/3/-/search", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)

			query := r.URL.Query()
			if got := query.Get("search"); got != "doe" {
				t.Errorf("expected query 'search=doe', got 'search=%s'", got)
			}

			mustWriteHTTPResponse(t, w, "testdata/search_users.json")
		})

		users, _, err := client.Search.UsersByGroup("3", "doe", nil)

		assert.NoError(t, err)

		want := []*User{{
			ID:        1,
			Username:  "user1",
			Name:      "John Doe1",
			State:     "active",
			AvatarURL: "http://www.gravatar.com/avatar/c922747a93b40d1ea88262bf1aebee62?s=80&d=identicon",
			WebURL:    "http://localhost/user1",
		}}
		assert.Equal(t, want, users)
	})

	t.Run("invalid group ID - returns 404", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/groups/invalid/-/search", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message": "404 Group Not Found"}`)
		})

		users, resp, err := client.Search.UsersByGroup("invalid", "doe", &SearchOptions{})
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestSearchService_UsersByProject(t *testing.T) {
	t.Parallel()

	t.Run("valid project ID", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/projects/6/-/search", func(w http.ResponseWriter, r *http.Request) {
			testParam(t, r, "search", "doe")
			testParam(t, r, "scope", "users")
			testMethod(t, r, http.MethodGet)
			mustWriteHTTPResponse(t, w, "testdata/search_users.json")
		})
		opts := &SearchOptions{ListOptions: ListOptions{PerPage: 2}}
		users, _, err := client.Search.UsersByProject("6", "doe", opts)

		assert.NoError(t, err)

		want := []*User{{
			ID:        1,
			Username:  "user1",
			Name:      "John Doe1",
			State:     "active",
			AvatarURL: "http://www.gravatar.com/avatar/c922747a93b40d1ea88262bf1aebee62?s=80&d=identicon",
			WebURL:    "http://localhost/user1",
		}}
		assert.Equal(t, want, users)
	})

	t.Run("invalid project ID", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/projects/invalid/-/search", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message": "404 Project Not Found"}`)
		})

		opts := &SearchOptions{ListOptions: ListOptions{PerPage: 2}}
		users, resp, err := client.Search.UsersByProject("invalid", "doe", opts)

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
