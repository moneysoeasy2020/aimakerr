package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupSecuritySettings_UpdateSecretPushProtectionEnabledSetting(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/security_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
			"secret_push_protection_enabled": true,
			"errors": []
		}`)
	})

	want := &GroupSecuritySettings{
		SecretPushProtectionEnabled: true,
		Errors:                      []string{},
	}

	d, resp, err := client.GroupSecuritySettings.UpdateSecretPushProtectionEnabledSetting(1, UpdateGroupSecuritySettingsOptions{
		SecretPushProtectionEnabled: Ptr(true),
		ProjectsToExclude:           Ptr([]int64{1, 2}),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, d)
}
