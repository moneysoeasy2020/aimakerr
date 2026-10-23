package gitlab

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGroupReleases_ListGroupReleases(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/releases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		[
			{
				"name": "standard release",
				"tag_name": "releasetag",
				"description": "describing things",
				"created_at": "2022-01-10T15:23:15.529Z",
				"released_at": "2022-01-10T15:23:15.529Z",
				"commit": {
					"id": "e8cbb845ae5a53a2fef2938cf63cf82efc10d993",
					"short_id": "e8cbb845",
					"created_at": "2022-01-10T15:23:15.529Z",
					"parent_ids": [],
					"title": "Update test",
					"message": "Update test",
					"author_name": "Administrator",
					"author_email": "admin@example.com",
					"authored_date": "2022-01-10T15:23:15.529Z",
					"committer_name": "Administrator",
					"committer_email": "admin@example.com",
					"committed_date": "2022-01-10T15:23:15.529Z",
					"web_url": "https://gitlab.com/groups/gitlab-org/-/commit/e8cbb845ae5a53a2fef2938cf63cf82efc10d993"
				},
				"upcoming_release": false,
				"commit_path": "/testgroup/test/-/commit/e8cbb845ae5a53a2fef2938cf63cf82efc10d993",
				"tag_path": "/testgroup/test/-/tags/testtag"
			}
		]`)
	})

	createdAt := time.Date(2022, time.January, 10, 15, 23, 15, 529000000, time.UTC)

	want := []*Release{{
		Name:        "standard release",
		TagName:     "releasetag",
		Description: "describing things",
		CreatedAt:   &createdAt,
		ReleasedAt:  &createdAt,
		Commit: Commit{
			ID:             "e8cbb845ae5a53a2fef2938cf63cf82efc10d993",
			ShortID:        "e8cbb845",
			CreatedAt:      &createdAt,
			ParentIDs:      []string{},
			Title:          "Update test",
			Message:        "Update test",
			AuthorName:     "Administrator",
			AuthorEmail:    "admin@example.com",
			AuthoredDate:   &createdAt,
			CommitterName:  "Administrator",
			CommitterEmail: "admin@example.com",
			CommittedDate:  &createdAt,
			WebURL:         "https://gitlab.com/groups/gitlab-org/-/commit/e8cbb845ae5a53a2fef2938cf63cf82efc10d993",
		},
		UpcomingRelease: false,
		CommitPath:      "/testgroup/test/-/commit/e8cbb845ae5a53a2fef2938cf63cf82efc10d993",
		TagPath:         "/testgroup/test/-/tags/testtag",
	}}

	releases, resp, err := client.GroupReleases.ListGroupReleases(1, &ListGroupReleasesOptions{Simple: Ptr(true)})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, releases)
}
