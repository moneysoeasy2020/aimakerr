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

func TestGetGlobalSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/notification_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{
			"level": "participating",
			"notification_email": "admin@example.com"
		  }`)
	})

	settings, resp, err := client.NotificationSettings.GetGlobalSettings()
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &NotificationSettings{
		Level:             1,
		NotificationEmail: "admin@example.com",
	}
	assert.Equal(t, want, settings)
}

func TestGetProjectSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/notification_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{
		"level":"custom",
		"events":{
			"new_note":true,
			"new_issue":true,
			"reopen_issue":true,
			"close_issue":true,
			"reassign_issue":true,
			"issue_due":true,
			"new_merge_request":true,
			"push_to_merge_request":true,
			"reopen_merge_request":true,
			"close_merge_request":true,
			"reassign_merge_request":true,
			"merge_merge_request":true,
			"failed_pipeline":true,
			"fixed_pipeline":true,
			"success_pipeline":true,
			"moved_project":true,
			"merge_when_pipeline_succeeds":true,
			"new_epic":true
			}
		}`)
	})

	settings, resp, err := client.NotificationSettings.GetSettingsForProject(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &NotificationSettings{
		Level: 5, // custom
		Events: &NotificationEvents{
			NewEpic:                   true,
			NewNote:                   true,
			NewIssue:                  true,
			ReopenIssue:               true,
			CloseIssue:                true,
			ReassignIssue:             true,
			IssueDue:                  true,
			NewMergeRequest:           true,
			PushToMergeRequest:        true,
			ReopenMergeRequest:        true,
			CloseMergeRequest:         true,
			ReassignMergeRequest:      true,
			MergeMergeRequest:         true,
			FailedPipeline:            true,
			FixedPipeline:             true,
			SuccessPipeline:           true,
			MovedProject:              true,
			MergeWhenPipelineSucceeds: true,
		},
	}
	assert.Equal(t, want, settings)
}

func TestUpdateProjectSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	customLevel := notificationLevelTypes["custom"]
	options := NotificationSettingsOptions{
		Level:        &customLevel,
		NewEpic:      Ptr(true),
		MovedProject: Ptr(true),
		CloseIssue:   Ptr(true),
	}

	// Handle the request on the server, and return a fully hydrated response
	mux.HandleFunc("/api/v4/projects/1/notification_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `{
		"level":"custom",
		"events":{
			"new_note":true,
			"new_issue":true,
			"reopen_issue":true,
			"close_issue":true,
			"reassign_issue":true,
			"issue_due":true,
			"new_merge_request":true,
			"push_to_merge_request":true,
			"reopen_merge_request":true,
			"close_merge_request":true,
			"reassign_merge_request":true,
			"merge_merge_request":true,
			"failed_pipeline":true,
			"fixed_pipeline":true,
			"success_pipeline":true,
			"moved_project":true,
			"merge_when_pipeline_succeeds":true,
			"new_epic":true
			}
		}`)
	})

	// Make the actual request
	settings, resp, err := client.NotificationSettings.UpdateSettingsForProject(1, &options)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Test the response and the request
	wantResponse := &NotificationSettings{
		Level: customLevel,
		Events: &NotificationEvents{
			NewEpic:                   true,
			NewNote:                   true,
			NewIssue:                  true,
			ReopenIssue:               true,
			CloseIssue:                true,
			ReassignIssue:             true,
			IssueDue:                  true,
			NewMergeRequest:           true,
			PushToMergeRequest:        true,
			ReopenMergeRequest:        true,
			CloseMergeRequest:         true,
			ReassignMergeRequest:      true,
			MergeMergeRequest:         true,
			FailedPipeline:            true,
			FixedPipeline:             true,
			SuccessPipeline:           true,
			MovedProject:              true,
			MergeWhenPipelineSucceeds: true,
		},
	}

	assert.Equal(t, wantResponse, settings)
}

func TestUpdateGlobalSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	options := NotificationSettingsOptions{
		NotificationEmail: Ptr("admin@example.com"),
	}
	mux.HandleFunc("/api/v4/notification_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `{
            "level": "participating",
            "notification_email": "admin@example.com"
        }`)
	})

	settings, resp, err := client.NotificationSettings.UpdateGlobalSettings(&options)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "admin@example.com", settings.NotificationEmail)
}

func TestGetSettingsForGroup(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/notification_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{
            "level": "participating",
            "notification_email": "group@example.com"
        }`)
	})

	settings, resp, err := client.NotificationSettings.GetSettingsForGroup(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "group@example.com", settings.NotificationEmail)
}

func TestUpdateSettingsForGroup(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	options := NotificationSettingsOptions{
		NotificationEmail: Ptr("group@example.com"),
	}
	mux.HandleFunc("/api/v4/groups/1/notification_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `{
            "level": "participating",
            "notification_email": "group@example.com"
        }`)
	})

	settings, resp, err := client.NotificationSettings.UpdateSettingsForGroup(1, &options)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "group@example.com", settings.NotificationEmail)
}
