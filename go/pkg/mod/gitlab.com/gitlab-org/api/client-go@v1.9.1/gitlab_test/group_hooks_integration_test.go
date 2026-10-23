//go:build integration

package gitlab_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Integration tests for the Group Hooks API.
// These tests require a GitLab instance running on localhost:8095.
// They also require a valid admin token in GITLAB_TOKEN environment variable.

func Test_GroupHooksListGroupHooks_Integration(t *testing.T) {
	client := SetupIntegrationClient(t)

	group := CreateTestGroup(t, client)
	hook, err := CreateTestGroupHook(t, group.ID, client)
	require.NoError(t, err, "Failed to create test hook")

	hooks, _, err := client.Groups.ListGroupHooks(group.ID, &gitlab.ListGroupHooksOptions{})
	require.NoError(t, err, "Failed to list group hooks")

	assert.NotNil(t, hooks)
	assert.GreaterOrEqual(t, len(hooks), 1)
	assert.Equal(t, hook.ID, hooks[0].ID)
}

func Test_GroupHooksGetGroupHook_Integration(t *testing.T) {
	client := SetupIntegrationClient(t)

	group := CreateTestGroup(t, client)
	hook, err := CreateTestGroupHook(t, group.ID, client)
	require.NoError(t, err, "Failed to create test hook")

	retrievedHook, _, err := client.Groups.GetGroupHook(group.ID, hook.ID)
	require.NoError(t, err, "Failed to get group hook")

	assert.Equal(t, hook.ID, retrievedHook.ID)
	assert.Equal(t, hook.URL, retrievedHook.URL)
	assert.True(t, retrievedHook.PushEvents)
}

func Test_GroupHooksAddGroupHook_Integration(t *testing.T) {
	client := SetupIntegrationClient(t)

	group := CreateTestGroup(t, client)

	suffix := time.Now().UnixNano()
	name := fmt.Sprintf("testhook%d", suffix)
	description := fmt.Sprintf("Test Hook %d", suffix)
	hookURL := fmt.Sprintf("https://example.com/%d", suffix)

	hook, _, err := client.Groups.AddGroupHook(group.ID, &gitlab.AddGroupHookOptions{
		URL:                       &hookURL,
		Name:                      &name,
		Description:               &description,
		PushEvents:                gitlab.Ptr(true),
		PushEventsBranchFilter:    gitlab.Ptr("main"),
		IssuesEvents:              gitlab.Ptr(true),
		ConfidentialIssuesEvents:  gitlab.Ptr(true),
		MergeRequestsEvents:       gitlab.Ptr(true),
		TagPushEvents:             gitlab.Ptr(true),
		NoteEvents:                gitlab.Ptr(true),
		ConfidentialNoteEvents:    gitlab.Ptr(true),
		JobEvents:                 gitlab.Ptr(true),
		PipelineEvents:            gitlab.Ptr(true),
		ProjectEvents:             gitlab.Ptr(true),
		WikiPageEvents:            gitlab.Ptr(true),
		DeploymentEvents:          gitlab.Ptr(true),
		FeatureFlagEvents:         gitlab.Ptr(true),
		ReleasesEvents:            gitlab.Ptr(true),
		MilestoneEvents:           gitlab.Ptr(true),
		SubGroupEvents:            gitlab.Ptr(true),
		EmojiEvents:               gitlab.Ptr(true),
		MemberEvents:              gitlab.Ptr(true),
		VulnerabilityEvents:       gitlab.Ptr(true),
		EnableSSLVerification:     gitlab.Ptr(true),
		Token:                     gitlab.Ptr("secret-token"),
		ResourceAccessTokenEvents: gitlab.Ptr(true),
	})
	require.NoError(t, err, "Failed to add group hook")

	assert.NotZero(t, hook.ID)
	assert.Equal(t, hookURL, hook.URL)
	assert.Equal(t, name, hook.Name)
	assert.Equal(t, description, hook.Description)
	assert.True(t, hook.PushEvents)
	assert.Equal(t, "main", hook.PushEventsBranchFilter)
	assert.True(t, hook.IssuesEvents)
	assert.True(t, hook.ConfidentialIssuesEvents)
	assert.True(t, hook.MergeRequestsEvents)
	assert.True(t, hook.TagPushEvents)
	assert.True(t, hook.NoteEvents)
	assert.True(t, hook.ConfidentialNoteEvents)
	assert.True(t, hook.JobEvents)
	assert.True(t, hook.PipelineEvents)
	assert.True(t, hook.ProjectEvents)
	assert.True(t, hook.WikiPageEvents)
	assert.True(t, hook.DeploymentEvents)
	assert.True(t, hook.FeatureFlagEvents)
	assert.True(t, hook.ReleasesEvents)
	assert.True(t, hook.MilestoneEvents)
	assert.True(t, hook.SubGroupEvents)
	assert.True(t, hook.EmojiEvents)
	assert.True(t, hook.MemberEvents)
	assert.True(t, hook.VulnerabilityEvents)
	assert.True(t, hook.EnableSSLVerification)
	assert.True(t, hook.ResourceAccessTokenEvents)
}

func Test_GroupHooksEditGroupHook_Integration(t *testing.T) {
	client := SetupIntegrationClient(t)

	group := CreateTestGroup(t, client)

	hook, err := CreateTestGroupHook(t, group.ID, client)
	require.NoError(t, err, "Failed to create test hook")
	suffix := time.Now().UnixNano()
	name := fmt.Sprintf("testhook%d", suffix)
	description := fmt.Sprintf("Test Hook %d", suffix)
	hookURL := fmt.Sprintf("https://example.com/%d", suffix)

	updatedHook, _, err := client.Groups.EditGroupHook(group.ID, hook.ID, &gitlab.EditGroupHookOptions{
		URL:                       &hookURL,
		Name:                      &name,
		Description:               &description,
		PushEvents:                gitlab.Ptr(true),
		PushEventsBranchFilter:    gitlab.Ptr("main"),
		IssuesEvents:              gitlab.Ptr(true),
		ConfidentialIssuesEvents:  gitlab.Ptr(true),
		MergeRequestsEvents:       gitlab.Ptr(true),
		TagPushEvents:             gitlab.Ptr(true),
		NoteEvents:                gitlab.Ptr(true),
		ConfidentialNoteEvents:    gitlab.Ptr(true),
		JobEvents:                 gitlab.Ptr(true),
		PipelineEvents:            gitlab.Ptr(true),
		ProjectEvents:             gitlab.Ptr(true),
		WikiPageEvents:            gitlab.Ptr(true),
		DeploymentEvents:          gitlab.Ptr(true),
		FeatureFlagEvents:         gitlab.Ptr(true),
		ReleasesEvents:            gitlab.Ptr(true),
		MilestoneEvents:           gitlab.Ptr(true),
		SubGroupEvents:            gitlab.Ptr(true),
		EmojiEvents:               gitlab.Ptr(true),
		MemberEvents:              gitlab.Ptr(true),
		VulnerabilityEvents:       gitlab.Ptr(true),
		EnableSSLVerification:     gitlab.Ptr(true),
		Token:                     gitlab.Ptr("secret-token"),
		ResourceAccessTokenEvents: gitlab.Ptr(true),
	})
	require.NoError(t, err, "Failed to edit group hook")

	assert.NotZero(t, updatedHook.ID)
	assert.Equal(t, hookURL, updatedHook.URL)
	assert.Equal(t, name, updatedHook.Name)
	assert.Equal(t, description, updatedHook.Description)
	assert.True(t, updatedHook.PushEvents)
	assert.Equal(t, "main", updatedHook.PushEventsBranchFilter)
	assert.True(t, updatedHook.IssuesEvents)
	assert.True(t, updatedHook.ConfidentialIssuesEvents)
	assert.True(t, updatedHook.MergeRequestsEvents)
	assert.True(t, updatedHook.TagPushEvents)
	assert.True(t, updatedHook.NoteEvents)
	assert.True(t, updatedHook.ConfidentialNoteEvents)
	assert.True(t, updatedHook.JobEvents)
	assert.True(t, updatedHook.PipelineEvents)
	assert.True(t, updatedHook.ProjectEvents)
	assert.True(t, updatedHook.WikiPageEvents)
	assert.True(t, updatedHook.DeploymentEvents)
	assert.True(t, updatedHook.FeatureFlagEvents)
	assert.True(t, updatedHook.ReleasesEvents)
	assert.True(t, updatedHook.MilestoneEvents)
	assert.True(t, updatedHook.SubGroupEvents)
	assert.True(t, updatedHook.EmojiEvents)
	assert.True(t, updatedHook.MemberEvents)
	assert.True(t, updatedHook.VulnerabilityEvents)
	assert.True(t, updatedHook.EnableSSLVerification)
	assert.True(t, updatedHook.ResourceAccessTokenEvents)
}

func Test_GroupHooksDeleteGroupHook_Integration(t *testing.T) {
	client := SetupIntegrationClient(t)

	group := CreateTestGroup(t, client)

	hook, err := CreateTestGroupHook(t, group.ID, client)
	require.NoError(t, err, "Failed to create test hook")

	_, err = client.Groups.DeleteGroupHook(group.ID, hook.ID)
	require.NoError(t, err, "Failed to delete group hook")
}
