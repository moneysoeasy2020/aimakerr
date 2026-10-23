package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectMirrorService_ListProjectMirror(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/42/remote_mirrors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			[
			  {
				"enabled": true,
				"id": 101486,
				"last_error": null,
				"only_protected_branches": true,
				"keep_divergent_refs": true,
				"update_status": "finished",
				"auth_method": "password",
				"url": "https://*****:*****@gitlab.com/gitlab-org/security/gitlab.git"
			  }
			]
		`)
	})

	want := []*ProjectMirror{{
		Enabled:               true,
		ID:                    101486,
		LastError:             "",
		OnlyProtectedBranches: true,
		KeepDivergentRefs:     true,
		UpdateStatus:          "finished",
		AuthMethod:            "password",
		URL:                   "https://*****:*****@gitlab.com/gitlab-org/security/gitlab.git",
	}}

	pms, resp, err := client.ProjectMirrors.ListProjectMirror(42, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pms)

	pms, resp, err = client.ProjectMirrors.ListProjectMirror(42.01, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pms)

	pms, resp, err = client.ProjectMirrors.ListProjectMirror(42, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pms)

	pms, resp, err = client.ProjectMirrors.ListProjectMirror(43, nil, nil)
	require.Error(t, err)
	require.Nil(t, pms)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMirrorService_GetProjectMirror(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/42/remote_mirrors/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"enabled": true,
				"id": 101486,
				"last_error": null,
				"only_protected_branches": true,
				"keep_divergent_refs": true,
				"update_status": "finished",
				"auth_method": "password",
				"url": "https://*****:*****@gitlab.com/gitlab-org/security/gitlab.git"
			}
		`)
	})

	want := &ProjectMirror{
		Enabled:               true,
		ID:                    101486,
		LastError:             "",
		OnlyProtectedBranches: true,
		KeepDivergentRefs:     true,
		UpdateStatus:          "finished",
		AuthMethod:            "password",
		URL:                   "https://*****:*****@gitlab.com/gitlab-org/security/gitlab.git",
	}

	pm, resp, err := client.ProjectMirrors.GetProjectMirror(42, 1, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pm)
}

func TestProjectMirrorService_GetProjectMirrorPublicKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/42/remote_mirrors/1/public_key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"public_key": "ssh-rsa AAAA..."
			}
		`)
	})

	want := &ProjectMirrorPublicKey{
		PublicKey: "ssh-rsa AAAA...",
	}

	pm, resp, err := client.ProjectMirrors.GetProjectMirrorPublicKey(42, 1, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pm)
}

func TestProjectMirrorService_AddProjectMirror(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/42/remote_mirrors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `
			{
				"enabled": false,
				"id": 101486,
				"last_error": null,
				"last_successful_update_at": null,
				"last_update_at": null,
				"last_update_started_at": null,
				"only_protected_branches": false,
				"keep_divergent_refs": false,
				"update_status": "none",
				"auth_method": "password",
				"url": "https://*****:*****@example.com/gitlab/example.git"
			}
		`)
	})

	want := &ProjectMirror{
		Enabled:                false,
		ID:                     101486,
		LastError:              "",
		LastSuccessfulUpdateAt: nil,
		LastUpdateAt:           nil,
		LastUpdateStartedAt:    nil,
		OnlyProtectedBranches:  false,
		KeepDivergentRefs:      false,
		UpdateStatus:           "none",
		AuthMethod:             "password",
		URL:                    "https://*****:*****@example.com/gitlab/example.git",
	}

	pm, resp, err := client.ProjectMirrors.AddProjectMirror(42, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pm)

	pm, resp, err = client.ProjectMirrors.AddProjectMirror(42.01, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMirrors.AddProjectMirror(42, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMirrors.AddProjectMirror(43, nil, nil)
	require.Error(t, err)
	require.Nil(t, pm)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMirrorService_EditProjectMirror(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/42/remote_mirrors/101486", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `
			{
				"enabled": false,
				"id": 101486,
				"last_error": null,
				"only_protected_branches": true,
				"keep_divergent_refs": true,
				"update_status": "finished",
				"auth_method": "password",
				"url": "https://*****:*****@gitlab.com/gitlab-org/security/gitlab.git"
			}
		`)
	})

	want := &ProjectMirror{
		Enabled:               false,
		ID:                    101486,
		LastError:             "",
		OnlyProtectedBranches: true,
		KeepDivergentRefs:     true,
		UpdateStatus:          "finished",
		AuthMethod:            "password",
		URL:                   "https://*****:*****@gitlab.com/gitlab-org/security/gitlab.git",
	}

	pm, resp, err := client.ProjectMirrors.EditProjectMirror(42, 101486, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pm)

	pm, resp, err = client.ProjectMirrors.EditProjectMirror(42.01, 101486, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMirrors.EditProjectMirror(42, 101486, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMirrors.EditProjectMirror(43, 101486, nil, nil)
	require.Error(t, err)
	require.Nil(t, pm)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
