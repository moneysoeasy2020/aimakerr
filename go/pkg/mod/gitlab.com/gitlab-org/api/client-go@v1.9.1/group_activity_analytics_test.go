package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupActivityAnalytics_GetRecentlyCreatedIssuesCount(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/analytics/group_activity/issues_count", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"issues_count": 42}`)
	})

	want := &IssuesCount{IssuesCount: 42}
	opt := &GetRecentlyCreatedIssuesCountOptions{GroupPath: "test-group"}

	issuesCount, resp, err := client.GroupActivityAnalytics.GetRecentlyCreatedIssuesCount(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, issuesCount)
}

func TestGroupActivityAnalytics_GetRecentlyCreatedMergeRequestsCount(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/analytics/group_activity/merge_requests_count", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"merge_requests_count": 42}`)
	})

	want := &MergeRequestsCount{MergeRequestsCount: 42}
	opt := &GetRecentlyCreatedMergeRequestsCountOptions{GroupPath: "test-group"}

	issuesCount, resp, err := client.GroupActivityAnalytics.GetRecentlyCreatedMergeRequestsCount(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, issuesCount)
}

func TestGroupActivityAnalytics_GetRecentlyAddedMembersCount(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/analytics/group_activity/new_members_count", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"new_members_count": 42}`)
	})

	want := &NewMembersCount{NewMembersCount: 42}
	opt := &GetRecentlyAddedMembersCountOptions{GroupPath: "test-group"}

	issuesCount, resp, err := client.GroupActivityAnalytics.GetRecentlyAddedMembersCount(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, issuesCount)
}
