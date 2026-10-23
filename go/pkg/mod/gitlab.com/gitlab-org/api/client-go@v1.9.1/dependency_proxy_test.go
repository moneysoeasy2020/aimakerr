package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencyProxy_PurgeGroupDependencyProxy(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/groups/1/dependency_proxy/cache", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.DependencyProxy.PurgeGroupDependencyProxy(1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
