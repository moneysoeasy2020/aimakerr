package gitlab

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListActiveGroupIntegrations(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
			{
				"id": 1,
				"title": "Microsoft Teams",
				"slug": "microsoft_teams",
				"created_at": "2023-01-01T00:00:00.000Z",
				"updated_at": "2023-01-02T00:00:00.000Z",
				"active": true,
				"commit_events": true,
				"push_events": true,
				"issues_events": true,
				"alert_events": false,
				"confidential_issues_events": false,
				"merge_requests_events": true,
				"tag_push_events": true,
				"deployment_events": false,
				"note_events": true,
				"confidential_note_events": false,
				"pipeline_events": true,
				"wiki_page_events": false,
				"job_events": false,
				"comment_on_event_enabled": true,
				"inherited": false,
				"vulnerability_events": false
			},
			{
				"id": 2,
				"title": "Slack",
				"slug": "slack",
				"created_at": "2023-01-03T00:00:00.000Z",
				"updated_at": "2023-01-04T00:00:00.000Z",
				"active": true,
				"commit_events": false,
				"push_events": true,
				"issues_events": true,
				"alert_events": true,
				"confidential_issues_events": true,
				"merge_requests_events": true,
				"tag_push_events": false,
				"deployment_events": true,
				"note_events": false,
				"confidential_note_events": false,
				"pipeline_events": false,
				"wiki_page_events": true,
				"job_events": true,
				"comment_on_event_enabled": false,
				"inherited": true,
				"vulnerability_events": true
			}
		]`)
	})

	integrations, resp, err := client.Integrations.ListActiveGroupIntegrations(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	createdAt1, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00.000Z")
	updatedAt1, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00.000Z")
	createdAt2, _ := time.Parse(time.RFC3339, "2023-01-03T00:00:00.000Z")
	updatedAt2, _ := time.Parse(time.RFC3339, "2023-01-04T00:00:00.000Z")

	want := []*Integration{
		{
			ID:                       1,
			Title:                    "Microsoft Teams",
			Slug:                     "microsoft_teams",
			CreatedAt:                &createdAt1,
			UpdatedAt:                &updatedAt1,
			Active:                   true,
			CommitEvents:             true,
			PushEvents:               true,
			IssuesEvents:             true,
			AlertEvents:              false,
			ConfidentialIssuesEvents: false,
			MergeRequestsEvents:      true,
			TagPushEvents:            true,
			DeploymentEvents:         false,
			NoteEvents:               true,
			ConfidentialNoteEvents:   false,
			PipelineEvents:           true,
			WikiPageEvents:           false,
			JobEvents:                false,
			CommentOnEventEnabled:    true,
			Inherited:                false,
			VulnerabilityEvents:      false,
		},
		{
			ID:                       2,
			Title:                    "Slack",
			Slug:                     "slack",
			CreatedAt:                &createdAt2,
			UpdatedAt:                &updatedAt2,
			Active:                   true,
			CommitEvents:             false,
			PushEvents:               true,
			IssuesEvents:             true,
			AlertEvents:              true,
			ConfidentialIssuesEvents: true,
			MergeRequestsEvents:      true,
			TagPushEvents:            false,
			DeploymentEvents:         true,
			NoteEvents:               false,
			ConfidentialNoteEvents:   false,
			PipelineEvents:           false,
			WikiPageEvents:           true,
			JobEvents:                true,
			CommentOnEventEnabled:    false,
			Inherited:                true,
			VulnerabilityEvents:      true,
		},
	}
	assert.Equal(t, want, integrations)
	assert.NotNil(t, resp)
}

func TestSetUpGroupHarbor(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/harbor", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
			"id": 1,
			"title": "Harbor",
			"slug": "harbor",
			"created_at": "2023-01-01T00:00:00.000Z",
			"updated_at": "2023-01-02T00:00:00.000Z",
			"active": true,
			"commit_events": true,
			"push_events": true,
			"issues_events": true,
			"alert_events": false,
			"confidential_issues_events": false,
			"merge_requests_events": true,
			"tag_push_events": true,
			"deployment_events": false,
			"note_events": true,
			"confidential_note_events": false,
			"pipeline_events": true,
			"wiki_page_events": false,
			"job_events": false,
			"comment_on_event_enabled": true,
			"inherited": false,
			"vulnerability_events": false
		}`)
	})
	integration, resp, err := client.Integrations.SetUpGroupHarbor(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00.000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00.000Z")
	want := &Integration{
		ID:                       1,
		Title:                    "Harbor",
		Slug:                     "harbor",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Active:                   true,
		CommitEvents:             true,
		PushEvents:               true,
		IssuesEvents:             true,
		AlertEvents:              false,
		ConfidentialIssuesEvents: false,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		DeploymentEvents:         false,
		NoteEvents:               true,
		ConfidentialNoteEvents:   false,
		PipelineEvents:           true,
		WikiPageEvents:           false,
		JobEvents:                false,
		CommentOnEventEnabled:    true,
		Inherited:                false,
		VulnerabilityEvents:      false,
	}
	assert.Equal(t, want, integration)
}

func TestDisableGroupHarbor(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/harbor", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})
	resp, err := client.Integrations.DisableGroupHarbor(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetGroupHarborSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/harbor", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id": 1,
			"title": "Harbor",
			"slug": "harbor",
			"created_at": "2023-01-01T00:00:00.000Z",
			"updated_at": "2023-01-02T00:00:00.000Z",
			"active": true,
			"commit_events": true,
			"push_events": true,
			"issues_events": true,
			"alert_events": false,
			"confidential_issues_events": false,
			"merge_requests_events": true,
			"tag_push_events": true,
			"deployment_events": false,
			"note_events": true,
			"confidential_note_events": false,
			"pipeline_events": true,
			"wiki_page_events": false,
			"job_events": false,
			"comment_on_event_enabled": true,
			"inherited": false,
			"vulnerability_events": false
		}`)
	})
	integration, resp, err := client.Integrations.GetGroupHarborSettings(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00.000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00.000Z")
	want := &Integration{
		ID:                       1,
		Title:                    "Harbor",
		Slug:                     "harbor",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Active:                   true,
		CommitEvents:             true,
		PushEvents:               true,
		IssuesEvents:             true,
		AlertEvents:              false,
		ConfidentialIssuesEvents: false,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		DeploymentEvents:         false,
		NoteEvents:               true,
		ConfidentialNoteEvents:   false,
		PipelineEvents:           true,
		WikiPageEvents:           false,
		JobEvents:                false,
		CommentOnEventEnabled:    true,
		Inherited:                false,
		VulnerabilityEvents:      false,
	}
	assert.Equal(t, want, integration)
}

func TestSetGroupMicrosoftTeamsNotifications(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/microsoft_teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
			"id": 1,
			"title": "Microsoft Teams",
			"slug": "microsoft_teams",
			"created_at": "2023-01-01T00:00:00.000Z",
			"updated_at": "2023-01-02T00:00:00.000Z",
			"active": true,
			"commit_events": true,
			"push_events": true,
			"issues_events": true,
			"alert_events": false,
			"confidential_issues_events": false,
			"merge_requests_events": true,
			"tag_push_events": true,
			"deployment_events": false,
			"note_events": true,
			"confidential_note_events": false,
			"pipeline_events": true,
			"wiki_page_events": false,
			"job_events": false,
			"comment_on_event_enabled": true,
			"inherited": false,
			"vulnerability_events": false
		}`)
	})
	integration, resp, err := client.Integrations.SetGroupMicrosoftTeamsNotifications(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00.000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00.000Z")
	want := &Integration{
		ID:                       1,
		Title:                    "Microsoft Teams",
		Slug:                     "microsoft_teams",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Active:                   true,
		CommitEvents:             true,
		PushEvents:               true,
		IssuesEvents:             true,
		AlertEvents:              false,
		ConfidentialIssuesEvents: false,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		DeploymentEvents:         false,
		NoteEvents:               true,
		ConfidentialNoteEvents:   false,
		PipelineEvents:           true,
		WikiPageEvents:           false,
		JobEvents:                false,
		CommentOnEventEnabled:    true,
		Inherited:                false,
		VulnerabilityEvents:      false,
	}
	assert.Equal(t, want, integration)
}

func TestDisableGroupMicrosoftTeamsNotifications(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/microsoft_teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})
	resp, err := client.Integrations.DisableGroupMicrosoftTeamsNotifications(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetGroupMicrosoftTeamsNotifications(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/microsoft_teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id": 1,
			"title": "Microsoft Teams",
			"slug": "microsoft_teams",
			"created_at": "2023-01-01T00:00:00.000Z",
			"updated_at": "2023-01-02T00:00:00.000Z",
			"active": true,
			"commit_events": true,
			"push_events": true,
			"issues_events": true,
			"alert_events": false,
			"confidential_issues_events": false,
			"merge_requests_events": true,
			"tag_push_events": true,
			"deployment_events": false,
			"note_events": true,
			"confidential_note_events": false,
			"pipeline_events": true,
			"wiki_page_events": false,
			"job_events": false,
			"comment_on_event_enabled": true,
			"inherited": false,
			"vulnerability_events": false
		}`)
	})
	integration, resp, err := client.Integrations.GetGroupMicrosoftTeamsNotifications(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00.000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00.000Z")
	want := &Integration{
		ID:                       1,
		Title:                    "Microsoft Teams",
		Slug:                     "microsoft_teams",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Active:                   true,
		CommitEvents:             true,
		PushEvents:               true,
		IssuesEvents:             true,
		AlertEvents:              false,
		ConfidentialIssuesEvents: false,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		DeploymentEvents:         false,
		NoteEvents:               true,
		ConfidentialNoteEvents:   false,
		PipelineEvents:           true,
		WikiPageEvents:           false,
		JobEvents:                false,
		CommentOnEventEnabled:    true,
		Inherited:                false,
		VulnerabilityEvents:      false,
	}
	assert.Equal(t, want, integration)
}

func TestSetUpGroupJira(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
			"id": 1,
			"title": "Jira",
			"slug": "jira",
			"created_at": "2025-01-01T00:00:00.000Z",
			"updated_at": "2025-01-02T00:00:00.000Z",
			"active": true,
			"commit_events": true,
			"push_events": true,
			"issues_events": true,
			"alert_events": false,
			"confidential_issues_events": false,
			"merge_requests_events": true,
			"tag_push_events": true,
			"deployment_events": false,
			"note_events": true,
			"confidential_note_events": false,
			"pipeline_events": true,
			"wiki_page_events": false,
			"job_events": false,
			"comment_on_event_enabled": true,
			"inherited": false,
			"vulnerability_events": false
		}`)
	})
	integration, resp, err := client.Integrations.SetUpGroupJira(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	createdAt, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00.000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00.000Z")
	want := &Integration{
		ID:                       1,
		Title:                    "Jira",
		Slug:                     "jira",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Active:                   true,
		CommitEvents:             true,
		PushEvents:               true,
		IssuesEvents:             true,
		AlertEvents:              false,
		ConfidentialIssuesEvents: false,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		DeploymentEvents:         false,
		NoteEvents:               true,
		ConfidentialNoteEvents:   false,
		PipelineEvents:           true,
		WikiPageEvents:           false,
		JobEvents:                false,
		CommentOnEventEnabled:    true,
		Inherited:                false,
		VulnerabilityEvents:      false,
	}
	assert.Equal(t, want, integration)
}

func TestDisableGroupJira(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})
	resp, err := client.Integrations.DisableGroupJira(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetGroupJiraSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/integrations/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id": 1,
			"title": "Jira",
			"slug": "jira",
			"created_at": "2025-01-01T00:00:00.000Z",
			"updated_at": "2025-01-02T00:00:00.000Z",
			"active": true,
			"commit_events": true,
			"push_events": true,
			"issues_events": true,
			"alert_events": false,
			"confidential_issues_events": false,
			"merge_requests_events": true,
			"tag_push_events": true,
			"deployment_events": false,
			"note_events": true,
			"confidential_note_events": false,
			"pipeline_events": true,
			"wiki_page_events": false,
			"job_events": false,
			"comment_on_event_enabled": true,
			"inherited": false,
			"vulnerability_events": false
		}`)
	})
	integration, resp, err := client.Integrations.GetGroupJiraSettings(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	createdAt, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00.000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00.000Z")
	want := &Integration{
		ID:                       1,
		Title:                    "Jira",
		Slug:                     "jira",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Active:                   true,
		CommitEvents:             true,
		PushEvents:               true,
		IssuesEvents:             true,
		AlertEvents:              false,
		ConfidentialIssuesEvents: false,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		DeploymentEvents:         false,
		NoteEvents:               true,
		ConfidentialNoteEvents:   false,
		PipelineEvents:           true,
		WikiPageEvents:           false,
		JobEvents:                false,
		CommentOnEventEnabled:    true,
		Inherited:                false,
		VulnerabilityEvents:      false,
	}
	assert.Equal(t, want, integration)
}
