//go:build integration

package gitlab_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Integration tests for the Users API.
// These tests require a GitLab instance running on localhost:8095.
// They also require a valid admin token in GITLAB_TOKEN environment variable.

// TestUsersListUsersIntegration tests the ListUsers function.
func Test_UsersListUsers_Integration(t *testing.T) {
	// GIVEN a GitLab instance running on localhost:8095
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the ListUsers function is called
	users, _, err := client.Users.ListUsers(&gitlab.ListUsersOptions{
		Username: &user.Username,
	})
	require.NoError(t, err, "Failed to list users")

	// THEN it should return a list of users
	assert.Len(t, users, 1) // Usernames are unique
	assert.Equal(t, users[0].ID, user.ID)
}

// Test_UsersGetUser_Integration tests the GetUser function with a regular user.
func Test_UsersGetUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the GetUser function is called with the user ID
	retrievedUser, _, err := client.Users.GetUser(user.ID, gitlab.GetUsersOptions{})
	require.NoError(t, err, "Failed to get user")

	// THEN it should return the user details
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Username, retrievedUser.Username)
	assert.Equal(t, user.Name, retrievedUser.Name)
	assert.Equal(t, user.Email, retrievedUser.Email)
}

// Test_UsersBlockUser_Integration tests the BlockUser function.
func Test_UsersBlockUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the BlockUser function is called
	err = client.Users.BlockUser(user.ID)
	assert.NoError(t, err)

	// THEN the user should be blocked successfully
	// Verify user is blocked by checking their state
	retrievedUser, _, err := client.Users.GetUser(user.ID, gitlab.GetUsersOptions{})
	require.NoError(t, err, "Failed to get user after blocking")

	assert.Equal(t, "blocked", retrievedUser.State)
}

// Test_UsersUnblockUser_Integration tests the UnblockUser function.
func Test_UsersUnblockUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a blocked test user
	client := SetupIntegrationClient(t)

	// Create and block a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	err = client.Users.BlockUser(user.ID)
	require.NoError(t, err, "Failed to block test user")

	// WHEN the UnblockUser function is called
	err = client.Users.UnblockUser(user.ID)
	assert.NoError(t, err)

	// THEN the user should be unblocked successfully
	// Verify user is unblocked by checking their state
	retrievedUser, _, err := client.Users.GetUser(user.ID, gitlab.GetUsersOptions{})
	require.NoError(t, err, "Failed to get user after unblocking")
	assert.Equal(t, "active", retrievedUser.State)
}

// Test_UsersBanUser_Integration tests the BanUser function.
func Test_UsersBanUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the BanUser function is called
	err = client.Users.BanUser(user.ID)
	assert.NoError(t, err)

	// THEN the user should be banned successfully
	// Verify user is banned by checking their state
	retrievedUser, _, err := client.Users.GetUser(user.ID, gitlab.GetUsersOptions{})
	require.NoError(t, err, "Failed to get user after banning")
	assert.Equal(t, "banned", retrievedUser.State)
}

// Test_UsersUnbanUser_Integration tests the UnbanUser function.
func Test_UsersUnbanUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a banned test user
	client := SetupIntegrationClient(t)

	// Create and ban a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	err = client.Users.BanUser(user.ID)
	require.NoError(t, err, "Failed to ban test user")

	// WHEN the UnbanUser function is called
	err = client.Users.UnbanUser(user.ID)
	assert.NoError(t, err)

	// THEN the user should be unbanned successfully
	// Verify user is unbanned by checking their state
	retrievedUser, _, err := client.Users.GetUser(user.ID, gitlab.GetUsersOptions{})
	require.NoError(t, err, "Failed to get user after unbanning")
	assert.Equal(t, "active", retrievedUser.State)
}

// Test_UsersDeactivateUser_Integration tests the DeactivateUser function.
func Test_UsersDeactivateUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the DeactivateUser function is called
	err = client.Users.DeactivateUser(user.ID)
	assert.NoError(t, err)

	// THEN the user should be deactivated successfully
	// Verify user is deactivated by checking their state
	retrievedUser, _, err := client.Users.GetUser(user.ID, gitlab.GetUsersOptions{})
	require.NoError(t, err, "Failed to get user after deactivating")
	assert.Equal(t, "deactivated", retrievedUser.State)
}

// Test_UsersActivateUser_Integration tests the ActivateUser function.
func Test_UsersActivateUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a deactivated test user
	client := SetupIntegrationClient(t)

	// Create and deactivate a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	err = client.Users.DeactivateUser(user.ID)
	require.NoError(t, err, "Failed to deactivate test user")

	// WHEN the ActivateUser function is called
	err = client.Users.ActivateUser(user.ID)
	assert.NoError(t, err)

	// THEN the user should be activated successfully
	// Verify user is activated by checking their state
	retrievedUser, _, err := client.Users.GetUser(user.ID, gitlab.GetUsersOptions{})
	require.NoError(t, err, "Failed to get user after activating")
	assert.Equal(t, "active", retrievedUser.State)
}

// Test_UsersCreateUser_Integration tests the CreateUser function.
func Test_UsersCreateUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance
	client := SetupIntegrationClient(t)

	// WHEN the CreateUser function is called with valid user data
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// THEN a new user should be created successfully
	// Verify the user was created with expected properties
	assert.NotZero(t, user.ID)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Name)
	assert.NotEmpty(t, user.Email)
	assert.Equal(t, "active", user.State)
}

// Test_UsersModifyUser_Integration tests the ModifyUser function.
func Test_UsersModifyUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the ModifyUser function is called with updated data
	newName := "Modified Test User"
	newBio := "This is a modified test user"
	modifiedUser, _, err := client.Users.ModifyUser(user.ID, &gitlab.ModifyUserOptions{
		Name: &newName,
		Bio:  &newBio,
	})
	assert.NoError(t, err)

	// THEN the user should be modified successfully
	// Verify the modifications
	assert.Equal(t, newName, modifiedUser.Name)
	assert.Equal(t, newBio, modifiedUser.Bio)
	assert.Equal(t, user.ID, modifiedUser.ID)
	assert.Equal(t, user.Username, modifiedUser.Username)
}

// Test_UsersDeleteUser_Integration tests the DeleteUser function.
func Test_UsersDeleteUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user (without cleanup since we're testing deletion)
	suffix := time.Now().UnixNano()
	username := fmt.Sprintf("testuser%d", suffix)
	email := fmt.Sprintf("testuser%d@example.com", suffix)
	name := fmt.Sprintf("Test User %d", suffix)

	user, _, err := client.Users.CreateUser(&gitlab.CreateUserOptions{
		Username:         &username,
		Email:            &email,
		Name:             &name,
		Password:         gitlab.Ptr("f0hYXux#yy2CFypKq!aV"),
		SkipConfirmation: gitlab.Ptr(true),
	})
	require.NoError(t, err, "Failed to create test user")

	// WHEN the DeleteUser function is called
	resp, err := client.Users.DeleteUser(user.ID)
	assert.NoError(t, err)

	// THEN the user should be deleted successfully
	// http 204 means the user is successfully deleted
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// We cannot perform a "read after write" check to ensure the
	// user is deleted, because the actual delete is async, so if we attempt
	// to immediately read, it will succeed (which isn't what we would normally expect.)
}

// Test_UsersGetUserAssociationsCount_Integration tests the GetUserAssociationsCount function.
func Test_UsersGetUserAssociationsCount_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the GetUserAssociationsCount function is called
	associationsCount, _, err := client.Users.GetUserAssociationsCount(user.ID)
	assert.NoError(t, err)

	// THEN it should return the user's association counts
	// Verify the response structure
	assert.NotNil(t, associationsCount)
	assert.GreaterOrEqual(t, associationsCount.GroupsCount, int64(0))
	assert.GreaterOrEqual(t, associationsCount.ProjectsCount, int64(0))
	assert.GreaterOrEqual(t, associationsCount.IssuesCount, int64(0))
	assert.GreaterOrEqual(t, associationsCount.MergeRequestsCount, int64(0))
}

// Test_UsersGetUserMemberships_Integration tests the GetUserMemberships function.
func Test_UsersGetUserMemberships_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the GetUserMemberships function is called
	memberships, _, err := client.Users.GetUserMemberships(user.ID, &gitlab.GetUserMembershipOptions{})
	assert.NoError(t, err)

	// THEN it should return the user's memberships
	// Verify the response (should be empty for a new user)
	assert.NotNil(t, memberships)
	// New users typically have no memberships initially
}

// Test_UsersGetUserStatus_Integration tests the GetUserStatus function.
func Test_UsersGetUserStatus_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the GetUserStatus function is called
	status, _, err := client.Users.GetUserStatus(user.ID)
	assert.NoError(t, err)

	// THEN it should return the user's status
	// Verify the response structure (new users typically have empty status)
	assert.NotNil(t, status)
}

// Test_UsersSetUserStatus_Integration tests the SetUserStatus function.
func Test_UsersSetUserStatus_Integration(t *testing.T) {
	// GIVEN a GitLab instance
	client := SetupIntegrationClient(t)

	// WHEN the SetUserStatus function is called with status data
	emoji := "coffee"
	message := "Working on integration tests"
	availability := gitlab.Busy

	status, _, err := client.Users.SetUserStatus(&gitlab.UserStatusOptions{
		Emoji:        &emoji,
		Message:      &message,
		Availability: &availability,
	})
	assert.NoError(t, err)

	// THEN the current user's status should be updated
	// Verify the status was set
	assert.Equal(t, emoji, status.Emoji)
	assert.Equal(t, message, status.Message)
	assert.Equal(t, availability, status.Availability)

	// Clean up by clearing the status
	_, _, err = client.Users.SetUserStatus(&gitlab.UserStatusOptions{
		Emoji:   gitlab.Ptr(""),
		Message: gitlab.Ptr(""),
	})
	assert.NoError(t, err)
}

// Test_UsersCreateServiceAccountUser_Integration tests the CreateServiceAccountUser function.
func Test_UsersCreateServiceAccountUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance
	client := SetupIntegrationClient(t)
	SkipIfNotLicensed(t, client)

	// WHEN the CreateServiceAccountUser function is called
	suffix := time.Now().UnixNano()
	name := fmt.Sprintf("TestSA%d", suffix)
	username := fmt.Sprintf("serviceaccount%d", suffix)
	email := fmt.Sprintf("serviceaccount%d@test.com", suffix)

	serviceAccount, _, err := client.Users.CreateServiceAccountUser(&gitlab.CreateServiceAccountUserOptions{
		Name:     &name,
		Username: &username,
		Email:    &email,
	})
	assert.NoError(t, err)

	// Clean up
	t.Cleanup(func() {
		_, err := client.Users.DeleteUser(serviceAccount.ID)
		if err != nil {
			t.Logf("Failed to delete service account user %d: %v", serviceAccount.ID, err)
		}
	})

	// THEN a service account user should be created successfully
	// Verify the service account was created
	assert.NotZero(t, serviceAccount.ID)
	assert.Equal(t, username, serviceAccount.Username)
	assert.Equal(t, name, serviceAccount.Name)
	assert.Equal(t, email, serviceAccount.Email)
}

// Test_UsersListServiceAccounts_Integration tests the ListServiceAccounts function.
func Test_UsersListServiceAccounts_Integration(t *testing.T) {
	// GIVEN a GitLab instance with service accounts
	client := SetupIntegrationClient(t)
	SkipIfNotLicensed(t, client)

	// Create a service account first
	suffix := time.Now().UnixNano()
	name := fmt.Sprintf("Test Service Account %d", suffix)
	username := fmt.Sprintf("serviceaccount%d", suffix)
	email := fmt.Sprintf("serviceaccount%d@test.com", suffix)

	serviceAccount, _, err := client.Users.CreateServiceAccountUser(&gitlab.CreateServiceAccountUserOptions{
		Name:     &name,
		Username: &username,
		Email:    &email,
	})
	require.NoError(t, err, "Failed to create service account user")

	// Clean up
	t.Cleanup(func() {
		_, err := client.Users.DeleteUser(serviceAccount.ID)
		if err != nil {
			t.Logf("Failed to delete service account user %d: %v", serviceAccount.ID, err)
		}
	})

	// WHEN the ListServiceAccounts function is called
	serviceAccounts, _, err := client.Users.ListServiceAccounts(&gitlab.ListServiceAccountsOptions{})
	assert.NoError(t, err)

	// THEN it should return a list of service accounts
	// Verify the response contains our service account
	assert.NotNil(t, serviceAccounts)
	found := false
	for _, sa := range serviceAccounts {
		if sa.ID == serviceAccount.ID {
			found = true
			assert.Equal(t, username, sa.Username)
			assert.Equal(t, name, sa.Name)
			break
		}
	}
	assert.True(t, found, "Created service account should be in the list")
}

// Test_UsersDeleteUserIdentity_Integration tests the DeleteUserIdentity function.
func Test_UsersDeleteUserIdentity_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user that has an identity
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// Note: This test may not work in all GitLab instances as it requires
	// external identity providers to be configured. We'll test the API call
	// but expect it might return an error for users without external identities.

	// WHEN the DeleteUserIdentity function is called
	_, err = client.Users.DeleteUserIdentity(user.ID, "github")

	// THEN the user's identity should be deleted successfully
	// We don't assert NoError here because the user likely doesn't have
	// a GitHub identity. The important thing is that the API call is made
	// without causing a panic or unexpected error format.
	assert.Error(t, err) // Expected to fail for users without external identities
}

// Test_UsersGetSSHKeyForUser_Integration tests the GetSSHKeyForUser function.
func Test_UsersGetSSHKeyForUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user that has SSH keys
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the GetSSHKeyForUser function is called
	// Test GetSSHKeyForUser function with a non-existent key ID
	// This should return an error since the user has no SSH keys
	_, _, err = client.Users.GetSSHKeyForUser(user.ID, 1)

	// THEN it should return the SSH key details or an appropriate error
	assert.Error(t, err) // Expected to fail for non-existent SSH key
}

// Test_UsersDisableTwoFactor_Integration tests the DisableTwoFactor function.
func Test_UsersDisableTwoFactor_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the DisableTwoFactor function is called
	// This will likely return an error since the user doesn't have 2FA enabled
	err = client.Users.DisableTwoFactor(user.ID)

	// THEN it should handle the request appropriately
	// We expect this to fail since the user doesn't have 2FA enabled
	// The important thing is that the API call doesn't panic
	assert.Error(t, err)
}

// Test_UsersCreateUserRunner_Integration tests the CreateUserRunner function.
func Test_UsersCreateUserRunner_Integration(t *testing.T) {
	// GIVEN a GitLab instance
	client := SetupIntegrationClient(t)

	// WHEN the CreateUserRunner function is called
	runnerType := "instance_type"
	description := "Test integration runner"

	runner, _, err := client.Users.CreateUserRunner(&gitlab.CreateUserRunnerOptions{
		RunnerType:  &runnerType,
		Description: &description,
	})
	assert.NoError(t, err)

	// THEN it should create a user runner successfully
	// Verify the runner was created
	assert.NotZero(t, runner.ID)
	assert.NotEmpty(t, runner.Token)
}

// Test_UsersCreatePersonalAccessTokenForCurrentUser_Integration tests the CreatePersonalAccessTokenForCurrentUser function.
func Test_UsersCreatePersonalAccessTokenForCurrentUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance
	client := SetupIntegrationClient(t)

	// WHEN the CreatePersonalAccessTokenForCurrentUser function is called
	tokenName := "integration-test-token"
	scopes := []string{"k8s_proxy"}

	token, _, err := client.Users.CreatePersonalAccessTokenForCurrentUser(&gitlab.CreatePersonalAccessTokenForCurrentUserOptions{
		Name:   &tokenName,
		Scopes: &scopes,
	})
	assert.NoError(t, err)

	// THEN it should create a personal access token for the current user
	assert.NotZero(t, token.ID)
	assert.Equal(t, tokenName, token.Name)
	assert.Equal(t, scopes, token.Scopes)
	assert.NotEmpty(t, token.Token)
}

// Test_UsersUploadAvatar_Integration tests the UploadAvatar function.
func Test_UsersUploadAvatar_Integration(t *testing.T) {
	// GIVEN a GitLab instance
	client := SetupIntegrationClient(t)

	// Create a simple test image (1x1 PNG)
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00,
		0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0xF8, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
		0xAE, 0x42, 0x60, 0x82,
	}

	// WHEN the UploadAvatar function is called with avatar data
	avatar := bytes.NewReader(pngData)
	user, _, err := client.Users.UploadAvatar(avatar, "test-avatar.png")
	assert.NoError(t, err)

	// THEN it should upload the avatar for the current user
	// Verify the avatar was uploaded
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.AvatarURL)
}

// Test_UsersApproveUser_Integration tests the ApproveUser function.
func Test_UsersApproveUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the ApproveUser function is called
	// This will likely return an error since the user is already active/approved
	err = client.Users.ApproveUser(user.ID)

	// THEN it should handle the request appropriately
	// We expect this to fail since the user is already active
	// The important thing is that the API call doesn't panic
	assert.Error(t, err)
}

// Test_UsersRejectUser_Integration tests the RejectUser function.
func Test_UsersRejectUser_Integration(t *testing.T) {
	// GIVEN a GitLab instance with a test user
	client := SetupIntegrationClient(t)

	// Create a test user
	user, err := CreateTestUser(t, client)
	require.NoError(t, err, "Failed to create test user")

	// WHEN the RejectUser function is called
	// This will likely return an error since the user is already active
	err = client.Users.RejectUser(user.ID)

	// THEN it should handle the request appropriately
	// We expect this to fail since the user is already active
	// The important thing is that the API call doesn't panic
	assert.Error(t, err)
}
