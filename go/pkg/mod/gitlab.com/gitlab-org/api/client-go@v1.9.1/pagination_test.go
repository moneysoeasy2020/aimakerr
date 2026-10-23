//go:build go1.23

package gitlab

import (
	"errors"
	"fmt"
	"iter"
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPagination_Scan_OffsetBased(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	handleTwoPagesSuccessfully(t, mux)

	opt := &ListProjectsOptions{}
	it, hasErr := Scan(func(p PaginationOptionFunc) ([]*Project, *Response, error) {
		return client.Projects.ListProjects(opt, p)
	})
	projects := slices.Collect(it)

	require.NoError(t, hasErr())

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestPagination_Scan_KeysetBased(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	handleTwoPagesSuccessfullyWithKeyset(t, mux)

	opt := &ListProjectsOptions{
		ListOptions: ListOptions{
			Pagination: "keyset",
			OrderBy:    "id",
		},
	}
	it, hasErr := Scan(func(p PaginationOptionFunc) ([]*Project, *Response, error) {
		return client.Projects.ListProjects(opt, p)
	})
	projects := slices.Collect(it)

	require.NoError(t, hasErr())

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestPagination_Scan_Error(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	handleTwoPagesWithFailure(t, mux)

	opt := &ListProjectsOptions{}
	it, hasErr := Scan(func(p PaginationOptionFunc) ([]*Project, *Response, error) {
		return client.Projects.ListProjects(opt, p)
	})
	projects := slices.Collect(it)

	require.Error(t, hasErr())

	want := []*Project{{ID: 1}}
	assert.Equal(t, want, projects)
}

func TestPagination_Scan_ExhaustedError(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	handleTwoPagesWithFailure(t, mux)

	opt := &ListProjectsOptions{}
	_, hasErr := Scan(func(p PaginationOptionFunc) ([]*Project, *Response, error) {
		return client.Projects.ListProjects(opt, p)
	})

	assert.Panics(t, func() {
		hasErr()
	})
}

func TestPagination_Scan2(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	handleTwoPagesSuccessfully(t, mux)

	opt := &ListProjectsOptions{}
	it := Scan2(func(p PaginationOptionFunc) ([]*Project, *Response, error) {
		return client.Projects.ListProjects(opt, p)
	})
	// collect
	projects := make([]*Project, 0, 2)
	for p, err := range it {
		require.NoError(t, err)
		projects = append(projects, p)
	}

	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestPagination_Scan2_Error(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	handleTwoPagesWithFailure(t, mux)

	opt := &ListProjectsOptions{}
	it := Scan2(func(p PaginationOptionFunc) ([]*Project, *Response, error) {
		return client.Projects.ListProjects(opt, p)
	})
	// collect
	projects := make([]*Project, 0, 2)
	next, stop := iter.Pull2(it)
	defer stop()

	// first item / page
	p, err, valid := next()
	require.True(t, valid)
	require.NoError(t, err)
	projects = append(projects, p)

	// second item / page
	_, err, valid = next()
	require.True(t, valid)
	require.Error(t, err)

	want := []*Project{{ID: 1}}
	assert.Equal(t, want, projects)
}

func TestPagination_Must(t *testing.T) {
	t.Parallel()

	it := Must(func(yield func(int, error) bool) { yield(42, nil) })
	xs := slices.Collect(it)

	require.Len(t, xs, 1)
	require.Equal(t, 42, xs[0])
}

func TestPagination_Must_Error(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		it := Must(func(yield func(int, error) bool) { yield(0, errors.New("sentinel")) })
		_ = slices.Collect(it)
	})
}

func TestPagination_ScanAndCollect(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	handleTwoPagesSuccessfully(t, mux)

	opt := &ListProjectsOptions{}
	projects, err := ScanAndCollect(func(p PaginationOptionFunc) ([]*Project, *Response, error) {
		return client.Projects.ListProjects(opt, p)
	})
	require.NoError(t, err)
	want := []*Project{{ID: 1}, {ID: 2}}
	assert.Equal(t, want, projects)
}

func TestPagination_ScanAndCollect_Error(t *testing.T) {
	t.Parallel()

	mux, client := setup(t)
	handleTwoPagesWithFailure(t, mux)

	opt := &ListProjectsOptions{}
	projects, err := ScanAndCollect(func(p PaginationOptionFunc) ([]*Project, *Response, error) {
		return client.Projects.ListProjects(opt, p)
	})
	require.Error(t, err)
	require.Nil(t, projects)
}

func handleTwoPagesSuccessfully(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "": // implicit first page
			w.Header().Add("X-Next-Page", "2")
			fmt.Fprint(w, `[{"id":1}]`)
		case "2":
			w.Header().Add("X-Next-Page", "0")
			fmt.Fprint(w, `[{"id":2}]`)
		default:
			assert.Fail(t, fmt.Sprintf("received request for unexpected page '%s'", page))
		}
	})
}

func handleTwoPagesSuccessfullyWithKeyset(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		idBefore := r.URL.Query().Get("id_before")
		switch idBefore {
		case "":
			w.Header().Add("Link", `<https://gitlab.example.com/api/v4/projects?id_before=2>; rel="next"`)
			fmt.Fprint(w, `[{"id":1}]`)
		case "2":
			fmt.Fprint(w, `[{"id":2}]`)
		default:
			assert.Fail(t, fmt.Sprintf("received request for unexpected page for id_before='%s'", idBefore))
		}
	})
}

func handleTwoPagesWithFailure(t *testing.T, mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "": // implicit first page
			w.Header().Add("X-Next-Page", "2")
			fmt.Fprint(w, `[{"id":1}]`)
		case "2":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			assert.Failf(t, "received request for unexpected page %d", page)
		}
	})
}
