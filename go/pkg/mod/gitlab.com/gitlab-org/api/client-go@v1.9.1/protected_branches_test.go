//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListProtectedBranches(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
	{
		"id":1,
		"name":"master",
		"push_access_levels":[{
			"id":1,
			"access_level":40,
			"access_level_description":"Maintainers",
			"deploy_key_id":null,
			"user_id":null,
			"group_id":null
		},{
			"id":2,
			"access_level":30,
			"access_level_description":"User name",
			"deploy_key_id":null,
			"user_id":123,
			"group_id":null
		},{
			"id":3,
			"access_level":40,
			"access_level_description":"deploy key",
			"deploy_key_id":456,
			"user_id":null,
			"group_id":null
		}],
		"merge_access_levels":[{
			"id":1,
			"access_level":40,
			"access_level_description":"Maintainers",
			"user_id":null,
			"group_id":null
		}],
		"code_owner_approval_required":false
	}
]`)
	})
	opt := &ListProtectedBranchesOptions{}
	protectedBranches, resp, err := client.ProtectedBranches.ListProtectedBranches("1", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	want := []*ProtectedBranch{
		{
			ID:   1,
			Name: "master",
			PushAccessLevels: []*BranchAccessDescription{
				{
					ID:                     1,
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
				{
					ID:                     2,
					AccessLevel:            30,
					AccessLevelDescription: "User name",
					UserID:                 123,
				},
				{
					ID:                     3,
					AccessLevel:            40,
					AccessLevelDescription: "deploy key",
					DeployKeyID:            456,
				},
			},
			MergeAccessLevels: []*BranchAccessDescription{
				{
					ID:                     1,
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
			AllowForcePush:            false,
			CodeOwnerApprovalRequired: false,
		},
	}
	assert.Equal(t, want, protectedBranches)
}

func TestListProtectedBranches_WithoutCodeOwnerApproval(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
	{
		"id":1,
		"name":"master",
		"push_access_levels":[{
			"access_level":40,
			"access_level_description":"Maintainers"
		}],
		"merge_access_levels":[{
			"access_level":40,
			"access_level_description":"Maintainers"
		}]
	}
]`)
	})
	opt := &ListProtectedBranchesOptions{}
	protectedBranches, resp, err := client.ProtectedBranches.ListProtectedBranches("1", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	want := []*ProtectedBranch{
		{
			ID:   1,
			Name: "master",
			PushAccessLevels: []*BranchAccessDescription{
				{
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
			MergeAccessLevels: []*BranchAccessDescription{
				{
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
			AllowForcePush:            false,
			CodeOwnerApprovalRequired: false,
		},
	}
	assert.Equal(t, want, protectedBranches)
}

func TestGetProtectedBranch(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_branches/main", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id":1,
			"name":"main",
			"push_access_levels":[{
				"id":1,
				"access_level":40,
				"access_level_description":"Maintainers",
				"deploy_key_id":null,
				"user_id":null,
				"group_id":null
			},{
				"id":2,
				"access_level":30,
				"access_level_description":"User name",
				"deploy_key_id":null,
				"user_id":123,
				"group_id":null
			},{
				"id":3,
				"access_level":40,
				"access_level_description":"deploy key",
				"deploy_key_id":456,
				"user_id":null,
				"group_id":null
			}],
			"merge_access_levels":[{
				"id":1,
				"access_level":40,
				"access_level_description":"Maintainers",
				"user_id":null,
				"group_id":null
			}],
			"code_owner_approval_required":false
		}`)
	})
	protectedBranch, resp, err := client.ProtectedBranches.GetProtectedBranch(1, "main")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	want := &ProtectedBranch{
		ID:   1,
		Name: "main",
		PushAccessLevels: []*BranchAccessDescription{
			{
				ID:                     1,
				AccessLevel:            40,
				AccessLevelDescription: "Maintainers",
			},
			{
				ID:                     2,
				AccessLevel:            30,
				AccessLevelDescription: "User name",
				UserID:                 123,
			},
			{
				ID:                     3,
				AccessLevel:            40,
				AccessLevelDescription: "deploy key",
				DeployKeyID:            456,
			},
		},
		MergeAccessLevels: []*BranchAccessDescription{
			{
				ID:                     1,
				AccessLevel:            40,
				AccessLevelDescription: "Maintainers",
			},
		},
		AllowForcePush:            false,
		CodeOwnerApprovalRequired: false,
	}
	assert.Equal(t, want, protectedBranch)
}

func TestProtectRepositoryBranches(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
	{
		"id":1,
		"name":"master",
		"push_access_levels":[{
			"access_level":40,
			"access_level_description":"Maintainers"
		}],
		"merge_access_levels":[{
			"access_level":40,
			"access_level_description":"Maintainers"
		}],
		"allow_force_push":true,
		"code_owner_approval_required":true
	}`)
	})
	opt := &ProtectRepositoryBranchesOptions{
		Name:                      Ptr("master"),
		PushAccessLevel:           Ptr(MaintainerPermissions),
		MergeAccessLevel:          Ptr(MaintainerPermissions),
		AllowForcePush:            Ptr(true),
		CodeOwnerApprovalRequired: Ptr(true),
	}
	protectedBranches, resp, err := client.ProtectedBranches.ProtectRepositoryBranches("1", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	want := &ProtectedBranch{
		ID:   1,
		Name: "master",
		PushAccessLevels: []*BranchAccessDescription{
			{
				AccessLevel:            40,
				AccessLevelDescription: "Maintainers",
			},
		},
		MergeAccessLevels: []*BranchAccessDescription{
			{
				AccessLevel:            40,
				AccessLevelDescription: "Maintainers",
			},
		},
		AllowForcePush:            true,
		CodeOwnerApprovalRequired: true,
	}
	assert.Equal(t, want, protectedBranches)
}

func TestUnprotectRepositoryBranches(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_branches/main", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})
	resp, err := client.ProtectedBranches.UnprotectRepositoryBranches("1", "main")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestUpdateRepositoryBranches(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_branches/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		testBodyJSON(t, r, map[string]bool{
			"code_owner_approval_required": true,
		})
		fmt.Fprintf(w, `{
			"name": "master",
			"code_owner_approval_required": true
		}`)
	})
	opt := &UpdateProtectedBranchOptions{
		CodeOwnerApprovalRequired: Ptr(true),
	}
	protectedBranch, resp, err := client.ProtectedBranches.UpdateProtectedBranch("1", "master", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProtectedBranch{
		Name:                      "master",
		CodeOwnerApprovalRequired: true,
	}
	assert.Equal(t, want, protectedBranch)
}
