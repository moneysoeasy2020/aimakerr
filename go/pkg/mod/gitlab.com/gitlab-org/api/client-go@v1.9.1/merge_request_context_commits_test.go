package gitlab

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListMergeRequestContextCommits(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN a project with merge request context commits
	mux.HandleFunc("/api/v4/projects/1/merge_requests/1/context_commits", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/list_merge_request_context_commits.json")
	})

	// WHEN listing the merge request context commits
	commits, resp, err := client.MergeRequestContextCommits.ListMergeRequestContextCommits(1, 1)

	// THEN the request should succeed and return the context commits
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, commits, 1)

	createdAt := time.Date(2017, time.April, 11, 10, 8, 59, 0, time.UTC)
	authoredDate := time.Date(2017, time.April, 11, 10, 8, 59, 0, time.UTC)
	committedDate := time.Date(2017, time.April, 11, 10, 8, 59, 0, time.UTC)
	want := []*Commit{
		{
			ID:             "4a24d82dbca5c11c61556f3b35ca472b7463187e",
			ShortID:        "4a24d82d",
			CreatedAt:      &createdAt,
			ParentIDs:      nil,
			Title:          "Update README.md to include `Usage in testing and development`",
			Message:        "Update README.md to include `Usage in testing and development`",
			AuthorName:     "Example \"Sample\" User",
			AuthorEmail:    "user@example.com",
			AuthoredDate:   &authoredDate,
			CommitterName:  "Example \"Sample\" User",
			CommitterEmail: "user@example.com",
			CommittedDate:  &committedDate,
		},
	}

	assert.Equal(t, want, commits)
}

func TestCreateMergeRequestContextCommits(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN a project with a merge request
	mux.HandleFunc("/api/v4/projects/15/merge_requests/12/context_commits", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		mustWriteHTTPResponse(t, w, "testdata/create_merge_request_context_commits.json")
	})

	// WHEN creating context commits for the merge request
	opt := &CreateMergeRequestContextCommitsOptions{
		Commits: Ptr([]string{"51856a574ac3302a95f82483d6c7396b1e0783cb"}),
	}
	commits, resp, err := client.MergeRequestContextCommits.CreateMergeRequestContextCommits(15, 12, opt)

	// THEN the request should succeed and return the created context commits
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, commits, 1)

	createdAt := time.Date(2014, time.February, 27, 10, 5, 10, 0, time.UTC)
	authoredDate := time.Date(2014, time.February, 27, 10, 5, 10, 0, time.UTC)
	committedDate := time.Date(2014, time.February, 27, 10, 5, 10, 0, time.UTC)
	want := []*Commit{
		{
			ID:             "51856a574ac3302a95f82483d6c7396b1e0783cb",
			ShortID:        "51856a57",
			CreatedAt:      &createdAt,
			ParentIDs:      []string{"57a82e2180507c9e12880c0747f0ea65ad489515"},
			Title:          "Commit title",
			Message:        "Commit message",
			AuthorName:     "Example User",
			AuthorEmail:    "user@example.com",
			AuthoredDate:   &authoredDate,
			CommitterName:  "Example User",
			CommitterEmail: "user@example.com",
			CommittedDate:  &committedDate,
			Trailers:       map[string]string{},
			WebURL:         "https://gitlab.example.com/project/path/-/commit/51856a574ac3302a95f82483d6c7396b1e0783cb",
		},
	}

	assert.Equal(t, want, commits)
}

func TestDeleteMergeRequestContextCommits(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN a project with merge request context commits
	mux.HandleFunc("/api/v4/projects/1/merge_requests/1/context_commits", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	// WHEN deleting context commits from the merge request
	opt := &DeleteMergeRequestContextCommitsOptions{
		Commits: Ptr([]string{"51856a574ac3302a95f82483d6c7396b1e0783cb"}),
	}
	resp, err := client.MergeRequestContextCommits.DeleteMergeRequestContextCommits(1, 1, opt)

	// THEN the request should succeed
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestListMergeRequestContextCommits_WithStringProjectID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN a project identified by namespace/path
	mux.HandleFunc("/api/v4/projects/namespace%2Fproject/merge_requests/1/context_commits", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/list_merge_request_context_commits.json")
	})

	// WHEN listing the merge request context commits using string project ID
	commits, resp, err := client.MergeRequestContextCommits.ListMergeRequestContextCommits("namespace/project", 1)

	// THEN the request should succeed
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, commits, 1)
}

func TestCreateMergeRequestContextCommits_WithStringProjectID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN a project identified by namespace/path
	mux.HandleFunc("/api/v4/projects/namespace%2Fproject/merge_requests/1/context_commits", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		mustWriteHTTPResponse(t, w, "testdata/create_merge_request_context_commits.json")
	})

	// WHEN creating context commits using string project ID
	opt := &CreateMergeRequestContextCommitsOptions{
		Commits: Ptr([]string{"51856a574ac3302a95f82483d6c7396b1e0783cb"}),
	}
	commits, resp, err := client.MergeRequestContextCommits.CreateMergeRequestContextCommits("namespace/project", 1, opt)

	// THEN the request should succeed
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, commits, 1)
}

func TestDeleteMergeRequestContextCommits_WithStringProjectID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// GIVEN a project identified by namespace/path
	mux.HandleFunc("/api/v4/projects/namespace%2Fproject/merge_requests/1/context_commits", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	// WHEN deleting context commits using string project ID
	opt := &DeleteMergeRequestContextCommitsOptions{
		Commits: Ptr([]string{"51856a574ac3302a95f82483d6c7396b1e0783cb"}),
	}
	resp, err := client.MergeRequestContextCommits.DeleteMergeRequestContextCommits("namespace/project", 1, opt)

	// THEN the request should succeed
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
