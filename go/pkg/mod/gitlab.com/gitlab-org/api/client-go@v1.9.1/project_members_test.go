package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectMembersService_ListProjectMembers(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			[
			{
				"id": 1,
				"username": "venkatesh_thalluri",
				"name": "Venkatesh Thalluri",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
				"web_url": "http://192.168.1.8:3000/root",
				"created_at": "2012-10-22T14:13:35Z",
				"created_by": {
					"id": 2,
					"username": "john_doe",
					"name": "John Doe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
					"web_url": "http://192.168.1.8:3000/root"
				},
				"access_level": 30,
				"group_saml_identity": null
			}
			]
		`)
	})

	want := []*ProjectMember{{
		ID:          1,
		Username:    "venkatesh_thalluri",
		Email:       "",
		Name:        "Venkatesh Thalluri",
		State:       "active",
		AccessLevel: 30,
		WebURL:      "http://192.168.1.8:3000/root",
		AvatarURL:   "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
		CreatedAt:   mustParseTime("2012-10-22T14:13:35Z"),
		CreatedBy: &MemberCreatedBy{
			ID:        2,
			Username:  "john_doe",
			Name:      "John Doe",
			State:     "active",
			AvatarURL: "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
			WebURL:    "http://192.168.1.8:3000/root",
		},
	}}

	pms, resp, err := client.ProjectMembers.ListProjectMembers(1, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pms)

	pms, resp, err = client.ProjectMembers.ListProjectMembers(1.01, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pms)

	pms, resp, err = client.ProjectMembers.ListProjectMembers(1, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pms)

	pms, resp, err = client.ProjectMembers.ListProjectMembers(2, nil, nil)
	require.Error(t, err)
	require.Nil(t, pms)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMembersService_ListAllProjectMembers(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/members/all", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			[
			{
				"id": 1,
				"username": "venkatesh_thalluri",
				"name": "Venkatesh Thalluri",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
				"web_url": "http://192.168.1.8:3000/root",
				"created_at": "2012-10-22T14:13:35Z",
				"created_by": {
					"id": 2,
					"username": "john_doe",
					"name": "John Doe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
					"web_url": "http://192.168.1.8:3000/root"
				},
				"access_level": 30,
				"group_saml_identity": null
			}
			]
		`)
	})

	want := []*ProjectMember{{
		ID:          1,
		Username:    "venkatesh_thalluri",
		Email:       "",
		Name:        "Venkatesh Thalluri",
		State:       "active",
		AccessLevel: 30,
		WebURL:      "http://192.168.1.8:3000/root",
		AvatarURL:   "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
		CreatedAt:   mustParseTime("2012-10-22T14:13:35Z"),
		CreatedBy: &MemberCreatedBy{
			ID:        2,
			Username:  "john_doe",
			Name:      "John Doe",
			State:     "active",
			AvatarURL: "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
			WebURL:    "http://192.168.1.8:3000/root",
		},
	}}

	pms, resp, err := client.ProjectMembers.ListAllProjectMembers(1, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pms)

	pms, resp, err = client.ProjectMembers.ListAllProjectMembers(1.01, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pms)

	pms, resp, err = client.ProjectMembers.ListAllProjectMembers(1, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pms)

	pms, resp, err = client.ProjectMembers.ListAllProjectMembers(2, nil, nil)
	require.Error(t, err)
	require.Nil(t, pms)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMembersService_GetProjectMember(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/members/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"id": 1,
				"username": "venkatesh_thalluri",
				"name": "Venkatesh Thalluri",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
				"web_url": "http://192.168.1.8:3000/root",
				"created_at": "2012-10-22T14:13:35Z",
				"created_by": {
					"id": 2,
					"username": "john_doe",
					"name": "John Doe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
					"web_url": "http://192.168.1.8:3000/root"
				},
				"access_level": 30,
				"email": "venkatesh.thalluri@example.com",
				"expires_at": null,
				"group_saml_identity": null
			}
		`)
	})

	want := &ProjectMember{
		ID:          1,
		Username:    "venkatesh_thalluri",
		Email:       "venkatesh.thalluri@example.com",
		Name:        "Venkatesh Thalluri",
		State:       "active",
		ExpiresAt:   nil,
		AccessLevel: 30,
		WebURL:      "http://192.168.1.8:3000/root",
		AvatarURL:   "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
		CreatedAt:   mustParseTime("2012-10-22T14:13:35Z"),
		CreatedBy: &MemberCreatedBy{
			ID:        2,
			Username:  "john_doe",
			Name:      "John Doe",
			State:     "active",
			AvatarURL: "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
			WebURL:    "http://192.168.1.8:3000/root",
		},
	}

	pm, resp, err := client.ProjectMembers.GetProjectMember(1, 1, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pm)

	pm, resp, err = client.ProjectMembers.GetProjectMember(1.01, 1, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMembers.GetProjectMember(1, 1, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMembers.GetProjectMember(2, 1, nil, nil)
	require.Error(t, err)
	require.Nil(t, pm)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMembersService_GetInheritedProjectMember(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/members/all/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `
			{
				"id": 1,
				"username": "venkatesh_thalluri",
				"name": "Venkatesh Thalluri",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
				"web_url": "http://192.168.1.8:3000/root",
				"created_at": "2012-10-22T14:13:35Z",
				"created_by": {
					"id": 2,
					"username": "john_doe",
					"name": "John Doe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
					"web_url": "http://192.168.1.8:3000/root"
				},
				"access_level": 30,
				"email": "venkatesh.thalluri@example.com",
				"expires_at": null,
				"group_saml_identity": null
			}
		`)
	})

	want := &ProjectMember{
		ID:          1,
		Username:    "venkatesh_thalluri",
		Email:       "venkatesh.thalluri@example.com",
		Name:        "Venkatesh Thalluri",
		State:       "active",
		ExpiresAt:   nil,
		AccessLevel: 30,
		WebURL:      "http://192.168.1.8:3000/root",
		AvatarURL:   "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
		CreatedAt:   mustParseTime("2012-10-22T14:13:35Z"),
		CreatedBy: &MemberCreatedBy{
			ID:        2,
			Username:  "john_doe",
			Name:      "John Doe",
			State:     "active",
			AvatarURL: "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
			WebURL:    "http://192.168.1.8:3000/root",
		},
	}

	pm, resp, err := client.ProjectMembers.GetInheritedProjectMember(1, 1, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pm)

	pm, resp, err = client.ProjectMembers.GetInheritedProjectMember(1.01, 1, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMembers.GetInheritedProjectMember(1, 1, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMembers.GetInheritedProjectMember(2, 1, nil, nil)
	require.Error(t, err)
	require.Nil(t, pm)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMembersService_AddProjectMember(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	var requestOpts AddProjectMemberOptions
	mux.HandleFunc("/api/v4/projects/1/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		// Read the input arguments and unmarshal them into requestOpts
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		err = json.Unmarshal(body, &requestOpts)
		if err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		fmt.Fprintf(w, `
			{
				"id": 1,
				"username": "venkatesh_thalluri",
				"name": "Venkatesh Thalluri",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
				"web_url": "http://192.168.1.8:3000/root",
				"created_at": "2012-10-22T14:13:35Z",
				"created_by": {
					"id": 2,
					"username": "john_doe",
					"name": "John Doe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
					"web_url": "http://192.168.1.8:3000/root"
				},
				"access_level": 30,
				"email": "venkatesh.thalluri@example.com",
				"expires_at": null,
				"group_saml_identity": null
			}
		`)
	})

	want := &ProjectMember{
		ID:          1,
		Username:    "venkatesh_thalluri",
		Email:       "venkatesh.thalluri@example.com",
		Name:        "Venkatesh Thalluri",
		State:       "active",
		ExpiresAt:   nil,
		AccessLevel: 30,
		WebURL:      "http://192.168.1.8:3000/root",
		AvatarURL:   "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
		CreatedAt:   mustParseTime("2012-10-22T14:13:35Z"),
		CreatedBy: &MemberCreatedBy{
			ID:        2,
			Username:  "john_doe",
			Name:      "John Doe",
			State:     "active",
			AvatarURL: "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
			WebURL:    "http://192.168.1.8:3000/root",
		},
	}

	pm, resp, err := client.ProjectMembers.AddProjectMember(1, &AddProjectMemberOptions{
		Username: Ptr("venkatesh_thalluri"),
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pm)
	require.Equal(t, "venkatesh_thalluri", *requestOpts.Username)

	pm, resp, err = client.ProjectMembers.AddProjectMember(1.01, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMembers.AddProjectMember(1, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMembers.AddProjectMember(2, nil, nil)
	require.Error(t, err)
	require.Nil(t, pm)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMembersService_EditProjectMember(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/members/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `
			{
				"id": 1,
				"username": "venkatesh_thalluri",
				"name": "Venkatesh Thalluri",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
				"web_url": "http://192.168.1.8:3000/root",
				"created_at": "2012-10-22T14:13:35Z",
				"created_by": {
					"id": 2,
					"username": "john_doe",
					"name": "John Doe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
					"web_url": "http://192.168.1.8:3000/root"
				},
				"access_level": 30,
				"email": "venkatesh.thalluri@example.com",
				"expires_at": null,
				"group_saml_identity": null
			}
		`)
	})

	want := &ProjectMember{
		ID:          1,
		Username:    "venkatesh_thalluri",
		Email:       "venkatesh.thalluri@example.com",
		Name:        "Venkatesh Thalluri",
		State:       "active",
		ExpiresAt:   nil,
		AccessLevel: 30,
		WebURL:      "http://192.168.1.8:3000/root",
		AvatarURL:   "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
		CreatedAt:   mustParseTime("2012-10-22T14:13:35Z"),
		CreatedBy: &MemberCreatedBy{
			ID:        2,
			Username:  "john_doe",
			Name:      "John Doe",
			State:     "active",
			AvatarURL: "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
			WebURL:    "http://192.168.1.8:3000/root",
		},
	}

	pm, resp, err := client.ProjectMembers.EditProjectMember(1, 1, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pm)

	pm, resp, err = client.ProjectMembers.EditProjectMember(1.01, 1, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMembers.EditProjectMember(1, 1, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)
	require.Nil(t, pm)

	pm, resp, err = client.ProjectMembers.EditProjectMember(2, 1, nil, nil)
	require.Error(t, err)
	require.Nil(t, pm)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMembersService_DeleteProjectMember(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/members/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprintf(w, `
			{
				"id": 1,
				"username": "venkatesh_thalluri",
				"name": "Venkatesh Thalluri",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
				"web_url": "http://192.168.1.8:3000/root",
				"access_level": 30,
				"email": "venkatesh.thalluri@example.com",
				"expires_at": null,
				"group_saml_identity": null
			}
		`)
	})

	resp, err := client.ProjectMembers.DeleteProjectMember(1, 1, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	resp, err = client.ProjectMembers.DeleteProjectMember(1.01, 1, nil, nil)
	require.ErrorIs(t, err, ErrInvalidIDType)
	require.Nil(t, resp)

	resp, err = client.ProjectMembers.DeleteProjectMember(1, 1, nil, errorOption)
	require.ErrorIs(t, err, errRequestOptionFunc)
	require.Nil(t, resp)

	resp, err = client.ProjectMembers.DeleteProjectMember(2, 1, nil, nil)
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProjectMembersService_CustomRole(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	path := fmt.Sprintf("/%sprojects/1/members/2", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		fmt.Fprint(w, `
		{
			"id":1,
			"username":"test",
			"name":"testName",
			"access_level":30,
			"member_role":{
				"id":1,
				"group_id":2,
				"name":"TestingCustomRole",
				"description":"",
				"base_access_level":30,
				"admin_cicd_variables":null,
				"admin_group_member":null,
				"admin_merge_request":null,
				"admin_push_rules":null,
				"admin_terraform_state":null,
				"admin_vulnerability":null,
				"archive_project":null,
				"manage_group_access_tokens":null,
				"manage_project_access_tokens":null,
				"read_code":true,
				"read_dependency":null,
				"read_vulnerability":null,
				"remove_group":null,
				"remove_project":null
			}
		}
		`)
	})

	want := &ProjectMember{
		ID:          1,
		Username:    "test",
		Name:        "testName",
		AccessLevel: AccessLevelValue(30),
		MemberRole: &MemberRole{
			ID:              1,
			GroupID:         2,
			Name:            "TestingCustomRole",
			Description:     "",
			BaseAccessLevel: AccessLevelValue(30),
			ReadCode:        true,
		},
	}
	member, _, err := client.ProjectMembers.GetProjectMember(1, 2)

	assert.NoError(t, err)
	assert.Equal(t, want, member)
}

func TestProjectMembersService_ListProjectMembersWithSeatInfo(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		// Verify that the show_seat_info parameter is passed correctly
		assert.Equal(t, "true", r.URL.Query().Get("show_seat_info"))
		fmt.Fprintf(w, `
			[
			{
				"id": 1,
				"username": "venkatesh_thalluri",
				"name": "Venkatesh Thalluri",
				"state": "active",
				"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
				"web_url": "http://192.168.1.8:3000/root",
				"created_at": "2012-10-22T14:13:35Z",
				"created_by": {
					"id": 2,
					"username": "john_doe",
					"name": "John Doe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
					"web_url": "http://192.168.1.8:3000/root"
				},
				"access_level": 30,
				"is_using_seat": true,
				"group_saml_identity": null
			}
			]
		`)
	})

	want := []*ProjectMember{{
		ID:          1,
		Username:    "venkatesh_thalluri",
		Email:       "",
		Name:        "Venkatesh Thalluri",
		State:       "active",
		AccessLevel: 30,
		WebURL:      "http://192.168.1.8:3000/root",
		AvatarURL:   "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
		CreatedAt:   mustParseTime("2012-10-22T14:13:35Z"),
		CreatedBy: &MemberCreatedBy{
			ID:        2,
			Username:  "john_doe",
			Name:      "John Doe",
			State:     "active",
			AvatarURL: "https://www.gravatar.com/avatar/c2525a7f58ae3776070e44c106c48e15?s=80&d=identicon",
			WebURL:    "http://192.168.1.8:3000/root",
		},
		IsUsingSeat: true,
	}}

	pms, resp, err := client.ProjectMembers.ListProjectMembers(1, &ListProjectMembersOptions{
		ShowSeatInfo: Ptr(true),
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, pms)
}
