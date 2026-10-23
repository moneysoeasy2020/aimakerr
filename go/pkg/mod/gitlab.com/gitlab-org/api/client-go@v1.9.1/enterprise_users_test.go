package gitlab

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnterpriseUsers_ListEnterpriseUsers(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/enterprise_users", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `[
			{
				"id": 66,
				"username": "user22",
				"name": "Sidney Jones22",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/xxx?s=80&d=identicon",
				"web_url": "http://my.gitlab.com/user22",
				"created_at": "2021-09-10T12:48:22.000Z",
				"confirmed_at": "2021-09-10T12:48:22.000Z",
				"email": "user22@example.org",
				"theme_id": 1,
				"color_scheme_id": 1,
				"projects_limit": 100000,
				"identities": [
					{
						"provider": "group_saml",
						"extern_uid": "2435223452345"
					}
				],
				"can_create_group": true,
				"can_create_project": true,
				"can_create_organization": true,
				"two_factor_enabled": false,
				"external": false,
				"private_profile": false,
				"commit_email": "user22@example.org"
			}
		]`)
	})

	date := time.Date(2021, time.September, 10, 12, 48, 22, 0, time.UTC)
	want := []*User{
		{
			ID:            66,
			Username:      "user22",
			Name:          "Sidney Jones22",
			State:         "active",
			AvatarURL:     "https://www.gravatar.com/avatar/xxx?s=80&d=identicon",
			WebURL:        "http://my.gitlab.com/user22",
			CreatedAt:     &date,
			ConfirmedAt:   &date,
			Email:         "user22@example.org",
			ThemeID:       1,
			ColorSchemeID: 1,
			ProjectsLimit: 100000,
			Identities: []*UserIdentity{
				{
					Provider:  "group_saml",
					ExternUID: "2435223452345",
				},
			},
			CanCreateGroup:        true,
			CanCreateProject:      true,
			CanCreateOrganization: true,
			TwoFactorEnabled:      false,
			External:              false,
			PrivateProfile:        false,
		},
	}

	users, resp, err := client.EnterpriseUsers.ListEnterpriseUsers(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, users)
}

func TestEnterpriseUsers_GetEnterpriseUser(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/enterprise_users/66", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"id": 66,
				"username": "user22",
				"name": "Sidney Jones22",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/xxx?s=80&d=identicon",
				"web_url": "http://my.gitlab.com/user22",
				"created_at": "2021-09-10T12:48:22.000Z",
				"confirmed_at": "2021-09-10T12:48:22.000Z",
				"email": "user22@example.org",
				"theme_id": 1,
				"color_scheme_id": 1,
				"projects_limit": 100000,
				"identities": [
					{
						"provider": "group_saml",
						"extern_uid": "2435223452345"
					}
				],
				"can_create_group": true,
				"can_create_project": true,
				"two_factor_enabled": false,
				"external": false,
				"private_profile": false,
				"commit_email": "user22@example.org"
			}
		`)
	})

	date := time.Date(2021, time.September, 10, 12, 48, 22, 0, time.UTC)
	want := &User{
		ID:            66,
		Username:      "user22",
		Name:          "Sidney Jones22",
		State:         "active",
		AvatarURL:     "https://www.gravatar.com/avatar/xxx?s=80&d=identicon",
		WebURL:        "http://my.gitlab.com/user22",
		CreatedAt:     &date,
		ConfirmedAt:   &date,
		Email:         "user22@example.org",
		ThemeID:       1,
		ColorSchemeID: 1,
		ProjectsLimit: 100000,
		Identities: []*UserIdentity{
			{
				Provider:  "group_saml",
				ExternUID: "2435223452345",
			},
		},
		CanCreateGroup:   true,
		CanCreateProject: true,
		TwoFactorEnabled: false,
		External:         false,
		PrivateProfile:   false,
	}

	user, resp, err := client.EnterpriseUsers.GetEnterpriseUser(1, 66)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, user)
}

func TestEnterpriseUsers_Disable2FAForEnterpriseUser(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/enterprise_users/66/disable_two_factor", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
	})

	resp, err := client.EnterpriseUsers.Disable2FAForEnterpriseUser(1, 66)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnterpriseUsers_DeleteEnterpriseUser(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/enterprise_users/66", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/groups/1/enterprise_users/66?hard_delete=true")
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	hard := true
	resp, err := client.EnterpriseUsers.DeleteEnterpriseUser(1, 66, &DeleteEnterpriseUserOptions{HardDelete: &hard})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
