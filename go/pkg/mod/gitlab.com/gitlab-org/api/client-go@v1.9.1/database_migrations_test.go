package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseMigrations_MarkMigrationAsSuccessful(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/admin/migrations/1/mark", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
	})
	opt := &MarkMigrationAsSuccessfulOptions{
		Database: "mydb",
	}
	resp, err := client.DatabaseMigrations.MarkMigrationAsSuccessful(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
