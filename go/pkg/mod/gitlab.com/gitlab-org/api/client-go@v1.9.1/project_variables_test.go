package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectVariablesService_ListVariables(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/variables", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteJSONResponse(t, w, []map[string]any{
			{
				"key":           "TEST_VARIABLE_1",
				"variable_type": "env_var",
				"value":         "TEST_1",
				"description":   "test variable 1",
			},
		})
	})

	want := []*ProjectVariable{{
		Key:              "TEST_VARIABLE_1",
		Value:            "TEST_1",
		VariableType:     "env_var",
		Protected:        false,
		Masked:           false,
		Hidden:           false,
		EnvironmentScope: "",
		Description:      "test variable 1",
	}}

	pvs, resp, err := client.ProjectVariables.ListVariables(1, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pvs)

	pvs, resp, err = client.ProjectVariables.ListVariables(1.01, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pvs)

	pvs, resp, err = client.ProjectVariables.ListVariables(1, nil, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pvs)

	pvs, resp, err = client.ProjectVariables.ListVariables(2, nil, nil)
	require.Error(t, err)
	require.Nil(t, pvs)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectVariablesService_GetVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/variables/TEST_VARIABLE_1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testParam(t, r, "filter[environment_scope]", "prod")
		mustWriteJSONResponse(t, w, map[string]any{
			"key":           "TEST_VARIABLE_1",
			"variable_type": "env_var",
			"value":         "TEST_1",
			"protected":     false,
			"masked":        true,
			"hidden":        true,
			"description":   "test variable 1",
		})
	})

	want := &ProjectVariable{
		Key:              "TEST_VARIABLE_1",
		Value:            "TEST_1",
		VariableType:     "env_var",
		Protected:        false,
		Masked:           true,
		Hidden:           true,
		EnvironmentScope: "",
		Description:      "test variable 1",
	}

	pv, resp, err := client.ProjectVariables.GetVariable(1, "TEST_VARIABLE_1", &GetProjectVariableOptions{Filter: &VariableFilter{EnvironmentScope: "prod"}}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pv)

	pv, resp, err = client.ProjectVariables.GetVariable(1.01, "TEST_VARIABLE_1", nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.GetVariable(1, "TEST_VARIABLE_1", nil, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.GetVariable(2, "TEST_VARIABLE_1", nil, nil)
	require.Error(t, err)
	require.Nil(t, pv)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectVariablesService_GetVariable_EmptyEnvironmentScopeAttribute(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/variables/TEST_VARIABLE_1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.False(t, r.URL.Query().Has("filter[environment_scope]"))
		mustWriteJSONResponse(t, w, map[string]any{
			"key":           "TEST_VARIABLE_1",
			"variable_type": "env_var",
			"value":         "TEST_1",
			"protected":     false,
			"masked":        true,
			"hidden":        true,
			"description":   "test variable 1",
		})
	})

	want := &ProjectVariable{
		Key:              "TEST_VARIABLE_1",
		Value:            "TEST_1",
		VariableType:     "env_var",
		Protected:        false,
		Masked:           true,
		Hidden:           true,
		EnvironmentScope: "",
		Description:      "test variable 1",
	}

	pv, resp, err := client.ProjectVariables.GetVariable(1, "TEST_VARIABLE_1", &GetProjectVariableOptions{Filter: &VariableFilter{EnvironmentScope: ""}}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pv)
}

func TestProjectVariablesService_CreateVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/variables", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBodyJSON(t, r, map[string]string{
			"description": "new variable",
		})
		mustWriteJSONResponse(t, w, map[string]any{
			"key":               "NEW_VARIABLE",
			"value":             "new value",
			"protected":         false,
			"variable_type":     "env_var",
			"masked":            false,
			"masked_and_hidden": false,
			"environment_scope": "*",
			"description":       "new variable",
		})
	})

	want := &ProjectVariable{
		Key:              "NEW_VARIABLE",
		Value:            "new value",
		VariableType:     "env_var",
		Protected:        false,
		Masked:           false,
		Hidden:           false,
		EnvironmentScope: "*",
		Description:      "new variable",
	}

	pv, resp, err := client.ProjectVariables.CreateVariable(1, &CreateProjectVariableOptions{Description: Ptr("new variable")}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pv)

	pv, resp, err = client.ProjectVariables.CreateVariable(1.01, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.CreateVariable(1, nil, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.CreateVariable(2, nil, nil)
	require.Error(t, err)
	require.Nil(t, pv)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectVariablesService_CreateVariable_MaskedAndHidden(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/variables", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBodyJSON(t, r, map[string]string{
			"description": "new variable",
		})
		mustWriteJSONResponse(t, w, map[string]any{
			"key":               "NEW_VARIABLE",
			"protected":         false,
			"variable_type":     "env_var",
			"masked":            true,
			"hidden":            true,
			"environment_scope": "*",
			"description":       "new variable",
		})
	})

	want := &ProjectVariable{
		Key:              "NEW_VARIABLE",
		VariableType:     "env_var",
		Protected:        false,
		Masked:           true,
		Hidden:           true,
		EnvironmentScope: "*",
		Description:      "new variable",
	}

	pv, resp, err := client.ProjectVariables.CreateVariable(1, &CreateProjectVariableOptions{Description: Ptr("new variable")}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pv)

	pv, resp, err = client.ProjectVariables.CreateVariable(1.01, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.CreateVariable(1, nil, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.CreateVariable(2, nil, nil)
	require.Error(t, err)
	require.Nil(t, pv)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectVariablesService_UpdateVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/variables/NEW_VARIABLE", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testBodyJSON(t, r, map[string]any{
			"description": "updated description",
			"filter":      map[string]any{"environment_scope": "prod"},
		})
		mustWriteJSONResponse(t, w, map[string]any{
			"key":               "NEW_VARIABLE",
			"value":             "updated value",
			"protected":         false,
			"variable_type":     "env_var",
			"masked":            false,
			"environment_scope": "*",
			"description":       "updated description",
		})
	})

	want := &ProjectVariable{
		Key:              "NEW_VARIABLE",
		Value:            "updated value",
		VariableType:     "env_var",
		Protected:        false,
		Masked:           false,
		Hidden:           false,
		EnvironmentScope: "*",
		Description:      "updated description",
	}

	pv, resp, err := client.ProjectVariables.UpdateVariable(1, "NEW_VARIABLE", &UpdateProjectVariableOptions{
		Filter:      &VariableFilter{EnvironmentScope: "prod"},
		Description: Ptr("updated description"),
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pv)

	pv, resp, err = client.ProjectVariables.UpdateVariable(1.01, "NEW_VARIABLE", nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.UpdateVariable(1, "NEW_VARIABLE", nil, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.UpdateVariable(2, "NEW_VARIABLE", nil, nil)
	require.Error(t, err)
	require.Nil(t, pv)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectVariablesService_UpdateVariable_MaskedAndHidden(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/variables/NEW_VARIABLE", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testBodyJSON(t, r, map[string]any{
			"description": "updated description",
			"filter":      map[string]any{"environment_scope": "prod"},
		})
		mustWriteJSONResponse(t, w, map[string]any{
			"key":               "NEW_VARIABLE",
			"value":             nil,
			"protected":         false,
			"variable_type":     "env_var",
			"masked":            true,
			"hidden":            true,
			"environment_scope": "*",
			"description":       "updated description",
		})
	})

	want := &ProjectVariable{
		Key:              "NEW_VARIABLE",
		VariableType:     "env_var",
		Protected:        false,
		Masked:           true,
		Hidden:           true,
		EnvironmentScope: "*",
		Description:      "updated description",
	}

	pv, resp, err := client.ProjectVariables.UpdateVariable(1, "NEW_VARIABLE", &UpdateProjectVariableOptions{
		Filter:      &VariableFilter{EnvironmentScope: "prod"},
		Description: Ptr("updated description"),
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pv)

	pv, resp, err = client.ProjectVariables.UpdateVariable(1.01, "NEW_VARIABLE", nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.UpdateVariable(1, "NEW_VARIABLE", nil, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pv)

	pv, resp, err = client.ProjectVariables.UpdateVariable(2, "NEW_VARIABLE", nil, nil)
	require.Error(t, err)
	require.Nil(t, pv)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectVariablesService_RemoveVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/variables/VARIABLE_1", func(_ http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		testParam(t, r, "filter[environment_scope]", "prod")
	})

	resp, err := client.ProjectVariables.RemoveVariable(1, "VARIABLE_1", &RemoveProjectVariableOptions{Filter: &VariableFilter{EnvironmentScope: "prod"}}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	resp, err = client.ProjectVariables.RemoveVariable(1.01, "VARIABLE_1", nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)

	resp, err = client.ProjectVariables.RemoveVariable(1, "VARIABLE_1", nil, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)

	resp, err = client.ProjectVariables.RemoveVariable(2, "VARIABLE_1", nil, nil)
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
