package gitlab

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupScheduleExport(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/export",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			fmt.Fprint(w, `{"message": "202 Accepted"}`)
		})

	resp, err := client.GroupImportExport.ScheduleExport(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGroupExportDownload(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/export/download",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, `fake content`)
		})

	export, resp, err := client.GroupImportExport.ExportDownload(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	data, err := io.ReadAll(export)
	assert.NoError(t, err)

	want := []byte("fake content")
	assert.Equal(t, want, data)
}

func TestGroupImport(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	content := []byte("temporary file's content")
	tmpfile := filepath.Join(t.TempDir(), "example.tar.gz")
	if err := os.WriteFile(tmpfile, content, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc("/api/v4/groups/import",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			fmt.Fprint(w, `{"message": "202 Accepted"}`)
		})

	opt := &GroupImportFileOptions{
		Name:     Ptr("test"),
		Path:     Ptr("path"),
		File:     Ptr(tmpfile),
		ParentID: Ptr(int64(1)),
	}

	resp, err := client.GroupImportExport.ImportFile(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
