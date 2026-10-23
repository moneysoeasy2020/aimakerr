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

func TestWebhookEventType(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest(http.MethodGet, "https://gitlab.com", nil)
	assert.NoError(t, err)

	req.Header.Set("X-Gitlab-Event", "Push Hook")

	eventType := HookEventType(req)
	assert.Equal(t, "Push Hook", string(eventType))
}

func TestWebhookEventToken(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest(http.MethodGet, "https://gitlab.com", nil)
	assert.NoError(t, err)

	req.Header.Set("X-Gitlab-Token", "798d3dd3-67f5-41df-ad19-7882cc6263bf")

	actualToken := HookEventToken(req)
	assert.Equal(t, "798d3dd3-67f5-41df-ad19-7882cc6263bf", actualToken)
}

func TestParseBuildHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/build.json")

	parsedEvent, err := ParseWebhook("Build Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*BuildEvent)
	assert.True(t, ok, "Expected BuildEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "build", event.ObjectKind)
	assert.Equal(t, int64(1977), event.BuildID)
	assert.False(t, event.BuildAllowFailure)
	assert.Equal(t, "2293ada6b400935a1378653304eaf6221e0fdb8f", event.Commit.SHA)
	assert.Equal(t, "2021-02-23T02:41:37.886Z", event.BuildCreatedAt)
}

func TestParseCommitCommentHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/note_commit.json")

	parsedEvent, err := ParseWebhook("Note Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*CommitCommentEvent)
	assert.True(t, ok, "Expected CommitCommentEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, string(NoteEventTargetType), event.ObjectKind)
	assert.Equal(t, int64(5), event.ProjectID)
	assert.Equal(t, "Commit", event.ObjectAttributes.NoteableType)
	assert.Equal(t, "cfe32cf61b73a0d5e9f13e774abde7ff789b1660", event.Commit.ID)
}

func TestParseFeatureFlagHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/feature_flag.json")

	parsedEvent, err := ParseWebhook("Feature Flag Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*FeatureFlagEvent)
	assert.True(t, ok, "Expected FeatureFlagEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "feature_flag", event.ObjectKind)
	assert.Equal(t, int64(1), event.Project.ID)
	assert.Equal(t, int64(1), event.User.ID)
	assert.Equal(t, "Administrator", event.User.Name)
	assert.Equal(t, int64(6), event.ObjectAttributes.ID)
	assert.Equal(t, "test-feature-flag", event.ObjectAttributes.Name)
	assert.Equal(t, "test-feature-flag-description", event.ObjectAttributes.Description)
	assert.True(t, event.ObjectAttributes.Active)
}

func TestParseGroupResourceAccessTokenHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/resource_access_token_group.json")

	parsedEvent, err := ParseWebhook("Resource Access Token Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*GroupResourceAccessTokenEvent)
	assert.True(t, ok, "Expected GroupResourceAccessTokenEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "expiring_access_token", event.EventName)
}

func TestParseHookWebHook(t *testing.T) {
	t.Parallel()
	parsedEvent1, err := ParseHook("Merge Request Hook", loadFixture(t, "testdata/webhooks/merge_request.json"))
	assert.NoError(t, err)

	parsedEvent2, err := ParseWebhook("Merge Request Hook", loadFixture(t, "testdata/webhooks/merge_request.json"))
	assert.NoError(t, err)

	assert.Equal(t, parsedEvent1, parsedEvent2)
}

func TestParseIssueCommentHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/note_issue.json")

	parsedEvent, err := ParseWebhook("Note Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*IssueCommentEvent)
	assert.True(t, ok, "Expected IssueCommentEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, string(NoteEventTargetType), event.ObjectKind)
	assert.Equal(t, int64(5), event.ProjectID)
	assert.Equal(t, "Issue", event.ObjectAttributes.NoteableType)
	assert.Equal(t, "test_issue", event.Issue.Title)
	assert.Len(t, event.Issue.Labels, 2)
}

func TestParseIssueHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/issue.json")

	parsedEvent, err := ParseWebhook("Issue Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*IssueEvent)
	assert.True(t, ok, "Expected IssueEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "issue", event.ObjectKind)
	assert.Equal(t, "GitLab Test", event.Project.Name)
	assert.Equal(t, "opened", event.ObjectAttributes.State)
	assert.Equal(t, "user1", event.Assignee.Username)
	assert.Len(t, event.Labels, 1)
	assert.Equal(t, int64(0), event.Changes.UpdatedByID.Previous)
	assert.Equal(t, int64(1), event.Changes.UpdatedByID.Current)
	assert.Len(t, event.Changes.Labels.Previous, 1)
	assert.Len(t, event.Changes.Labels.Current, 1)
	assert.Empty(t, event.Changes.Description.Previous)
	assert.Equal(t, "New description", event.Changes.Description.Current)
	assert.Empty(t, event.Changes.Title.Previous)
	assert.Equal(t, "New title", event.Changes.Title.Current)
}

func TestParseMergeRequestCommentHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/note_merge_request.json")

	parsedEvent, err := ParseWebhook("Note Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*MergeCommentEvent)
	assert.True(t, ok, "Expected MergeCommentEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, string(NoteEventTargetType), event.ObjectKind)
	assert.Equal(t, int64(5), event.ProjectID)
	assert.Equal(t, "MergeRequest", event.ObjectAttributes.NoteableType)
	assert.Equal(t, int64(7), event.MergeRequest.ID)
	assert.Equal(t, "Merge branch 'another-branch' into 'master'", event.MergeRequest.LastCommit.Title)
}

func TestParseMemberHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/member.json")

	parsedEvent, err := ParseWebhook("Member Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*MemberEvent)
	assert.True(t, ok, "Expected MemberEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "user_add_to_group", event.EventName)
}

func TestParseMergeRequestHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/merge_request.json")

	parsedEvent, err := ParseWebhook("Merge Request Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*MergeEvent)
	assert.True(t, ok, "Expected MergeEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "merge_request", event.ObjectKind)
	assert.Equal(t, "unchecked", event.ObjectAttributes.MergeStatus)
	assert.Equal(t, "da1560886d4f094c3e6c9ef40349f7d38b5d27d7", event.ObjectAttributes.LastCommit.ID)
	assert.False(t, event.ObjectAttributes.WorkInProgress)
	assert.Len(t, event.Labels, 1)
	assert.Equal(t, int64(0), event.Changes.UpdatedByID.Previous)
	assert.Equal(t, int64(1), event.Changes.UpdatedByID.Current)
	assert.Len(t, event.Changes.Labels.Previous, 1)
	assert.Len(t, event.Changes.Labels.Current, 1)
}

func TestParsePipelineHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/pipeline.json")

	parsedEvent, err := ParseWebhook("Pipeline Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*PipelineEvent)
	assert.True(t, ok, "Expected PipelineEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "pipeline", event.ObjectKind)
	assert.Equal(t, int64(63), event.ObjectAttributes.Duration)
	assert.Equal(t, "bcbb5ec396a2c0f828686f14fac9b80b780504f2", event.Commit.ID)
	assert.Equal(t, int64(380), event.Builds[0].ID)
	assert.Equal(t, "instance_type", event.Builds[0].Runner.RunnerType)
}

func TestParseProjectResourceAccessTokenHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/resource_access_token_project.json")

	parsedEvent, err := ParseWebhook("Resource Access Token Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*ProjectResourceAccessTokenEvent)
	assert.True(t, ok, "Expected ProjectResourceAccessTokenEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "expiring_access_token", event.EventName)
}

func TestParsePushHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/push.json")

	parsedEvent, err := ParseWebhook("Push Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*PushEvent)
	assert.True(t, ok, "Expected PushEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, eventObjectKindPush, event.ObjectKind)
	assert.Equal(t, int64(15), event.ProjectID)
	assert.Equal(t, exampleEventUserName, event.UserName)
	assert.NotNil(t, event.Commits[0])
	assert.NotNil(t, event.Commits[0].Timestamp)
	assert.Equal(t, "Jordi Mallach", event.Commits[0].Author.Name)
}

func TestParseReleaseHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/release.json")

	parsedEvent, err := ParseWebhook("Release Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*ReleaseEvent)
	assert.True(t, ok, "Expected ReleaseEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "release", event.ObjectKind)
	assert.Equal(t, "Project Name", event.Project.Name)
}

func TestParseServiceWebHook(t *testing.T) {
	t.Parallel()
	parsedEvent, err := ParseWebhook("Service Hook", loadFixture(t, "testdata/webhooks/service_merge_request.json"))
	assert.NoError(t, err)

	event, ok := parsedEvent.(*MergeEvent)
	assert.True(t, ok, "Expected MergeEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, &EventUser{
		ID:        2,
		Name:      "the test",
		Username:  "test",
		Email:     "test@test.test",
		AvatarURL: "https://www.gravatar.com/avatar/dd46a756faad4727fb679320751f6dea?s=80&d=identicon",
	}, event.User)
	assert.Equal(t, "unchecked", event.ObjectAttributes.MergeStatus)
	assert.Equal(t, "next-feature", event.ObjectAttributes.SourceBranch)
	assert.Equal(t, "master", event.ObjectAttributes.TargetBranch)
}

func TestParseSnippetCommentHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/note_snippet.json")

	parsedEvent, err := ParseWebhook("Note Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*SnippetCommentEvent)
	assert.True(t, ok, "Expected SnippetCommentEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, string(NoteEventTargetType), event.ObjectKind)
	assert.Equal(t, int64(5), event.ProjectID)
	assert.Equal(t, "Snippet", event.ObjectAttributes.NoteableType)
	assert.Equal(t, "test", event.Snippet.Title)
}

func TestParseSubGroupHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/subgroup.json")

	parsedEvent, err := ParseWebhook("Subgroup Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*SubGroupEvent)
	assert.True(t, ok, "Expected SubGroupEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "subgroup_create", event.EventName)
}

func TestParseTagHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/tag_push.json")

	parsedEvent, err := ParseWebhook("Tag Push Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*TagEvent)
	assert.True(t, ok, "Expected TagEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, eventObjectKindTagPush, event.ObjectKind)
	assert.Equal(t, int64(1), event.ProjectID)
	assert.Equal(t, exampleEventUserName, event.UserName)
	assert.Equal(t, exampleEventUserUsername, event.UserUsername)
	assert.Equal(t, "refs/tags/v1.0.0", event.Ref)
}

func TestParseWikiPageHook(t *testing.T) {
	t.Parallel()
	raw := loadFixture(t, "testdata/webhooks/wiki_page.json")

	parsedEvent, err := ParseWebhook("Wiki Page Hook", raw)
	assert.NoError(t, err)

	event, ok := parsedEvent.(*WikiPageEvent)
	assert.True(t, ok, "Expected WikiPageEvent, but parsing produced %T", parsedEvent)

	assert.Equal(t, "wiki_page", event.ObjectKind)
	assert.Equal(t, "awesome-project", event.Project.Name)
	assert.Equal(t, "http://example.com/root/awesome-project/wikis/home", event.Wiki.WebURL)
	assert.Equal(t, "adding an awesome page to the wiki", event.ObjectAttributes.Message)
}
