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

func TestListProtectedTags(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"name":"1.0.0", "create_access_levels": [{"access_level": 40, "access_level_description": "Maintainers"}]},{"name":"*-release", "create_access_levels": [{"access_level": 30, "access_level_description": "Developers + Maintainers"}]}]`)
	})

	expected := []*ProtectedTag{
		{
			Name: "1.0.0",
			CreateAccessLevels: []*TagAccessDescription{
				{
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
		},
		{
			Name: "*-release",
			CreateAccessLevels: []*TagAccessDescription{
				{
					AccessLevel:            30,
					AccessLevelDescription: "Developers + Maintainers",
				},
			},
		},
	}

	opt := &ListProtectedTagsOptions{}
	tags, _, err := client.ProtectedTags.ListProtectedTags(1, opt)
	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tags)
}

func TestListProtectedTagsWithDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"name":"release-1-0", "create_access_levels": [{"id":1,"access_level": 40, "access_level_description": "Maintainers"},{"id":2,"access_level": 40, "access_level_description": "Deploy key", "deploy_key_id": 1}]}]`)
	})

	expected := []*ProtectedTag{
		{
			Name: "release-1-0",
			CreateAccessLevels: []*TagAccessDescription{
				{
					ID:                     1,
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
				{
					ID:                     2,
					AccessLevel:            40,
					DeployKeyID:            1,
					AccessLevelDescription: "Deploy key",
				},
			},
		},
	}

	opt := &ListProtectedTagsOptions{}
	tags, _, err := client.ProtectedTags.ListProtectedTags(1, opt)
	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tags)
}

func TestGetProtectedTag(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	tagName := "my-awesome-tag"

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/protected_tags/%s", tagName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"name":"my-awesome-tag", "create_access_levels": [{"access_level": 30, "access_level_description": "Developers + Maintainers"},{"access_level": 40, "access_level_description": "Sample Group", "group_id": 300}]}`)
	})

	expected := &ProtectedTag{
		Name: tagName,
		CreateAccessLevels: []*TagAccessDescription{
			{
				AccessLevel:            30,
				AccessLevelDescription: "Developers + Maintainers",
			},
			{
				AccessLevel:            40,
				GroupID:                300,
				AccessLevelDescription: "Sample Group",
			},
		},
	}

	tag, _, err := client.ProtectedTags.GetProtectedTag(1, tagName)

	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tag)
}

func TestGetProtectedTagWithDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	tagName := "v1.0.0"

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/protected_tags/%s", tagName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"name":"v1.0.0", "create_access_levels": [{"id": 1, "access_level": 40, "access_level_description": "Maintainers"},{"id": 2, "access_level": 40, "access_level_description": "Deploy key", "deploy_key_id": 5}]}`)
	})

	expected := &ProtectedTag{
		Name: tagName,
		CreateAccessLevels: []*TagAccessDescription{
			{
				ID:                     1,
				AccessLevel:            40,
				AccessLevelDescription: "Maintainers",
			},
			{
				ID:                     2,
				AccessLevel:            40,
				DeployKeyID:            5,
				AccessLevelDescription: "Deploy key",
			},
		},
	}

	tag, _, err := client.ProtectedTags.GetProtectedTag(1, tagName)

	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tag)
}

func TestProtectRepositoryTags(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"name":"my-awesome-tag", "create_access_levels": [{"access_level": 30, "access_level_description": "Developers + Maintainers"},{"access_level": 40, "access_level_description": "Sample Group", "group_id": 300}]}`)
	})

	expected := &ProtectedTag{
		Name: "my-awesome-tag",
		CreateAccessLevels: []*TagAccessDescription{
			{
				AccessLevel:            30,
				AccessLevelDescription: "Developers + Maintainers",
			},
			{
				AccessLevel:            40,
				GroupID:                300,
				AccessLevelDescription: "Sample Group",
			},
		},
	}

	opt := &ProtectRepositoryTagsOptions{
		Name:              Ptr("my-awesome-tag"),
		CreateAccessLevel: Ptr(AccessLevelValue(30)),
		AllowedToCreate: &[]*TagsPermissionOptions{
			{
				GroupID: Ptr(int64(300)),
			},
		},
	}
	tag, _, err := client.ProtectedTags.ProtectRepositoryTags(1, opt)

	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tag)
}

func TestProtectRepositoryTagsWithDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"name":"*-stable", "create_access_levels": [{"id": 1, "access_level": 30, "user_id": 10, "access_level_description": "Administrator"},{"id": 2, "access_level": 40, "deploy_key_id": 20, "access_level_description": "Deploy key"}]}`)
	})

	expected := &ProtectedTag{
		Name: "*-stable",
		CreateAccessLevels: []*TagAccessDescription{
			{
				ID:                     1,
				AccessLevel:            30,
				UserID:                 10,
				AccessLevelDescription: "Administrator",
			},
			{
				ID:                     2,
				AccessLevel:            40,
				DeployKeyID:            20,
				AccessLevelDescription: "Deploy key",
			},
		},
	}

	opt := &ProtectRepositoryTagsOptions{
		Name:              Ptr("*-stable"),
		CreateAccessLevel: Ptr(AccessLevelValue(30)),
		AllowedToCreate: &[]*TagsPermissionOptions{
			{
				UserID: Ptr(int64(10)),
			},
			{
				DeployKeyID: Ptr(int64(20)),
			},
		},
	}
	tag, _, err := client.ProtectedTags.ProtectRepositoryTags(1, opt)

	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tag)
}

func TestUnprotectRepositoryTags(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/protected_tags/my-awesome-tag", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.ProtectedTags.UnprotectRepositoryTags(1, "my-awesome-tag")
	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
