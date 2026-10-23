package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphQL_Do_Success(t *testing.T) {
	t.Parallel()

	// GIVEN
	mux, client := setup(t)
	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testJSONBody(t, r, `{ "query": "query { project(fullPath: \"gitlab-org/gitlab\") { id } }" }`)
		fmt.Fprint(w, `
			{
				"data": {
					"project": {
						"id": "any-id"
					}
				}
			}
		`)
	})

	// WHEN
	var response struct {
		Data struct {
			Project struct {
				ID string `json:"id"`
			} `json:"project"`
		} `json:"data"`
	}
	_, err := client.GraphQL.Do(GraphQLQuery{Query: `query { project(fullPath: "gitlab-org/gitlab") { id } }`}, &response)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, "any-id", response.Data.Project.ID)
}

func TestGraphQL_Do_Success_With_Variables(t *testing.T) {
	t.Parallel()

	// GIVEN
	mux, client := setup(t)
	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testJSONBody(t, r, `{ "query": "query ($projectPath: ID!) { project(fullPath: $projectPath) { id } }", "variables": { "projectPath": "gitlab-org/gitlab" } }`)
		fmt.Fprint(w, `
			{
				"data": {
					"project": {
						"id": "any-id"
					}
				}
			}
		`)
	})

	// WHEN
	var response struct {
		Data struct {
			Project struct {
				ID string `json:"id"`
			} `json:"project"`
		} `json:"data"`
	}
	_, err := client.GraphQL.Do(
		GraphQLQuery{
			Query:     `query ($projectPath: ID!) { project(fullPath: $projectPath) { id } }`,
			Variables: map[string]any{"projectPath": "gitlab-org/gitlab"},
		},
		&response)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, "any-id", response.Data.Project.ID)
}

func TestGraphQL_Do_ErrorWithMessages(t *testing.T) {
	t.Parallel()

	// GIVEN
	mux, client := setup(t)
	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"errors": [{"message": "bad request"}]}`)
	})

	// WHEN
	var response struct {
		Data struct {
			Project struct {
				ID string `json:"id"`
			} `json:"project"`
		} `json:"data"`
	}
	_, err := client.GraphQL.Do(GraphQLQuery{Query: `query { project(fullPath: "gitlab-org/gitlab") { id } }`}, &response)

	// THEN
	assert.ErrorContains(t, err, "GraphQL errors: bad request")
}

func TestGraphQL_Do_ErrorNoMessages(t *testing.T) {
	t.Parallel()

	// GIVEN
	mux, client := setup(t)
	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"key": "whuat"}`)
	})

	// WHEN
	var response struct {
		Data struct {
			Project struct {
				ID string `json:"id"`
			} `json:"project"`
		} `json:"data"`
	}
	_, err := client.GraphQL.Do(GraphQLQuery{Query: `query { project(fullPath: "gitlab-org/gitlab") { id } }`}, &response)

	// THEN
	assert.ErrorContains(t, err, `{key: whuat} (no additional error messages)`)
}
