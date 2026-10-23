package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGroupMergeRequestApprovalSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/merge_request_approval_setting", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			{
				"allow_author_approval": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"allow_committer_approval": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"allow_overrides_to_approver_list_per_merge_request": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"retain_approvals_on_push": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"selective_code_owner_removals": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"require_password_to_approve": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"require_reauthentication_to_approve": {
					"value": false,
					"locked": false,
					"inherited_from": null
				}
			}
		`)

		want := &MergeRequestApprovalSettings{
			AllowAuthorApproval: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			AllowCommitterApproval: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			AllowOverridesToApproverListPerMergeRequest: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			RetainApprovalsOnPush: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			SelectiveCodeOwnerRemovals: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			RequirePasswordToApprove: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			RequireReauthenticationToApprove: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
		}

		settings, resp, err := client.MergeRequestApprovalSettings.GetGroupMergeRequestApprovalSettings(1)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, want, settings)
	})
}

func TestUpdateGroupMergeRequestApprovalSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/merge_request_approval_setting", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
			{
				"allow_author_approval": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"allow_committer_approval": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"allow_overrides_to_approver_list_per_merge_request": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"retain_approvals_on_push": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"selective_code_owner_removals": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"require_password_to_approve": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"require_reauthentication_to_approve": {
					"value": true,
					"locked": false,
					"inherited_from": null
				}
			}
		`)

		want := &MergeRequestApprovalSettings{
			AllowAuthorApproval: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			AllowCommitterApproval: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			AllowOverridesToApproverListPerMergeRequest: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			RetainApprovalsOnPush: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			SelectiveCodeOwnerRemovals: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			RequirePasswordToApprove: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			RequireReauthenticationToApprove: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
		}

		settings, resp, err := client.MergeRequestApprovalSettings.UpdateGroupMergeRequestApprovalSettings(1, &UpdateGroupMergeRequestApprovalSettingsOptions{
			AllowAuthorApproval:                         Ptr(false),
			AllowCommitterApproval:                      Ptr(false),
			AllowOverridesToApproverListPerMergeRequest: Ptr(false),
			RetainApprovalsOnPush:                       Ptr(true),
			RequireReauthenticationToApprove:            Ptr(true),
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, want, settings)
	})
}

func TestGetProjectMergeRequestApprovalSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_request_approval_setting", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			{
				"allow_author_approval": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"allow_committer_approval": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"allow_overrides_to_approver_list_per_merge_request": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"retain_approvals_on_push": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"selective_code_owner_removals": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"require_password_to_approve": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"require_reauthentication_to_approve": {
					"value": false,
					"locked": false,
					"inherited_from": null
				}
			}
		`)

		want := &MergeRequestApprovalSettings{
			AllowAuthorApproval: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			AllowCommitterApproval: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			AllowOverridesToApproverListPerMergeRequest: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			RetainApprovalsOnPush: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			SelectiveCodeOwnerRemovals: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			RequirePasswordToApprove: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			RequireReauthenticationToApprove: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
		}

		settings, resp, err := client.MergeRequestApprovalSettings.GetProjectMergeRequestApprovalSettings(1)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, want, settings)
	})
}

func TestUpdateProjectMergeRequestApprovalSettings(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_request_approval_setting", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			{
				"allow_author_approval": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"allow_committer_approval": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"allow_overrides_to_approver_list_per_merge_request": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"retain_approvals_on_push": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"selective_code_owner_removals": {
					"value": true,
					"locked": false,
					"inherited_from": null
				},
				"require_password_to_approve": {
					"value": false,
					"locked": false,
					"inherited_from": null
				},
				"require_reauthentication_to_approve": {
					"value": true,
					"locked": false,
					"inherited_from": null
				}
			}
		`)

		want := &MergeRequestApprovalSettings{
			AllowAuthorApproval: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			AllowCommitterApproval: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			AllowOverridesToApproverListPerMergeRequest: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			RetainApprovalsOnPush: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			SelectiveCodeOwnerRemovals: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
			RequirePasswordToApprove: MergeRequestApprovalSetting{
				Value:         false,
				Locked:        false,
				InheritedFrom: "",
			},
			RequireReauthenticationToApprove: MergeRequestApprovalSetting{
				Value:         true,
				Locked:        false,
				InheritedFrom: "",
			},
		}

		settings, resp, err := client.MergeRequestApprovalSettings.UpdateProjectMergeRequestApprovalSettings(1, &UpdateProjectMergeRequestApprovalSettingsOptions{
			AllowAuthorApproval:                         Ptr(false),
			AllowCommitterApproval:                      Ptr(false),
			AllowOverridesToApproverListPerMergeRequest: Ptr(false),
			RetainApprovalsOnPush:                       Ptr(true),
			SelectiveCodeOwnerRemovals:                  Ptr(true),
			RequireReauthenticationToApprove:            Ptr(true),
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, want, settings)
	})
}
