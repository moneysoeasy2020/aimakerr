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
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListTodos(t *testing.T) {
	t.Parallel()

	const testdataFile = "testdata/list_todos.json"

	tests := []struct {
		name    string
		opts    *ListTodosOptions
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
		wantErr string
	}{
		{
			name: "with action",
			opts: &ListTodosOptions{
				Action: Ptr(TodoMentioned),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testParam(t, r, "action", string(TodoMentioned))
				mustWriteHTTPResponse(t, w, testdataFile)
			},
		},
		{
			name: "with author_id",
			opts: &ListTodosOptions{
				AuthorID: Ptr(int64(1)),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testParam(t, r, "author_id", "1")
				mustWriteHTTPResponse(t, w, testdataFile)
			},
		},
		{
			name: "with project_id",
			opts: &ListTodosOptions{
				ProjectID: Ptr(int64(1)),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testParam(t, r, "project_id", "1")
				mustWriteHTTPResponse(t, w, testdataFile)
			},
		},
		{
			name: "with group_id",
			opts: &ListTodosOptions{
				GroupID: Ptr(int64(1)),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testParam(t, r, "group_id", "1")
				mustWriteHTTPResponse(t, w, testdataFile)
			},
		},
		{
			name: "with state",
			opts: &ListTodosOptions{
				State: Ptr("pending"),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testParam(t, r, "state", "pending")
				mustWriteHTTPResponse(t, w, testdataFile)
			},
		},
		{
			name: "with type",
			opts: &ListTodosOptions{
				Type: Ptr("Issue"),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testParam(t, r, "type", "Issue")
				mustWriteHTTPResponse(t, w, testdataFile)
			},
		},
		{
			name: "with server error",
			opts: &ListTodosOptions{},
			handler: func(t *testing.T, w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				mustWriteErrorResponse(t, w, errors.New("Internal Server Error"))
			},
			wantErr: "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mux, client := setup(t)

			// Create a closure that captures the correct t for this test case
			mux.HandleFunc("/api/v4/todos", func(w http.ResponseWriter, r *http.Request) {
				tt.handler(t, w, r)
			})

			todos, _, err := client.Todos.ListTodos(tt.opts)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)

			want := []*Todo{
				{ID: 1, State: "pending", Target: &TodoTarget{ID: float64(1), ApprovalsBeforeMerge: 2}},
				{ID: 2, State: "pending", Target: &TodoTarget{ID: "1d76d1b2e3e886108f662765c97f4687f4134d8c"}},
			}

			require.Equal(t, want, todos)
		})
	}
}

func TestMarkAllTodosAsDone(t *testing.T) {
	t.Parallel()

	t.Run("successful request", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/todos/mark_as_done", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			w.WriteHeader(http.StatusNoContent)
		})

		resp, err := client.Todos.MarkAllTodosAsDone()
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("error handling", func(t *testing.T) {
		t.Parallel()
		mux, client := setup(t)

		mux.HandleFunc("/api/v4/todos/mark_as_done", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		resp, err := client.Todos.MarkAllTodosAsDone()
		require.Error(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestMarkTodoAsDone(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/todos/1/mark_as_done", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
	})

	_, err := client.Todos.MarkTodoAsDone(1)
	require.NoError(t, err)
}
