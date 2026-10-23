//
// Copyright 2021, Sune Keller
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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishPackageFile(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1234/packages/generic/foo/0.1.2/bar-baz.txt", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
		{
			"message": "201 Created"
		}
	`)
	})

	_, _, err := client.GenericPackages.PublishPackageFile(1234, "foo", "0.1.2", "bar-baz.txt", strings.NewReader("bar = baz"), &PublishPackageFileOptions{})
	if err != nil {
		t.Errorf("GenericPackages.PublishPackageFile returned error: %v", err)
	}
}

func TestDownloadPackageFile(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1234/packages/generic/foo/0.1.2/bar-baz.txt", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, strings.TrimSpace(`
		bar = baz
	`))
	})

	packageBytes, _, err := client.GenericPackages.DownloadPackageFile(1234, "foo", "0.1.2", "bar-baz.txt")
	if err != nil {
		t.Errorf("GenericPackages.DownloadPackageFile returned error: %v", err)
	}

	want := []byte("bar = baz")
	if !reflect.DeepEqual(want, packageBytes) {
		t.Errorf("GenericPackages.DownloadPackageFile returned %+v, want %+v", packageBytes, want)
	}
}

func TestFormatPackageURL(t *testing.T) {
	t.Parallel()
	_, client := setup(t)

	url, err := client.GenericPackages.FormatPackageURL(1234, "foo", "0.1.2", "bar-baz.txt")
	assert.NoError(t, err)

	want := "projects/1234/packages/generic/foo/0%2E1%2E2/bar-baz%2Etxt"
	assert.Equal(t, want, url)
}
