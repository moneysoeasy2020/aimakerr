package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtectedPackagesService_ListPackageProtectionRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/packages/protection/rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			[
			  {
				"id": 1,
				"project_id": 1,
				"package_name_pattern": "@my-scope/my-package-*",
				"package_type": "npm",
				"minimum_access_level_for_delete": "owner",
				"minimum_access_level_for_push": "maintainer"
			  }
			]
		`)
	})

	want := []*PackageProtectionRule{{
		ID:                          1,
		ProjectID:                   1,
		PackageNamePattern:          "@my-scope/my-package-*",
		PackageType:                 "npm",
		MinimumAccessLevelForDelete: "owner",
		MinimumAccessLevelForPush:   "maintainer",
	}}

	rules, resp, err := client.ProtectedPackages.ListPackageProtectionRules(1, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, rules)
}

func TestProtectedPackagesService_CreatePackageProtectionRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/packages/protection/rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
			{
				"id": 1,
				"project_id": 1,
				"package_name_pattern": "@my-scope/my-package-*",
				"package_type": "npm",
				"minimum_access_level_for_delete": "owner",
				"minimum_access_level_for_push": "maintainer"
			}
		`)
	})

	want := &PackageProtectionRule{
		ID:                          1,
		ProjectID:                   1,
		PackageNamePattern:          "@my-scope/my-package-*",
		PackageType:                 "npm",
		MinimumAccessLevelForDelete: "owner",
		MinimumAccessLevelForPush:   "maintainer",
	}

	opts := &CreatePackageProtectionRulesOptions{
		PackageNamePattern:          Ptr("@my-scope/my-package-*"),
		PackageType:                 Ptr("npm"),
		MinimumAccessLevelForDelete: Ptr(int64(MaintainerPermissions)),
		MinimumAccessLevelForPush:   Ptr(int64(OwnerPermissions)),
	}

	rule, resp, err := client.ProtectedPackages.CreatePackageProtectionRules(1, opts)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, rule)
}

func TestProtectedPackagesService_UpdatePackageProtectionRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/packages/protection/rules/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		fmt.Fprint(w, `
			{
				"id": 123,
				"project_id": 1,
				"package_name_pattern": "@my-scope/my-package-updated",
				"package_type": "npm",
				"minimum_access_level_for_delete": "owner",
				"minimum_access_level_for_push": "owner"
			}
		`)
	})

	want := &PackageProtectionRule{
		ID:                          123,
		ProjectID:                   1,
		PackageNamePattern:          "@my-scope/my-package-updated",
		PackageType:                 "npm",
		MinimumAccessLevelForDelete: "owner",
		MinimumAccessLevelForPush:   "owner",
	}

	opts := &UpdatePackageProtectionRulesOptions{
		PackageNamePattern:        Ptr("@my-scope/my-package-updated"),
		MinimumAccessLevelForPush: Ptr(int64(OwnerPermissions)),
	}

	rule, resp, err := client.ProtectedPackages.UpdatePackageProtectionRules(1, int64(123), opts)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, rule)
}

func TestProtectedPackagesService_DeletePackageProtectionRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/packages/protection/rules/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.ProtectedPackages.DeletePackageProtectionRules(1, int64(123))
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp, err = client.ProtectedPackages.DeletePackageProtectionRules(1.23, int64(123))
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)

	resp, err = client.ProtectedPackages.DeletePackageProtectionRules(1, int64(123), errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)

	resp, err = client.ProtectedPackages.DeletePackageProtectionRules(7, int64(123))
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
