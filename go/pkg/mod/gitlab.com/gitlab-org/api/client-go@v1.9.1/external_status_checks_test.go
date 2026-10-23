package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListMergeStatusChecks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/1/status_checks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, exampleProjectMergeRequestStatusChecksList)
	})

	statusChecks, _, err := client.ExternalStatusChecks.ListMergeStatusChecks(1, 1, nil)
	if err != nil {
		t.Fatalf("ExternalStatusChecks.ListMergeStatusChecks returns an error: %v", err)
	}

	expectedStatusChecks := []*MergeStatusCheck{
		{
			ID:          2,
			Name:        "Rule 1",
			ExternalURL: "https://gitlab.com/test-endpoint",
			Status:      "approved",
		},
		{
			ID:          1,
			Name:        "Rule 2",
			ExternalURL: "https://gitlab.com/test-endpoint-2",
			Status:      "pending",
		},
	}

	assert.Equal(t, expectedStatusChecks, statusChecks)
}

func TestListProjectStatusChecks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/external_status_checks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, exampleProjectStatusChecksList)
	})

	projectStatusChecks, _, err := client.ExternalStatusChecks.ListProjectStatusChecks(1, nil)
	if err != nil {
		t.Fatalf("ExternalStatusChecks.ListProjectStatusChecks returns an error: %v", err)
	}

	expectedProjectStatusChecks := []*ProjectStatusCheck{
		{
			ID:          1,
			Name:        "Compliance Check",
			ProjectID:   6,
			ExternalURL: "https://gitlab.com/example/test.json",
			HMAC:        false,
			ProtectedBranches: []StatusCheckProtectedBranch{
				{
					ID:                        14,
					ProjectID:                 6,
					Name:                      "master",
					CreatedAt:                 mustParseTime("2020-10-12T14:04:50.787Z"),
					UpdatedAt:                 mustParseTime("2020-10-12T14:04:50.787Z"),
					CodeOwnerApprovalRequired: false,
				},
			},
		},
	}

	assert.Equal(t, expectedProjectStatusChecks, projectStatusChecks)
}

func TestRetryFailedStatusCheckForAMergeRequest(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/2/status_checks/3/retry", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"message": "202 Accepted"}`)
	})

	resp, err := client.ExternalStatusChecks.RetryFailedStatusCheckForAMergeRequest(1, 2, 3)
	if err != nil {
		t.Fatalf("ExternalStatusChecks.RetryFailedStatusCheckForAMergeRequest returns an error: %v", err)
	}

	assert.NotNil(t, resp)
}

func TestCreateProjectExternalStatusChecks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/6/external_status_checks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, exampleProjectStatusCheck)
	})

	projectStatusCheck, _, err := client.ExternalStatusChecks.CreateProjectExternalStatusCheck(6, &CreateProjectExternalStatusCheckOptions{
		Name:         Ptr("Compliance Check"),
		ExternalURL:  Ptr("https://gitlab.com/example/test.json"),
		SharedSecret: Ptr("HMAC"),
	})
	if err != nil {
		t.Fatalf("ExternalStatusChecks.CreateProjectExternalStatusCheck returns an error: %v", err)
	}

	expectedProjectStatusCheck := &ProjectStatusCheck{
		ID:          1,
		Name:        "Compliance Check",
		ProjectID:   6,
		ExternalURL: "https://gitlab.com/example/test.json",
		HMAC:        true,
		ProtectedBranches: []StatusCheckProtectedBranch{
			{
				ID:                        14,
				ProjectID:                 6,
				Name:                      "master",
				CreatedAt:                 mustParseTime("2020-10-12T14:04:50.787Z"),
				UpdatedAt:                 mustParseTime("2020-10-12T14:04:50.787Z"),
				CodeOwnerApprovalRequired: false,
			},
		},
	}

	assert.Equal(t, expectedProjectStatusCheck, projectStatusCheck)
}

func TestUpdateProjectExternalStatusChecks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/6/external_status_checks/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, exampleProjectStatusCheck)
	})

	projectStatusCheck, _, err := client.ExternalStatusChecks.UpdateProjectExternalStatusCheck(6, 1, &UpdateProjectExternalStatusCheckOptions{
		Name:         Ptr("Compliance Check"),
		ExternalURL:  Ptr("https://gitlab.com/example/test.json"),
		SharedSecret: Ptr("HMAC"),
	})
	if err != nil {
		t.Fatalf("ExternalStatusChecks.UpdateProjectExternalStatusCheck returns an error: %v", err)
	}

	expectedProjectStatusCheck := &ProjectStatusCheck{
		ID:          1,
		Name:        "Compliance Check",
		ProjectID:   6,
		ExternalURL: "https://gitlab.com/example/test.json",
		HMAC:        true,
		ProtectedBranches: []StatusCheckProtectedBranch{
			{
				ID:                        14,
				ProjectID:                 6,
				Name:                      "master",
				CreatedAt:                 mustParseTime("2020-10-12T14:04:50.787Z"),
				UpdatedAt:                 mustParseTime("2020-10-12T14:04:50.787Z"),
				CodeOwnerApprovalRequired: false,
			},
		},
	}

	assert.Equal(t, expectedProjectStatusCheck, projectStatusCheck)
}

func TestListProjectMergeRequestExternalStatusChecks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/1/status_checks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, exampleProjectMergeRequestStatusChecksList)
	})

	statusChecks, _, err := client.ExternalStatusChecks.ListProjectMergeRequestExternalStatusChecks(1, 1, &ListProjectMergeRequestExternalStatusChecksOptions{})
	if err != nil {
		t.Fatalf("ExternalStatusChecks.ListProjectMergeRequestExternalStatusChecks returns an error: %v", err)
	}

	expectedStatusChecks := []*MergeStatusCheck{
		{
			ID:          2,
			Name:        "Rule 1",
			ExternalURL: "https://gitlab.com/test-endpoint",
			Status:      "approved",
		},
		{
			ID:          1,
			Name:        "Rule 2",
			ExternalURL: "https://gitlab.com/test-endpoint-2",
			Status:      "pending",
		},
	}

	assert.Equal(t, expectedStatusChecks, statusChecks)
}

func TestListProjectExternalStatusChecks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/external_status_checks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, exampleProjectStatusChecksList)
	})

	projectStatusChecks, _, err := client.ExternalStatusChecks.ListProjectExternalStatusChecks(1, &ListProjectExternalStatusChecksOptions{})
	if err != nil {
		t.Fatalf("ExternalStatusChecks.ListProjectExternalStatusChecks returns an error: %v", err)
	}

	expectedProjectStatusChecks := []*ProjectStatusCheck{
		{
			ID:          1,
			Name:        "Compliance Check",
			ProjectID:   6,
			ExternalURL: "https://gitlab.com/example/test.json",
			HMAC:        false,
			ProtectedBranches: []StatusCheckProtectedBranch{
				{
					ID:                        14,
					ProjectID:                 6,
					Name:                      "master",
					CreatedAt:                 mustParseTime("2020-10-12T14:04:50.787Z"),
					UpdatedAt:                 mustParseTime("2020-10-12T14:04:50.787Z"),
					CodeOwnerApprovalRequired: false,
				},
			},
		},
	}

	assert.Equal(t, expectedProjectStatusChecks, projectStatusChecks)
}

func TestRetryFailedExternalStatusCheckForProjectMergeRequest(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/2/status_checks/3/retry", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"message": "202 Accepted"}`)
	})

	resp, err := client.ExternalStatusChecks.RetryFailedExternalStatusCheckForProjectMergeRequest(1, 2, 3, &RetryFailedExternalStatusCheckForProjectMergeRequestOptions{})
	if err != nil {
		t.Fatalf("ExternalStatusChecks.RetryFailedExternalStatusCheckForProjectMergeRequest returns an error: %v", err)
	}

	assert.NotNil(t, resp)
}
