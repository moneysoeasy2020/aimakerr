package gitlab

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownUploads_UploadProjectMarkdown(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/uploads", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("Projects.UploadFile request content-type %+v want multipart/form-data;", r.Header.Get("Content-Type"))
		}
		if r.ContentLength == -1 {
			t.Fatalf("Projects.UploadFile request content-length is -1")
		}
		fmt.Fprint(w, `
			{
				"id": 5,
				"alt": "dk",
				"url": "/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png",
				"full_path": "/-/project/1234/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png",
				"markdown": "![dk](/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png)"
			}
		`)
	})

	want := &ProjectMarkdownUploadedFile{
		ID:       5,
		Alt:      "dk",
		URL:      "/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png",
		FullPath: "/-/project/1234/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png",
		Markdown: "![dk](/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png)",
	}

	b := strings.NewReader("dummy")
	upload, resp, err := client.ProjectMarkdownUploads.UploadProjectMarkdown(1, b, "test.txt")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, upload)
}

func TestMarkdownUploads_UploadProjectMarkdown_Retry(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	isFirstRequest := true
	mux.HandleFunc("/api/v4/projects/1/uploads", func(w http.ResponseWriter, r *http.Request) {
		if isFirstRequest {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			isFirstRequest = false
			return
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data;") {
			t.Fatalf("Projects.UploadFile request content-type %+v want multipart/form-data;", r.Header.Get("Content-Type"))
		}
		if r.ContentLength == -1 {
			t.Fatalf("Projects.UploadFile request content-length is -1")
		}
		fmt.Fprint(w, `
			{
				"id": 5,
				"alt": "dk",
				"url": "/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png",
				"full_path": "/-/project/1234/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png",
				"markdown": "![dk](/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png)"
			}
		`)
	})

	want := &ProjectMarkdownUploadedFile{
		ID:       5,
		Alt:      "dk",
		URL:      "/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png",
		FullPath: "/-/project/1234/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png",
		Markdown: "![dk](/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png)",
	}

	b := strings.NewReader("dummy")
	upload, resp, err := client.ProjectMarkdownUploads.UploadProjectMarkdown(1, b, "test.txt")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, upload)
}

func TestMarkdownUploads_ListProjectMarkdownUploads(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/uploads", func(w http.ResponseWriter, r *http.Request) {
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
	want := []*ProjectMarkdownUpload{
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

	uploads, resp, err := client.ProjectMarkdownUploads.ListProjectMarkdownUploads(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, uploads)
}

func TestMarkdownUploads_DownloadProjectMarkdownUploadByID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/uploads/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, strings.TrimSpace(`
			bar = baz
		`))
	})

	want := []byte("bar = baz")

	bytes, resp, err := client.ProjectMarkdownUploads.DownloadProjectMarkdownUploadByID(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, bytes)
}

func TestMarkdownUploads_DownloadProjectMarkdownUploadBySecretAndFilename(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/uploads/secret/filename", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, strings.TrimSpace(`
			bar = baz
		`))
	})

	want := []byte("bar = baz")

	bytes, resp, err := client.ProjectMarkdownUploads.DownloadProjectMarkdownUploadBySecretAndFilename(1, "secret", "filename")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, bytes)
}

func TestMarkdownUploads_DeleteProjectMarkdownUploadByID(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/uploads/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204)
	})

	resp, err := client.ProjectMarkdownUploads.DeleteProjectMarkdownUploadByID(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestMarkdownUploads_DeleteProjectMarkdownUploadBySecretAndFilename(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/uploads/secret/filename", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204)
	})

	resp, err := client.ProjectMarkdownUploads.DeleteProjectMarkdownUploadBySecretAndFilename(1, "secret", "filename")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
}
