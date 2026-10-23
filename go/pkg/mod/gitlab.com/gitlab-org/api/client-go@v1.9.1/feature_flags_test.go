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

func TestListFeatureFlags(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/features", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		[
			{
			  "name": "experimental_feature",
			  "state": "off",
			  "gates": [
				{
				  "key": "boolean",
				  "value": false
				}
			  ]
			},
			{
			  "name": "new_library",
			  "state": "on"
			}
		  ]
	`)
	})

	features, resp, err := client.Features.ListFeatures()
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Feature{
		{Name: "experimental_feature", State: "off", Gates: []Gate{
			{Key: "boolean", Value: false},
		}},
		{Name: "new_library", State: "on"},
	}
	assert.Equal(t, want, features)
}

func TestListFeatureDefinitions(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/features/definitions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		[
			{
				"name": "geo_pages_deployment_replication",
				"introduced_by_url": "https://gitlab.com/gitlab-org/gitlab/-/merge_requests/68662",
				"rollout_issue_url": "https://gitlab.com/gitlab-org/gitlab/-/issues/337676",
				"milestone": "14.3",
				"log_state_changes": null,
				"type": "development",
				"group": "group::geo",
				"default_enabled": true
			}
		]
		`)
	})

	definitions, resp, err := client.Features.ListFeatureDefinitions()
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*FeatureDefinition{
		{
			Name:            "geo_pages_deployment_replication",
			IntroducedByURL: "https://gitlab.com/gitlab-org/gitlab/-/merge_requests/68662",
			RolloutIssueURL: "https://gitlab.com/gitlab-org/gitlab/-/issues/337676",
			Milestone:       "14.3",
			Type:            "development",
			Group:           "group::geo",
			DefaultEnabled:  true,
		},
	}

	assert.Equal(t, want, definitions)
}

func TestSetFeatureFlag(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/features/new_library", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
		{
			"name": "new_library",
			"state": "conditional",
			"gates": [
			  {
				"key": "boolean",
				"value": false
			  },
			  {
				"key": "percentage_of_time",
				"value": 30
			  }
			]
		  }
		`)
	})

	feature, resp, err := client.Features.SetFeatureFlag("new_library", &SetFeatureFlagOptions{
		Value:        false,
		Key:          "boolean",
		FeatureGroup: "experiment",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &Feature{
		Name:  "new_library",
		State: "conditional",
		Gates: []Gate{
			{Key: "boolean", Value: false},
			{Key: "percentage_of_time", Value: 30.0},
		},
	}
	assert.Equal(t, want, feature)
}

func TestDeleteFeatureFlag(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/features/new_library", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Features.DeleteFeatureFlag("new_library")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
