package gitlab

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProjectSecuritySettings_ListProjectSecuritySettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	createdAt := time.Date(2024, time.October, 22, 14, 13, 35, 0, time.UTC)

	mux.HandleFunc("/api/v4/projects/1/security_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"project_id": 7,
			"created_at": "2024-10-22T14:13:35.000Z",
			"updated_at": "2024-10-22T14:13:35.000Z",
			"auto_fix_container_scanning": true,
			"auto_fix_dast": true,
			"auto_fix_dependency_scanning": true,
			"auto_fix_sast": true,
			"continuous_vulnerability_scans_enabled": true,
			"container_scanning_for_registry_enabled": false,
			"secret_push_protection_enabled": true
		}`)
	})

	want := &ProjectSecuritySettings{
		ProjectID:                           7,
		CreatedAt:                           &createdAt,
		UpdatedAt:                           &createdAt,
		AutoFixContainerScanning:            true,
		AutoFixDAST:                         true,
		AutoFixDependencyScanning:           true,
		AutoFixSAST:                         true,
		ContinuousVulnerabilityScansEnabled: true,
		ContainerScanningForRegistryEnabled: false,
		SecretPushProtectionEnabled:         true,
	}

	d, resp, err := client.ProjectSecuritySettings.ListProjectSecuritySettings(1)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, d)
}

func TestProjectSecuritySettings_UpdateSecretPushProtectionEnabledSetting(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	createdAt := time.Date(2024, time.October, 22, 14, 13, 35, 0, time.UTC)

	mux.HandleFunc("/api/v4/projects/1/security_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
			"project_id": 7,
			"created_at": "2024-10-22T14:13:35.000Z",
			"updated_at": "2024-10-22T14:13:35.000Z",
			"auto_fix_container_scanning": true,
			"auto_fix_dast": true,
			"auto_fix_dependency_scanning": true,
			"auto_fix_sast": true,
			"continuous_vulnerability_scans_enabled": true,
			"container_scanning_for_registry_enabled": false,
			"secret_push_protection_enabled": true
		}`)
	})

	want := &ProjectSecuritySettings{
		ProjectID:                           7,
		CreatedAt:                           &createdAt,
		UpdatedAt:                           &createdAt,
		AutoFixContainerScanning:            true,
		AutoFixDAST:                         true,
		AutoFixDependencyScanning:           true,
		AutoFixSAST:                         true,
		ContinuousVulnerabilityScansEnabled: true,
		ContainerScanningForRegistryEnabled: false,
		SecretPushProtectionEnabled:         true,
	}

	d, resp, err := client.ProjectSecuritySettings.UpdateSecretPushProtectionEnabledSetting(1, UpdateProjectSecuritySettingsOptions{
		SecretPushProtectionEnabled: Ptr(true),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, d)
}
