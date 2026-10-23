package gitlab

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupRelationsExportService_ScheduleExport(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/export_relations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		// This endpoint returns a 202 Accepted with no body on success.
		w.WriteHeader(http.StatusAccepted)
	})

	resp, err := client.GroupRelationsExport.ScheduleExport(1, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestGroupRelationsExportService_ListExportStatus(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/export_relations/status", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			[
			  {
				"relation": "labels",
				"status": 1,
				"error": null,
				"updated_at": "2021-05-04T11:25:20.423Z",
				"batched": true,
				"batches_count": 1,
				"batches": [
				  {
					"status": 1,
					"batch_number": 1,
					"objects_count": 1,
					"error": null,
					"updated_at": "2021-05-04T11:25:20.423Z"
				  }
				]
			  },
			  {
				"relation": "milestones",
				"status": 1,
				"error": null,
				"updated_at": "2021-05-04T11:25:20.085Z",
				"batched": false,
				"batches_count": 0
			  }
			]
		`)
	})

	updatedAt1, err := time.Parse(time.RFC3339, "2021-05-04T11:25:20.423Z")
	require.NoError(t, err)
	updatedAt2, err := time.Parse(time.RFC3339, "2021-05-04T11:25:20.085Z")
	require.NoError(t, err)

	want := []*GroupRelationStatus{
		{
			Relation:     "labels",
			Status:       1,
			Error:        "",
			UpdatedAt:    updatedAt1,
			Batched:      true,
			BatchesCount: 1,
			Batches: []Batch{
				{
					Status:       1,
					BatchNumber:  1,
					ObjectsCount: 1,
					Error:        "",
					UpdatedAt:    updatedAt1,
				},
			},
		},
		{
			Relation:     "milestones",
			Status:       1,
			Error:        "",
			UpdatedAt:    updatedAt2,
			Batched:      false,
			BatchesCount: 0,
			Batches:      nil,
		},
	}

	stats, resp, err := client.GroupRelationsExport.ListExportStatus(1, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, want, stats)
}

func TestGroupRelationsExportService_ExportDownload(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/export_relations/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/octet-stream")
		fmt.Fprint(w, "fake-export-data")
	})

	reader, resp, err := client.GroupRelationsExport.ExportDownload(1, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, reader)

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, []byte("fake-export-data"), data)
}
