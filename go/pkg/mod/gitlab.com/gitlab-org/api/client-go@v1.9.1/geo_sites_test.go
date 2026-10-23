package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoSites_CreateGeoSite(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/geo_sites", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"id": 3,
			"name": "Test Site 1",
			"url": "https://secondary.example.com/",
			"internal_url": "https://secondary.example.com/",
			"primary": false,
			"enabled": true,
			"current": false,
			"files_max_capacity": 10,
			"repos_max_capacity": 25,
			"verification_max_capacity": 100,
			"container_repositories_max_capacity": 10,
			"selective_sync_type": "namespaces",
			"selective_sync_shards": [],
			"selective_sync_namespace_ids": [1, 25],
			"minimum_reverification_interval": 7,
			"sync_object_storage": false,
			"web_edit_url": "https://primary.example.com/admin/geo/sites/3/edit",
			"web_geo_replication_details_url": "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
			"_links": {
				"self": "https://primary.example.com/api/v4/geo_sites/3",
				"status": "https://primary.example.com/api/v4/geo_sites/3/status",
				"repair": "https://primary.example.com/api/v4/geo_sites/3/repair"
			}
		}`)
	})

	want := &GeoSite{
		ID:                               3,
		Name:                             "Test Site 1",
		URL:                              "https://secondary.example.com/",
		InternalURL:                      "https://secondary.example.com/",
		Primary:                          false,
		Enabled:                          true,
		Current:                          false,
		FilesMaxCapacity:                 10,
		ReposMaxCapacity:                 25,
		VerificationMaxCapacity:          100,
		ContainerRepositoriesMaxCapacity: 10,
		SelectiveSyncType:                "namespaces",
		SelectiveSyncShards:              []string{},
		SelectiveSyncNamespaceIDs:        []int64{1, 25},
		MinimumReverificationInterval:    7,
		SyncObjectStorage:                false,
		WebEditURL:                       "https://primary.example.com/admin/geo/sites/3/edit",
		WebGeoReplicationDetailsURL:      "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
		Links: GeoSiteLinks{
			Self:   "https://primary.example.com/api/v4/geo_sites/3",
			Status: "https://primary.example.com/api/v4/geo_sites/3/status",
			Repair: "https://primary.example.com/api/v4/geo_sites/3/repair",
		},
	}

	site, resp, err := client.GeoSites.CreateGeoSite(&CreateGeoSitesOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, site)
}

func TestGeoSites_ListGeoSite(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/geo_sites", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{
			"id": 3,
			"name": "Test Site 1",
			"url": "https://secondary.example.com/",
			"internal_url": "https://secondary.example.com/",
			"primary": false,
			"enabled": true,
			"current": false,
			"files_max_capacity": 10,
			"repos_max_capacity": 25,
			"verification_max_capacity": 100,
			"container_repositories_max_capacity": 10,
			"selective_sync_type": "namespaces",
			"selective_sync_shards": [],
			"selective_sync_namespace_ids": [1, 25],
			"minimum_reverification_interval": 7,
			"sync_object_storage": false,
			"web_edit_url": "https://primary.example.com/admin/geo/sites/3/edit",
			"web_geo_replication_details_url": "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
			"_links": {
				"self": "https://primary.example.com/api/v4/geo_sites/3",
				"status": "https://primary.example.com/api/v4/geo_sites/3/status",
				"repair": "https://primary.example.com/api/v4/geo_sites/3/repair"
			}
		}]`)
	})

	want := []*GeoSite{{
		ID:                               3,
		Name:                             "Test Site 1",
		URL:                              "https://secondary.example.com/",
		InternalURL:                      "https://secondary.example.com/",
		Primary:                          false,
		Enabled:                          true,
		Current:                          false,
		FilesMaxCapacity:                 10,
		ReposMaxCapacity:                 25,
		VerificationMaxCapacity:          100,
		ContainerRepositoriesMaxCapacity: 10,
		SelectiveSyncType:                "namespaces",
		SelectiveSyncShards:              []string{},
		SelectiveSyncNamespaceIDs:        []int64{1, 25},
		MinimumReverificationInterval:    7,
		SyncObjectStorage:                false,
		WebEditURL:                       "https://primary.example.com/admin/geo/sites/3/edit",
		WebGeoReplicationDetailsURL:      "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
		Links: GeoSiteLinks{
			Self:   "https://primary.example.com/api/v4/geo_sites/3",
			Status: "https://primary.example.com/api/v4/geo_sites/3/status",
			Repair: "https://primary.example.com/api/v4/geo_sites/3/repair",
		},
	}}

	sites, resp, err := client.GeoSites.ListGeoSites(&ListGeoSitesOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, sites)
}

func TestGeoSites_GetGeoSite(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/geo_sites/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id": 3,
			"name": "Test Site 1",
			"url": "https://secondary.example.com/",
			"internal_url": "https://secondary.example.com/",
			"primary": false,
			"enabled": true,
			"current": false,
			"files_max_capacity": 10,
			"repos_max_capacity": 25,
			"verification_max_capacity": 100,
			"container_repositories_max_capacity": 10,
			"selective_sync_type": "namespaces",
			"selective_sync_shards": [],
			"selective_sync_namespace_ids": [1, 25],
			"minimum_reverification_interval": 7,
			"sync_object_storage": false,
			"web_edit_url": "https://primary.example.com/admin/geo/sites/3/edit",
			"web_geo_replication_details_url": "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
			"_links": {
				"self": "https://primary.example.com/api/v4/geo_sites/3",
				"status": "https://primary.example.com/api/v4/geo_sites/3/status",
				"repair": "https://primary.example.com/api/v4/geo_sites/3/repair"
			}
		}`)
	})

	want := &GeoSite{
		ID:                               3,
		Name:                             "Test Site 1",
		URL:                              "https://secondary.example.com/",
		InternalURL:                      "https://secondary.example.com/",
		Primary:                          false,
		Enabled:                          true,
		Current:                          false,
		FilesMaxCapacity:                 10,
		ReposMaxCapacity:                 25,
		VerificationMaxCapacity:          100,
		ContainerRepositoriesMaxCapacity: 10,
		SelectiveSyncType:                "namespaces",
		SelectiveSyncShards:              []string{},
		SelectiveSyncNamespaceIDs:        []int64{1, 25},
		MinimumReverificationInterval:    7,
		SyncObjectStorage:                false,
		WebEditURL:                       "https://primary.example.com/admin/geo/sites/3/edit",
		WebGeoReplicationDetailsURL:      "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
		Links: GeoSiteLinks{
			Self:   "https://primary.example.com/api/v4/geo_sites/3",
			Status: "https://primary.example.com/api/v4/geo_sites/3/status",
			Repair: "https://primary.example.com/api/v4/geo_sites/3/repair",
		},
	}

	site, resp, err := client.GeoSites.GetGeoSite(3)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, site)
}

func TestGeoSites_EditGeoSite(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/geo_sites/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
			"id": 3,
			"name": "Test Site 1",
			"url": "https://secondary.example.com/",
			"internal_url": "https://secondary.example.com/",
			"primary": false,
			"enabled": true,
			"current": false,
			"files_max_capacity": 10,
			"repos_max_capacity": 25,
			"verification_max_capacity": 100,
			"container_repositories_max_capacity": 10,
			"selective_sync_type": "namespaces",
			"selective_sync_shards": [],
			"selective_sync_namespace_ids": [1, 25],
			"minimum_reverification_interval": 7,
			"sync_object_storage": false,
			"web_edit_url": "https://primary.example.com/admin/geo/sites/3/edit",
			"web_geo_replication_details_url": "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
			"_links": {
				"self": "https://primary.example.com/api/v4/geo_sites/3",
				"status": "https://primary.example.com/api/v4/geo_sites/3/status",
				"repair": "https://primary.example.com/api/v4/geo_sites/3/repair"
			}
		}`)
	})

	want := &GeoSite{
		ID:                               3,
		Name:                             "Test Site 1",
		URL:                              "https://secondary.example.com/",
		InternalURL:                      "https://secondary.example.com/",
		Primary:                          false,
		Enabled:                          true,
		Current:                          false,
		FilesMaxCapacity:                 10,
		ReposMaxCapacity:                 25,
		VerificationMaxCapacity:          100,
		ContainerRepositoriesMaxCapacity: 10,
		SelectiveSyncType:                "namespaces",
		SelectiveSyncShards:              []string{},
		SelectiveSyncNamespaceIDs:        []int64{1, 25},
		MinimumReverificationInterval:    7,
		SyncObjectStorage:                false,
		WebEditURL:                       "https://primary.example.com/admin/geo/sites/3/edit",
		WebGeoReplicationDetailsURL:      "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
		Links: GeoSiteLinks{
			Self:   "https://primary.example.com/api/v4/geo_sites/3",
			Status: "https://primary.example.com/api/v4/geo_sites/3/status",
			Repair: "https://primary.example.com/api/v4/geo_sites/3/repair",
		},
	}

	site, resp, err := client.GeoSites.EditGeoSite(3, &EditGeoSiteOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, site)
}

func TestGeoSites_DeleteGeoSite(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/geo_sites/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.GeoSites.DeleteGeoSite(3)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGeoSites_RepairGeoSite(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/geo_sites/3/repair", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"id": 3,
			"name": "Test Site 1",
			"url": "https://secondary.example.com/",
			"internal_url": "https://secondary.example.com/",
			"primary": false,
			"enabled": true,
			"current": false,
			"files_max_capacity": 10,
			"repos_max_capacity": 25,
			"verification_max_capacity": 100,
			"container_repositories_max_capacity": 10,
			"selective_sync_type": "namespaces",
			"selective_sync_shards": [],
			"selective_sync_namespace_ids": [1, 25],
			"minimum_reverification_interval": 7,
			"sync_object_storage": false,
			"web_edit_url": "https://primary.example.com/admin/geo/sites/3/edit",
			"web_geo_replication_details_url": "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
			"_links": {
				"self": "https://primary.example.com/api/v4/geo_sites/3",
				"status": "https://primary.example.com/api/v4/geo_sites/3/status",
				"repair": "https://primary.example.com/api/v4/geo_sites/3/repair"
			}
		}`)
	})

	want := &GeoSite{
		ID:                               3,
		Name:                             "Test Site 1",
		URL:                              "https://secondary.example.com/",
		InternalURL:                      "https://secondary.example.com/",
		Primary:                          false,
		Enabled:                          true,
		Current:                          false,
		FilesMaxCapacity:                 10,
		ReposMaxCapacity:                 25,
		VerificationMaxCapacity:          100,
		ContainerRepositoriesMaxCapacity: 10,
		SelectiveSyncType:                "namespaces",
		SelectiveSyncShards:              []string{},
		SelectiveSyncNamespaceIDs:        []int64{1, 25},
		MinimumReverificationInterval:    7,
		SyncObjectStorage:                false,
		WebEditURL:                       "https://primary.example.com/admin/geo/sites/3/edit",
		WebGeoReplicationDetailsURL:      "https://secondary.example.com/admin/geo/sites/3/replication/lfs_objects",
		Links: GeoSiteLinks{
			Self:   "https://primary.example.com/api/v4/geo_sites/3",
			Status: "https://primary.example.com/api/v4/geo_sites/3/status",
			Repair: "https://primary.example.com/api/v4/geo_sites/3/repair",
		},
	}

	site, resp, err := client.GeoSites.RepairGeoSite(3)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, site)
}

func TestGeoSites_ListStatusesOfAllGeoSites(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/geo_sites/status", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
		{
			"geo_node_id": 1,
			"projects_count": 19,
			"package_files_count": 25,
			"package_files_checksum_total_count": 25,
			"package_files_checksummed_count": 25,
			"terraform_state_versions_count": 18,
			"terraform_state_versions_checksum_total_count": 18,
			"terraform_state_versions_checksummed_count": 18,
			"snippet_repositories_count": 20,
			"snippet_repositories_checksum_total_count": 20,
			"snippet_repositories_checksummed_count": 20,
			"uploads_count": 51,
			"uploads_checksum_total_count": 51,
			"uploads_checksummed_count": 51,
			"job_artifacts_count": 205,
			"job_artifacts_checksum_total_count": 205,
			"job_artifacts_checksummed_count": 205,
			"project_wiki_repositories_count": 19,
			"project_wiki_repositories_checksum_total_count": 19,
			"project_wiki_repositories_checksummed_count": 19,
			"replication_slots_used_in_percentage": "100.00%",
			"replication_slots_count": 1,
			"replication_slots_used_count": 1,
			"healthy": true,
			"health": "Healthy",
			"health_status": "Healthy",
			"missing_oauth_application": false,
			"last_event_id": 357,
			"last_event_timestamp": 1683127088,
			"last_successful_status_check_timestamp": 1683129788,
			"version": "16.0.0-pre",
			"revision": "129eb954664",
			"namespaces": [],
			"storage_shards_match": true,
			"_links": {
				"self": "https://primary.example.com/api/v4/geo_sites/1/status",
				"site": "https://primary.example.com/api/v4/geo_sites/1"
			}
		}]`)
	})

	want := []*GeoSiteStatus{{
		GeoNodeID:                                 1,
		ProjectsCount:                             19,
		PackageFilesCount:                         25,
		PackageFilesChecksumTotalCount:            25,
		PackageFilesChecksummedCount:              25,
		TerraformStateVersionsCount:               18,
		TerraformStateVersionsChecksumTotalCount:  18,
		TerraformStateVersionsChecksummedCount:    18,
		SnippetRepositoriesCount:                  20,
		SnippetRepositoriesChecksumTotalCount:     20,
		SnippetRepositoriesChecksummedCount:       20,
		UploadsCount:                              51,
		UploadsChecksumTotalCount:                 51,
		UploadsChecksummedCount:                   51,
		JobArtifactsCount:                         205,
		JobArtifactsChecksumTotalCount:            205,
		JobArtifactsChecksummedCount:              205,
		ProjectWikiRepositoriesCount:              19,
		ProjectWikiRepositoriesChecksumTotalCount: 19,
		ProjectWikiRepositoriesChecksummedCount:   19,
		ReplicationSlotsUsedInPercentage:          "100.00%",
		ReplicationSlotsCount:                     1,
		ReplicationSlotsUsedCount:                 1,
		Healthy:                                   true,
		Health:                                    "Healthy",
		HealthStatus:                              "Healthy",
		MissingOAuthApplication:                   false,
		LastEventID:                               357,
		LastEventTimestamp:                        1683127088,
		LastSuccessfulStatusCheckTimestamp:        1683129788,
		Version:                                   "16.0.0-pre",
		Revision:                                  "129eb954664",
		Namespaces:                                []string{},
		StorageShardsMatch:                        true,
		Links: GeoSiteStatusLink{
			Self: "https://primary.example.com/api/v4/geo_sites/1/status",
			Site: "https://primary.example.com/api/v4/geo_sites/1",
		},
	}}

	statuses, resp, err := client.GeoSites.ListStatusOfAllGeoSites(&ListStatusOfAllGeoSitesOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, statuses)
}

func TestGeoSites_GetStatusOfGeoSite(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/geo_sites/1/status", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		{
			"geo_node_id": 1,
			"projects_count": 19,
			"package_files_count": 25,
			"package_files_checksum_total_count": 25,
			"package_files_checksummed_count": 25,
			"terraform_state_versions_count": 18,
			"terraform_state_versions_checksum_total_count": 18,
			"terraform_state_versions_checksummed_count": 18,
			"snippet_repositories_count": 20,
			"snippet_repositories_checksum_total_count": 20,
			"snippet_repositories_checksummed_count": 20,
			"uploads_count": 51,
			"uploads_checksum_total_count": 51,
			"uploads_checksummed_count": 51,
			"job_artifacts_count": 205,
			"job_artifacts_checksum_total_count": 205,
			"job_artifacts_checksummed_count": 205,
			"project_wiki_repositories_count": 19,
			"project_wiki_repositories_checksum_total_count": 19,
			"project_wiki_repositories_checksummed_count": 19,
			"replication_slots_used_in_percentage": "100.00%",
			"replication_slots_count": 1,
			"replication_slots_used_count": 1,
			"healthy": true,
			"health": "Healthy",
			"health_status": "Healthy",
			"missing_oauth_application": false,
			"last_event_id": 357,
			"last_event_timestamp": 1683127088,
			"last_successful_status_check_timestamp": 1683129788,
			"version": "16.0.0-pre",
			"revision": "129eb954664",
			"namespaces": [],
			"storage_shards_match": true,
			"_links": {
				"self": "https://primary.example.com/api/v4/geo_sites/1/status",
				"site": "https://primary.example.com/api/v4/geo_sites/1"
			}
		}`)
	})

	want := &GeoSiteStatus{
		GeoNodeID:                                 1,
		ProjectsCount:                             19,
		PackageFilesCount:                         25,
		PackageFilesChecksumTotalCount:            25,
		PackageFilesChecksummedCount:              25,
		TerraformStateVersionsCount:               18,
		TerraformStateVersionsChecksumTotalCount:  18,
		TerraformStateVersionsChecksummedCount:    18,
		SnippetRepositoriesCount:                  20,
		SnippetRepositoriesChecksumTotalCount:     20,
		SnippetRepositoriesChecksummedCount:       20,
		UploadsCount:                              51,
		UploadsChecksumTotalCount:                 51,
		UploadsChecksummedCount:                   51,
		JobArtifactsCount:                         205,
		JobArtifactsChecksumTotalCount:            205,
		JobArtifactsChecksummedCount:              205,
		ProjectWikiRepositoriesCount:              19,
		ProjectWikiRepositoriesChecksumTotalCount: 19,
		ProjectWikiRepositoriesChecksummedCount:   19,
		ReplicationSlotsUsedInPercentage:          "100.00%",
		ReplicationSlotsCount:                     1,
		ReplicationSlotsUsedCount:                 1,
		Healthy:                                   true,
		Health:                                    "Healthy",
		HealthStatus:                              "Healthy",
		MissingOAuthApplication:                   false,
		LastEventID:                               357,
		LastEventTimestamp:                        1683127088,
		LastSuccessfulStatusCheckTimestamp:        1683129788,
		Version:                                   "16.0.0-pre",
		Revision:                                  "129eb954664",
		Namespaces:                                []string{},
		StorageShardsMatch:                        true,
		Links: GeoSiteStatusLink{
			Self: "https://primary.example.com/api/v4/geo_sites/1/status",
			Site: "https://primary.example.com/api/v4/geo_sites/1",
		},
	}

	status, resp, err := client.GeoSites.GetStatusOfGeoSite(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, status)
}
