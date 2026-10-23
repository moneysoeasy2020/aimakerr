package gitlab

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testUser struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type testProject struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func TestDoRequestSuccess(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/users/1"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/get_user.json")
	})

	// WHEN
	user, resp, err := do[*testUser](
		client,
		withPath("users/1"),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "John Smith", user.Name)
}

func TestDoRequestPOSTWithBody(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/projects"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var reqBody testProject
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "New Project", reqBody.Name)

		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": 1, "name": "New Project", "description": "Test project"}`))
	})

	requestBody := &testProject{
		Name:        "New Project",
		Description: "Test project",
	}

	// WHEN
	project, resp, err := do[*testProject](
		client,
		withMethod(http.MethodPost),
		withPath("projects"),
		withAPIOpts(requestBody),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "New Project", project.Name)
}

func TestDoRequestErrorResponse(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/users/999"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(404)
		w.Write([]byte(`{"message": "Not found"}`))
	})

	// WHEN
	user, resp, err := do[*testUser](
		client,
		withPath("users/99"),
	)

	// THEN
	assert.Error(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Nil(t, user)
}

func TestDoRequestSliceSuccess(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/users"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/list_users.json")
	})

	// WHEN
	users, resp, err := do[[]testUser](
		client,
		withPath("users"),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, users, 3)

	expectedUsers := []testUser{
		{ID: 1, Name: "Example User 1"},
		{ID: 2, Name: "Example User 2"},
		{ID: 3, Name: "Example User 3"},
	}
	for i, user := range users {
		assert.Equal(t, expectedUsers[i].ID, user.ID)
		assert.Equal(t, expectedUsers[i].Name, user.Name)
	}
}

func TestDoRequestSliceEmptySlice(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/users"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})

	// WHEN
	users, resp, err := do[[]testUser](
		client,
		withPath("users"),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Empty(t, users)
}

func TestDoRequestSliceErrorResponse(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/users"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(500)
		w.Write([]byte(`{"message": "Internal server error"}`))
	})

	// WHEN
	users, resp, err := do[[]testUser](
		client,
		withPath("users"),
	)

	// THEN
	assert.Error(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Nil(t, users)
}

func TestDoRequestVoidSuccessDELETE(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/users/1"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204) // No Content
	})

	// WHEN
	_, resp, err := do[none](
		client,
		withMethod(http.MethodDelete),
		withPath("users/1"),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestDoRequestVoidSuccessPUT(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/merge_requests/1/approve"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		var reqBody map[string]any
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "approve", reqBody["action"])

		w.WriteHeader(200)
	})

	requestBody := map[string]string{"action": "approve"}

	// WHEN
	_, resp, err := do[none](
		client,
		withMethod(http.MethodPut),
		withPath("merge_requests/1/approve"),
		withAPIOpts(requestBody),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDoRequestVoidErrorResponse(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/users/1"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(403)
		w.Write([]byte(`{"message": "Forbidden"}`))
	})

	// WHEN
	_, resp, err := do[none](
		client,
		withMethod(http.MethodDelete),
		withPath("users/1"),
	)

	// THEN
	assert.Error(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestRequestHandlerWithOptions(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	path := "/api/v4/users"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		// Check that query parameters from options are included
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))

		mustWriteHTTPResponse(t, w, "testdata/list_users_public_email.json")
	})

	options := []RequestOptionFunc{
		WithOffsetPaginationParameters(2),
		WithHeader("X-Test-Header", "test-value"),
	}

	// WHEN
	users, resp, err := do[[]testUser](
		client,
		withPath("users"),
		withRequestOpts(options...),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, users, 1)
}

func TestDoRequestProjectID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	mux.HandleFunc("/api/v4/projects/group%2Fproject", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	// WHEN
	_, resp, err := do[none](
		client,
		withPath("projects/%s", ProjectID{"group/project"}),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDoRequestGroupID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	mux.HandleFunc("/api/v4/groups/sub%2Fgroup", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	// WHEN
	_, resp, err := do[none](
		client,
		withPath("groups/%s", GroupID{"sub/group"}),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDoRequestRunnerID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	mux.HandleFunc("/api/v4/runners/some%2Frunner", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	// WHEN
	_, resp, err := do[none](
		client,
		withPath("runners/%s", RunnerID{"some/runner"}),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDoRequestUserID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN
	mux.HandleFunc("/api/v4/users/test%2Fuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	// WHEN
	_, resp, err := do[none](
		client,
		withPath("users/%s", UserID{"test/user"}),
	)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
