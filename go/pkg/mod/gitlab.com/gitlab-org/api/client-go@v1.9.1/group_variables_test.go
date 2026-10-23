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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListGroupVariables(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			mustWriteJSONResponse(t, w, []map[string]any{
				{
					"key":       "TEST_VARIABLE_1",
					"value":     "test1",
					"protected": false,
					"masked":    true,
					"hidden":    true,
				},
			})
		})

	variables, resp, err := client.GroupVariables.ListVariables(1, &ListGroupVariablesOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*GroupVariable{
		{
			Key:       "TEST_VARIABLE_1",
			Value:     "test1",
			Protected: false,
			Masked:    true,
			Hidden:    true,
		},
	}

	assert.ElementsMatch(t, want, variables)
}

func TestGetGroupVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/variables/TEST_VARIABLE_1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			testParam(t, r, "filter[environment_scope]", "prod")
			mustWriteJSONResponse(t, w, map[string]any{
				"key":       "TEST_VARIABLE_1",
				"value":     "test1",
				"protected": false,
				"masked":    true,
				"hidden":    false,
			})
		})

	variable, resp, err := client.GroupVariables.GetVariable(1, "TEST_VARIABLE_1", &GetGroupVariableOptions{Filter: &VariableFilter{EnvironmentScope: "prod"}})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Value: "test1", Protected: false, Masked: true, Hidden: false}
	assert.Equal(t, want, variable)
}

func TestCreateGroupVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			mustWriteJSONResponse(t, w, map[string]any{
				"key":       "TEST_VARIABLE_1",
				"value":     "test1",
				"protected": false,
				"masked":    true,
				"hidden":    false,
			})
		})

	opt := &CreateGroupVariableOptions{
		Key:             Ptr("TEST_VARIABLE_1"),
		Value:           Ptr("test1"),
		Protected:       Ptr(false),
		Masked:          Ptr(true),
		MaskedAndHidden: Ptr(false),
	}

	variable, resp, err := client.GroupVariables.CreateVariable(1, opt, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Value: "test1", Protected: false, Masked: true, Hidden: false}
	assert.Equal(t, want, variable)
}

func TestCreateGroupVariable_MaskedAndHidden(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			mustWriteJSONResponse(t, w, map[string]any{
				"key":       "TEST_VARIABLE_1",
				"protected": false,
				"masked":    true,
				"hidden":    true,
			})
		})

	opt := &CreateGroupVariableOptions{
		Key:             Ptr("TEST_VARIABLE_1"),
		Value:           Ptr("test1"),
		Protected:       Ptr(false),
		Masked:          Ptr(true),
		MaskedAndHidden: Ptr(true),
	}

	variable, resp, err := client.GroupVariables.CreateVariable(1, opt, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Protected: false, Masked: true, Hidden: true}
	assert.Equal(t, want, variable)
}

func TestDeleteGroupVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/variables/TEST_VARIABLE_1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodDelete)
			w.WriteHeader(http.StatusAccepted)
		})

	resp, err := client.GroupVariables.RemoveVariable(1, "TEST_VARIABLE_1", &RemoveGroupVariableOptions{Filter: &VariableFilter{EnvironmentScope: "prod"}})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := http.StatusAccepted
	got := resp.StatusCode
	assert.Equal(t, want, got)
}

func TestUpdateGroupVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/variables/TEST_VARIABLE_1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPut)
			mustWriteJSONResponse(t, w, map[string]any{
				"key":       "TEST_VARIABLE_1",
				"value":     "test1",
				"protected": false,
				"masked":    true,
				"hidden":    false,
			})
		})

	variable, resp, err := client.GroupVariables.UpdateVariable(1, "TEST_VARIABLE_1", &UpdateGroupVariableOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Value: "test1", Protected: false, Masked: true, Hidden: false}
	assert.Equal(t, want, variable)
}

func TestUpdateGroupVariable_Filter(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/variables/TEST_VARIABLE_1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPut)
			mustWriteJSONResponse(t, w, map[string]any{
				"key":       "TEST_VARIABLE_1",
				"value":     "test1",
				"protected": false,
				"masked":    true,
				"hidden":    false,
			})
		})

	variable, resp, err := client.GroupVariables.UpdateVariable(1, "TEST_VARIABLE_1", &UpdateGroupVariableOptions{Filter: &VariableFilter{EnvironmentScope: "prod"}})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Value: "test1", Protected: false, Masked: true, Hidden: false}
	assert.Equal(t, want, variable)
}

func TestUpdateGroupVariable_MaskedAndHidden(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/variables/TEST_VARIABLE_1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPut)
			mustWriteJSONResponse(t, w, map[string]any{
				"key":       "TEST_VARIABLE_1",
				"protected": false,
				"masked":    true,
				"hidden":    true,
			})
		})

	variable, resp, err := client.GroupVariables.UpdateVariable(1, "TEST_VARIABLE_1", &UpdateGroupVariableOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Protected: false, Masked: true, Hidden: true}
	assert.Equal(t, want, variable)
}
