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
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPipelineJobs(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/1/jobs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	jobs, resp, err := client.Jobs.ListPipelineJobs(1, 1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Job{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, jobs)
}

func TestJobsService_ListProjectJobs(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/jobs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
  {
    "commit": {
      "author_email": "admin@example.com",
      "author_name": "Administrator",
      "id": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
      "short_id": "0ff3ae19",
      "title": "Test the CI integration."
    },
    "allow_failure": false,
    "tag_list": [
      "docker runner",
      "ubuntu18"
    ],
    "id": 7,
    "name": "teaspoon",
    "pipeline": {
      "id": 6,
      "project_id": 1,
      "ref": "master",
      "sha": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
      "status": "pending"
    },
    "ref": "master",
    "stage": "test",
    "status": "failed",
	  "failure_reason": "script_failure",
    "tag": false,
    "web_url": "https://example.com/foo/bar/-/jobs/7"
  },
  {
    "commit": {
      "author_email": "admin@example.com",
      "author_name": "Administrator",
      "id": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
      "message": "Test the CI integration.",
      "short_id": "0ff3ae19",
      "title": "Test the CI integration."
    },
    "allow_failure": false,
    "duration": 0.192,
    "tag_list": [
      "docker runner",
      "win10-2004"
    ],
    "id": 6,
    "name": "rspec:other",
    "pipeline": {
      "id": 6,
      "project_id": 1,
      "ref": "master",
      "sha": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
      "status": "pending"
    },
    "ref": "master",
    "runner": null,
    "stage": "test",
    "status": "failed",
    "tag": false,
    "web_url": "https://example.com/foo/bar/-/jobs/6"
  }
]`)
	})

	jobs, resp, err := client.Jobs.ListProjectJobs(1, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Job{
		{
			Commit: &Commit{
				ID:          "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
				ShortID:     "0ff3ae19",
				Title:       "Test the CI integration.",
				AuthorName:  "Administrator",
				AuthorEmail: "admin@example.com",
			},
			AllowFailure: false,
			ID:           7,
			Name:         "teaspoon",
			TagList:      []string{"docker runner", "ubuntu18"},
			Pipeline: JobPipeline{
				ID:        6,
				ProjectID: 1,
				Ref:       "master",
				Sha:       "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
				Status:    "pending",
			},
			Ref:           "master",
			Stage:         "test",
			Status:        "failed",
			FailureReason: "script_failure",
			Tag:           false,
			WebURL:        "https://example.com/foo/bar/-/jobs/7",
		},
		{
			Commit: &Commit{
				ID:          "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
				ShortID:     "0ff3ae19",
				Title:       "Test the CI integration.",
				AuthorName:  "Administrator",
				AuthorEmail: "admin@example.com",
				Message:     "Test the CI integration.",
			},
			AllowFailure: false,
			Duration:     0.192,
			ID:           6,
			Name:         "rspec:other",
			TagList:      []string{"docker runner", "win10-2004"},
			Pipeline: JobPipeline{
				ID:        6,
				ProjectID: 1,
				Ref:       "master",
				Sha:       "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
				Status:    "pending",
			},
			Ref:    "master",
			Stage:  "test",
			Status: "failed",
			Tag:    false,
			WebURL: "https://example.com/foo/bar/-/jobs/6",
		},
	}
	assert.Equal(t, want, jobs)
}

func TestDownloadArtifactsFile(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	wantContent := []byte("This is the archive content")
	mux.HandleFunc("/api/v4/projects/9/jobs/artifacts/abranch/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testParam(t, r, "job", "publish")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `This is the archive content`)
	})

	opt := &DownloadArtifactsFileOptions{Job: Ptr("publish")}
	reader, resp, err := client.Jobs.DownloadArtifactsFile(9, "abranch", opt)
	assert.NoError(t, err)

	content, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, wantContent, content)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDownloadSingleArtifactsFileByTagOrBranch(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	wantContent := []byte("This is the file content")
	mux.HandleFunc("/api/v4/projects/9/jobs/artifacts/abranch/raw/foo/bar.pdf", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `This is the file content`)
	})

	opt := &DownloadArtifactsFileOptions{Job: Ptr("publish")}
	reader, resp, err := client.Jobs.DownloadSingleArtifactsFileByTagOrBranch(9, "abranch", "foo/bar.pdf", opt)
	assert.NoError(t, err)

	content, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, wantContent, content)
	assert.Equal(t, 200, resp.StatusCode)
}
