//
// Copyright 2021, Eric Stevens
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
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestListGroupHooks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/hooks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
[
	{
		"id": 1,
		"url": "http://example.com/hook",
		"group_id": 3,
		"push_events": true,
		"push_events_branch_filter": "main",
		"issues_events": true,
		"confidential_issues_events": true,
		"merge_requests_events": true,
		"tag_push_events": true,
		"note_events": true,
		"job_events": true,
		"pipeline_events": true,
		"wiki_page_events": true,
		"deployment_events": true,
		"releases_events": true,
		"subgroup_events": true,
		"emoji_events": true,
		"member_events": true,
		"enable_ssl_verification": true,
		"alert_status": "executable",
		"created_at": "2012-10-12T17:04:47Z",
		"resource_access_token_events": true,
		"project_events": true,
		"milestone_events": true,
		"vulnerability_events": true,
		"custom_headers": [
			{"key": "Authorization"},
			{"key": "OtherHeader"}
		]
	}
]`)
	})

	groupHooks, resp, err := client.Groups.ListGroupHooks(1, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	datePointer := time.Date(2012, time.October, 12, 17, 4, 47, 0, time.UTC)
	want := []*GroupHook{{
		ID:                        1,
		URL:                       "http://example.com/hook",
		GroupID:                   3,
		PushEvents:                true,
		PushEventsBranchFilter:    "main",
		IssuesEvents:              true,
		ConfidentialIssuesEvents:  true,
		MergeRequestsEvents:       true,
		TagPushEvents:             true,
		NoteEvents:                true,
		JobEvents:                 true,
		PipelineEvents:            true,
		WikiPageEvents:            true,
		DeploymentEvents:          true,
		ReleasesEvents:            true,
		SubGroupEvents:            true,
		EmojiEvents:               true,
		MemberEvents:              true,
		EnableSSLVerification:     true,
		AlertStatus:               "executable",
		CreatedAt:                 &datePointer,
		ResourceAccessTokenEvents: true,
		ProjectEvents:             true,
		MilestoneEvents:           true,
		VulnerabilityEvents:       true,
		CustomHeaders: []*HookCustomHeader{
			{
				Key: "Authorization",
			},
			{
				Key: "OtherHeader",
			},
		},
	}}

	if !reflect.DeepEqual(groupHooks, want) {
		t.Errorf("listGroupHooks returned \ngot:\n%v\nwant:\n%v", Stringify(groupHooks), Stringify(want))
	}
}

func TestGetGroupHook(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
{
	"id": 1,
	"url": "http://example.com/hook",
	"group_id": 3,
	"push_events": true,
	"push_events_branch_filter": "main",
	"issues_events": true,
	"confidential_issues_events": true,
	"merge_requests_events": true,
	"tag_push_events": true,
	"note_events": true,
	"job_events": true,
	"pipeline_events": true,
	"wiki_page_events": true,
	"deployment_events": true,
	"releases_events": true,
	"subgroup_events": true,
	"emoji_events": true,
	"member_events": true,
	"enable_ssl_verification": true,
	"alert_status": "executable",
	"created_at": "2012-10-12T17:04:47Z",
	"resource_access_token_events": true,
	"project_events": true,
	"milestone_events": true,
	"vulnerability_events": true,
	"custom_headers": [
		{"key": "Authorization"},
		{"key": "OtherHeader"}
	]
}`)
	})

	groupHook, resp, err := client.Groups.GetGroupHook(1, 1)
	require.NoError(t, err)
	require.NotNil(t, resp)

	datePointer := time.Date(2012, time.October, 12, 17, 4, 47, 0, time.UTC)
	want := &GroupHook{
		ID:                        1,
		URL:                       "http://example.com/hook",
		GroupID:                   3,
		PushEvents:                true,
		PushEventsBranchFilter:    "main",
		IssuesEvents:              true,
		ConfidentialIssuesEvents:  true,
		MergeRequestsEvents:       true,
		TagPushEvents:             true,
		NoteEvents:                true,
		JobEvents:                 true,
		PipelineEvents:            true,
		WikiPageEvents:            true,
		DeploymentEvents:          true,
		ReleasesEvents:            true,
		SubGroupEvents:            true,
		EmojiEvents:               true,
		MemberEvents:              true,
		EnableSSLVerification:     true,
		AlertStatus:               "executable",
		CreatedAt:                 &datePointer,
		ResourceAccessTokenEvents: true,
		ProjectEvents:             true,
		MilestoneEvents:           true,
		VulnerabilityEvents:       true,
		CustomHeaders: []*HookCustomHeader{
			{
				Key: "Authorization",
			},
			{
				Key: "OtherHeader",
			},
		},
	}

	if !reflect.DeepEqual(groupHook, want) {
		t.Errorf("getGroupHooks returned \ngot:\n%v\nwant:\n%v", Stringify(groupHook), Stringify(want))
	}
}

func TestResendGroupHookEvent(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/hooks/1/events/1/resend", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"response_status": 200}`)
	})

	resp, err := client.Groups.ResendGroupHookEvent(1, 1, 1)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestAddGroupHook(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/hooks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
{
	"id": 1,
	"url": "http://example.com/hook",
	"group_id": 3,
	"push_events": true,
	"push_events_branch_filter": "main",
	"issues_events": true,
	"confidential_issues_events": true,
	"merge_requests_events": true,
	"tag_push_events": true,
	"note_events": true,
	"job_events": true,
	"pipeline_events": true,
	"wiki_page_events": true,
	"deployment_events": true,
	"releases_events": true,
	"subgroup_events": true,
	"emoji_events": true,
	"member_events": true,
	"enable_ssl_verification": true,
	"created_at": "2012-10-12T17:04:47Z",
	"custom_webhook_template": "addTestValue",
	"resource_access_token_events": true,
	"project_events": true,
	"milestone_events": true,
	"vulnerability_events": true,
	"custom_headers": [
		{"key": "Authorization", "value": "testMe"},
		{"key": "OtherHeader", "value": "otherTest"}
	]
}`)
	})

	url := "http://www.example.com/hook"
	opt := &AddGroupHookOptions{
		URL: &url,
	}

	groupHooks, resp, err := client.Groups.AddGroupHook(1, opt)
	require.NoError(t, err)
	require.NotNil(t, resp)

	datePointer := time.Date(2012, time.October, 12, 17, 4, 47, 0, time.UTC)
	want := &GroupHook{
		ID:                        1,
		URL:                       "http://example.com/hook",
		GroupID:                   3,
		PushEvents:                true,
		PushEventsBranchFilter:    "main",
		IssuesEvents:              true,
		ConfidentialIssuesEvents:  true,
		ConfidentialNoteEvents:    false,
		MergeRequestsEvents:       true,
		TagPushEvents:             true,
		NoteEvents:                true,
		JobEvents:                 true,
		PipelineEvents:            true,
		WikiPageEvents:            true,
		DeploymentEvents:          true,
		ReleasesEvents:            true,
		SubGroupEvents:            true,
		EmojiEvents:               true,
		MemberEvents:              true,
		EnableSSLVerification:     true,
		CreatedAt:                 &datePointer,
		CustomWebhookTemplate:     "addTestValue",
		ResourceAccessTokenEvents: true,
		ProjectEvents:             true,
		MilestoneEvents:           true,
		VulnerabilityEvents:       true,
		CustomHeaders: []*HookCustomHeader{
			{
				Key:   "Authorization",
				Value: "testMe",
			},
			{
				Key:   "OtherHeader",
				Value: "otherTest",
			},
		},
	}

	if !reflect.DeepEqual(groupHooks, want) {
		t.Errorf("AddGroupHook returned \ngot:\n%v\nwant:\n%v", Stringify(groupHooks), Stringify(want))
	}
}

func TestEditGroupHook(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
{
	"id": 1,
	"url": "http://example.com/hook",
	"group_id": 3,
	"push_events": true,
	"push_events_branch_filter": "main",
	"issues_events": true,
	"confidential_issues_events": true,
	"merge_requests_events": true,
	"tag_push_events": true,
	"note_events": true,
	"job_events": true,
	"pipeline_events": true,
	"wiki_page_events": true,
	"deployment_events": true,
	"releases_events": true,
	"subgroup_events": true,
	"emoji_events": true,
	"member_events": true,
	"enable_ssl_verification": true,
	"created_at": "2012-10-12T17:04:47Z",
	"custom_webhook_template": "testValue",
	"resource_access_token_events": true,
	"project_events": true,
	"milestone_events": true,
	"vulnerability_events": true,
	"custom_headers": [
		{"key": "Authorization", "value": "testMe"},
		{"key": "OtherHeader", "value": "otherTest"}
	]
}`)
	})

	url := "http://www.example.com/hook"
	opt := &EditGroupHookOptions{
		URL: &url,
	}

	groupHooks, resp, err := client.Groups.EditGroupHook(1, 1, opt)
	require.NoError(t, err)
	require.NotNil(t, resp)

	datePointer := time.Date(2012, time.October, 12, 17, 4, 47, 0, time.UTC)
	want := &GroupHook{
		ID:                        1,
		URL:                       "http://example.com/hook",
		GroupID:                   3,
		PushEvents:                true,
		PushEventsBranchFilter:    "main",
		IssuesEvents:              true,
		ConfidentialIssuesEvents:  true,
		ConfidentialNoteEvents:    false,
		MergeRequestsEvents:       true,
		TagPushEvents:             true,
		NoteEvents:                true,
		JobEvents:                 true,
		PipelineEvents:            true,
		WikiPageEvents:            true,
		DeploymentEvents:          true,
		ReleasesEvents:            true,
		SubGroupEvents:            true,
		EmojiEvents:               true,
		MemberEvents:              true,
		EnableSSLVerification:     true,
		CreatedAt:                 &datePointer,
		CustomWebhookTemplate:     "testValue",
		ResourceAccessTokenEvents: true,
		ProjectEvents:             true,
		MilestoneEvents:           true,
		VulnerabilityEvents:       true,
		CustomHeaders: []*HookCustomHeader{
			{
				Key:   "Authorization",
				Value: "testMe",
			},
			{
				Key:   "OtherHeader",
				Value: "otherTest",
			},
		},
	}

	if !reflect.DeepEqual(groupHooks, want) {
		t.Errorf("EditGroupHook returned \ngot:\n%v\nwant:\n%v", Stringify(groupHooks), Stringify(want))
	}
}

func TestDeleteGroupHook(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Groups.DeleteGroupHook(1, 1)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestTriggerTestGroupHook(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/hooks/1/test/push_events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"message":"201 Created"}`)
	})

	mux.HandleFunc("/api/v4/groups/1/hooks/1/test/invalid_trigger", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "trigger does not have a valid value"}`)
	})

	tests := []struct {
		name       string
		groupID    any
		hookID     int64
		trigger    GroupHookTrigger
		wantErr    bool
		wantStatus int
		wantErrMsg string
	}{
		{
			name:       "Valid trigger",
			groupID:    1,
			hookID:     1,
			trigger:    GroupHookTriggerPush,
			wantErr:    false,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Invalid group ID",
			groupID:    "invalid",
			hookID:     1,
			trigger:    GroupHookTriggerPush,
			wantErr:    true,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Invalid trigger type",
			groupID:    1,
			hookID:     1,
			trigger:    "invalid_trigger",
			wantErr:    true,
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "trigger does not have a valid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.Groups.TriggerTestGroupHook(tt.groupID, tt.hookID, tt.trigger)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantStatus != 0 {
					require.Equal(t, tt.wantStatus, resp.StatusCode)
				}
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

func TestSetGroupWebhookHeader(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	var bodyJSON map[string]any

	// Removed most of the arguments to keep test slim
	mux.HandleFunc("/api/v4/groups/1/hooks/1/custom_headers/Authorization", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.WriteHeader(http.StatusNoContent)

		// validate that the `value` body is sent properly
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Unable to read body properly. Error: %v", err)
		}

		// Unmarshal the body into JSON so we can check it
		_ = json.Unmarshal(body, &bodyJSON)

		fmt.Fprint(w, ``)
	})

	req, err := client.Groups.SetGroupCustomHeader(1, 1, "Authorization", &SetHookCustomHeaderOptions{Value: Ptr("testValue")})
	if err != nil {
		t.Errorf("Groups.SetGroupCustomHeader returned error: %v", err)
	}

	require.Equal(t, "testValue", bodyJSON["value"])
	require.Equal(t, http.StatusNoContent, req.StatusCode)
}

func TestDeleteGroupCustomHeader(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/groups/1/hooks/1/custom_headers/Authorization", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Groups.DeleteGroupCustomHeader(1, 1, "Authorization")
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSetGroupHookURLVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/groups/1/hooks/1/url_variables/KEY", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	resp, err := client.Groups.SetGroupHookURLVariable(1, 1, "KEY", &SetHookURLVariableOptions{Value: Ptr("VALUE")})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestDeleteGroupHookURLVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/groups/1/hooks/1/url_variables/KEY", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Groups.DeleteGroupHookURLVariable(1, 1, "KEY")
	require.NoError(t, err)
	require.NotNil(t, resp)
}
