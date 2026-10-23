package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupSCIM_GetSCIMIdentitiesForGroup(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/scim/identities", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
			{
				"external_uid": "be20d8dcc028677c931e04f387",
				"user_id": 48,
				"active": true
			}
		]`)
	})

	want := &GroupSCIMIdentity{
		ExternalUID: "be20d8dcc028677c931e04f387",
		UserID:      48,
		Active:      true,
	}

	identities, resp, err := client.GroupSCIM.GetSCIMIdentitiesForGroup(1)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, identities[0])
}

func TestGroupSCIM_GetSCIMIdentity(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/scim/be20d8dcc028677c931e04f387", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"external_uid": "be20d8dcc028677c931e04f387",
			"user_id": 48,
			"active": true
		}`)
	})

	want := &GroupSCIMIdentity{
		ExternalUID: "be20d8dcc028677c931e04f387",
		UserID:      48,
		Active:      true,
	}

	identity, resp, err := client.GroupSCIM.GetSCIMIdentity(1, "be20d8dcc028677c931e04f387")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, identity)
}

func TestGroupSCIM_UpdateSCIMIdentity(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/scim/be20d8dcc028677c931e04f387", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
	})

	resp, err := client.GroupSCIM.UpdateSCIMIdentity(1, "be20d8dcc028677c931e04f387", &UpdateSCIMIdentityOptions{ExternUID: Ptr("fa299f2409f25863347")})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestGroupSCIM_DeleteSCIMIdentity(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/scim/be20d8dcc028677c931e04f387", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.GroupSCIM.DeleteSCIMIdentity(1, "be20d8dcc028677c931e04f387")
	require.NoError(t, err)
	require.NotNil(t, resp)
}
