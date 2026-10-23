package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplicationStatistics_GetApplicationStatistics(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/application/statistics", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"forks": 2,
				"issues": 1,
				"merge_requests": 3,
				"notes": 4,
				"snippets": 5,
				"ssh_keys": 6,
				"milestones": 7,
				"users": 8,
				"groups": 9,
				"projects": 10,
				"active_users": 11
			}
		`)
	})

	want := &ApplicationStatistics{
		Forks:         2,
		Issues:        1,
		MergeRequests: 3,
		Notes:         4,
		Snippets:      5,
		SSHKeys:       6,
		Milestones:    7,
		Users:         8,
		Groups:        9,
		Projects:      10,
		ActiveUsers:   11,
	}
	statistics, resp, err := client.ApplicationStatistics.GetApplicationStatistics()
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, statistics)
}
