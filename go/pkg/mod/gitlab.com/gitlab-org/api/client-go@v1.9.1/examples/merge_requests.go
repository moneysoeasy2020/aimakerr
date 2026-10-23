package main

import (
	"log"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// This example demonstrates how to manage merge requests in GitLab using the GitLab API client.
// It covers the following steps:
// 1. Initialize the GitLab client with a personal access token.
// 2. Create a new merge request for a project.
// 3. List all merge requests for a project.
// 4. Approve or reject a merge request.
// 5. Add comments to a merge request.
func mergeRequestsExample() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	projectID := 12345 // Replace with your project ID

	createOpts := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr("Add new feature"),
		Description:  gitlab.Ptr("This merge request adds a new feature."),
		SourceBranch: gitlab.Ptr("feature-branch"),
		TargetBranch: gitlab.Ptr("main"),
	}
	mergeRequest, _, err := git.MergeRequests.CreateMergeRequest(projectID, createOpts)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created merge request: %v\n", mergeRequest)

	listOpts := &gitlab.ListProjectMergeRequestsOptions{}
	mergeRequests, _, err := git.MergeRequests.ListProjectMergeRequests(projectID, listOpts)
	if err != nil {
		log.Fatal(err)
	}
	for _, mr := range mergeRequests {
		log.Printf("Found merge request: %v\n", mr)
	}

	approveOpts := &gitlab.ApproveMergeRequestOptions{}
	_, _, err = git.MergeRequestApprovals.ApproveMergeRequest(projectID, mergeRequest.IID, approveOpts)
	if err != nil {
		log.Printf("Failed to approve merge request: %v\n", err)
	} else {
		log.Printf("Approved merge request: %v\n", mergeRequest.IID)
	}

	commentOpts := &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.Ptr("This is a comment on the merge request."),
	}
	comment, _, err := git.Notes.CreateMergeRequestNote(projectID, mergeRequest.IID, commentOpts)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Added comment to merge request: %v\n", comment)
}
