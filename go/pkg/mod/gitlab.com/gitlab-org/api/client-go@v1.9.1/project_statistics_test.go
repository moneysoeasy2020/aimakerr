package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectStatisticsService_Last30DaysStatistics(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/statistics", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			{
			  "fetches": {
				"total": 50,
				"days": [
				  {
					"count": 10,
					"date": "2025-10-07"
				  },
				  {
					"count": 40,
					"date": "2025-10-06"
				  }
				]
			  }
			}
		`)
	})

	// This is the expected struct that the JSON response should unmarshal into.
	want := &ProjectStatistics{
		Fetches: FetchStats{
			Total: 50,
			Days: []DayStats{
				{
					Count: 10,
					Date:  "2025-10-07",
				},
				{
					Count: 40,
					Date:  "2025-10-06",
				},
			},
		},
	}

	// Test the happy path.
	stats, resp, err := client.ProjectStatistics.Last30DaysStatistics(1)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, want, stats)
}
