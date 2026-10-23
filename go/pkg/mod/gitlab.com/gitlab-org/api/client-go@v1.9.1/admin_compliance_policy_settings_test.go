package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminCompliancePolicySettingsService_GetCompliancePolicySettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/admin/security/compliance_policy_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"csp_namespace_id": 42
			}
		`)
	})

	cspNamespaceID := int64(42)
	want := &AdminCompliancePolicySettings{
		CSPNamespaceID: &cspNamespaceID,
	}

	settings, resp, err := client.AdminCompliancePolicySettings.GetCompliancePolicySettings()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, settings)

	settings, resp, err = client.AdminCompliancePolicySettings.GetCompliancePolicySettings(errorOption)
	require.EqualError(t, err, "RequestOptionFunc returns an error")
	require.Nil(t, resp)
	require.Nil(t, settings)
}

func TestAdminCompliancePolicySettingsService_GetCompliancePolicySettings_NullNamespace(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/admin/security/compliance_policy_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"csp_namespace_id": null
			}
		`)
	})

	want := &AdminCompliancePolicySettings{
		CSPNamespaceID: nil,
	}

	settings, resp, err := client.AdminCompliancePolicySettings.GetCompliancePolicySettings()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, settings)
}

func TestAdminCompliancePolicySettingsService_GetCompliancePolicySettings_StatusNotFound(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/admin/security/compliance_policy_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusNotFound)
	})

	settings, resp, err := client.AdminCompliancePolicySettings.GetCompliancePolicySettings()
	require.Error(t, err)
	require.Nil(t, settings)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAdminCompliancePolicySettingsService_UpdateCompliancePolicySettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/admin/security/compliance_policy_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `
			{
				"csp_namespace_id": 42
			}
		`)
	})

	cspNamespaceID := int64(42)
	want := &AdminCompliancePolicySettings{
		CSPNamespaceID: &cspNamespaceID,
	}

	opt := &UpdateAdminCompliancePolicySettingsOptions{
		CSPNamespaceID: &cspNamespaceID,
	}

	settings, resp, err := client.AdminCompliancePolicySettings.UpdateCompliancePolicySettings(opt)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, settings)

	settings, resp, err = client.AdminCompliancePolicySettings.UpdateCompliancePolicySettings(opt, errorOption)
	require.EqualError(t, err, "RequestOptionFunc returns an error")
	require.Nil(t, resp)
	require.Nil(t, settings)
}

func TestAdminCompliancePolicySettingsService_UpdateCompliancePolicySettings_ClearNamespace(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/admin/security/compliance_policy_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `
			{
				"csp_namespace_id": null
			}
		`)
	})

	want := &AdminCompliancePolicySettings{
		CSPNamespaceID: nil,
	}

	opt := &UpdateAdminCompliancePolicySettingsOptions{
		CSPNamespaceID: nil,
	}

	settings, resp, err := client.AdminCompliancePolicySettings.UpdateCompliancePolicySettings(opt)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, settings)
}

func TestAdminCompliancePolicySettingsService_UpdateCompliancePolicySettings_StatusInternalServerError(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/admin/security/compliance_policy_settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.WriteHeader(http.StatusInternalServerError)
	})

	cspNamespaceID := int64(42)
	opt := &UpdateAdminCompliancePolicySettingsOptions{
		CSPNamespaceID: &cspNamespaceID,
	}

	settings, resp, err := client.AdminCompliancePolicySettings.UpdateCompliancePolicySettings(opt)
	require.Error(t, err)
	require.Nil(t, settings)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
