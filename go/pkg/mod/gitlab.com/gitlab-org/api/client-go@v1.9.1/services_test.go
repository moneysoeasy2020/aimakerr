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

func TestListServices(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})
	want := []*Service{{ID: 1}, {ID: 2}}

	services, resp, err := client.Services.ListServices(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, services)
}

func TestCustomIssueTrackerService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/custom-issue-tracker", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
      "id": 1,
      "title": "5",
      "push_events": true,
      "properties": {
        "new_issue_url": "1",
        "issues_url": "2",
        "project_url": "3"
      }
    }`)
	})
	want := &CustomIssueTrackerService{
		Service: Service{
			ID:         1,
			Title:      "5",
			PushEvents: true,
		},
		Properties: &CustomIssueTrackerServiceProperties{
			NewIssueURL: "1",
			IssuesURL:   "2",
			ProjectURL:  "3",
		},
	}

	service, resp, err := client.Services.GetCustomIssueTrackerService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetCustomIssueTrackerService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/custom-issue-tracker", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetCustomIssueTrackerServiceOptions{
		NewIssueURL: Ptr("1"),
		IssuesURL:   Ptr("2"),
		ProjectURL:  Ptr("3"),
	}

	_, resp, err := client.Services.SetCustomIssueTrackerService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteCustomIssueTrackerService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/custom-issue-tracker", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteCustomIssueTrackerService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetDataDogService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/datadog", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
      "id": 1,
      "active": true,
      "properties": {
        "api_url": "",
        "datadog_env": "production",
        "datadog_service": "gitlab",
        "datadog_site": "datadoghq.com",
        "datadog_tags": "country=canada\nprovince=ontario",
        "archive_trace_events": true
      }
    }`)
	})
	want := &DataDogService{
		Service: Integration{ID: 1, Active: true},
		Properties: &DataDogServiceProperties{
			APIURL:             "",
			DataDogEnv:         "production",
			DataDogService:     "gitlab",
			DataDogSite:        "datadoghq.com",
			DataDogTags:        "country=canada\nprovince=ontario",
			ArchiveTraceEvents: true,
		},
	}

	service, resp, err := client.Services.GetDataDogService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetDataDogService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/datadog", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetDataDogServiceOptions{
		APIKey:             Ptr("secret"),
		APIURL:             Ptr("https://some-api.com"),
		DataDogEnv:         Ptr("sandbox"),
		DataDogService:     Ptr("source-code"),
		DataDogSite:        Ptr("datadoghq.eu"),
		DataDogTags:        Ptr("country=france"),
		ArchiveTraceEvents: Ptr(false),
	}

	_, resp, err := client.Services.SetDataDogService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteDataDogService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/datadog", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteDataDogService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetDiscordService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/discord", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &DiscordService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetDiscordService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetDiscordService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/discord", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetDiscordServiceOptions{
		WebHook: Ptr("webhook_uri"),
	}

	_, resp, err := client.Services.SetDiscordService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteDiscordService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/discord", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteDiscordService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetDroneCIService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/drone-ci", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &DroneCIService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetDroneCIService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetDroneCIService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/drone-ci", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetDroneCIServiceOptions{Ptr("token"), Ptr("drone-url"), Ptr(true), nil, nil, nil}

	_, resp, err := client.Services.SetDroneCIService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteDroneCIService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/drone-ci", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteDroneCIService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetEmailsOnPushService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/emails-on-push", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &EmailsOnPushService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetEmailsOnPushService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetEmailsOnPushService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/emails-on-push", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetEmailsOnPushServiceOptions{Ptr("t"), Ptr(true), Ptr(true), Ptr(true), Ptr(true), Ptr("t")}

	_, resp, err := client.Services.SetEmailsOnPushService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteEmailsOnPushService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/emails-on-push", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteEmailsOnPushService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetHarborService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/harbor", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &HarborService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetHarborService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetHarborService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/harbor", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetHarborServiceOptions{
		URL:                  Ptr("url"),
		ProjectName:          Ptr("project"),
		Username:             Ptr("user"),
		Password:             Ptr("pass"),
		UseInheritedSettings: Ptr(false),
	}

	_, resp, err := client.Services.SetHarborService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteHarborService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/harbor", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteHarborService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetSlackApplication(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/gitlab-slack-application", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &SlackApplication{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetSlackApplication(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetSlackApplication(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/gitlab-slack-application", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetSlackApplicationOptions{Channel: Ptr("#channel1"), NoteEvents: Ptr(true), AlertEvents: Ptr(true)}

	_, resp, err := client.Services.SetSlackApplication(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDisableSlackApplication(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/gitlab-slack-application", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DisableSlackApplication(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetJiraService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/0/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	mux.HandleFunc("/api/v4/projects/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1, "properties": {"jira_issue_transition_id": "2"}}`)
	})

	mux.HandleFunc("/api/v4/projects/2/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1, "properties": {"jira_issue_transition_id": 2}}`)
	})

	mux.HandleFunc("/api/v4/projects/3/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1, "properties": {"jira_issue_transition_id": "2,3"}}`)
	})

	mux.HandleFunc("/api/v4/projects/4/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1, "properties": {"jira_auth_type": 1}}`)
	})

	want := []*JiraService{
		{
			Service:    Service{ID: 1},
			Properties: &JiraServiceProperties{},
		},
		{
			Service: Service{ID: 1},
			Properties: &JiraServiceProperties{
				JiraIssueTransitionID: "2",
			},
		},
		{
			Service: Service{ID: 1},
			Properties: &JiraServiceProperties{
				JiraIssueTransitionID: "2",
			},
		},
		{
			Service: Service{ID: 1},
			Properties: &JiraServiceProperties{
				JiraIssueTransitionID: "2,3",
			},
		},
		{
			Service: Service{ID: 1},
			Properties: &JiraServiceProperties{
				JiraAuthType: 1,
			},
		},
	}

	for testcase := range want {
		service, resp, err := client.Services.GetJiraService(testcase)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, want[testcase], service)
	}
}

func TestSetJiraService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetJiraServiceOptions{
		URL:                          Ptr("asd"),
		APIURL:                       Ptr("asd"),
		Username:                     Ptr("aas"),
		Password:                     Ptr("asd"),
		Active:                       Ptr(true),
		JiraIssuePrefix:              Ptr("ASD"),
		JiraIssueRegex:               Ptr("ASD"),
		JiraIssueTransitionAutomatic: Ptr(true),
		JiraIssueTransitionID:        Ptr("2,3"),
		CommitEvents:                 Ptr(true),
		MergeRequestsEvents:          Ptr(true),
		CommentOnEventEnabled:        Ptr(true),
		IssuesEnabled:                Ptr(true),
		UseInheritedSettings:         Ptr(true),
	}

	_, resp, err := client.Services.SetJiraService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSetJiraServiceProjectKeys(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetJiraServiceOptions{
		URL:                          Ptr("asd"),
		APIURL:                       Ptr("asd"),
		Username:                     Ptr("aas"),
		Password:                     Ptr("asd"),
		Active:                       Ptr(true),
		JiraIssuePrefix:              Ptr("ASD"),
		JiraIssueRegex:               Ptr("ASD"),
		JiraIssueTransitionAutomatic: Ptr(true),
		JiraIssueTransitionID:        Ptr("2,3"),
		CommitEvents:                 Ptr(true),
		MergeRequestsEvents:          Ptr(true),
		CommentOnEventEnabled:        Ptr(true),
		IssuesEnabled:                Ptr(true),
		ProjectKeys:                  Ptr([]string{"as"}),
		UseInheritedSettings:         Ptr(true),
	}

	_, resp, err := client.Services.SetJiraService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSetJiraServiceAuthTypeBasicAuth(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetJiraServiceOptions{
		URL:          Ptr("asd"),
		Username:     Ptr("aas"),
		Password:     Ptr("asd"),
		JiraAuthType: Ptr(int64(0)),
	}

	_, resp, err := client.Services.SetJiraService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSetJiraServiceAuthTypeTokenAuth(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetJiraServiceOptions{
		URL:          Ptr("asd"),
		Password:     Ptr("asd"),
		JiraAuthType: Ptr(int64(1)),
	}

	_, resp, err := client.Services.SetJiraService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteJiraService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteJiraService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetMatrixService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/matrix", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})

	want := &MatrixService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetMatrixService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetMatrixService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/matrix", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testJSONBody(t, r, `{
			"hostname": "https://matrix.example.com",
			"room": "!roomid:matrix.example.com",
			"token": "mxp_test_token",
			"use_inherited_settings": false
		}`)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetMatrixServiceOptions{
		Hostname:             Ptr("https://matrix.example.com"),
		Room:                 Ptr("!roomid:matrix.example.com"),
		Token:                Ptr("mxp_test_token"),
		UseInheritedSettings: Ptr(false),
	}

	_, resp, err := client.Services.SetMatrixService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteMatrixService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/matrix", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteMatrixService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetMattermostService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/mattermost", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &MattermostService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetMattermostService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetMattermostService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/mattermost", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetMattermostServiceOptions{
		WebHook:  Ptr("webhook_uri"),
		Username: Ptr("username"),
		Channel:  Ptr("#development"),
	}

	_, resp, err := client.Services.SetMattermostService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteMattermostService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/mattermost", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteMattermostService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetMattermostSlashCommandsService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/mattermost-slash-commands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &MattermostSlashCommandsService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetMattermostSlashCommandsService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetMattermostSlashCommandsService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/mattermost-slash-commands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetMattermostSlashCommandsServiceOptions{
		Token:    Ptr("token"),
		Username: Ptr("username"),
	}

	_, resp, err := client.Services.SetMattermostSlashCommandsService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteMattermostSlashCommandsService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/mattermost-slash-commands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteMattermostSlashCommandsService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetPipelinesEmailService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/pipelines-email", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &PipelinesEmailService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetPipelinesEmailService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetPipelinesEmailService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/pipelines-email", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetPipelinesEmailServiceOptions{
		Recipients:                Ptr("test@email.com"),
		NotifyOnlyBrokenPipelines: Ptr(true),
		NotifyOnlyDefaultBranch:   Ptr(false),
		AddPusher:                 nil,
		BranchesToBeNotified:      nil,
		PipelineEvents:            nil,
	}

	_, resp, err := client.Services.SetPipelinesEmailService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeletePipelinesEmailService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/pipelines-email", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeletePipelinesEmailService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetRedmineService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/redmine", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &RedmineService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetRedmineService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetRedmineService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/redmine", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetRedmineServiceOptions{Ptr("t"), Ptr("u"), Ptr("a"), Ptr(false)}

	_, resp, err := client.Services.SetRedmineService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteRedmineService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/integrations/redmine", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteRedmineService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetSlackService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/slack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &SlackService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetSlackService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetSlackService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/slack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetSlackServiceOptions{
		WebHook:  Ptr("webhook_uri"),
		Username: Ptr("username"),
		Channel:  Ptr("#development"),
	}

	_, resp, err := client.Services.SetSlackService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteSlackService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/slack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteSlackService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetSlackSlashCommandsService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/slack-slash-commands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &SlackSlashCommandsService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetSlackSlashCommandsService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetSlackSlashCommandsService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/slack-slash-commands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetSlackSlashCommandsServiceOptions{
		Token: Ptr("token"),
	}

	_, resp, err := client.Services.SetSlackSlashCommandsService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteSlackSlashCommandsService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/slack-slash-commands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteSlackSlashCommandsService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetTelegramService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/telegram", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			{
			  "id": 1,
			  "title": "Telegram",
			  "slug": "telegram",
			  "created_at": "2023-12-16T20:21:03.117Z",
			  "updated_at": "2023-12-16T20:22:19.140Z",
			  "active": true,
			  "commit_events": true,
			  "push_events": false,
			  "issues_events": false,
			  "incident_events": false,
			  "alert_events": true,
			  "confidential_issues_events": false,
			  "merge_requests_events": false,
			  "tag_push_events": false,
			  "deployment_events": false,
			  "note_events": false,
			  "confidential_note_events": false,
			  "pipeline_events": true,
			  "wiki_page_events": false,
			  "job_events": true,
			  "comment_on_event_enabled": true,
			  "vulnerability_events": false,
			  "properties": {
				"room": "-1000000000000",
				"notify_only_broken_pipelines": false,
				"branches_to_be_notified": "all"
			  }
			}
		`)
	})

	want := &TelegramService{
		Service: Service{
			ID:                       1,
			Title:                    "Telegram",
			Slug:                     "telegram",
			CreatedAt:                mustParseTime("2023-12-16T20:21:03.117Z"),
			UpdatedAt:                mustParseTime("2023-12-16T20:22:19.140Z"),
			Active:                   true,
			CommitEvents:             true,
			PushEvents:               false,
			IssuesEvents:             false,
			AlertEvents:              true,
			ConfidentialIssuesEvents: false,
			MergeRequestsEvents:      false,
			TagPushEvents:            false,
			DeploymentEvents:         false,
			NoteEvents:               false,
			ConfidentialNoteEvents:   false,
			PipelineEvents:           true,
			WikiPageEvents:           false,
			JobEvents:                true,
			CommentOnEventEnabled:    true,
			VulnerabilityEvents:      false,
		},
		Properties: &TelegramServiceProperties{
			Room:                      "-1000000000000",
			NotifyOnlyBrokenPipelines: false,
			BranchesToBeNotified:      "all",
		},
	}

	service, resp, err := client.Services.GetTelegramService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetTelegramService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/telegram", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetTelegramServiceOptions{
		Token:                     Ptr("token"),
		Room:                      Ptr("-1000"),
		NotifyOnlyBrokenPipelines: Ptr(true),
		BranchesToBeNotified:      Ptr("all"),
		PushEvents:                Ptr(true),
		IssuesEvents:              Ptr(true),
		ConfidentialIssuesEvents:  Ptr(true),
		MergeRequestsEvents:       Ptr(true),
		TagPushEvents:             Ptr(true),
		NoteEvents:                Ptr(true),
		ConfidentialNoteEvents:    Ptr(true),
		PipelineEvents:            Ptr(true),
		WikiPageEvents:            Ptr(true),
	}

	_, resp, err := client.Services.SetTelegramService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteTelegramService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/telegram", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteTelegramService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetYouTrackService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/youtrack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &YouTrackService{Service: Service{ID: 1}}

	service, resp, err := client.Services.GetYouTrackService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, service)
}

func TestSetYouTrackService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/youtrack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	opt := &SetYouTrackServiceOptions{
		IssuesURL:   Ptr("https://example.org/youtrack/issue/:id"),
		ProjectURL:  Ptr("https://example.org/youtrack/projects/1"),
		Description: Ptr("description"),
		PushEvents:  Ptr(true),
	}

	_, resp, err := client.Services.SetYouTrackService(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteYouTrackService(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/services/youtrack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Services.DeleteYouTrackService(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
