package main

import (
	"context"
	"log"
	"net/http"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func createMergeRequestAndSetAutoMerge() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	projectName := "example/example"

	// Create a new Merge Request
	mr, _, err := git.MergeRequests.CreateMergeRequest(projectName, &gitlab.CreateMergeRequestOptions{
		SourceBranch:       gitlab.Ptr("my-topic-branch"),
		TargetBranch:       gitlab.Ptr("main"),
		Title:              gitlab.Ptr("New MergeRequest"),
		Description:        gitlab.Ptr("New MergeRequest"),
		RemoveSourceBranch: gitlab.Ptr(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Set auto-merge to created Merge Request
	// c.f. https://docs.gitlab.com/user/project/merge_requests/auto_merge/
	_, _, err = git.MergeRequests.AcceptMergeRequest(
		projectName, mr.IID, &gitlab.AcceptMergeRequestOptions{AutoMerge: gitlab.Ptr(true)},

		// client-go provides retries on rate limit (429) and server (>= 500) errors by default.
		//
		// But Method Not Allowed (405) and Unprocessable Content (422) errors will be returned
		// when AcceptMergeRequest is called immediately after CreateMergeRequest.
		//
		// c.f. https://docs.gitlab.com/api/merge_requests/#merge-a-merge-request
		//
		// Therefore, add a retryable status code only for AcceptMergeRequest calls
		gitlab.WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			if err != nil {
				return false, err
			}
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusUnprocessableEntity {
				return true, nil
			}
			return false, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
