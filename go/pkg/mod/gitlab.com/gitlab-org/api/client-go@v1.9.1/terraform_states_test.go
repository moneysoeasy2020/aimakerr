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

func TestTerraformState_List(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
			{
				"data": {
					"project": {
						"terraformStates": {
							"nodes": [
								{
									"name": "production",
									"createdAt": "2025-05-25T13:47:41Z",
									"deletedAt": null,
									"latestVersion": {
										"createdAt": "2025-05-25T13:47:41Z",
										"updatedAt": "2025-05-25T13:47:41Z",
										"downloadPath": "/api/v4/projects/54006670/terraform/state/production/versions/1",
										"serial": 1
									},
									"updatedAt": "2025-05-25T13:47:41Z",
									"lockedAt": null
								},
								{
									"name": "staging",
									"createdAt": "2025-05-25T13:47:41Z",
									"deletedAt": null,
									"latestVersion": {
										"createdAt": "2025-05-25T13:47:41Z",
										"updatedAt": "2025-05-25T13:47:41Z",
										"downloadPath": "/api/v4/projects/54006670/terraform/state/staging/versions/7",
										"serial": 7
									},
									"updatedAt": "2025-05-25T13:47:41Z",
									"lockedAt": null
								}
							]
						}
					}
				},
				"correlationId": "bdee71750f5428cb8bcfdbd88ef2ef7a"
			}
		`)
	})

	response, _, err := client.TerraformStates.List("timofurrer/opentofu-test")
	require.NoError(t, err)

	testTime, err := time.Parse(time.RFC3339, "2025-05-25T13:47:41Z")
	require.NoError(t, err)

	want := []TerraformState{
		{
			Name:      "production",
			CreatedAt: testTime,
			LatestVersion: TerraformStateVersion{
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
				DownloadPath: "/api/v4/projects/54006670/terraform/state/production/versions/1",
				Serial:       1,
			},
			UpdatedAt: testTime,
		},
		{
			Name:      "staging",
			CreatedAt: testTime,
			LatestVersion: TerraformStateVersion{
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
				DownloadPath: "/api/v4/projects/54006670/terraform/state/staging/versions/7",
				Serial:       7,
			},
			UpdatedAt: testTime,
		},
	}
	assert.Equal(t, want, response)
}

func TestTerraformState_Get(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
			{
				"data": {
					"project": {
						"terraformState": {
							"name": "production",
							"createdAt": "2025-05-25T13:47:41Z",
							"deletedAt": null,
							"latestVersion": {
								"createdAt": "2025-05-25T13:47:41Z",
								"updatedAt": "2025-05-25T13:47:41Z",
								"downloadPath": "/api/v4/projects/54006670/terraform/state/production/versions/1",
								"serial": 1
							},
							"updatedAt": "2025-05-25T13:47:41Z",
							"lockedAt": null
						}
					}
				},
				"correlationId": "504812e16a2d0ca50f1316b625d7a899"
			}
		`)
	})

	response, _, err := client.TerraformStates.Get("timofurrer/opentofu-test", "production")
	require.NoError(t, err)

	testTime, err := time.Parse(time.RFC3339, "2025-05-25T13:47:41Z")
	require.NoError(t, err)

	want := &TerraformState{
		Name:      "production",
		CreatedAt: testTime,
		LatestVersion: TerraformStateVersion{
			CreatedAt:    testTime,
			UpdatedAt:    testTime,
			DownloadPath: "/api/v4/projects/54006670/terraform/state/production/versions/1",
			Serial:       1,
		},
		UpdatedAt: testTime,
	}
	assert.Equal(t, want, response)
}

func TestTerraformState_DownloadLatest(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/20/terraform/state/production", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"some": "value"}`)
	})

	r, _, err := client.TerraformStates.DownloadLatest(20, "production")
	require.NoError(t, err)

	data, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.JSONEq(t, `{"some": "value"}`, string(data))
}

func TestTerraformState_Download(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/20/terraform/state/production/versions/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"some": "value"}`)
	})

	r, _, err := client.TerraformStates.Download(20, "production", 42)
	require.NoError(t, err)

	data, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.JSONEq(t, `{"some": "value"}`, string(data))
}

func TestTerraformState_Delete(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/20/terraform/state/production", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.TerraformStates.Delete(20, "production")
	assert.NoError(t, err)
}

func TestTerraformState_DeleteVersion(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/20/terraform/state/production/versions/42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.TerraformStates.DeleteVersion(20, "production", uint64(42))
	assert.NoError(t, err)
}

func TestTerraformState_Lock(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/20/terraform/state/production/lock", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
	})

	_, err := client.TerraformStates.Lock(20, "production")
	assert.NoError(t, err)
}

func TestTerraformState_Unlock(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/20/terraform/state/production/lock", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.TerraformStates.Unlock(20, "production")
	assert.NoError(t, err)
}
