package gitlab

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectAliasesService_CreateProjectAlias(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/project_aliases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "Failed to read request body")

		var payload CreateProjectAliasOptions
		err = json.Unmarshal(body, &payload)
		assert.NoError(t, err, "Failed to unmarshal JSON")

		assert.NotNil(t, payload.Name, "Expected Name to be non-nil")
		assert.Equal(t, "my-alias", *payload.Name)
		assert.Equal(t, int64(1), payload.ProjectID)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 10, "name": "my-alias", "project_id": 1}`))
	})

	s := client.ProjectAliases
	opt := &CreateProjectAliasOptions{
		Name:      Ptr("my-alias"),
		ProjectID: int64(1),
	}
	alias, resp, err := s.CreateProjectAlias(opt)
	require.NoError(t, err)
	assert.Equal(t, int64(10), alias.ID)
	assert.Equal(t, "my-alias", alias.Name)
	assert.Equal(t, int64(1), alias.ProjectID)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestProjectAliasesService_DeleteProjectAlias(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/project_aliases/my-alias", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	s := client.ProjectAliases
	resp, err := s.DeleteProjectAlias("my-alias")
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestProjectAliasesService_GetProjectAlias(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/project_aliases/my-alias", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": 10, "name": "my-alias", "project_id": 1}`))
	})

	s := client.ProjectAliases
	alias, resp, err := s.GetProjectAlias("my-alias")
	require.NoError(t, err)
	assert.Equal(t, int64(10), alias.ID)
	assert.Equal(t, "my-alias", alias.Name)
	assert.Equal(t, int64(1), alias.ProjectID)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestProjectAliasesService_ListProjectAliases(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/project_aliases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": 10, "name": "my-alias", "project_id": 1}]`))
	})

	s := client.ProjectAliases
	aliases, resp, err := s.ListProjectAliases()
	require.NoError(t, err)
	require.Len(t, aliases, 1)
	assert.Equal(t, int64(10), aliases[0].ID)
	assert.Equal(t, "my-alias", aliases[0].Name)
	assert.Equal(t, int64(1), aliases[0].ProjectID)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestProjectAliasesService_GetProjectAlias_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	aliasName := "my/alias?with=special&chars"
	expectedEscaped := "my%2Falias%3Fwith%3Dspecial%26chars"

	mux.HandleFunc("/api/v4/project_aliases/"+expectedEscaped, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": 10, "name": "my/alias?with=special&chars", "project_id": 1}`))
	})

	s := client.ProjectAliases
	alias, resp, err := s.GetProjectAlias(aliasName)
	require.NoError(t, err)
	assert.Equal(t, int64(10), alias.ID)
	assert.Equal(t, "my/alias?with=special&chars", alias.Name)
	assert.Equal(t, int64(1), alias.ProjectID)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestProjectAliasesService_DeleteProjectAlias_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	aliasName := "my/alias?with=special&chars"
	expectedEscaped := "my%2Falias%3Fwith%3Dspecial%26chars"

	mux.HandleFunc("/api/v4/project_aliases/"+expectedEscaped, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	s := client.ProjectAliases
	resp, err := s.DeleteProjectAlias(aliasName)
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestProjectAliasesService_GetProjectAlias_WithSpacesAndDots(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	aliasName := "my alias.name"
	expectedEscaped := "my%20alias%2Ename"

	mux.HandleFunc("/api/v4/project_aliases/"+expectedEscaped, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": 20, "name": "my alias.name", "project_id": 2}`))
	})

	s := client.ProjectAliases
	alias, resp, err := s.GetProjectAlias(aliasName)
	require.NoError(t, err)
	assert.Equal(t, int64(20), alias.ID)
	assert.Equal(t, "my alias.name", alias.Name)
	assert.Equal(t, int64(2), alias.ProjectID)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestProjectAliasesService_DeleteProjectAlias_WithSpacesAndDots(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	aliasName := "my alias.name"
	expectedEscaped := "my%20alias%2Ename"

	mux.HandleFunc("/api/v4/project_aliases/"+expectedEscaped, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	s := client.ProjectAliases
	resp, err := s.DeleteProjectAlias(aliasName)
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}
