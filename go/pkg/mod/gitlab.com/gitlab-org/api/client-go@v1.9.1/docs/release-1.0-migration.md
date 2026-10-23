# Client-go Version 1.0 Migration Guide

Release 1.0 marks a major milestone for the GitLab client-go project, representing its first release which
provides a breaking change guarantee. Moving forward, expect client-go to provide a major version migration with each
update to the Go language version to align supported Go versions to the language support guarantee.

This guide will walk through how to consume the breaking changes from the release. It's likely (though not guaranteed)
that future versions will have fewer breaking changes, as the library has been in version 0.X for a while.

## Struct Alias Updates

Some structs have had alias or names updated or removed to more clearly align them to Golang naming conventions. 
In these cases, update struct references to align to the new names.

- `ListRegistryRepositoriesOptions` has been renamed to `ListProjectRegistryRepositoriesOptions` to clearly express that this
is for project registries as opposed to group registries
- `UpdateMergeRequestApprovalSettingsOptions` has been renamed to `UpdateProjectMergeRequestApprovalSettingsOptions`
- The Approval Settings Options struct for updating Group Approval Settings has been separated from the Project struct and is
now `UpdateGroupMergeRequestApprovalSettingsOptions`
- Several attributes in existing structs have had their casing updated to align to repository standards, including:
  - `PipelineId` is now `PipelineID`
  - `SelectiveSyncNamespaceIds` is now `SelectiveSyncNamespaceIDs`
  - `RefsUrl` is now `RefsURL`
  - `BitbucketServerUrl` is now `BitbucketServerURL`
  - `CrlUrl` is now `CrlURL`
  - `ServicePingNonSqlMetrics` is now `ServicePingNonSQLMetrics`

Since each of the above struct and attribute changes are in-place renames, consuming the changes is as simple as updating the
references wherever they are used in code:

```go
// old code
myRepositoryOptions := ListRegistryRepositoriesOptions{}

// new code
myNewRepositoryOptions := ListProjectRegistryRepositoriesOptions{}
```

### Merge Requests That Implement This Change

- [refactor!: standardize Go naming conventions for ID and URL fields](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2462) by @elC0mpa
- [refactor: decouple group and project request approval settings](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2465) by @elC0mpa

## Updates to `ListOptions` to avoid Aliases

Areas that used `ListOptions` before now use dedicated Options structs instead. This prevents us from needing to introduce
breaking changes in the future when individual APIs gain additional arguments. The following are new Structs that are used
instead of `ListOptions` in APIs. Each struct is composed of only the `ListOptions` struct right now, but may add new fields in
the future as needed.
  - `ListAccessRequestsOptions` 
  - `ListApplicationsOptions`
  - `ListAwardEmojiOptions`
  - `ListIssueBoardsOptions`
  - `GetIssueBoardListsOptions`
  - `ListBroadcastMessagesOptions`
  - `ListCIYMLTemplatesOptions`
  - `ListAgentsOptions`
  - `ListAgentTokensOptions`
  - `GetCommitCommentsOptions`
  - `ListGroupRegistryRepositoriesOptions`
  - `ListRegistryRepositoryTagsOptions`
  - `ListProjectDeployKeysOptions`
  - `ListUserProjectDeployKeysOptions`
  - `ListProjectDeployTokensOptions`
  - `ListGroupDeployTokensOptions`
  - `ListIssueDiscussionsOptions`
  - `ListSnippetDiscussionsOptions`
  - `ListGroupEpicDiscussionsOptions`
  - `ListMergeRequestDiscussionsOptions`
  - `ListCommitDiscussionsOptions`
  - `ListDockerfileTemplatesOptions`
  - `ListClientKeysOptions`
  - `ListFreezePeriodsOptions`
  - `ListGeoNodesOptions`
  - `ListGeoSitesOptions`
  - `ListStatusOfAllGeoSitesOptions`
  - `ListTemplatesOptions`
  - `ListGroupIssueBoardsOptions`
  - `ListGroupIssueBoardListsOptions`
  - `ListGroupEpicBoardsOptions`
  - `ListGroupHooksOptions`
  - `ListMembershipsForBillableGroupMemberOptions`
  - `GetGroupMilestoneIssuesOptions`
  - `GetGroupMilestoneMergeRequestsOptions`
  - `GetGroupMilestoneBurndownChartEventsOptions`
  - `ListGroupProtectedEnvironmentsOptions`
  - `RetrieveAllGroupStorageMovesOptions`
  - `ListGroupVariablesOptions`
  - `ListInstanceVariablesOptions`
  - `ListMergeRequestsClosingIssueOptions`
  - `ListMergeRequestsRelatedToIssueOptions`
  - `GetMergeRequestCommitsOptions`
  - `GetIssuesClosedOnMergeOptions`
  - `ListRelatedIssuesOptions`
  - `GetMergeRequestDiffVersionsOptions`
  - `GetMilestoneIssuesOptions`
  - `GetMilestoneMergeRequestsOptions`
  - `ListPackageFilesOptions`
  - `ListPagesDomainsOptions`
  - `ListPipelinesTriggeredByScheduleOptions`
  - `ListPipelineTriggersOptions`
  - `ListProjectMirrorOptions`
  - `RetrieveAllProjectStorageMovesOptions`
  - `ListProjectSnippetsOptions`
  - `ListProjectVariablesOptions`
  - `ListProjectHooksOptions`
  - `GetProjectApprovalRulesListsOptions`
  - `ListProtectedEnvironmentsOptions`
  - `ListProtectedTagsOptions`
  - `ListReleaseLinksOptions`
  - `ListProjectSecureFilesOptions`
  - `RetrieveAllSnippetStorageMovesOptions`
  - `ListSnippetsOptions`
  - `ExploreSnippetsOptions`
  - `ListSSHKeysOptions`
  - `ListSSHKeysForUserOptions`
  - `ListEmailsForUserOptions`
  
To consume these updates, the code that currently uses `ListOptions` will need to be updated
to instead use the struct named above, depending on which function is being called. The above
functions are all composed with ListOptions, and wrapping the existing `ListOptions` in the new
struct will be required.

```go
// old code
reqOptions := ListOptions{
    Page: 1,
    PerPage: 20,
}

// new code
snippetsListOptions :=  ExploreSnippetsOptions{
    ListOptions: ListOptions{
        Page:    1,
        PerPage: 10,
    },
}
```

### Merge Requests That Implement This Change

- [feat!(ListOptions): Update ListOptions to use composition instead of aliasing](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2430) by @PatrickRice

## Updates Inline struct values to Named Types

Many structs within the codebase would previously use inlined structs as part of the struct definition.
This caused several problems:
  1. Users who used these structs in tests or code would encounter breaking changes when we added new attributes; 
something that shouldn't normally be a breaking change.
  2. Initializing these structs was difficult, requiring repeating the whole struct definition

Users who are only consuming API responses will be unaffected by this change, but users
who manually create structs may need to update to use the new named structs instead. 

The following structs were impacted, and use the new named structs as specified:

- AwardEmoji
  - BasicUser (reused existing)
- IssueBoard
  - BasicUser (reused existing)
- BoardList
  - BoardListAssignee
- Deployment
  - DeploymentDeployable
  - DeploymentDeployablePipeline
- NoteEvent
  - NoteEventObjectAttributes
- PushSystemEvent
  - PushSystemEventProject
  - PushSystemEventCommit
  - PushSystemEventCommitAuthor
- TagPushSystemEvent
  - TagPushSystemEventProject
  - TagPushSystemEventCommit
  - TagPushSystemEventCommitAuthor
- RepositoryUpdateSystemEvent
  - RepositoryUpdateSystemEventProject
  - RepositoryUpdateSystemEventChange
- ContributionEvent
  - ContributionEventPushData
  - BasicUser (reused existing)
- ProjectEvent
  - BasicUser (reused existing)
  - ProjectEventData
  - ProjectEventNote
  - ProjectEventNoteAuthor
  - ProjectEventPushData
- GenericPackagesFile
  - GenericPackagesFileURL
- ImportRepositoryFromGitHubOptions
  - ImportRepositoryFromGitHubOptionalStagesOptions
- IssuesStatistics
  - IssuesStatisticsStatistics
  - IssuesStatisticsCounts
- Job
  - JobPipeline
  - JobArtifact
  - JobArtifactsFile
  - JobRunner
- License
  - LicenseLicensee
  - LicenseAddOns


```go
// old code:
awardEmoji := &AwardEmoji{
    User: struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	}{
        Name: "test1"
    }
}

// new code
newAwardEmoji := &AwardEmoji{
    User: &BasicUser {
        Name: "test1"
    }
}
```

### Merge Requests That Implement This Change

- [refactor(no-release): refactor inline structs to reusable types](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2477) by @heidi.berry

## Fix Several Structs with Spelling Errors

Several structs used throughout the code base had spelling errors in their names. While this doesn't impact
functionality, it can make the structs difficult to search for. These spelling errors have been fixed, which
will require an update when initializing the struct in code.

- `CISeperateCache` -> `CISeparatedCaches` 
- `SharedVisiableOnly` -> `SharedVisibleOnly`
- `IncludeDescendantGrouops` -> `IncludeDescendantGroups`
- `ProjectReposityStorage` -> `ProjectRepositoryStorage`
- `UpdateEpicIsssueAssignmentOptions` -> `UpdateEpicIssueAssignmentOptions`
- `ListProjectInvidedGroupOptions` -> `ListProjectInvitedGroupOptions`

### Merge Requests That Implement This Change

- [refactor(no-release): fix revive.var-naming lint issues](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2437) by @alexandear
- [refactor!: Fix typos in struct names](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2429) by @PatrickRice
- [fix!(epics): remove UpdateEpicIsssueAssignmentOptions in favor to UpdateEpicIssueAssignmentOptions](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2407) by @alexandear
- [fix!: remove ProjectReposityStorage in favor to ProjectRepositoryStorage](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2366) by @alexandear


## Migrate `int` to `int64`

All structs and functions in `client-go` have been updated to use `int64` types where they previously used `int`.
This ensures there is no ambiguity about the fact that GitLab can use large integers to all integer values, and
prevents some cross-platform issues.

When referencing `int64` returns from `client-go`, either update the referencing code to use `int64`, or cast it 
to an `int` instead.

```go
// old code
var myNum int
myNum = gitlabProject.ID

// new code
var myNewNum int64
myNewNum = gitlabProject.ID
```

### Merge Requests That Implement This Change

- [chore(no-release): Update `int` in the first batch of files to `int64`](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2505) by @PatrickRice
- [chore(no-release): Round 2 of int-> int64 refactoring](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2530) by @PatrickRice
- [chore: Batch 3 of int -> int64 conversations](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2538) by @PatrickRice
- [Finalize migration from int -> int64](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2540) by @PatrickRice

## Update `AssigneeID` from an `int` to `AssigneeIDValue` in `ListProjectIssuesOptions`

Previously, there was no way to search for `any` or `none` explicitly in the List Project Issues API, since those are not valid
integer values. The `AssigneeIDValue` type allows using those values, so the `AssigneeID` argument has been updated to use that
type instead.

When setting or referencing that type, cast from the `int` value to `AssigneeIDValue` instead

```go
// old code
listProjectIssue := &ListProjectIssuesOptions{
  AuthorID:   Ptr(int(1)),
  AssigneeID: Ptr(int(2)),
}

// new code
listProjectIssue := &ListProjectIssuesOptions{
  AuthorID:   Ptr(int64(1)),
  AssigneeID: AssigneeID(2),
}
```

### Merge Requests That Implement This Change

- [fix(issues): use AssigneeIDValue for ListProjectIssuesOptions.AssigneeID](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2571) by @claytonrcarter

## Header Value Casing has Been Updated

In several cases, the casing on HTTP Headers has been updated to better align to GitLab's documentation. While
this shouldn't be a breaking change, tests that compare headers in a case sensitive fashion may break.

The following header values have been updated:
- `RateLimit-Limit` -> `Ratelimit-Limit`
- `RateLimit-Reset` -> `Ratelimit-Reset`
- `PRIVATE-TOKEN` -> `Private-Token`
- `JOB-TOKEN` -> `Job-Token`

### Merge Requests That Implement This Change

- [refactor!: Canonicalize request headers](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2531) by @alexandear

## Added Integration Tests for some tests

As part of continuing to improve comprehensive testing, a new integration test capability has
been added that runs tests against a locally running GitLab Ultimate instance. The initial 
set of User API tests runs on every merge to ensure that the APIs work as expected when run against
GitLab.

### Merge Requests That Implement This Change

- [chore(no-release): Add integration test support using a real instance](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2466) by @PatrickRice

## Go Version Upgrades

This update increases the version of Go that's required from 1.23 to 1.24. This follows Golang's supported language versions. 

### Merge Requests That Implement This Change

- [chore(deps): update dependency go to v1.24.0](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2424) by @gitlab-dependency-update-bot
