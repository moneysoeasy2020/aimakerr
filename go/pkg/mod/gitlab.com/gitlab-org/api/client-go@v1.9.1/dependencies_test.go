package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencies_ListProjectDependencies(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/dependencies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		[
			{
				"name": "rails",
				"version": "5.0.1",
				"package_manager": "bundler",
				"dependency_file_path": "Gemfile.lock",
				"vulnerabilities": [
					{
						"name": "DDoS",
						"severity": "unknown",
						"id": 144827,
						"url": "https://gitlab.example.com/group/project/-/security/vulnerabilities/144827"
					}
				],
				"licenses": [
					{
						"name": "MIT",
						"url": "https://opensource.org/licenses/MIT"
					}
				]
			},
			{
				"name": "hanami",
				"version": "1.3.1",
				"package_manager": "bundler",
				"dependency_file_path": "Gemfile.lock",
				"vulnerabilities": [],
				"licenses": [
					{
						"name": "MIT",
						"url": "https://opensource.org/licenses/MIT"
					}
				]
			}
		]
		`)
	})

	want := []*Dependency{
		{
			Name:               "rails",
			Version:            "5.0.1",
			PackageManager:     "bundler",
			DependencyFilePath: "Gemfile.lock",
			Vulnerabilities: []*DependencyVulnerability{
				{
					Name:     "DDoS",
					Severity: "unknown",
					ID:       144827,
					URL:      "https://gitlab.example.com/group/project/-/security/vulnerabilities/144827",
				},
			},
			Licenses: []*DependencyLicense{
				{
					Name: "MIT",
					URL:  "https://opensource.org/licenses/MIT",
				},
			},
		},
		{
			Name:               "hanami",
			Version:            "1.3.1",
			PackageManager:     "bundler",
			DependencyFilePath: "Gemfile.lock",
			Vulnerabilities:    []*DependencyVulnerability{},
			Licenses: []*DependencyLicense{
				{
					Name: "MIT",
					URL:  "https://opensource.org/licenses/MIT",
				},
			},
		},
	}

	dependencies, resp, err := client.Dependencies.ListProjectDependencies(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, dependencies)
}

func TestDependencies_ListProjectDependencies_WithPackageManagers(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/dependencies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "package_manager=go%2Cbundler", r.URL.RawQuery)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[]`)
	})

	_, resp, err := client.Dependencies.ListProjectDependencies(1, &ListProjectDependenciesOptions{
		PackageManager: []*DependencyPackageManagerValue{Ptr(Go), Ptr(Bundler)},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
