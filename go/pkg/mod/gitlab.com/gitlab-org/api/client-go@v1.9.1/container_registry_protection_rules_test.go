package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListContainerRegistryProtectionRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/7/registry/protection/repository/rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `[
				{
					"id": 1,
					"project_id": 7,
					"repository_path_pattern": "flightjs/flight0",
					"minimum_access_level_for_push": "maintainer",
					"minimum_access_level_for_delete": "owner"
				},
				{
					"id": 2,
					"project_id": 7,
					"repository_path_pattern": "flightjs/flight1",
					"minimum_access_level_for_push": "maintainer",
					"minimum_access_level_for_delete": "owner"
				}
			]`)
	})

	want := []*ContainerRegistryProtectionRule{
		{
			ID:                          1,
			ProjectID:                   7,
			RepositoryPathPattern:       "flightjs/flight0",
			MinimumAccessLevelForPush:   "maintainer",
			MinimumAccessLevelForDelete: "owner",
		},
		{
			ID:                          2,
			ProjectID:                   7,
			RepositoryPathPattern:       "flightjs/flight1",
			MinimumAccessLevelForPush:   "maintainer",
			MinimumAccessLevelForDelete: "owner",
		},
	}

	rules, resp, err := client.ContainerRegistryProtectionRules.ListContainerRegistryProtectionRules(7)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, rules)
}

func TestCreateContainerRegistryProtectionRule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/7/registry/protection/repository/rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `
			{
				"id": 1,
				"project_id": 7,
				"repository_path_pattern": "flightjs/flight0",
				"minimum_access_level_for_push": "maintainer",
				"minimum_access_level_for_delete": "admin"
			}
		`)
	})

	want := &ContainerRegistryProtectionRule{
		ID:                          1,
		ProjectID:                   7,
		RepositoryPathPattern:       "flightjs/flight0",
		MinimumAccessLevelForPush:   "maintainer",
		MinimumAccessLevelForDelete: "admin",
	}

	rule, resp, err := client.ContainerRegistryProtectionRules.CreateContainerRegistryProtectionRule(7, &CreateContainerRegistryProtectionRuleOptions{
		RepositoryPathPattern:       Ptr("flightjs/flight0"),
		MinimumAccessLevelForPush:   Ptr(ProtectionRuleAccessLevelMaintainer),
		MinimumAccessLevelForDelete: Ptr(ProtectionRuleAccessLevelAdmin),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, rule)
}

func TestUpdateContainerRegistryProtectionRule(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/7/registry/protection/repository/rules/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		fmt.Fprintf(w, `
			{
				"id": 1,
				"project_id": 7,
				"repository_path_pattern": "flightjs/flight0",
				"minimum_access_level_for_push": "maintainer",
				"minimum_access_level_for_delete": "owner"
			}
		`)
	})

	want := &ContainerRegistryProtectionRule{
		ID:                          1,
		ProjectID:                   7,
		RepositoryPathPattern:       "flightjs/flight0",
		MinimumAccessLevelForPush:   "maintainer",
		MinimumAccessLevelForDelete: "owner",
	}

	rule, resp, err := client.ContainerRegistryProtectionRules.UpdateContainerRegistryProtectionRule(7, 1, &UpdateContainerRegistryProtectionRuleOptions{
		RepositoryPathPattern:       Ptr("flightjs/flight0"),
		MinimumAccessLevelForPush:   Ptr(ProtectionRuleAccessLevelMaintainer),
		MinimumAccessLevelForDelete: Ptr(ProtectionRuleAccessLevelOwner),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, rule)
}

func TestDeleteContainerRegistryProtectionRule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/7/registry/protection/repository/rules/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.ContainerRegistryProtectionRules.DeleteContainerRegistryProtectionRule(7, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
