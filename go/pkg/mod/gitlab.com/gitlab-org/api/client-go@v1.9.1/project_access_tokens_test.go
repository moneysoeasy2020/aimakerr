//
// Copyright 2021, Patrick Webster
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
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestListProjectAccessTokens(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/list_project_access_tokens.json")
	})

	projectAccessTokens, _, err := client.ProjectAccessTokens.ListProjectAccessTokens(1, &ListProjectAccessTokensOptions{State: Ptr("active"), ListOptions: ListOptions{Page: 1, PerPage: 20}})
	if err != nil {
		t.Errorf("ProjectAccessTokens.ListProjectAccessTokens returned error: %v", err)
	}

	want := []*ProjectAccessToken{
		{
			PersonalAccessToken: PersonalAccessToken{
				ID:          1876,
				UserID:      2453,
				Active:      true,
				Name:        "token 10",
				Description: "Describe token 10",
				Scopes:      []string{"api", "read_api", "read_repository", "write_repository"},
				CreatedAt:   mustParseTime("2021-03-09T21:11:47.271Z"),
				LastUsedAt:  mustParseTime("2021-03-10T21:11:47.271Z"),
				Revoked:     false,
			},
			AccessLevel: AccessLevelValue(40),
		},
		{
			PersonalAccessToken: PersonalAccessToken{
				ID:          1877,
				UserID:      2456,
				Active:      true,
				Name:        "token 8",
				Description: "Describe token 8",
				Scopes:      []string{"api", "read_api", "read_repository", "write_repository"},
				CreatedAt:   mustParseTime("2021-03-09T21:11:47.340Z"),
				Revoked:     false,
			},
			AccessLevel: AccessLevelValue(30),
		},
	}

	if !reflect.DeepEqual(want, projectAccessTokens) {
		t.Errorf("ProjectAccessTokens.ListProjectAccessTokens returned %+v, want %+v", projectAccessTokens, want)
	}
}

func TestGetProjectAccessToken(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/access_tokens/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/get_project_access_token.json")
	})

	projectAccessToken, _, err := client.ProjectAccessTokens.GetProjectAccessToken(1, 1)
	if err != nil {
		t.Errorf("ProjectAccessTokens.GetProjectAccessToken returned error: %v", err)
	}

	want := &ProjectAccessToken{
		PersonalAccessToken: PersonalAccessToken{
			ID:          1,
			UserID:      2453,
			Active:      true,
			Name:        "token 10",
			Description: "Describe token 10",
			Scopes:      []string{"api", "read_api", "read_repository", "write_repository"},
			CreatedAt:   mustParseTime("2021-03-09T21:11:47.271Z"),
			Revoked:     false,
		},
		AccessLevel: AccessLevelValue(40),
	}

	if !reflect.DeepEqual(want, projectAccessToken) {
		t.Errorf("ProjectAccessTokens.GetProjectAccessToken returned %+v, want %+v", projectAccessToken, want)
	}
}

func TestCreateProjectAccessToken(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		mustWriteHTTPResponse(t, w, "testdata/create_project_access_token.json")
	})

	projectAccessToken, _, err := client.ProjectAccessTokens.CreateProjectAccessToken(1, nil)
	if err != nil {
		t.Errorf("ProjectAccessTokens.CreateProjectAccessToken returned error: %v", err)
	}

	want := &ProjectAccessToken{
		PersonalAccessToken: PersonalAccessToken{
			ID:          1876,
			UserID:      2453,
			Active:      true,
			Name:        "token 10",
			Description: "Describe token 10",
			Scopes:      []string{"api", "read_api", "read_repository", "write_repository"},
			CreatedAt:   mustParseTime("2021-03-09T21:11:47.271Z"),
			Revoked:     false,
			Token:       "2UsevZE1x1ZdFZW4MNzH",
			ExpiresAt:   nil,
		},
		AccessLevel: AccessLevelValue(40),
	}

	if !reflect.DeepEqual(want, projectAccessToken) {
		t.Errorf("ProjectAccessTokens.CreateProjectAccessToken returned %+v, want %+v", projectAccessToken, want)
	}
}

func TestRotateProjectAccessToken(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/projects/1/access_tokens/42/rotate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		mustWriteHTTPResponse(t, w, "testdata/rotate_project_access_token.json")
	})

	expiration := ISOTime(time.Date(2023, time.August, 15, 0, 0, 0, 0, time.UTC))
	opts := &RotateProjectAccessTokenOptions{ExpiresAt: &expiration}
	rotatedToken, _, err := client.ProjectAccessTokens.RotateProjectAccessToken(1, 42, opts)
	if err != nil {
		t.Errorf("ProjectAccessTokens.RotateProjectAccessToken returned error: %v", err)
	}

	want := &ProjectAccessToken{
		PersonalAccessToken: PersonalAccessToken{
			ID:          42,
			UserID:      1337,
			Active:      true,
			Name:        "Rotated Token",
			Description: "Describe rotated token",
			Scopes:      []string{"api"},
			ExpiresAt:   &expiration,
			CreatedAt:   mustParseTime("2023-08-01T15:00:00.000Z"),
			Revoked:     false,
			Token:       "s3cr3t",
		},
		AccessLevel: AccessLevelValue(30),
	}

	if !reflect.DeepEqual(want, rotatedToken) {
		t.Errorf("ProjectAccessTokens.RotateProjectAccessTokens returned %+v, want %+v", rotatedToken, want)
	}
}

func TestRotateProjectAccessTokenSelf(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	mux.HandleFunc("/api/v4/projects/1/access_tokens/self/rotate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		mustWriteHTTPResponse(t, w, "testdata/rotate_project_access_token.json")
	})

	expiration := ISOTime(time.Date(2023, time.August, 15, 0, 0, 0, 0, time.UTC))
	opts := &RotateProjectAccessTokenOptions{ExpiresAt: &expiration}
	rotatedToken, _, err := client.ProjectAccessTokens.RotateProjectAccessTokenSelf(1, opts)
	if err != nil {
		t.Errorf("ProjectAccessTokens.RotateProjectAccessTokenSelf returned error: %v", err)
	}

	want := &ProjectAccessToken{
		PersonalAccessToken: PersonalAccessToken{
			ID:          42,
			UserID:      1337,
			Active:      true,
			Name:        "Rotated Token",
			Description: "Describe rotated token",
			Scopes:      []string{"api"},
			ExpiresAt:   &expiration,
			CreatedAt:   mustParseTime("2023-08-01T15:00:00.000Z"),
			Revoked:     false,
			Token:       "s3cr3t",
		},
		AccessLevel: AccessLevelValue(30),
	}

	if !reflect.DeepEqual(want, rotatedToken) {
		t.Errorf("ProjectAccessTokens.RotateProjectAccessTokenSelf returned %+v, want %+v", rotatedToken, want)
	}
}

func TestRevokeProjectAccessToken(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/access_tokens/1234", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.ProjectAccessTokens.RevokeProjectAccessToken("1", 1234)
	if err != nil {
		t.Errorf("ProjectAccessTokens.RevokeProjectAccessToken returned error: %v", err)
	}
}
