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

func TestSecureFiles_ListProjectSecureFiles(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/secure_files", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			[
				{
					"id": 1,
					"name": "myfile.jks",
					"checksum": "16630b189ab34b2e3504f4758e1054d2e478deda510b2b08cc0ef38d12e80aac",
					"checksum_algorithm": "sha256",
					"created_at": "2022-02-22T22:22:22.000Z",
					"expires_at": null,
					"metadata": null
				},
				{
					"id": 2,
					"name": "myfile.cer",
					"checksum": "16630b189ab34b2e3504f4758e1054d2e478deda510b2b08cc0ef38d12e80aa2",
					"checksum_algorithm": "sha256",
					"created_at": "2022-02-22T22:22:22.000Z",
					"expires_at": "2023-09-21T14:55:59.000Z",
					"metadata": {
						"id":"75949910542696343243264405377658443914",
						"issuer": {
							"C":"US",
							"O":"Apple Inc.",
							"CN":"Apple Worldwide Developer Relations Certification Authority",
							"OU":"G3"
						},
						"subject": {
							"C":"US",
							"O":"Organization Name",
							"CN":"Apple Distribution: Organization Name (ABC123XYZ)",
							"OU":"ABC123XYZ",
							"UID":"ABC123XYZ"
						},
						"expires_at":"2023-09-21T14:55:59.000Z"
					}
				}
			]
		`)
	})

	createdAt := time.Date(2022, time.February, 22, 22, 22, 22, 0, time.UTC)
	expiresAt := time.Date(2023, time.September, 21, 14, 55, 59, 0, time.UTC)
	want := []*SecureFile{
		{
			ID:                1,
			Name:              "myfile.jks",
			Checksum:          "16630b189ab34b2e3504f4758e1054d2e478deda510b2b08cc0ef38d12e80aac",
			ChecksumAlgorithm: "sha256",
			CreatedAt:         &createdAt,
			ExpiresAt:         nil,
			Metadata:          nil,
		},
		{
			ID:                2,
			Name:              "myfile.cer",
			Checksum:          "16630b189ab34b2e3504f4758e1054d2e478deda510b2b08cc0ef38d12e80aa2",
			ChecksumAlgorithm: "sha256",
			CreatedAt:         &createdAt,
			ExpiresAt:         &expiresAt,
			Metadata: &SecureFileMetadata{
				ID: "75949910542696343243264405377658443914",
				Issuer: SecureFileIssuer{
					C:  "US",
					O:  "Apple Inc.",
					CN: "Apple Worldwide Developer Relations Certification Authority",
					OU: "G3",
				},
				Subject: SecureFileSubject{
					C:   "US",
					O:   "Organization Name",
					CN:  "Apple Distribution: Organization Name (ABC123XYZ)",
					OU:  "ABC123XYZ",
					UID: "ABC123XYZ",
				},
				ExpiresAt: &expiresAt,
			},
		},
	}

	files, resp, err := client.SecureFiles.ListProjectSecureFiles(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, files)
}

func TestSecureFiles_ShowSecureFileDetails(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/secure_files/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			{
				"id": 1,
				"name": "myfile.jks",
				"checksum": "16630b189ab34b2e3504f4758e1054d2e478deda510b2b08cc0ef38d12e80aac",
				"checksum_algorithm": "sha256",
				"created_at": "2022-02-22T22:22:22.000Z",
				"expires_at": null,
				"metadata": null
			}
		`)
	})

	createdAt := time.Date(2022, time.February, 22, 22, 22, 22, 0, time.UTC)
	want := &SecureFile{
		ID:                1,
		Name:              "myfile.jks",
		Checksum:          "16630b189ab34b2e3504f4758e1054d2e478deda510b2b08cc0ef38d12e80aac",
		ChecksumAlgorithm: "sha256",
		CreatedAt:         &createdAt,
		ExpiresAt:         nil,
		Metadata:          nil,
	}

	file, resp, err := client.SecureFiles.ShowSecureFileDetails(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, file)
}

func TestSecureFiles_CreateSecureFile(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/secure_files", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testFormBody(t, r, "name", "test.jks")

		fmt.Fprint(w, `
			{
				"id": 1,
				"name": "myfile.jks",
				"checksum": "16630b189ab34b2e3504f4758e1054d2e478deda510b2b08cc0ef38d12e80aac",
				"checksum_algorithm": "sha256",
				"created_at": "2022-02-22T22:22:22.0Z",
				"expires_at": null,
				"metadata": null
			}
		`)
	})

	createdAt := time.Date(2022, time.February, 22, 22, 22, 22, 0, time.UTC)
	want := &SecureFile{
		ID:                1,
		Name:              "myfile.jks",
		Checksum:          "16630b189ab34b2e3504f4758e1054d2e478deda510b2b08cc0ef38d12e80aac",
		ChecksumAlgorithm: "sha256",
		CreatedAt:         &createdAt,
		ExpiresAt:         nil,
		Metadata:          nil,
	}

	b := strings.NewReader("dummy")
	file, resp, err := client.SecureFiles.CreateSecureFile(1, b, &CreateSecureFileOptions{Name: Ptr("test.jks")})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, file)
}

func TestSecureFiles_DownloadSecureFile(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/secure_files/2/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, strings.TrimSpace(`
			bar = baz
		`))
	})

	var want bytes.Buffer
	want.Write([]byte("bar = baz"))

	bytes, resp, err := client.SecureFiles.DownloadSecureFile(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, &want, bytes)
}

func TestSecureFiles_RemoveSecureFile(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/secure_files/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(204)
	})

	resp, err := client.SecureFiles.RemoveSecureFile(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 204, resp.StatusCode)
}
