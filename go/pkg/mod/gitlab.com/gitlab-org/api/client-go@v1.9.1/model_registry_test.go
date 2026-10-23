package gitlab

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadMachineLearningModelPackage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	// Mock out the API request and response
	mux.HandleFunc("/api/v4/projects/1/packages/ml_models/2/files/path/filename", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		// Creating test data JSON files is optional, you can also include the JSON inline.
		fmt.Fprint(w, `fake content`)
	})

	// Call the function being tested
	registry, resp, err := client.ModelRegistry.DownloadMachineLearningModelPackage(1, 2, "path", "filename")

	require.NoError(t, err)
	assert.NotNil(t, resp)

	data, err := io.ReadAll(registry)
	require.NoError(t, err)

	want := []byte("fake content")

	assert.Equal(t, want, data)
}
