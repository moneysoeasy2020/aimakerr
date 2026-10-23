package gitlab

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGroupMarkdownUploads_ListGroupMarkdownUploads(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/uploads", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			[
				{
					"id": 1,
					"size": 1024,
					"filename": "image.png",
					"created_at":"2024-06-20T15:53:03.000Z",
					"uploaded_by": {
						"id": 18,
						"name" : "Alexandra Bashirian",
						"username" : "eileen.lowe"
					}
				},
				{
					"id": 2,
					"size": 512,
					"filename": "other-image.png",
					"created_at":"2024-06-19T15:53:03.000Z",
					"uploaded_by": null
				}
			]
		`)
	})

	created1 := time.Date(2024, time.June, 20, 15, 53, 3, 0, time.UTC)
	created2 := time.Date(2024, time.June, 19, 15, 53, 3, 0, time.UTC)
	want := []*GroupMarkdownUpload{
		{
			ID:        1,
			Size:      1024,
			Filename:  "image.png",
			CreatedAt: &created1,
			UploadedBy: &User{
				ID:       18,
				Name:     "Alexandra Bashirian",
				Username: "eileen.lowe",
			},
		},
		{
			ID:        2,
			Size:      512,
			Filename:  "other-image.png",
			CreatedAt: &created2,
		},
	}

	uploads, resp, err := client.GroupMarkdownUploads.ListGroupMarkdownUploads(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, uploads)
}

func TestGroupMarkdownUploads_DownloadGroupMarkdownUploadByID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/uploads/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, strings.TrimSpace(`
			bar = baz
		`))
	})

	var want bytes.Buffer
	want.Write([]byte("bar = baz"))

	bytes, resp, err := client.GroupMarkdownUploads.DownloadGroupMarkdownUploadByID(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, &want, bytes)
}

func TestGroupMarkdownUploads_DownloadGroupMarkdownUploadBySecretAndFilename(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/uploads/secret/filename", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, strings.TrimSpace(`
			bar = baz
		`))
	})

	var want bytes.Buffer
	want.Write([]byte("bar = baz"))

	bytes, resp, err := client.GroupMarkdownUploads.DownloadGroupMarkdownUploadBySecretAndFilename(1, "secret", "filename")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, &want, bytes)
}

func TestGroupMarkdownUploads_DeleteGroupMarkdownUploadByID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/uploads/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204)
	})

	resp, err := client.GroupMarkdownUploads.DeleteGroupMarkdownUploadByID(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestGroupMarkdownUploads_DeleteGroupMarkdownUploadBySecretAndFilename(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/uploads/secret/filename", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204)
	})

	resp, err := client.GroupMarkdownUploads.DeleteGroupMarkdownUploadBySecretAndFilename(1, "secret", "filename")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
}
