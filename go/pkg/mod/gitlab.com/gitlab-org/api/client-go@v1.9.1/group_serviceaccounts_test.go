//
// Copyright 2023, James Hong
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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGroupServiceAccounts(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/service_accounts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
			{
				"id": 57,
				"username": "service_account_group_345_6018816a18e515214e0c34c2b33523fc",
				"name": "Service account user"
			},
			{
				"id": 58,
				"username": "service_account_group_346_7129927b29f626325f1d45d3c44634fd",
				"name": "Another service account"
			}
		]`)
	})

	serviceAccounts, resp, err := client.Groups.ListServiceAccounts(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*GroupServiceAccount{
		{
			ID:       57,
			UserName: "service_account_group_345_6018816a18e515214e0c34c2b33523fc",
			Name:     "Service account user",
		},
		{
			ID:       58,
			UserName: "service_account_group_346_7129927b29f626325f1d45d3c44634fd",
			Name:     "Another service account",
		},
	}
	assert.Equal(t, want, serviceAccounts)
}

func TestCreateServiceAccount(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/service_accounts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
      {
	      "id": 57,
	      "username": "service_account_group_345_6018816a18e515214e0c34c2b33523fc",
	      "name": "Service account user",
		  "email": "test@test.com"
      }`)
	})

	sa, resp, err := client.Groups.CreateServiceAccount(1, &CreateServiceAccountOptions{
		Name:     Ptr("Service account user"),
		Username: Ptr("service_account_group_345_6018816a18e515214e0c34c2b33523fc"),
		Email:    Ptr("test@test.com"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &GroupServiceAccount{
		ID:       57,
		UserName: "service_account_group_345_6018816a18e515214e0c34c2b33523fc",
		Name:     "Service account user",
		Email:    "test@test.com",
	}
	assert.Equal(t, want, sa)
}

func TestUpdateServiceAccount(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/service_accounts/57", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		fmt.Fprint(w, `
      {
	      "id": 57,
	      "username": "service_account_group_345_6018816a18e515214e0c34c2b33523fc",
	      "name": "Service account user",
		  "email": "service_account@example.com"
      }`)
	})

	sa, resp, err := client.Groups.UpdateServiceAccount(1, 57, &UpdateServiceAccountOptions{
		Name:     Ptr("Service account user"),
		Username: Ptr("service_account_group_345_6018816a18e515214e0c34c2b33523fc"),
		Email:    Ptr("service_account@example.com"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &GroupServiceAccount{
		ID:       57,
		UserName: "service_account_group_345_6018816a18e515214e0c34c2b33523fc",
		Name:     "Service account user",
		Email:    "service_account@example.com",
	}
	assert.Equal(t, want, sa)
}

func TestDeleteServiceAccount(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/service_accounts/57", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Groups.DeleteServiceAccount(1, 57, &DeleteServiceAccountOptions{HardDelete: Ptr(true)})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestListServiceAccountPersonalAccessTokens(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/service_accounts/57/personal_access_tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
      {
      	"id":6,
      	"name":"service_account_token",
      	"revoked":false,
      	"created_at":"2023-06-13T07:47:13.000Z",
		"description":"A token",
      	"scopes":["api"],
      	"user_id":71,
      	"last_used_at":null,
      	"active":true,
      	"expires_at":"2024-06-12",
      	"token":"random_token"
      }]`)
	})

	options := &ListServiceAccountPersonalAccessTokensOptions{
		State: Ptr("active"),
	}
	pats, resp, err := client.Groups.ListServiceAccountPersonalAccessTokens(1, 57, options)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	datePointer := time.Date(2023, time.June, 13, 7, 47, 13, 0, time.UTC)
	expiresAt := ISOTime(time.Date(2024, time.June, 12, 0, 0, 0, 0, time.UTC))

	want := []*PersonalAccessToken{{
		ID:          6,
		Name:        "service_account_token",
		Revoked:     false,
		CreatedAt:   &datePointer,
		Description: "A token",
		Scopes:      []string{"api"},
		UserID:      71,
		LastUsedAt:  nil,
		Active:      true,
		ExpiresAt:   &expiresAt,
		Token:       "random_token",
	}}
	assert.Equal(t, want, pats)
}

func TestCreateServiceAccountPersonalAccessToken(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/service_accounts/57/personal_access_tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
      {
      	"id":6,
      	"name":"service_account_token",
      	"revoked":false,
      	"created_at":"2023-06-13T07:47:13.000Z",
		"description":"A token",
      	"scopes":["api"],
      	"user_id":71,
      	"last_used_at":null,
      	"active":true,
      	"expires_at":"2024-06-12",
      	"token":"random_token"
      }`)
	})

	expireTime, err := ParseISOTime("2024-06-12")
	require.NoError(t, err)

	options := &CreateServiceAccountPersonalAccessTokenOptions{
		Scopes:    Ptr([]string{"api"}),
		Name:      Ptr("service_account_token"),
		ExpiresAt: Ptr(expireTime),
	}
	pat, resp, err := client.Groups.CreateServiceAccountPersonalAccessToken(1, 57, options)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	datePointer := time.Date(2023, time.June, 13, 7, 47, 13, 0, time.UTC)
	expiresAt := ISOTime(time.Date(2024, time.June, 12, 0, 0, 0, 0, time.UTC))

	want := &PersonalAccessToken{
		ID:          6,
		Name:        "service_account_token",
		Revoked:     false,
		CreatedAt:   &datePointer,
		Description: "A token",
		Scopes:      []string{"api"},
		UserID:      71,
		LastUsedAt:  nil,
		Active:      true,
		ExpiresAt:   &expiresAt,
		Token:       "random_token",
	}
	assert.Equal(t, want, pat)
}

func TestRevokeServiceAccountPersonalAccessToken(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/service_accounts/57/personal_access_tokens/6", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})
	resp, err := client.Groups.RevokeServiceAccountPersonalAccessToken(1, 57, 6)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestRotateServiceAccountPersonalAccessToken(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/service_accounts/57/personal_access_tokens/6/rotate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
      {
      	"id":7,
      	"name":"service_account_token",
      	"revoked":false,
      	"created_at":"2023-06-13T07:54:49.000Z",
		"description":"A token",
      	"scopes":["api"],
      	"user_id":71,
      	"last_used_at":null,
      	"active":true,
      	"expires_at":"2025-06-20",
      	"token":"random_token_2"
      }`)
	})

	datePointer := time.Date(2023, time.June, 13, 7, 54, 49, 0, time.UTC)
	expiresAt := ISOTime(time.Date(2025, time.June, 20, 0, 0, 0, 0, time.UTC))
	opts := &RotateServiceAccountPersonalAccessTokenOptions{ExpiresAt: &expiresAt}
	pat, resp, err := client.Groups.RotateServiceAccountPersonalAccessToken(1, 57, 6, opts)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &PersonalAccessToken{
		ID:          7,
		Name:        "service_account_token",
		Revoked:     false,
		CreatedAt:   &datePointer,
		Description: "A token",
		Scopes:      []string{"api"},
		UserID:      71,
		LastUsedAt:  nil,
		Active:      true,
		ExpiresAt:   &expiresAt,
		Token:       "random_token_2",
	}
	assert.Equal(t, want, pat)
}
