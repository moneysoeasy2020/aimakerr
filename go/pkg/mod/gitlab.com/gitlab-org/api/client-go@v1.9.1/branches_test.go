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
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const DefaultBranch = "master"

func TestGetBranch(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/branches/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/get_branch.json")
	})

	branch, resp, err := client.Branches.GetBranch(1, "master")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	authoredDate := time.Date(2012, time.June, 27, 5, 51, 39, 0, time.UTC)
	committedDate := time.Date(2012, time.June, 28, 3, 44, 20, 0, time.UTC)
	want := &Branch{
		Name:               DefaultBranch,
		Merged:             false,
		Protected:          true,
		Default:            true,
		DevelopersCanPush:  false,
		DevelopersCanMerge: false,
		CanPush:            true,
		Commit: &Commit{
			AuthorEmail:    "john@example.com",
			AuthorName:     exampleEventUserName,
			AuthoredDate:   &authoredDate,
			CommittedDate:  &committedDate,
			CommitterEmail: "john@example.com",
			CommitterName:  exampleEventUserName,
			ID:             "7b5c3cc8be40ee161ae89a06bba6229da1032a0c",
			ShortID:        "7b5c3cc",
			Title:          "add projects API",
			Message:        "add projects API",
			ParentIDs:      []string{"4ad91d3c1144c406e50c7b33bae684bd6837faf8"},
		},
	}

	assert.Equal(t, want, branch)
}

func TestBranchesService_ListBranches(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/repository/branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/list_branches.json")
	})

	authoredDate := time.Date(2012, time.June, 27, 5, 51, 39, 0, time.UTC)
	committedDate := time.Date(2012, time.June, 28, 3, 44, 20, 0, time.UTC)
	want := []*Branch{
		{
			Name:               DefaultBranch,
			Merged:             false,
			Protected:          true,
			Default:            true,
			DevelopersCanPush:  false,
			DevelopersCanMerge: false,
			CanPush:            true,
			WebURL:             "https://gitlab.example.com/my-group/my-project/-/tree/master",
			Commit: &Commit{
				AuthorEmail:    "john@example.com",
				AuthorName:     exampleEventUserName,
				AuthoredDate:   &authoredDate,
				CommittedDate:  &committedDate,
				CommitterEmail: "john@example.com",
				CommitterName:  exampleEventUserName,
				ID:             "7b5c3cc8be40ee161ae89a06bba6229da1032a0c",
				ShortID:        "7b5c3cc",
				Title:          "add projects API",
				Message:        "add projects API",
				ParentIDs:      []string{"4ad91d3c1144c406e50c7b33bae684bd6837faf8"},
			},
		},
		{
			Name:               "develop",
			Merged:             false,
			Protected:          true,
			Default:            true,
			DevelopersCanPush:  false,
			DevelopersCanMerge: false,
			CanPush:            true,
			WebURL:             "https://gitlab.example.com/my-group/my-project/-/tree/develop",
			Commit: &Commit{
				AuthorEmail:    "john@example.com",
				AuthorName:     exampleEventUserName,
				AuthoredDate:   &authoredDate,
				CommittedDate:  &committedDate,
				CommitterEmail: "john@example.com",
				CommitterName:  exampleEventUserName,
				ID:             "7b5c3cc8be40ee161ae89a06bba6229da1032a0c",
				ShortID:        "7b5c3cc",
				Title:          "add projects API",
				Message:        "add projects API",
				ParentIDs:      []string{"4ad91d3c1144c406e50c7b33bae684bd6837faf8"},
			},
		},
	}

	b, resp, err := client.Branches.ListBranches(5, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, b)

	b, resp, err = client.Branches.ListBranches(5.01, nil)
	assert.ErrorIs(t, err, ErrInvalidIDType)
	assert.Nil(t, resp)
	assert.Nil(t, b)

	b, resp, err = client.Branches.ListBranches(5, nil, errorOption)
	assert.ErrorIs(t, err, errRequestOptionFunc)
	assert.Nil(t, resp)
	assert.Nil(t, b)

	b, resp, err = client.Branches.ListBranches(3, nil)
	assert.Error(t, err)
	assert.Nil(t, b)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestBranchesService_ListBranchesWithSearch(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/repository/branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testURL(t, r, "/api/v4/projects/5/repository/branches?search=feature")
		mustWriteHTTPResponse(t, w, "testdata/list_branches.json")
	})

	opts := &ListBranchesOptions{
		Search: Ptr("feature"),
	}

	b, resp, err := client.Branches.ListBranches(5, opts)
	assert.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, b)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	authoredDate := time.Date(2012, time.June, 27, 5, 51, 39, 0, time.UTC)
	committedDate := time.Date(2012, time.June, 28, 3, 44, 20, 0, time.UTC)
	expectedBranches := []*Branch{
		{
			Name:               DefaultBranch,
			Merged:             false,
			Protected:          true,
			Default:            true,
			DevelopersCanPush:  false,
			DevelopersCanMerge: false,
			CanPush:            true,
			WebURL:             "https://gitlab.example.com/my-group/my-project/-/tree/master",
			Commit: &Commit{
				AuthorEmail:    "john@example.com",
				AuthorName:     exampleEventUserName,
				AuthoredDate:   &authoredDate,
				CommittedDate:  &committedDate,
				CommitterEmail: "john@example.com",
				CommitterName:  exampleEventUserName,
				ID:             "7b5c3cc8be40ee161ae89a06bba6229da1032a0c",
				ShortID:        "7b5c3cc",
				Title:          "add projects API",
				Message:        "add projects API",
				ParentIDs:      []string{"4ad91d3c1144c406e50c7b33bae684bd6837faf8"},
			},
		},
		{
			Name:               "develop",
			Merged:             false,
			Protected:          true,
			Default:            true,
			DevelopersCanPush:  false,
			DevelopersCanMerge: false,
			CanPush:            true,
			WebURL:             "https://gitlab.example.com/my-group/my-project/-/tree/develop",
			Commit: &Commit{
				AuthorEmail:    "john@example.com",
				AuthorName:     exampleEventUserName,
				AuthoredDate:   &authoredDate,
				CommittedDate:  &committedDate,
				CommitterEmail: "john@example.com",
				CommitterName:  exampleEventUserName,
				ID:             "7b5c3cc8be40ee161ae89a06bba6229da1032a0c",
				ShortID:        "7b5c3cc",
				Title:          "add projects API",
				Message:        "add projects API",
				ParentIDs:      []string{"4ad91d3c1144c406e50c7b33bae684bd6837faf8"},
			},
		},
	}
	assert.Equal(t, expectedBranches, b)
}

func TestBranchesService_ListBranchesWithRegex(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/repository/branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testParam(t, r, "regex", "^feature-.*")
		mustWriteHTTPResponse(t, w, "testdata/list_branches.json")
	})

	opts := &ListBranchesOptions{
		Regex: Ptr("^feature-.*"),
	}

	b, resp, err := client.Branches.ListBranches(5, opts)
	assert.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, b)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	authoredDate := time.Date(2012, time.June, 27, 5, 51, 39, 0, time.UTC)
	committedDate := time.Date(2012, time.June, 28, 3, 44, 20, 0, time.UTC)
	expectedBranch := []*Branch{
		{
			Name:               DefaultBranch,
			Merged:             false,
			Protected:          true,
			Default:            true,
			DevelopersCanPush:  false,
			DevelopersCanMerge: false,
			CanPush:            true,
			WebURL:             "https://gitlab.example.com/my-group/my-project/-/tree/master",
			Commit: &Commit{
				AuthorEmail:    "john@example.com",
				AuthorName:     exampleEventUserName,
				AuthoredDate:   &authoredDate,
				CommittedDate:  &committedDate,
				CommitterEmail: "john@example.com",
				CommitterName:  exampleEventUserName,
				ID:             "7b5c3cc8be40ee161ae89a06bba6229da1032a0c",
				ShortID:        "7b5c3cc",
				Title:          "add projects API",
				Message:        "add projects API",
				ParentIDs:      []string{"4ad91d3c1144c406e50c7b33bae684bd6837faf8"},
			},
		},
		{
			Name:               "develop",
			Merged:             false,
			Protected:          true,
			Default:            true,
			DevelopersCanPush:  false,
			DevelopersCanMerge: false,
			CanPush:            true,
			WebURL:             "https://gitlab.example.com/my-group/my-project/-/tree/develop",
			Commit: &Commit{
				AuthorEmail:    "john@example.com",
				AuthorName:     exampleEventUserName,
				AuthoredDate:   &authoredDate,
				CommittedDate:  &committedDate,
				CommitterEmail: "john@example.com",
				CommitterName:  exampleEventUserName,
				ID:             "7b5c3cc8be40ee161ae89a06bba6229da1032a0c",
				ShortID:        "7b5c3cc",
				Title:          "add projects API",
				Message:        "add projects API",
				ParentIDs:      []string{"4ad91d3c1144c406e50c7b33bae684bd6837faf8"},
			},
		},
	}
	assert.Equal(t, expectedBranch, b)
}

func TestBranchesService_CreateBranch(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		mustWriteHTTPResponse(t, w, "testdata/get_branch.json")
	})

	authoredDate := time.Date(2012, time.June, 27, 5, 51, 39, 0, time.UTC)
	committedDate := time.Date(2012, time.June, 28, 3, 44, 20, 0, time.UTC)
	want := &Branch{
		Name:               "master",
		Merged:             false,
		Protected:          true,
		Default:            true,
		DevelopersCanPush:  false,
		DevelopersCanMerge: false,
		CanPush:            true,
		Commit: &Commit{
			AuthorEmail:    "john@example.com",
			AuthorName:     exampleEventUserName,
			AuthoredDate:   &authoredDate,
			CommittedDate:  &committedDate,
			CommitterEmail: "john@example.com",
			CommitterName:  exampleEventUserName,
			ID:             "7b5c3cc8be40ee161ae89a06bba6229da1032a0c",
			ShortID:        "7b5c3cc",
			Title:          "add projects API",
			Message:        "add projects API",
			ParentIDs:      []string{"4ad91d3c1144c406e50c7b33bae684bd6837faf8"},
		},
	}

	b, resp, err := client.Branches.CreateBranch(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, b)

	b, resp, err = client.Branches.CreateBranch(1.01, nil)
	assert.ErrorIs(t, err, ErrInvalidIDType)
	assert.Nil(t, resp)
	assert.Nil(t, b)

	b, resp, err = client.Branches.CreateBranch(1, nil, errorOption)
	assert.ErrorIs(t, err, errRequestOptionFunc)
	assert.Nil(t, resp)
	assert.Nil(t, b)

	b, resp, err = client.Branches.CreateBranch(3, nil)
	assert.Error(t, err)
	assert.Nil(t, b)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestBranchesService_DeleteBranch(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/branches/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Branches.DeleteBranch(1, "master", nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	resp, err = client.Branches.DeleteBranch(1.01, "master", nil)
	assert.ErrorIs(t, err, ErrInvalidIDType)
	assert.Nil(t, resp)

	resp, err = client.Branches.DeleteBranch(1, "master", nil, errorOption)
	assert.ErrorIs(t, err, errRequestOptionFunc)
	assert.Nil(t, resp)

	resp, err = client.Branches.DeleteBranch(3, "master", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestBranchesService_DeleteMergedBranches(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/merged_branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Branches.DeleteMergedBranches(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	resp, err = client.Branches.DeleteMergedBranches(1.01, nil)
	assert.ErrorIs(t, err, ErrInvalidIDType)
	assert.Nil(t, resp)

	resp, err = client.Branches.DeleteMergedBranches(1, nil, errorOption)
	assert.ErrorIs(t, err, errRequestOptionFunc)
	assert.Nil(t, resp)

	resp, err = client.Branches.DeleteMergedBranches(3, nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
