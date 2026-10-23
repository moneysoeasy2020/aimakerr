package gitlab

import (
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkImportsService_StartMigration(t *testing.T) {
	t.Parallel()
	startMigrationOptions := &BulkImportStartMigrationOptions{
		Configuration: &BulkImportStartMigrationConfiguration{
			URL:         Ptr("https://source-gitlab-instance.example.com"),
			AccessToken: Ptr("source-gitlab-instance-access-token"),
		},
		Entities: []BulkImportStartMigrationEntity{
			{
				SourceType:           Ptr("group_entity"),
				SourceFullPath:       Ptr("gitlab-org/gitlab"),
				DestinationSlug:      Ptr("destination_slug"),
				DestinationNamespace: Ptr("destination/namespace/path"),
				MigrateProjects:      Ptr(true),
				MigrateMemberships:   Ptr(true),
			},
		},
	}
	wantStartMigrationResponse := &BulkImportStartMigrationResponse{
		ID:          1337,
		Status:      "created",
		SourceType:  "group_entity",
		SourceURL:   "https://source-gitlab-instance.example.com",
		CreatedAt:   time.Date(2021, time.June, 18, 9, 45, 55, 358000000, time.UTC),
		UpdatedAt:   time.Date(2021, time.June, 18, 9, 46, 27, 3000000, time.UTC),
		HasFailures: false,
	}
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/bulk_imports", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		gotBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		wantBody, err := os.ReadFile("testdata/post_bulk_imports_request.json")
		assert.NoError(t, err)
		assert.JSONEq(t, string(wantBody), string(gotBody))
		mustWriteHTTPResponse(t, w, "testdata/post_bulk_imports_response.json")
	})

	gotStartMigrationResponse, _, err := client.BulkImports.StartMigration(startMigrationOptions, nil)

	require.NoError(t, err)
	assert.Equal(t, wantStartMigrationResponse, gotStartMigrationResponse)
}
