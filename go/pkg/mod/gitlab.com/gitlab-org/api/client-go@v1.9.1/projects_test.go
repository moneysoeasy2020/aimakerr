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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListProjects(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name  string
		input string
		want  []*Project
	}{
		{"only id", `[{"id":1},{"id":2}]`, []*Project{{ID: 1}, {ID: 2}}},
		{
			"with ci_delete_pipelines_in_seconds",
			`[{"id":1, "ci_delete_pipelines_in_seconds": 14},{"id":2}]`,
			[]*Project{{ID: 1, CIDeletePipelinesInSeconds: 14}, {ID: 2}},
		},
		{
			"with ci_id_token_sub_claim_components",
			`[{"id":1, "ci_id_token_sub_claim_components": ["project_path", "ref_type"]},{"id":2}]`,
			[]*Project{{ID: 1, CIIdTokenSubClaimComponents: []string{"project_path", "ref_type"}}, {ID: 2}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mux, client := setup(t)

			mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				fmt.Fprint(w, testCase.input)
			})

			opt := &ListProjectsOptions{
				ListOptions: ListOptions{Page: 2, PerPage: 3},
				Archived:    Ptr(true),
				OrderBy:     Ptr("name"),
				Sort:        Ptr("asc"),
				Search:      Ptr("query"),
				Simple:      Ptr(true),
				Visibility:  Ptr(PublicVisibility),
			}

			projects, resp, err := client.Projects.ListProjects(opt)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, testCase.want, projects)
		})
	}
}

func TestListUserProjects(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/users/1/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectsOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Archived:    Ptr(true),
		OrderBy:     Ptr("name"),
		Sort:        Ptr("asc"),
		Search:      Ptr("query"),
		Simple:      Ptr(true),
		Visibility:  Ptr(PublicVisibility),
	}

	projects, resp, err := client.Projects.ListUserProjects(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestListUserContributedProjects(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/users/1/contributed_projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectsOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Archived:    Ptr(true),
		OrderBy:     Ptr("name"),
		Sort:        Ptr("asc"),
		Search:      Ptr("query"),
		Simple:      Ptr(true),
		Visibility:  Ptr(PublicVisibility),
	}

	projects, resp, err := client.Projects.ListUserContributedProjects(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestListUserStarredProjects(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/users/1/starred_projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectsOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Archived:    Ptr(true),
		OrderBy:     Ptr("name"),
		Sort:        Ptr("asc"),
		Search:      Ptr("query"),
		Simple:      Ptr(true),
		Visibility:  Ptr(PublicVisibility),
	}

	projects, resp, err := client.Projects.ListUserStarredProjects(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestListProjectsUsersByID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/projects/1/users?page=2&per_page=3&search=query")
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectUserOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Search:      Ptr("query"),
	}

	projects, resp, err := client.Projects.ListProjectsUsers(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*ProjectUser{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestListProjectsUsersByName(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/projects/namespace%2Fname/users?page=2&per_page=3&search=query")
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectUserOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Search:      Ptr("query"),
	}

	projects, resp, err := client.Projects.ListProjectsUsers("namespace/name", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*ProjectUser{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestListProjectsGroupsByID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/projects/1/groups?page=2&per_page=3&search=query")
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectGroupOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Search:      Ptr("query"),
	}

	groups, resp, err := client.Projects.ListProjectsGroups(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*ProjectGroup{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, groups)
}

func TestListProjectsGroupsByName(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/projects/namespace%2Fname/groups?page=2&per_page=3&search=query")
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectGroupOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Search:      Ptr("query"),
	}

	groups, resp, err := client.Projects.ListProjectsGroups("namespace/name", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*ProjectGroup{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, groups)
}

func TestListOwnedProjects(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectsOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Archived:    Ptr(true),
		OrderBy:     Ptr("name"),
		Sort:        Ptr("asc"),
		Search:      Ptr("query"),
		Simple:      Ptr(true),
		Owned:       Ptr(true),
		Visibility:  Ptr(PublicVisibility),
	}

	projects, resp, err := client.Projects.ListProjects(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestListProjectsActiveFlag(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "true", query.Get("active"))
		assert.Equal(t, "2", query.Get("page"))
		assert.Equal(t, "3", query.Get("per_page"))
		assert.Equal(t, "true", query.Get("archived"))
		assert.Equal(t, "name", query.Get("order_by"))
		assert.Equal(t, "asc", query.Get("sort"))
		assert.Equal(t, "query", query.Get("search"))
		assert.Equal(t, "true", query.Get("simple"))
		assert.Equal(t, "true", query.Get("owned"))
		assert.Equal(t, "public", query.Get("visibility"))

		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectsOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Archived:    Ptr(true),
		OrderBy:     Ptr("name"),
		Sort:        Ptr("asc"),
		Search:      Ptr("query"),
		Simple:      Ptr(true),
		Owned:       Ptr(true),
		Visibility:  Ptr(PublicVisibility),
		Active:      Ptr(true),
	}

	projects, resp, err := client.Projects.ListProjects(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestEditProject(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	var developerRole AccessControlValue = "developer"
	developerPipelineVariablesRole := CIPipelineVariablesDeveloperRole
	opt := &EditProjectOptions{
		CIRestrictPipelineCancellationRole:     Ptr(developerRole),
		CIPipelineVariablesMinimumOverrideRole: Ptr(developerPipelineVariablesRole),
	}

	// Store whether we've seen all the attributes we set
	attributesFound := false

	mux.HandleFunc("/api/v4/projects/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		// Check that our request properly included ci_restrict_pipeline_cancellation_role
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Unable to read body properly. Error: %v", err)
		}

		// Set the value to check if our value is included
		attributesFound = strings.Contains(string(body), "ci_restrict_pipeline_cancellation_role") &&
			strings.Contains(string(body), "ci_pipeline_variables_minimum_override_role")

		// Print the start of the mock example from https://docs.gitlab.com/api/projects/#edit-a-project
		// including the attribute we edited
		fmt.Fprint(w, `
		{
			"id": 1,
			"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			"description_html": "<p data-sourcepos=\"1:1-1:56\" dir=\"auto\">Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>",
			"default_branch": "main",
			"visibility": "private",
			"ssh_url_to_repo": "git@example.com:diaspora/diaspora-project-site.git",
			"http_url_to_repo": "http://example.com/diaspora/diaspora-project-site.git",
			"web_url": "http://example.com/diaspora/diaspora-project-site",
			"readme_url": "http://example.com/diaspora/diaspora-project-site/blob/main/README.md",
			"ci_restrict_pipeline_cancellation_role": "developer",
			"ci_pipeline_variables_minimum_override_role": "developer",
			"ci_delete_pipelines_in_seconds": 14
		}`)
	})

	project, resp, err := client.Projects.EditProject(1, opt)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, attributesFound)
	assert.Equal(t, developerRole, project.CIRestrictPipelineCancellationRole)
	assert.Equal(t, developerPipelineVariablesRole, project.CIPipelineVariablesMinimumOverrideRole)
	assert.Equal(t, int64(14), project.CIDeletePipelinesInSeconds)
}

func TestListStarredProjects(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectsOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Archived:    Ptr(true),
		OrderBy:     Ptr("name"),
		Sort:        Ptr("asc"),
		Search:      Ptr("query"),
		Simple:      Ptr(true),
		Starred:     Ptr(true),
		Visibility:  Ptr(PublicVisibility),
	}

	projects, resp, err := client.Projects.ListProjects(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestGetProjectByID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id": 1,
			"container_registry_enabled": true,
			"container_expiration_policy": {
			  "cadence": "7d",
			  "enabled": false,
			  "keep_n": null,
			  "older_than": null,
			  "name_regex_delete": null,
			  "name_regex_keep": null,
			  "next_run_at": "2020-01-07T21:42:58.658Z"
			},
			"ci_forward_deployment_enabled": true,
			"ci_forward_deployment_rollback_allowed": true,
			"ci_push_repository_for_job_token_allowed": true,
			"ci_restrict_pipeline_cancellation_role": "developer",
			"ci_pipeline_variables_minimum_override_role": "no_one_allowed",
			"packages_enabled": false,
			"build_coverage_regex": "Total.*([0-9]{1,3})%",
			"ci_delete_pipelines_in_seconds": 14,
			"resource_group_default_process_mode": "oldest_first"
		  }`)
	})

	wantTimestamp := time.Date(2020, time.January, 7, 21, 42, 58, 658000000, time.UTC)
	want := &Project{
		ID:                       1,
		ContainerRegistryEnabled: true,
		ContainerExpirationPolicy: &ContainerExpirationPolicy{
			Cadence:   "7d",
			NextRunAt: &wantTimestamp,
		},
		PackagesEnabled:                        false,
		BuildCoverageRegex:                     `Total.*([0-9]{1,3})%`,
		CIForwardDeploymentEnabled:             true,
		CIForwardDeploymentRollbackAllowed:     true,
		CIPushRepositoryForJobTokenAllowed:     true,
		CIRestrictPipelineCancellationRole:     "developer",
		CIPipelineVariablesMinimumOverrideRole: "no_one_allowed",
		CIDeletePipelinesInSeconds:             14,
		ResourceGroupDefaultProcessMode:        OldestFirst,
	}

	project, resp, err := client.Projects.GetProject(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, project)
}

func TestGetProjectByName(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/projects/namespace%2Fname")
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &Project{ID: 1}

	project, resp, err := client.Projects.GetProject("namespace/name", nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, project)
}

func TestGetProjectWithOptions(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id":1,
			"statistics": {
				"commit_count": 37,
				"storage_size": 1038090,
				"repository_size": 1038090,
				"wiki_size": 10,
				"lfs_objects_size": 0,
				"job_artifacts_size": 0,
				"pipeline_artifacts_size": 0,
				"packages_size": 238906167,
				"snippets_size": 146800,
				"uploads_size": 6523619,
				"container_registry_size": 284453
			}}`)
	})
	want := &Project{ID: 1, Statistics: &Statistics{
		CommitCount:           37,
		StorageSize:           1038090,
		RepositorySize:        1038090,
		WikiSize:              10,
		LFSObjectsSize:        0,
		JobArtifactsSize:      0,
		PipelineArtifactsSize: 0,
		PackagesSize:          238906167,
		SnippetsSize:          146800,
		UploadsSize:           6523619,
		ContainerRegistrySize: 284453,
	}}

	project, resp, err := client.Projects.GetProject(1, &GetProjectOptions{Statistics: Ptr(true)})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, project)
}

func TestCreateProject(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"id":1}`)
	})

	opt := &CreateProjectOptions{
		Name:        Ptr("n"),
		MergeMethod: Ptr(RebaseMerge),
	}

	project, resp, err := client.Projects.CreateProject(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &Project{ID: 1}
	assert.Equal(t, want, project)
}

func TestUploadAvatar(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("Projects.UploadAvatar request content-type %+v want multipart/form-data;", r.Header.Get("Content-Type"))
		}
		if r.ContentLength == -1 {
			t.Fatalf("Projects.UploadAvatar request content-length is -1")
		}
		fmt.Fprint(w, `{}`)
	})

	avatar := new(bytes.Buffer)
	_, resp, err := client.Projects.UploadAvatar(1, avatar, "avatar.png")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestUploadAvatar_Retry(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	isFirstRequest := true
	mux.HandleFunc("/api/v4/projects/1", func(w http.ResponseWriter, r *http.Request) {
		if isFirstRequest {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			isFirstRequest = false
			return
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("Projects.UploadAvatar request content-type %+v want multipart/form-data;", r.Header.Get("Content-Type"))
		}
		if r.ContentLength == -1 {
			t.Fatalf("Projects.UploadAvatar request content-length is -1")
		}
		fmt.Fprint(w, `{}`)
	})

	avatar := new(bytes.Buffer)
	_, resp, err := client.Projects.UploadAvatar(1, avatar, "avatar.png")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDownloadAvatar(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	ico, _ := base64.StdEncoding.DecodeString("AAABAAEAEBAAAAEAGABoAwAAFgAAACgAAAAQAAAAIAAAAAEAGAAAAAAAAAAAACABAAAgAQAAAAAAAAAAAAD9/f39/f39/f39/f39/f39/f3y9/x+u+9qsO3l8Pr9/f39/f39/f39/f39/f39/f39/f39/f39/f39/f39/f3c7Plfq+xFnepFnepSo+vI4ff9/f39/f39/f39/f39/f39/f39/f39/f39/f26z/VLkupFnepFnepFnepFnepFlOmevPL7/P39/f39/f39/f39/f39/f34+vyPsvBAe+hAe+hCh+lFm+pFnepDjOlAe+hAe+h2oO3v8/v9/f39/f39/f3o7/pqmOxAe+hAe+hAe+hAe+hBf+dBgedAe+hAe+hAe+hAe+hYi+rX4/j9/f3u8/tXi+pAe+hAe+hAe+hAe+g/deU7X9w6Xds+ceRAe+hAe+hAe+hAe+hIgenZ5fmVtvFAe+hAe+hAe+hAe+g+b+M6XNs6W9o6W9o6W9o9a+FAe+hAe+hAe+hAe+hyne1hketAe+hAe+hAeug9aOA6W9o6W9o6W9o6W9o6W9o6W9o8ZN5AeedAe+hAe+hDfehajepAe+g/d+Y7Yt06W9o6W9o6W9o6W9o6W9o6W9o6W9o6W9o7X9w/dOVAe+hAe+iAoew8Z986XNo6W9o6W9o6W9o6W9o6W9o6W9o6W9o6W9o6W9o6W9o6W9o8ZN5chufDzfI6W9o6W9o6W9o6W9pTb95Wct9Wct9Wct9Wct9Wct88Xdo6W9o6W9o6W9qfr+z6+vxMat06W9o6W9pKaN37+/z9/f39/f39/f39/f39/f1sheM6W9o6W9o7XNrm6vj9/f2Qo+k6W9o6W9qFmef9/f39/f39/f39/f39/f39/f2puO46W9o6W9psheL9/f39/f3Y3/Y6W9o6W9rDzfL9/f39/f39/f39/f39/f39/f3m6vk7XNo6W9q0wO/9/f39/f39/f1eeeBDY9v3+Pz9/f39/f39/f39/f39/f39/f39/f1ifOFDYtvz9fv9/f39/f39/f2vvO6InOf9/f39/f39/f39/f39/f39/f39/f39/f2quO2NoOj9/f39/f0AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	mux.HandleFunc("/api/v4/projects/1/avatar",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			w.Header().Add("Content-Length", strconv.Itoa(len(ico)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(ico)
		},
	)

	_, resp, err := client.Projects.DownloadAvatar(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, len(ico), int(resp.ContentLength))
}

func TestListProjectForks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/projects/namespace%2Fname/forks?archived=true&order_by=name&page=2&per_page=3&search=query&simple=true&sort=asc&visibility=public")
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectsOptions{}
	opt.ListOptions = ListOptions{Page: 2, PerPage: 3}
	opt.Archived = Ptr(true)
	opt.OrderBy = Ptr("name")
	opt.Sort = Ptr("asc")
	opt.Search = Ptr("query")
	opt.Simple = Ptr(true)
	opt.Visibility = Ptr(PublicVisibility)

	projects, resp, err := client.Projects.ListProjectForks("namespace/name", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1", func(_ http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	opt := &DeleteProjectOptions{
		FullPath:          Ptr("group/project"),
		PermanentlyRemove: Ptr(true),
	}

	resp, err := client.Projects.DeleteProject(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShareProjectWithGroup(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/share", func(_ http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
	})

	opt := &ShareWithGroupOptions{
		GroupID:     Ptr(int64(1)),
		GroupAccess: Ptr(AccessLevelValue(50)),
	}

	resp, err := client.Projects.ShareProjectWithGroup(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteSharedProjectFromGroup(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/share/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Projects.DeleteSharedProjectFromGroup(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetApprovalConfiguration(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/approvals", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"approvers": [],
			"approver_groups": [],
			"approvals_before_merge": 3,
			"reset_approvals_on_push": false,
			"disable_overriding_approvers_per_merge_request": false,
			"merge_requests_author_approval": true,
			"merge_requests_disable_committers_approval": true,
			"require_password_to_approve": true
		}`)
	})

	approvals, resp, err := client.Projects.GetApprovalConfiguration(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectApprovals{
		Approvers:            []*MergeRequestApproverUser{},
		ApproverGroups:       []*MergeRequestApproverGroup{},
		ApprovalsBeforeMerge: 3,
		ResetApprovalsOnPush: false,
		DisableOverridingApproversPerMergeRequest: false,
		MergeRequestsAuthorApproval:               true,
		MergeRequestsDisableCommittersApproval:    true,
		RequirePasswordToApprove:                  true,
	}

	assert.Equal(t, want, approvals)
}

func TestChangeApprovalConfiguration(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/approvals", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBodyJSON(t, r, map[string]int{
			"approvals_before_merge": 3,
		})
		fmt.Fprint(w, `{
			"approvers": [],
			"approver_groups": [],
			"approvals_before_merge": 3,
			"reset_approvals_on_push": false,
			"disable_overriding_approvers_per_merge_request": false,
			"merge_requests_author_approval": true,
			"merge_requests_disable_committers_approval": true,
			"require_password_to_approve": true
		}`)
	})

	opt := &ChangeApprovalConfigurationOptions{
		ApprovalsBeforeMerge: Ptr(int64(3)),
	}

	approvals, resp, err := client.Projects.ChangeApprovalConfiguration(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectApprovals{
		Approvers:            []*MergeRequestApproverUser{},
		ApproverGroups:       []*MergeRequestApproverGroup{},
		ApprovalsBeforeMerge: 3,
		ResetApprovalsOnPush: false,
		DisableOverridingApproversPerMergeRequest: false,
		MergeRequestsAuthorApproval:               true,
		MergeRequestsDisableCommittersApproval:    true,
		RequirePasswordToApprove:                  true,
	}

	assert.Equal(t, want, approvals)
}

func TestForkProject(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	namespaceID := int64(42)
	name := "myreponame"
	path := "myrepopath"

	mux.HandleFunc("/api/v4/projects/1/fork", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBodyJSON(t, r, struct {
			Branches    string `json:"branches"`
			Name        string `json:"name"`
			NamespaceID int64  `json:"namespace_id"`
			Path        string `json:"path"`
		}{"main", name, namespaceID, path})
		fmt.Fprint(w, `{"id":2}`)
	})

	project, resp, err := client.Projects.ForkProject(1, &ForkProjectOptions{
		Branches:    Ptr("main"),
		NamespaceID: Ptr(namespaceID),
		Name:        Ptr(name),
		Path:        Ptr(path),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &Project{ID: 2}
	assert.Equal(t, want, project)
}

func TestGetProjectApprovalRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/approval_rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
			{
				"id": 1,
				"name": "security",
				"rule_type": "regular",
				"eligible_approvers": [
					{
						"id": 5,
						"name": "John Doe",
						"username": "jdoe",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/jdoe"
					},
					{
						"id": 50,
						"name": "Group Member 1",
						"username": "group_member_1",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/group_member_1"
					}
				],
				"approvals_required": 3,
				"users": [
					{
						"id": 5,
						"name": "John Doe",
						"username": "jdoe",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/jdoe"
					}
				],
				"groups": [
					{
						"id": 5,
						"name": "group1",
						"path": "group1",
						"description": "",
						"visibility": "public",
						"lfs_enabled": false,
						"avatar_url": null,
						"web_url": "http://localhost/groups/group1",
						"request_access_enabled": false,
						"full_name": "group1",
						"full_path": "group1",
						"parent_id": null,
						"ldap_cn": null,
						"ldap_access": null
					}
				],
				"protected_branches": [
					  {
						"id": 1,
						"name": "master",
						"push_access_levels": [
						  {
							"access_level": 30,
							"access_level_description": "Developers + Maintainers"
						  }
						],
						"merge_access_levels": [
						  {
							"access_level": 30,
							"access_level_description": "Developers + Maintainers"
						  }
						],
						"unprotect_access_levels": [
						  {
							"access_level": 40,
							"access_level_description": "Maintainers"
						  }
						],
						"code_owner_approval_required": false
					  }
                ],
				"contains_hidden_groups": false
			}
		]`)
	})

	approvals, resp, err := client.Projects.GetProjectApprovalRules(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*ProjectApprovalRule{
		{
			ID:       1,
			Name:     "security",
			RuleType: "regular",
			EligibleApprovers: []*BasicUser{
				{
					ID:        5,
					Name:      "John Doe",
					Username:  "jdoe",
					State:     "active",
					AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					WebURL:    "http://localhost/jdoe",
				},
				{
					ID:        50,
					Name:      "Group Member 1",
					Username:  "group_member_1",
					State:     "active",
					AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					WebURL:    "http://localhost/group_member_1",
				},
			},
			ApprovalsRequired: 3,
			Users: []*BasicUser{
				{
					ID:        5,
					Name:      "John Doe",
					Username:  "jdoe",
					State:     "active",
					AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					WebURL:    "http://localhost/jdoe",
				},
			},
			Groups: []*Group{
				{
					ID:                   5,
					Name:                 "group1",
					Path:                 "group1",
					Description:          "",
					Visibility:           PublicVisibility,
					LFSEnabled:           false,
					AvatarURL:            "",
					WebURL:               "http://localhost/groups/group1",
					RequestAccessEnabled: false,
					FullName:             "group1",
					FullPath:             "group1",
				},
			},
			ProtectedBranches: []*ProtectedBranch{
				{
					ID:   1,
					Name: "master",
					PushAccessLevels: []*BranchAccessDescription{
						{
							AccessLevel:            30,
							AccessLevelDescription: "Developers + Maintainers",
						},
					},
					MergeAccessLevels: []*BranchAccessDescription{
						{
							AccessLevel:            30,
							AccessLevelDescription: "Developers + Maintainers",
						},
					},
					UnprotectAccessLevels: []*BranchAccessDescription{
						{
							AccessLevel:            40,
							AccessLevelDescription: "Maintainers",
						},
					},
					AllowForcePush:            false,
					CodeOwnerApprovalRequired: false,
				},
			},
		},
	}

	assert.Equal(t, want, approvals)
}

func TestGetProjectApprovalRule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/approval_rules/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id": 1,
			"name": "security",
			"rule_type": "regular",
			"eligible_approvers": [
				{
					"id": 5,
					"name": "John Doe",
					"username": "jdoe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/jdoe"
				},
				{
					"id": 50,
					"name": "Group Member 1",
					"username": "group_member_1",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/group_member_1"
				}
			],
			"approvals_required": 3,
			"users": [
				{
					"id": 5,
					"name": "John Doe",
					"username": "jdoe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/jdoe"
				}
			],
			"groups": [
				{
					"id": 5,
					"name": "group1",
					"path": "group1",
					"description": "",
					"visibility": "public",
					"lfs_enabled": false,
					"avatar_url": null,
					"web_url": "http://localhost/groups/group1",
					"request_access_enabled": false,
					"full_name": "group1",
					"full_path": "group1",
					"parent_id": null,
					"ldap_cn": null,
					"ldap_access": null
				}
			],
			"protected_branches": [
					{
					"id": 1,
					"name": "master",
					"push_access_levels": [
						{
						"access_level": 30,
						"access_level_description": "Developers + Maintainers"
						}
					],
					"merge_access_levels": [
						{
						"access_level": 30,
						"access_level_description": "Developers + Maintainers"
						}
					],
					"unprotect_access_levels": [
						{
						"access_level": 40,
						"access_level_description": "Maintainers"
						}
					],
					"code_owner_approval_required": false
					}
			],
			"contains_hidden_groups": false
		}`)
	})

	approvals, resp, err := client.Projects.GetProjectApprovalRule(1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectApprovalRule{
		ID:       1,
		Name:     "security",
		RuleType: "regular",
		EligibleApprovers: []*BasicUser{
			{
				ID:        5,
				Name:      "John Doe",
				Username:  "jdoe",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/jdoe",
			},
			{
				ID:        50,
				Name:      "Group Member 1",
				Username:  "group_member_1",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/group_member_1",
			},
		},
		ApprovalsRequired: 3,
		Users: []*BasicUser{
			{
				ID:        5,
				Name:      "John Doe",
				Username:  "jdoe",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/jdoe",
			},
		},
		Groups: []*Group{
			{
				ID:                   5,
				Name:                 "group1",
				Path:                 "group1",
				Description:          "",
				Visibility:           PublicVisibility,
				LFSEnabled:           false,
				AvatarURL:            "",
				WebURL:               "http://localhost/groups/group1",
				RequestAccessEnabled: false,
				FullName:             "group1",
				FullPath:             "group1",
			},
		},
		ProtectedBranches: []*ProtectedBranch{
			{
				ID:   1,
				Name: "master",
				PushAccessLevels: []*BranchAccessDescription{
					{
						AccessLevel:            30,
						AccessLevelDescription: "Developers + Maintainers",
					},
				},
				MergeAccessLevels: []*BranchAccessDescription{
					{
						AccessLevel:            30,
						AccessLevelDescription: "Developers + Maintainers",
					},
				},
				UnprotectAccessLevels: []*BranchAccessDescription{
					{
						AccessLevel:            40,
						AccessLevelDescription: "Maintainers",
					},
				},
				AllowForcePush:            false,
				CodeOwnerApprovalRequired: false,
			},
		},
	}

	assert.Equal(t, want, approvals)
}

func TestCreateProjectApprovalRule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/approval_rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"id": 1,
			"name": "security",
			"rule_type": "regular",
			"eligible_approvers": [
				{
					"id": 5,
					"name": "John Doe",
					"username": "jdoe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/jdoe"
				},
				{
					"id": 6,
					"name": "Cool user",
					"username": "some-cool-user",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/cool-user"
				},
				{
					"id": 50,
					"name": "Group Member 1",
					"username": "group_member_1",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/group_member_1"
				}
			],
			"approvals_required": 3,
			"users": [
				{
					"id": 5,
					"name": "John Doe",
					"username": "jdoe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/jdoe"
				},
				{
					"id": 6,
					"name": "Cool user",
					"username": "some-cool-user",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/cool-user"
				}
			],
			"groups": [
				{
					"id": 5,
					"name": "group1",
					"path": "group1",
					"description": "",
					"visibility": "public",
					"lfs_enabled": false,
					"avatar_url": null,
					"web_url": "http://localhost/groups/group1",
					"request_access_enabled": false,
					"full_name": "group1",
					"full_path": "group1",
					"parent_id": null,
					"ldap_cn": null,
					"ldap_access": null
				}
			],
			"protected_branches": [
				{
				  "id": 1,
				  "name": "master",
				  "push_access_levels": [
					{
					  "access_level": 30,
					  "access_level_description": "Developers + Maintainers"
					}
				  ],
				  "merge_access_levels": [
					{
					  "access_level": 30,
					  "access_level_description": "Developers + Maintainers"
					}
				  ],
				  "unprotect_access_levels": [
					{
					  "access_level": 40,
					  "access_level_description": "Maintainers"
					}
				  ],
				  "code_owner_approval_required": false
				}
			],
			"contains_hidden_groups": false
		}`)
	})

	opt := &CreateProjectLevelRuleOptions{
		Name:              Ptr("security"),
		ApprovalsRequired: Ptr(int64(3)),
		UserIDs:           &[]int64{5, 50},
		GroupIDs:          &[]int64{5},
		ReportType:        Ptr("code_coverage"),
		Usernames:         &([]string{"some-cool-user"}),
	}

	rule, resp, err := client.Projects.CreateProjectApprovalRule(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectApprovalRule{
		ID:       1,
		Name:     "security",
		RuleType: "regular",
		EligibleApprovers: []*BasicUser{
			{
				ID:        5,
				Name:      "John Doe",
				Username:  "jdoe",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/jdoe",
			},
			{
				ID:        6,
				Name:      "Cool user",
				Username:  "some-cool-user",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/cool-user",
			},
			{
				ID:        50,
				Name:      "Group Member 1",
				Username:  "group_member_1",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/group_member_1",
			},
		},
		ApprovalsRequired: 3,
		Users: []*BasicUser{
			{
				ID:        5,
				Name:      "John Doe",
				Username:  "jdoe",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/jdoe",
			},
			{
				ID:        6,
				Name:      "Cool user",
				Username:  "some-cool-user",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/cool-user",
			},
		},
		Groups: []*Group{
			{
				ID:                   5,
				Name:                 "group1",
				Path:                 "group1",
				Description:          "",
				Visibility:           PublicVisibility,
				LFSEnabled:           false,
				AvatarURL:            "",
				WebURL:               "http://localhost/groups/group1",
				RequestAccessEnabled: false,
				FullName:             "group1",
				FullPath:             "group1",
			},
		},
		ProtectedBranches: []*ProtectedBranch{
			{
				ID:   1,
				Name: "master",
				PushAccessLevels: []*BranchAccessDescription{
					{
						AccessLevel:            30,
						AccessLevelDescription: "Developers + Maintainers",
					},
				},
				MergeAccessLevels: []*BranchAccessDescription{
					{
						AccessLevel:            30,
						AccessLevelDescription: "Developers + Maintainers",
					},
				},
				UnprotectAccessLevels: []*BranchAccessDescription{
					{
						AccessLevel:            40,
						AccessLevelDescription: "Maintainers",
					},
				},
				AllowForcePush:            false,
				CodeOwnerApprovalRequired: false,
			},
		},
	}

	assert.Equal(t, want, rule)
}

func TestGetProjectPullMirrorDetails(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/mirror/pull", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
		  "id": 101486,
		  "last_error": null,
		  "last_successful_update_at": "2020-01-06T17:32:02.823Z",
		  "last_update_at": "2020-01-06T17:32:02.823Z",
		  "last_update_started_at": "2020-01-06T17:31:55.864Z",
		  "update_status": "finished",
		  "url": "https://*****:*****@gitlab.com/gitlab-org/security/gitlab.git",
		  "enabled": true,
		  "mirror_trigger_builds": true,
		  "only_mirror_protected_branches": null,
		  "mirror_overwrites_diverged_branches": false,
		  "mirror_branch_regex": null
		}`)
	})

	pullMirror, resp, err := client.Projects.GetProjectPullMirrorDetails(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	wantLastSuccessfulUpdateAtTimestamp := time.Date(2020, time.January, 6, 17, 32, 2, 823000000, time.UTC)
	wantLastUpdateAtTimestamp := time.Date(2020, time.January, 6, 17, 32, 2, 823000000, time.UTC)
	wantLastUpdateStartedAtTimestamp := time.Date(2020, time.January, 6, 17, 31, 55, 864000000, time.UTC)
	want := &ProjectPullMirrorDetails{
		ID:                               101486,
		LastError:                        "",
		LastSuccessfulUpdateAt:           &wantLastSuccessfulUpdateAtTimestamp,
		LastUpdateAt:                     &wantLastUpdateAtTimestamp,
		LastUpdateStartedAt:              &wantLastUpdateStartedAtTimestamp,
		UpdateStatus:                     "finished",
		URL:                              "https://*****:*****@gitlab.com/gitlab-org/security/gitlab.git",
		Enabled:                          true,
		MirrorTriggerBuilds:              true,
		OnlyMirrorProtectedBranches:      false,
		MirrorOverwritesDivergedBranches: false,
		MirrorBranchRegex:                "",
	}

	assert.Equal(t, want, pullMirror)
}

func TestConfigureProjectPullMirror(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/mirror/pull", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
		  "id": 101486,
		  "last_error": null,
		  "last_successful_update_at": "2020-01-06T17:32:02.823Z",
		  "last_update_at": "2020-01-06T17:32:02.823Z",
		  "last_update_started_at": "2020-01-06T17:31:55.864Z",
		  "update_status": "finished",
		  "url": "https://gitlab.example.com/group/project.git",
		  "enabled": true,
		  "mirror_trigger_builds": false,
		  "only_mirror_protected_branches": false,
		  "mirror_overwrites_diverged_branches": true,
		  "mirror_branch_regex": "releases/*"
		}`)
	})

	options := &ConfigureProjectPullMirrorOptions{
		Enabled:                          Ptr(true),
		URL:                              Ptr("https://gitlab.com/gitlab-org/security/gitlab.git"),
		AuthUser:                         Ptr("username"),
		AuthPassword:                     Ptr("secret"),
		MirrorTriggerBuilds:              Ptr(false),
		OnlyMirrorProtectedBranches:      Ptr(false),
		MirrorOverwritesDivergedBranches: Ptr(true),
		MirrorBranchRegex:                Ptr("releases/*"),
	}

	pullMirror, resp, err := client.Projects.ConfigureProjectPullMirror(1, options)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	wantLastSuccessfulUpdateAtTimestamp := time.Date(2020, time.January, 6, 17, 32, 2, 823000000, time.UTC)
	wantLastUpdateAtTimestamp := time.Date(2020, time.January, 6, 17, 32, 2, 823000000, time.UTC)
	wantLastUpdateStartedAtTimestamp := time.Date(2020, time.January, 6, 17, 31, 55, 864000000, time.UTC)
	want := &ProjectPullMirrorDetails{
		ID:                               101486,
		LastError:                        "",
		LastSuccessfulUpdateAt:           &wantLastSuccessfulUpdateAtTimestamp,
		LastUpdateAt:                     &wantLastUpdateAtTimestamp,
		LastUpdateStartedAt:              &wantLastUpdateStartedAtTimestamp,
		UpdateStatus:                     "finished",
		URL:                              "https://gitlab.example.com/group/project.git",
		Enabled:                          true,
		MirrorTriggerBuilds:              false,
		OnlyMirrorProtectedBranches:      false,
		MirrorOverwritesDivergedBranches: true,
		MirrorBranchRegex:                "releases/*",
	}

	assert.Equal(t, want, pullMirror)
}

func TestStartMirroringProject(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/mirror/pull", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
	})

	resp, err := client.Projects.StartMirroringProject(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCreateProjectApprovalRuleEligibleApprovers(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/approval_rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"id": 1,
			"name": "Any name",
			"rule_type": "any_approver",
			"eligible_approvers": [],
			"approvals_required": 1,
			"users": [],
			"groups": [],
			"contains_hidden_groups": false,
			"protected_branches": []
		}`)
	})

	opt := &CreateProjectLevelRuleOptions{
		Name:              Ptr("Any name"),
		ApprovalsRequired: Ptr(int64(1)),
	}

	rule, resp, err := client.Projects.CreateProjectApprovalRule(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectApprovalRule{
		ID:                1,
		Name:              "Any name",
		RuleType:          "any_approver",
		EligibleApprovers: []*BasicUser{},
		ApprovalsRequired: 1,
		Users:             []*BasicUser{},
		Groups:            []*Group{},
		ProtectedBranches: []*ProtectedBranch{},
	}

	assert.Equal(t, want, rule)
}

func TestProjectModelsOptionalMergeAttribute(t *testing.T) {
	t.Parallel()
	// Create a `CreateProjectOptions` struct, ensure that merge attribute doesn't serialize
	jsonString, err := json.Marshal(&CreateProjectOptions{
		Name: Ptr("testProject"),
	})
	assert.NoError(t, err)
	assert.NotContains(t, string(jsonString), "only_allow_merge_if_all_status_checks_passed")

	// Test the same thing but for `EditProjectOptions` struct
	jsonString, err = json.Marshal(&EditProjectOptions{
		Name: Ptr("testProject"),
	})
	assert.NoError(t, err)
	assert.NotContains(t, string(jsonString), "only_allow_merge_if_all_status_checks_passed")
}

func TestListProjectHooks(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/hooks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
	{
		"id": 1,
		"url": "http://example.com/hook",
		"name": "This is the name of an example hook",
		"description": "This is the description of an example hook",
		"confidential_note_events": true,
		"project_id": 1,
		"push_events": true,
		"push_events_branch_filter": "main",
		"issues_events": true,
		"confidential_issues_events": true,
		"merge_requests_events": true,
		"tag_push_events": true,
		"note_events": true,
		"job_events": true,
		"pipeline_events": true,
		"wiki_page_events": true,
		"deployment_events": true,
		"releases_events": true,
		"enable_ssl_verification": true,
		"alert_status": "executable",
		"created_at": "2024-10-13T13:37:00Z",
		"resource_access_token_events": true,
		"custom_webhook_template": "my custom template",
		"custom_headers": [
			{"key": "Authorization"},
			{"key": "OtherHeader"}
		]
	}
]`)
	})

	hooks, resp, err := client.Projects.ListProjectHooks(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	createdAt := time.Date(2024, time.October, 13, 13, 37, 0, 0, time.UTC)
	want := []*ProjectHook{{
		ID:                        1,
		URL:                       "http://example.com/hook",
		Name:                      "This is the name of an example hook",
		Description:               "This is the description of an example hook",
		ConfidentialNoteEvents:    true,
		ProjectID:                 1,
		PushEvents:                true,
		PushEventsBranchFilter:    "main",
		IssuesEvents:              true,
		ConfidentialIssuesEvents:  true,
		MergeRequestsEvents:       true,
		TagPushEvents:             true,
		NoteEvents:                true,
		JobEvents:                 true,
		PipelineEvents:            true,
		WikiPageEvents:            true,
		DeploymentEvents:          true,
		ReleasesEvents:            true,
		EnableSSLVerification:     true,
		CreatedAt:                 &createdAt,
		AlertStatus:               "executable",
		ResourceAccessTokenEvents: true,
		CustomWebhookTemplate:     "my custom template",
		CustomHeaders: []*HookCustomHeader{
			{
				Key: "Authorization",
			},
			{
				Key: "OtherHeader",
			},
		},
	}}

	assert.Equal(t, want, hooks)
}

// Test that the "CustomWebhookTemplate" serializes properly
func TestAddProjectHook_CustomTemplateStuff(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	customWebhookSet := false
	authValueSet := false

	mux.HandleFunc("/api/v4/projects/1/hooks",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			w.WriteHeader(http.StatusCreated)

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Unable to read body properly. Error: %v", err)
			}
			customWebhookSet = strings.Contains(string(body), "custom_webhook_template")
			authValueSet = strings.Contains(string(body), `"value":"stuff"`)

			fmt.Fprint(w, `{
				"custom_webhook_template": "testValue",
				"custom_headers": [
					{
						"key": "Authorization"
					},
					{
						"key": "Favorite-Pet"
					}
				]
			}`)
		},
	)

	hook, resp, err := client.Projects.AddProjectHook(1, &AddProjectHookOptions{
		CustomWebhookTemplate: Ptr(`{"example":"{{object_kind}}"}`),
		CustomHeaders: &[]*HookCustomHeader{
			{
				Key:   "Authorization",
				Value: "stuff",
			},
			{
				Key:   "Favorite-Pet",
				Value: "Cats",
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.True(t, customWebhookSet)
	assert.True(t, authValueSet)
	assert.Equal(t, "testValue", hook.CustomWebhookTemplate)
	assert.Len(t, hook.CustomHeaders, 2)
}

// Test that the "CustomWebhookTemplate" serializes properly when editing
func TestEditProjectHook_CustomTemplateStuff(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	customWebhookSet := false
	authValueSet := false

	mux.HandleFunc("/api/v4/projects/1/hooks/1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPut)
			w.WriteHeader(http.StatusOK)

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Unable to read body properly. Error: %v", err)
			}
			customWebhookSet = strings.Contains(string(body), "custom_webhook_template")
			authValueSet = strings.Contains(string(body), `"value":"stuff"`)

			fmt.Fprint(w, `{
				"custom_webhook_template": "testValue",
				"custom_headers": [
					{
						"key": "Authorization"
					},
					{
						"key": "Favorite-Pet"
					}
				]}`)
		},
	)

	hook, resp, err := client.Projects.EditProjectHook(1, 1, &EditProjectHookOptions{
		CustomWebhookTemplate: Ptr(`{"example":"{{object_kind}}"}`),
		CustomHeaders: &[]*HookCustomHeader{
			{
				Key:   "Authorization",
				Value: "stuff",
			},
			{
				Key:   "Favorite-Pet",
				Value: "Cats",
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, customWebhookSet)
	assert.True(t, authValueSet)
	assert.Equal(t, "testValue", hook.CustomWebhookTemplate)
	assert.Len(t, hook.CustomHeaders, 2)
}

func TestDeleteProjectHook(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/hooks/1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodDelete)
			w.WriteHeader(http.StatusNoContent)
		},
	)

	resp, err := client.Projects.DeleteProjectHook(1, 1)

	assert.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestSetProjectWebhookURLVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/hooks/2/url_variables/TEST_KEY", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	opt := &SetProjectWebhookURLVariableOptions{
		Value: Ptr("testValue"),
	}
	resp, err := client.Projects.SetProjectWebhookURLVariable(1, 2, "TEST_KEY", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteProjectWebhookURLVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/hooks/2/url_variables/TEST_KEY", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Projects.DeleteProjectWebhookURLVariable(1, 2, "TEST_KEY")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetProjectPushRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/push_rule", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id": 1,
			"commit_message_regex": "Fixes \\d+\\..*",
			"commit_message_negative_regex": "ssh\\:\\/\\/",
			"branch_name_regex": "(feat|fix)\\/*",
			"deny_delete_tag": false,
			"member_check": false,
			"prevent_secrets": false,
			"author_email_regex": "@company.com$",
			"file_name_regex": "(jar|exe)$",
			"max_file_size": 5,
			"commit_committer_check": false,
			"commit_committer_name_check": false,
			"reject_unsigned_commits": false,
			"reject_non_dco_commits": false
		  }`)
	})

	rule, resp, err := client.Projects.GetProjectPushRules(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectPushRules{
		ID:                         1,
		CommitMessageRegex:         "Fixes \\d+\\..*",
		CommitMessageNegativeRegex: "ssh\\:\\/\\/",
		BranchNameRegex:            "(feat|fix)\\/*",
		DenyDeleteTag:              false,
		MemberCheck:                false,
		PreventSecrets:             false,
		AuthorEmailRegex:           "@company.com$",
		FileNameRegex:              "(jar|exe)$",
		MaxFileSize:                5,
		CommitCommitterCheck:       false,
		CommitCommitterNameCheck:   false,
		RejectUnsignedCommits:      false,
		RejectNonDCOCommits:        false,
	}

	assert.Equal(t, want, rule)
}

func TestAddProjectPushRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/push_rule", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"id": 1,
			"commit_message_regex": "Fixes \\d+\\..*",
			"commit_message_negative_regex": "ssh\\:\\/\\/",
			"branch_name_regex": "(feat|fix)\\/*",
			"deny_delete_tag": false,
			"member_check": false,
			"prevent_secrets": false,
			"author_email_regex": "@company.com$",
			"file_name_regex": "(jar|exe)$",
			"max_file_size": 5,
			"commit_committer_check": false,
			"commit_committer_name_check": false,
			"reject_unsigned_commits": false,
			"reject_non_dco_commits": false
		  }`)
	})

	opt := &AddProjectPushRuleOptions{
		CommitMessageRegex:         Ptr("Fixes \\d+\\..*"),
		CommitMessageNegativeRegex: Ptr("ssh\\:\\/\\/"),
		BranchNameRegex:            Ptr("(feat|fix)\\/*"),
		DenyDeleteTag:              Ptr(false),
		MemberCheck:                Ptr(false),
		PreventSecrets:             Ptr(false),
		AuthorEmailRegex:           Ptr("@company.com$"),
		FileNameRegex:              Ptr("(jar|exe)$"),
		MaxFileSize:                Ptr(int64(5)),
		CommitCommitterCheck:       Ptr(false),
		CommitCommitterNameCheck:   Ptr(false),
		RejectUnsignedCommits:      Ptr(false),
		RejectNonDCOCommits:        Ptr(false),
	}

	rule, resp, err := client.Projects.AddProjectPushRule(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectPushRules{
		ID:                         1,
		CommitMessageRegex:         "Fixes \\d+\\..*",
		CommitMessageNegativeRegex: "ssh\\:\\/\\/",
		BranchNameRegex:            "(feat|fix)\\/*",
		DenyDeleteTag:              false,
		MemberCheck:                false,
		PreventSecrets:             false,
		AuthorEmailRegex:           "@company.com$",
		FileNameRegex:              "(jar|exe)$",
		MaxFileSize:                5,
		CommitCommitterCheck:       false,
		CommitCommitterNameCheck:   false,
		RejectUnsignedCommits:      false,
		RejectNonDCOCommits:        false,
	}

	assert.Equal(t, want, rule)
}

func TestEditProjectPushRules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/push_rule", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
			"id": 1,
			"commit_message_regex": "Fixes \\d+\\..*",
			"commit_message_negative_regex": "ssh\\:\\/\\/",
			"branch_name_regex": "(feat|fix)\\/*",
			"deny_delete_tag": false,
			"member_check": false,
			"prevent_secrets": false,
			"author_email_regex": "@company.com$",
			"file_name_regex": "(jar|exe)$",
			"max_file_size": 5,
			"commit_committer_check": false,
			"commit_committer_name_check": false,
			"reject_unsigned_commits": false,
			"reject_non_dco_commits": false
		  }`)
	})

	opt := &EditProjectPushRuleOptions{
		CommitMessageRegex:         Ptr("Fixes \\d+\\..*"),
		CommitMessageNegativeRegex: Ptr("ssh\\:\\/\\/"),
		BranchNameRegex:            Ptr("(feat|fix)\\/*"),
		DenyDeleteTag:              Ptr(false),
		MemberCheck:                Ptr(false),
		PreventSecrets:             Ptr(false),
		AuthorEmailRegex:           Ptr("@company.com$"),
		FileNameRegex:              Ptr("(jar|exe)$"),
		MaxFileSize:                Ptr(int64(5)),
		CommitCommitterCheck:       Ptr(false),
		CommitCommitterNameCheck:   Ptr(false),
		RejectUnsignedCommits:      Ptr(false),
		RejectNonDCOCommits:        Ptr(false),
	}

	rule, resp, err := client.Projects.EditProjectPushRule(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectPushRules{
		ID:                         1,
		CommitMessageRegex:         "Fixes \\d+\\..*",
		CommitMessageNegativeRegex: "ssh\\:\\/\\/",
		BranchNameRegex:            "(feat|fix)\\/*",
		DenyDeleteTag:              false,
		MemberCheck:                false,
		PreventSecrets:             false,
		AuthorEmailRegex:           "@company.com$",
		FileNameRegex:              "(jar|exe)$",
		MaxFileSize:                5,
		CommitCommitterCheck:       false,
		CommitCommitterNameCheck:   false,
		RejectUnsignedCommits:      false,
		RejectNonDCOCommits:        false,
	}

	assert.Equal(t, want, rule)
}

func TestGetProjectWebhookHeader(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// Removed most of the arguments to keep test slim
	mux.HandleFunc("/api/v4/projects/1/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"id": 1,
			"custom_webhook_template": "{\"event\":\"{{object_kind}}\"}",
			"custom_headers": [
			  {
				"key": "Authorization"
			  },
			  {
				"key": "OtherKey"
			  }
			]
		  }`)
	})

	hook, resp, err := client.Projects.GetProjectHook(1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectHook{
		ID:                    1,
		CustomWebhookTemplate: "{\"event\":\"{{object_kind}}\"}",
		CustomHeaders: []*HookCustomHeader{
			{
				Key: "Authorization",
			},
			{
				Key: "OtherKey",
			},
		},
	}

	assert.Equal(t, want, hook)
}

func TestSetProjectWebhookHeader(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)
	var bodyJSON map[string]any

	// Removed most of the arguments to keep test slim
	mux.HandleFunc("/api/v4/projects/1/hooks/1/custom_headers/Authorization", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.WriteHeader(http.StatusNoContent)

		// validate that the `value` body is sent properly
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Unable to read body properly. Error: %v", err)
		}

		// Unmarshal the body into JSON so we can check it
		_ = json.Unmarshal(body, &bodyJSON)

		fmt.Fprint(w, ``)
	})

	resp, err := client.Projects.SetProjectCustomHeader(1, 1, "Authorization", &SetHookCustomHeaderOptions{Value: Ptr("testValue")})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "testValue", bodyJSON["value"])
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteProjectWebhookHeader(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// Removed most of the arguments to keep test slim
	mux.HandleFunc("/api/v4/projects/1/hooks/1/custom_headers/Authorization", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprint(w, ``)
	})

	resp, err := client.Projects.DeleteProjectCustomHeader(1, 1, "Authorization")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestStartHousekeepingProject(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/housekeeping", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusAccepted)
	})

	resp, err := client.Projects.StartHousekeepingProject(1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestGetRepositoryStorage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/storage", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"project_id":1,"disk_path":"path/to/repo","repository_storage":"default"}`)
	})

	storage, _, err := client.Projects.GetRepositoryStorage(1)

	assert.NoError(t, err)
	assert.Equal(t, &ProjectRepositoryStorage{
		ProjectID:         1,
		DiskPath:          "path/to/repo",
		CreatedAt:         nil,
		RepositoryStorage: "default",
	}, storage)
}

func TestTransferProject(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/transfer", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testBodyJSON(t, r, map[string]string{
			"namespace": "new-namespace",
		})
		fmt.Fprint(w, `{"id":1}`)
	})

	opt := &TransferProjectOptions{Namespace: Ptr("new-namespace")}
	project, _, err := client.Projects.TransferProject(1, opt)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), project.ID)
}

func TestDeleteProjectPushRule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/push_rule", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Projects.DeleteProjectPushRule(1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestRestoreProject(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/restore", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"id": 1,
			"name": "test-project",
			"path": "test-project",
			"name_with_namespace": "namespace/test-project",
			"path_with_namespace": "namespace/test-project",
			"description": "Test project that was marked for deletion",
			"default_branch": "main",
			"visibility": "private",
			"ssh_url_to_repo": "git@gitlab.com:namespace/test-project.git",
			"http_url_to_repo": "https://example.gitlab.com/namespace/test-project.git",
			"web_url": "https://example.gitlab.com/namespace/test-project",
			"marked_for_deletion_on": null
		}`)
	})

	project, resp, err := client.Projects.RestoreProject(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	want := &Project{
		ID:                  1,
		Name:                "test-project",
		Path:                "test-project",
		NameWithNamespace:   "namespace/test-project",
		PathWithNamespace:   "namespace/test-project",
		Description:         "Test project that was marked for deletion",
		DefaultBranch:       "main",
		Visibility:          PrivateVisibility,
		SSHURLToRepo:        "git@gitlab.com:namespace/test-project.git",
		HTTPURLToRepo:       "https://example.gitlab.com/namespace/test-project.git",
		WebURL:              "https://example.gitlab.com/namespace/test-project",
		MarkedForDeletionOn: nil,
	}

	assert.Equal(t, want, project)
}

func TestRestoreProjectByName(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/projects/namespace%2Ftest-project/restore")
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"id": 2,
			"name": "test-project",
			"path": "test-project",
			"name_with_namespace": "namespace/test-project",
			"path_with_namespace": "namespace/test-project",
			"marked_for_deletion_on": null
		}`)
	})

	project, resp, err := client.Projects.RestoreProject("namespace/test-project")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &Project{
		ID:                  2,
		Name:                "test-project",
		Path:                "test-project",
		NameWithNamespace:   "namespace/test-project",
		PathWithNamespace:   "namespace/test-project",
		MarkedForDeletionOn: nil,
	}

	assert.Equal(t, want, project)
}

func TestEditProject_DuoReviewEnabledSetting(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	opt := &EditProjectOptions{
		AutoDuoCodeReviewEnabled: Ptr(true),
	}

	// Store whether we've seen all the attributes we set
	attributeFound := false

	mux.HandleFunc("/api/v4/projects/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		// Check that our request properly included auto_duo_code_review_enabled
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Unable to read body properly. Error: %v", err)
		}

		// Set the value to check if our value is included
		attributeFound = strings.Contains(string(body), "auto_duo_code_review_enabled")

		// Print the start of the mock example from https://docs.gitlab.com/api/projects/#edit-a-project
		// including the attribute we edited
		fmt.Fprint(w, `
		{
			"id": 1,
			"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			"description_html": "<p data-sourcepos=\"1:1-1:56\" dir=\"auto\">Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>",
			"default_branch": "main",
			"visibility": "private",
			"ssh_url_to_repo": "git@example.com:diaspora/diaspora-project-site.git",
			"http_url_to_repo": "http://example.com/diaspora/diaspora-project-site.git",
			"web_url": "http://example.com/diaspora/diaspora-project-site",
			"readme_url": "http://example.com/diaspora/diaspora-project-site/blob/main/README.md",
			"auto_duo_code_review_enabled": true
		}`)
	})

	project, resp, err := client.Projects.EditProject(1, opt)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, attributeFound)
	assert.True(t, project.AutoDuoCodeReviewEnabled)
}

func TestEditProject_ResourceGroupDefaultProcessModeSetting(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	opt := &EditProjectOptions{
		ResourceGroupDefaultProcessMode: Ptr(OldestFirst),
	}

	// Store whether we've seen all the attributes we set
	attributeFound := false

	mux.HandleFunc("/api/v4/projects/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		// Check that our request properly included resource_group_default_process_mode
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Unable to read body properly. Error: %v", err)
		}

		// Set the value to check if our value is included
		attributeFound = strings.Contains(string(body), "resource_group_default_process_mode")

		// Print the start of the mock example from https://docs.gitlab.com/api/projects/#edit-a-project
		// including the attribute we edited
		fmt.Fprint(w, `
		{
			"id": 1,
			"description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			"description_html": "<p data-sourcepos=\"1:1-1:56\" dir=\"auto\">Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>",
			"default_branch": "main",
			"visibility": "private",
			"ssh_url_to_repo": "git@example.com:diaspora/diaspora-project-site.git",
			"http_url_to_repo": "http://example.com/diaspora/diaspora-project-site.git",
			"web_url": "http://example.com/diaspora/diaspora-project-site",
			"readme_url": "http://example.com/diaspora/diaspora-project-site/blob/main/README.md",
			"resource_group_default_process_mode": "oldest_first"
		}`)
	})

	project, resp, err := client.Projects.EditProject(1, opt)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, attributeFound)
	assert.Equal(t, OldestFirst, project.ResourceGroupDefaultProcessMode)
}

func TestListProjectStarrers(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/123/starrers", func(w http.ResponseWriter, r *http.Request) {
		testURL(t, r, "/api/v4/projects/123/starrers?page=2&per_page=3&search=jane_smith")
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
			{
				"starred_since": "2019-01-28T14:47:30.642Z",
				"user": {
					"id": 1,
					"username": "jane_smith",
					"name": "Jane Smith",
					"state": "active",
					"avatar_url": "http://localhost:3000/uploads/user/avatar/1/cd8.jpeg",
					"web_url": "http://localhost:3000/jane_smith"
				}
			}
		]`)
	})

	opts := &ListProjectStarrersOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 3},
		Search:      Ptr("jane_smith"),
	}

	starrers, resp, err := client.Projects.ListProjectStarrers(123, opts)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, starrers, 1)

	want := []*ProjectStarrer{
		{
			StarredSince: *mustParseTime("2019-01-28T14:47:30.642Z"),
			User: ProjectUser{
				ID:        1,
				Username:  "jane_smith",
				Name:      "Jane Smith",
				State:     "active",
				AvatarURL: "http://localhost:3000/uploads/user/avatar/1/cd8.jpeg",
				WebURL:    "http://localhost:3000/jane_smith",
			},
		},
	}

	assert.Equal(t, want, starrers)
}
