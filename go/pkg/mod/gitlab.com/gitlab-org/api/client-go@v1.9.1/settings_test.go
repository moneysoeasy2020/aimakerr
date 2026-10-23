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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/application/settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1,    "default_projects_limit" : 100000}`)
	})

	settings, _, err := client.Settings.GetSettings()
	if err != nil {
		t.Fatal(err)
	}

	want := &Settings{ID: 1, DefaultProjectsLimit: 100000}
	if !reflect.DeepEqual(settings, want) {
		t.Errorf("Settings.GetSettings returned %+v, want %+v", settings, want)
	}
}

func TestUpdateSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/application/settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"default_projects_limit" : 100}`)
	})

	options := &UpdateSettingsOptions{
		DefaultProjectsLimit: Ptr(int64(100)),
	}
	settings, _, err := client.Settings.UpdateSettings(options)
	if err != nil {
		t.Fatal(err)
	}

	want := &Settings{DefaultProjectsLimit: 100}
	if !reflect.DeepEqual(settings, want) {
		t.Errorf("Settings.UpdateSettings returned %+v, want %+v", settings, want)
	}
}

func TestSettingsWithEmptyContainerRegistry(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/application/settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1, "container_registry_import_created_before": ""}`)
	})

	settings, _, err := client.Settings.GetSettings()
	if err != nil {
		t.Fatal(err)
	}

	want := &Settings{ID: 1, ContainerRegistryImportCreatedBefore: nil}
	if !reflect.DeepEqual(settings, want) {
		t.Errorf("Settings.UpdateSettings returned %+v, want %+v", settings, want)
	}
}

func TestSettingsDefaultBranchProtectionDefaults(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	var requestBody map[string]any
	mux.HandleFunc("/api/v4/application/settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		// Read the request body into `requestBody` by unmarshalling it
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprint(w, `{"id":1,    "default_projects_limit" : 100000}`)
	})

	_, _, err := client.Settings.UpdateSettings(&UpdateSettingsOptions{
		DefaultBranchProtectionDefaults: &DefaultBranchProtectionDefaultsOptions{
			AllowedToPush: &[]*GroupAccessLevel{
				{AccessLevel: Ptr(DeveloperPermissions)},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// This is the payload that should be produced. Float vs int won't matter when converted to a JSON string, so don't bother investigating why
	// it uses float instead of int when unmarshalled.
	want := map[string]any{
		"default_branch_protection_defaults": map[string]any{
			"allowed_to_push": []any{
				map[string]any{"access_level": float64(30)},
			},
		},
	}

	assert.Equal(t, want["default_branch_protection_defaults"], requestBody["default_branch_protection_defaults"])
}

func TestSettings_RequestBody(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	var requestBody map[string]any
	mux.HandleFunc("/api/v4/application/settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		// Read the request body into `requestBody` by unmarshalling it
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprint(w, `{"id":1, "default_projects_limit" : 100000, "enforce_ci_inbound_job_token_scope_enabled": true}`)
	})

	_, _, err := client.Settings.UpdateSettings(&UpdateSettingsOptions{
		EnforceCIInboundJobTokenScopeEnabled: Ptr(true),
	})
	if err != nil {
		t.Fatal(err)
	}

	// This is the payload that should be produced. This allows us to test that the request produced matches our options input.
	want := map[string]any{
		"enforce_ci_inbound_job_token_scope_enabled": true,
	}

	assert.Equal(t, want["enforce_ci_inbound_job_token_scope_enabled"], requestBody["enforce_ci_inbound_job_token_scope_enabled"])
}
