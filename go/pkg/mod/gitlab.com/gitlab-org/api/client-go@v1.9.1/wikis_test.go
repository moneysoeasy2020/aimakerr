//
// Copyright 2021, Stany MARCEL
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

package gitlab

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListWikis(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/wikis", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `[
			{
			  "content": "Here is an instruction how to deploy this project.",
			  "format": "markdown",
			  "slug": "deploy",
			  "title": "deploy"
			},
			{
			  "content": "Our development process is described here.",
			  "format": "markdown",
			  "slug": "development",
			  "title": "development"
			},
			{
			  "content": "*  [Deploy](deploy)\n*  [Development](development)",
			  "format": "markdown",
			  "slug": "home",
			  "title": "home"
			}
		  ]`)
	})

	wikis, resp, err := client.Wikis.ListWikis(1, &ListWikisOptions{WithContent: Ptr(true)})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Wiki{
		{
			Content: "Here is an instruction how to deploy this project.",
			Format:  "markdown",
			Slug:    "deploy",
			Title:   "deploy",
		},
		{
			Content: "Our development process is described here.",
			Format:  "markdown",
			Slug:    "development",
			Title:   "development",
		},
		{
			Content: "*  [Deploy](deploy)\n*  [Development](development)",
			Format:  "markdown",
			Slug:    "home",
			Title:   "home",
		},
	}

	assert.Equal(t, want, wikis)
}

func TestGetWikiPage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/wikis/home", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{
			"content": "home page",
			"format": "markdown",
			"slug": "home",
			"title": "home",
			"encoding": "UTF-8"
		  }`)
	})

	wiki, resp, err := client.Wikis.GetWikiPage(1, "home", &GetWikiPageOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &Wiki{
		Content:  "home page",
		Encoding: "UTF-8",
		Format:   "markdown",
		Slug:     "home",
		Title:    "home",
	}

	assert.Equal(t, want, wiki)
}

func TestCreateWikiPage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/wikis", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{
			"content": "Hello world",
			"format": "markdown",
			"slug": "Hello",
			"title": "Hello"
		  }`)
	})

	opt := &CreateWikiPageOptions{
		Content: Ptr("Hello world"),
		Title:   Ptr("Hello"),
		Format:  Ptr(WikiFormatMarkdown),
	}
	wiki, resp, err := client.Wikis.CreateWikiPage(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &Wiki{
		Content: "Hello world",
		Format:  "markdown",
		Slug:    "Hello",
		Title:   "Hello",
	}

	assert.Equal(t, want, wiki)
}

func TestEditWikiPage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/wikis/foo", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `{
			"content": "documentation",
			"format": "markdown",
			"slug": "Docs",
			"title": "Docs"
		  }`)
	})

	opt := &EditWikiPageOptions{
		Content: Ptr("documentation"),
		Format:  Ptr(WikiFormatMarkdown),
		Title:   Ptr("Docs"),
	}
	wiki, resp, err := client.Wikis.EditWikiPage(1, "foo", opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &Wiki{
		Content: "documentation",
		Format:  "markdown",
		Slug:    "Docs",
		Title:   "Docs",
	}

	assert.Equal(t, want, wiki)
}

func TestDeleteWikiPage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/wikis/foo", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Wikis.DeleteWikiPage(1, "foo")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestUploadWikiAttachment(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/wikis/attachments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
			{
				"file_name" : "dk.png",
				"file_path" : "uploads/6a061c4cf9f1c28cb22c384b4b8d4e3c/dk.png",
				"branch" : "main",
				"link" : {
					"url" : "uploads/6a061c4cf9f1c28cb22c384b4b8d4e3c/dk.png",
					"markdown" : "![A description of the attachment](uploads/6a061c4cf9f1c28cb22c384b4b8d4e3c/dk.png)"
				}
			}
		`)
	})

	want := &WikiAttachment{
		FileName: "dk.png",
		FilePath: "uploads/6a061c4cf9f1c28cb22c384b4b8d4e3c/dk.png",
		Branch:   "main",
		Link: WikiAttachmentLink{
			URL:      "uploads/6a061c4cf9f1c28cb22c384b4b8d4e3c/dk.png",
			Markdown: "![A description of the attachment](uploads/6a061c4cf9f1c28cb22c384b4b8d4e3c/dk.png)",
		},
	}

	b := strings.NewReader("dummy")
	attachment, resp, err := client.Wikis.UploadWikiAttachment(1, b, "dk.png", &UploadWikiAttachmentOptions{Branch: Ptr("main")})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, attachment)
}
