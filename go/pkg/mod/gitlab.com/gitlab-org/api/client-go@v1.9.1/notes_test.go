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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNotes_ListIssueNotes(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/issues/2/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			[
				{
					"id": 302,
					"body": "closed",
					"attachment": null,
					"author": {
						"id": 1,
						"username": "pipin",
						"email": "admin@example.com",
						"name": "Pip",
						"state": "active"
					},
					"created_at": "2013-10-02T09:56:03.0Z",
					"updated_at": "2013-10-02T09:56:03.0Z",
					"system": true,
					"noteable_id": 377,
					"noteable_type": "Issue",
					"project_id": 5,
					"noteable_iid": 377,
					"resolvable": false,
					"confidential": false,
					"internal": false
				},
				{
					"id": 305,
					"body": "Text of the comment\r\n",
					"attachment": null,
					"author": {
						"id": 1,
						"username": "pipin",
						"email": "admin@example.com",
						"name": "Pip",
						"state": "active"
					},
					"created_at": "2013-10-02T09:56:03.0Z",
					"updated_at": "2013-10-02T09:56:03.0Z",
					"system": true,
					"noteable_id": 121,
					"noteable_type": "Issue",
					"project_id": 5,
					"noteable_iid": 121,
					"resolvable": false,
					"confidential": true,
					"internal": true
				}
			]
		`)
	})

	createdAt := time.Date(2013, time.October, 2, 9, 56, 3, 0, time.UTC)
	want := []*Note{
		{
			ID:         302,
			Body:       "closed",
			Attachment: "",
			Author: NoteAuthor{
				ID:       1,
				Username: "pipin",
				Email:    "admin@example.com",
				Name:     "Pip",
				State:    "active",
			},
			CreatedAt:    &createdAt,
			UpdatedAt:    &createdAt,
			System:       true,
			NoteableID:   377,
			NoteableType: "Issue",
			ProjectID:    5,
			NoteableIID:  377,
			Resolvable:   false,
			Confidential: false,
			Internal:     false,
		},
		{
			ID:         305,
			Body:       "Text of the comment\r\n",
			Attachment: "",
			Author: NoteAuthor{
				ID:       1,
				Username: "pipin",
				Email:    "admin@example.com",
				Name:     "Pip",
				State:    "active",
			},
			CreatedAt:    &createdAt,
			UpdatedAt:    &createdAt,
			System:       true,
			NoteableID:   121,
			NoteableType: "Issue",
			ProjectID:    5,
			NoteableIID:  121,
			Resolvable:   false,
			Confidential: true,
			Internal:     true,
		},
	}

	notes, resp, err := client.Notes.ListIssueNotes(1, 2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, notes)
}

func TestNotes_GetIssueNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/issues/2/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			{
				"id": 302,
				"body": "closed",
				"attachment": null,
				"author": {
					"id": 1,
					"username": "pipin",
					"email": "admin@example.com",
					"name": "Pip",
					"state": "active"
				},
				"created_at": "2013-10-02T09:56:03.0Z",
				"updated_at": "2013-10-02T09:56:03.0Z",
				"system": true,
				"noteable_id": 377,
				"noteable_type": "Issue",
				"project_id": 5,
				"noteable_iid": 377,
				"resolvable": false,
				"confidential": false,
				"internal": false
			},
		`)
	})

	createdAt := time.Date(2013, time.October, 2, 9, 56, 3, 0, time.UTC)
	want := &Note{
		ID:         302,
		Body:       "closed",
		Attachment: "",
		Author: NoteAuthor{
			ID:       1,
			Username: "pipin",
			Email:    "admin@example.com",
			Name:     "Pip",
			State:    "active",
		},
		CreatedAt:    &createdAt,
		UpdatedAt:    &createdAt,
		System:       true,
		NoteableID:   377,
		NoteableType: "Issue",
		ProjectID:    5,
		NoteableIID:  377,
		Resolvable:   false,
		Confidential: false,
		Internal:     false,
	}

	note, resp, err := client.Notes.GetIssueNote(1, 2, 3)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestNotes_CreateIssueNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/issues/2/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
			{
				"id": 302,
				"body": "closed",
				"attachment": null,
				"author": {
					"id": 1,
					"username": "pipin",
					"email": "admin@example.com",
					"name": "Pip",
					"state": "active"
				},
				"created_at": "2013-10-02T09:56:03.0Z",
				"updated_at": "2013-10-02T09:56:03.0Z",
				"system": true,
				"noteable_id": 377,
				"noteable_type": "Issue",
				"project_id": 5,
				"noteable_iid": 377,
				"resolvable": false,
				"confidential": false,
				"internal": false
			},
		`)
	})

	createdAt := time.Date(2013, time.October, 2, 9, 56, 3, 0, time.UTC)
	want := &Note{
		ID:         302,
		Body:       "closed",
		Attachment: "",
		Author: NoteAuthor{
			ID:       1,
			Username: "pipin",
			Email:    "admin@example.com",
			Name:     "Pip",
			State:    "active",
		},
		CreatedAt:    &createdAt,
		UpdatedAt:    &createdAt,
		System:       true,
		NoteableID:   377,
		NoteableType: "Issue",
		ProjectID:    5,
		NoteableIID:  377,
		Resolvable:   false,
		Confidential: false,
		Internal:     false,
	}

	note, resp, err := client.Notes.CreateIssueNote(1, 2, &CreateIssueNoteOptions{Body: Ptr("closed"), Internal: Ptr(false), CreatedAt: &createdAt})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestNotes_UpdateIssueNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/issues/2/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
			{
				"id": 302,
				"body": "closed2",
				"attachment": null,
				"author": {
					"id": 1,
					"username": "pipin",
					"email": "admin@example.com",
					"name": "Pip",
					"state": "active"
				},
				"created_at": "2013-10-02T09:56:03.0Z",
				"updated_at": "2013-10-02T09:56:03.0Z",
				"system": true,
				"noteable_id": 377,
				"noteable_type": "Issue",
				"project_id": 5,
				"noteable_iid": 377,
				"resolvable": false,
				"confidential": false,
				"internal": false
			},
		`)
	})

	createdAt := time.Date(2013, time.October, 2, 9, 56, 3, 0, time.UTC)
	want := &Note{
		ID:         302,
		Body:       "closed2",
		Attachment: "",
		Author: NoteAuthor{
			ID:       1,
			Username: "pipin",
			Email:    "admin@example.com",
			Name:     "Pip",
			State:    "active",
		},
		CreatedAt:    &createdAt,
		UpdatedAt:    &createdAt,
		System:       true,
		NoteableID:   377,
		NoteableType: "Issue",
		ProjectID:    5,
		NoteableIID:  377,
		Resolvable:   false,
		Confidential: false,
		Internal:     false,
	}

	note, resp, err := client.Notes.UpdateIssueNote(1, 2, 3, &UpdateIssueNoteOptions{Body: Ptr("closed")})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestNotes_DeleteIssueNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/issues/2/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204)
	})

	resp, err := client.Notes.DeleteIssueNote(1, 2, 3)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestNotes_ListSnippetNotes(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/snippets/2/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			[
				{
					"id": 302,
					"body": "closed",
					"attachment": null,
					"author": {
						"id": 1,
						"username": "pipin",
						"email": "admin@example.com",
						"name": "Pip",
						"state": "active"
					},
					"created_at": "2013-10-02T09:56:03.0Z",
					"updated_at": "2013-10-02T09:56:03.0Z",
					"system": true,
					"noteable_id": 377,
					"noteable_type": "Issue",
					"project_id": 5,
					"noteable_iid": 377,
					"resolvable": false,
					"confidential": false,
					"internal": false
				},
				{
					"id": 305,
					"body": "Text of the comment\r\n",
					"attachment": null,
					"author": {
						"id": 1,
						"username": "pipin",
						"email": "admin@example.com",
						"name": "Pip",
						"state": "active"
					},
					"created_at": "2013-10-02T09:56:03.0Z",
					"updated_at": "2013-10-02T09:56:03.0Z",
					"system": true,
					"noteable_id": 121,
					"noteable_type": "Issue",
					"project_id": 5,
					"noteable_iid": 121,
					"resolvable": false,
					"confidential": true,
					"internal": true
				}
			]
		`)
	})

	createdAt := time.Date(2013, time.October, 2, 9, 56, 3, 0, time.UTC)
	want := []*Note{
		{
			ID:         302,
			Body:       "closed",
			Attachment: "",
			Author: NoteAuthor{
				ID:       1,
				Username: "pipin",
				Email:    "admin@example.com",
				Name:     "Pip",
				State:    "active",
			},
			CreatedAt:    &createdAt,
			UpdatedAt:    &createdAt,
			System:       true,
			NoteableID:   377,
			NoteableType: "Issue",
			ProjectID:    5,
			NoteableIID:  377,
			Resolvable:   false,
			Confidential: false,
			Internal:     false,
		},
		{
			ID:         305,
			Body:       "Text of the comment\r\n",
			Attachment: "",
			Author: NoteAuthor{
				ID:       1,
				Username: "pipin",
				Email:    "admin@example.com",
				Name:     "Pip",
				State:    "active",
			},
			CreatedAt:    &createdAt,
			UpdatedAt:    &createdAt,
			System:       true,
			NoteableID:   121,
			NoteableType: "Issue",
			ProjectID:    5,
			NoteableIID:  121,
			Resolvable:   false,
			Confidential: true,
			Internal:     true,
		},
	}

	notes, resp, err := client.Notes.ListSnippetNotes(1, 2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, notes)
}

func TestNotes_CreateSnippetNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/snippets/2/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
			{
				"id": 302,
				"body": "closed",
				"attachment": null,
				"author": {
					"id": 1,
					"username": "pipin",
					"email": "admin@example.com",
					"name": "Pip",
					"state": "active"
				},
				"created_at": "2013-10-02T09:56:03.0Z",
				"updated_at": "2013-10-02T09:56:03.0Z",
				"system": true,
				"noteable_id": 377,
				"noteable_type": "Issue",
				"project_id": 5,
				"noteable_iid": 377,
				"resolvable": false,
				"confidential": false,
				"internal": false
			},
		`)
	})

	createdAt := time.Date(2013, time.October, 2, 9, 56, 3, 0, time.UTC)
	want := &Note{
		ID:         302,
		Body:       "closed",
		Attachment: "",
		Author: NoteAuthor{
			ID:       1,
			Username: "pipin",
			Email:    "admin@example.com",
			Name:     "Pip",
			State:    "active",
		},
		CreatedAt:    &createdAt,
		UpdatedAt:    &createdAt,
		System:       true,
		NoteableID:   377,
		NoteableType: "Issue",
		ProjectID:    5,
		NoteableIID:  377,
		Resolvable:   false,
		Confidential: false,
		Internal:     false,
	}

	note, resp, err := client.Notes.CreateSnippetNote(1, 2, &CreateSnippetNoteOptions{Body: Ptr("closed"), CreatedAt: &createdAt})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestNotes_UpdateSnippetNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/snippets/2/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
			{
				"id": 302,
				"body": "closed2",
				"attachment": null,
				"author": {
					"id": 1,
					"username": "pipin",
					"email": "admin@example.com",
					"name": "Pip",
					"state": "active"
				},
				"created_at": "2013-10-02T09:56:03.0Z",
				"updated_at": "2013-10-02T09:56:03.0Z",
				"system": true,
				"noteable_id": 377,
				"noteable_type": "Issue",
				"project_id": 5,
				"noteable_iid": 377,
				"resolvable": false,
				"confidential": false,
				"internal": false
			},
		`)
	})

	createdAt := time.Date(2013, time.October, 2, 9, 56, 3, 0, time.UTC)
	want := &Note{
		ID:         302,
		Body:       "closed2",
		Attachment: "",
		Author: NoteAuthor{
			ID:       1,
			Username: "pipin",
			Email:    "admin@example.com",
			Name:     "Pip",
			State:    "active",
		},
		CreatedAt:    &createdAt,
		UpdatedAt:    &createdAt,
		System:       true,
		NoteableID:   377,
		NoteableType: "Issue",
		ProjectID:    5,
		NoteableIID:  377,
		Resolvable:   false,
		Confidential: false,
		Internal:     false,
	}

	note, resp, err := client.Notes.UpdateSnippetNote(1, 2, 3, &UpdateSnippetNoteOptions{Body: Ptr("closed")})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestNotes_DeleteSnippetNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/snippets/2/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204)
	})

	resp, err := client.Notes.DeleteSnippetNote(1, 2, 3)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestNotes_ListMergeRequestNotes(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/4329/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
		{
			"id": 3,
			"type": "DiffNote",
			"body": "foo bar",
			"attachment": null,
			"system": false,
			"noteable_id": 4392,
			"noteable_type": "Epic",
			"resolvable": false,
			"noteable_iid": null
		}]`)
	})

	want := []*Note{{
		ID:           3,
		Type:         DiffNote,
		Body:         "foo bar",
		System:       false,
		NoteableID:   4392,
		NoteableType: "Epic",
	}}

	notes, resp, err := client.Notes.ListMergeRequestNotes("1", 4329, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, notes)
}

func TestNotes_GetMergeRequestNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/4329/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		{
			"id": 3,
			"type": "DiffNote",
			"body": "foo bar",
			"attachment": null,
			"system": false,
			"noteable_id": 4392,
			"noteable_type": "Epic",
			"resolvable": false,
			"noteable_iid": null
		}`)
	})

	want := &Note{
		ID:           3,
		Type:         DiffNote,
		Body:         "foo bar",
		System:       false,
		NoteableID:   4392,
		NoteableType: "Epic",
	}

	note, resp, err := client.Notes.GetMergeRequestNote("1", 4329, 3, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestNotes_CreateMergeRequestNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/4329/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
		{
			"id": 3,
			"type": "DiffNote",
			"body": "foo bar",
			"attachment": null,
			"system": false,
			"noteable_id": 4392,
			"noteable_type": "Epic",
			"resolvable": false,
			"noteable_iid": null
		}`)
	})

	want := &Note{
		ID:           3,
		Type:         DiffNote,
		Body:         "foo bar",
		System:       false,
		NoteableID:   4392,
		NoteableType: "Epic",
	}

	note, resp, err := client.Notes.CreateMergeRequestNote("1", 4329, &CreateMergeRequestNoteOptions{Body: Ptr("foo bar")})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestNotes_UpdateMergeRequestNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/4329/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
		{
			"id": 3,
			"type": "DiffNote",
			"body": "foo bar",
			"attachment": null,
			"system": false,
			"noteable_id": 4392,
			"noteable_type": "Epic",
			"resolvable": false,
			"noteable_iid": null
		}`)
	})

	want := &Note{
		ID:           3,
		Type:         DiffNote,
		Body:         "foo bar",
		System:       false,
		NoteableID:   4392,
		NoteableType: "Epic",
	}

	note, resp, err := client.Notes.UpdateMergeRequestNote("1", 4329, 3, &UpdateMergeRequestNoteOptions{Body: Ptr("foo bar")})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestNotes_DeleteMergeRequestNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/2/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204)
	})

	resp, err := client.Notes.DeleteMergeRequestNote(1, 2, 3)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestNotes_GetEpicNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/epics/4329/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		{
			"id": 3,
			"type": null,
			"body": "foo bar",
			"attachment": null,
			"system": false,
			"noteable_id": 4392,
			"noteable_type": "Epic",
			"resolvable": false,
			"noteable_iid": null
		}`)
	})

	want := &Note{
		ID:           3,
		Body:         "foo bar",
		Attachment:   "",
		Title:        "",
		FileName:     "",
		System:       false,
		NoteableID:   4392,
		NoteableType: "Epic",
	}

	note, resp, err := client.Notes.GetEpicNote("1", 4329, 3, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, note)
}

func TestCreateNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/issues/1/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"id": 1, "body": "Body of note", "author" : {"id" : 1, "name": "snehal", "username": "snehal", "state": "active", "email": "snehal@example.com"}}`)
	})

	createNoteOptions := &CreateIssueNoteOptions{
		Body: Ptr("Body of note"),
	}

	note, _, err := client.Notes.CreateIssueNote("1", 1, createNoteOptions)
	if err != nil {
		t.Fatal(err)
	}

	want := &Note{
		ID:   1,
		Body: "Body of note",
		Author: NoteAuthor{
			ID: 1, Username: "snehal", Name: "snehal", Email: "snehal@example.com", State: "active", AvatarURL: "", WebURL: "",
		},
		Internal: false,
	}

	if !reflect.DeepEqual(want, note) {
		t.Errorf("Notes.CreateNote returned %+v, want %+v", note, want)
	}
}

func TestCreateInternalNote(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/issues/1/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"id": 1, "body": "Body of internal note", "internal": true}`)
	})

	createNoteOptions := &CreateIssueNoteOptions{
		Body: Ptr("Body of internal note"),
	}

	note, _, err := client.Notes.CreateIssueNote("1", 1, createNoteOptions)
	if err != nil {
		t.Fatal(err)
	}

	want := &Note{
		ID:       1,
		Body:     "Body of internal note",
		Internal: true,
	}

	if !reflect.DeepEqual(want, note) {
		t.Errorf("Notes.CreateNote returned %+v, want %+v", note, want)
	}
}
