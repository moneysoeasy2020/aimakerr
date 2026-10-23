package gitlab

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFeatureFlagUserLists_ListFeatureFlagUserLists(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/feature_flags_user_lists", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `[
			{
				"name": "user_list",
				"user_xids": "user1,user2",
				"id": 1,
				"iid": 1,
				"project_id": 1,
				"created_at": "2020-02-04T08:13:51.000Z",
				"updated_at": "2020-02-04T08:13:51.000Z"
			},
			{
				"name": "test_users",
				"user_xids": "user3,user4,user5",
				"id": 2,
				"iid": 2,
				"project_id": 1,
				"created_at": "2020-02-04T08:13:51.000Z",
				"updated_at": "2020-02-04T08:13:51.000Z"
			}
		]`)
	})

	createdAt := time.Date(2020, time.February, 4, 8, 13, 51, 0, time.UTC)
	want := []*FeatureFlagUserList{
		{
			Name:      "user_list",
			UserXIDs:  "user1,user2",
			ID:        1,
			IID:       1,
			ProjectID: 1,
			CreatedAt: &createdAt,
			UpdatedAt: &createdAt,
		},
		{
			Name:      "test_users",
			UserXIDs:  "user3,user4,user5",
			ID:        2,
			IID:       2,
			ProjectID: 1,
			CreatedAt: &createdAt,
			UpdatedAt: &createdAt,
		},
	}

	lists, resp, err := client.FeatureFlagUserLists.ListFeatureFlagUserLists(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, lists)
}

func TestFeatureFlagUserLists_CreateFeatureFlagUserList(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/feature_flags_user_lists", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `
			{
				"name": "user_list",
				"user_xids": "user1,user2",
				"id": 1,
				"iid": 1,
				"project_id": 1,
				"created_at": "2020-02-04T08:13:51.000Z",
				"updated_at": "2020-02-04T08:13:51.000Z"
			}
		`)
	})

	createdAt := time.Date(2020, time.February, 4, 8, 13, 51, 0, time.UTC)
	want := &FeatureFlagUserList{
		Name:      "user_list",
		UserXIDs:  "user1,user2",
		ID:        1,
		IID:       1,
		ProjectID: 1,
		CreatedAt: &createdAt,
		UpdatedAt: &createdAt,
	}

	list, resp, err := client.FeatureFlagUserLists.CreateFeatureFlagUserList(1, &CreateFeatureFlagUserListOptions{
		Name:     "user_list",
		UserXIDs: "user1,user2",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, list)
}

func TestFeatureFlagUserLists_GetFeatureFlagUserList(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/feature_flags_user_lists/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"name": "user_list",
				"user_xids": "user1,user2",
				"id": 1,
				"iid": 1,
				"project_id": 1,
				"created_at": "2020-02-04T08:13:51.000Z",
				"updated_at": "2020-02-04T08:13:51.000Z"
			}
		`)
	})

	createdAt := time.Date(2020, time.February, 4, 8, 13, 51, 0, time.UTC)
	want := &FeatureFlagUserList{
		Name:      "user_list",
		UserXIDs:  "user1,user2",
		ID:        1,
		IID:       1,
		ProjectID: 1,
		CreatedAt: &createdAt,
		UpdatedAt: &createdAt,
	}

	list, resp, err := client.FeatureFlagUserLists.GetFeatureFlagUserList(1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, list)
}

func TestFeatureFlagUserLists_UpdateFeatureFlagUserList(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/feature_flags_user_lists/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `
			{
				"name": "user_list",
				"user_xids": "user1,user2",
				"id": 1,
				"iid": 1,
				"project_id": 1,
				"created_at": "2020-02-04T08:13:51.000Z",
				"updated_at": "2020-02-04T08:13:51.000Z"
			}
		`)
	})

	createdAt := time.Date(2020, time.February, 4, 8, 13, 51, 0, time.UTC)
	want := &FeatureFlagUserList{
		Name:      "user_list",
		UserXIDs:  "user1,user2",
		ID:        1,
		IID:       1,
		ProjectID: 1,
		CreatedAt: &createdAt,
		UpdatedAt: &createdAt,
	}

	list, resp, err := client.FeatureFlagUserLists.UpdateFeatureFlagUserList(1, 1, &UpdateFeatureFlagUserListOptions{
		Name:     "user_list",
		UserXIDs: "user1,user2",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, list)
}

func TestFeatureFlagUserLists_DeleteFeatureFlagUserList(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/feature_flags_user_lists/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.FeatureFlagUserLists.DeleteFeatureFlagUserList(1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
