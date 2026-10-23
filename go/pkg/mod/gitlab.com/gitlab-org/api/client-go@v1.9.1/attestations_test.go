package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAttestations(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1337/attestations/76c34666f719ef14bd2b124a7db51e9c05e4db2e12a84800296d559064eebe2c", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `[
		{
			"build_id": 1337,
			"created_at": "2025-10-07T20:59:27.085Z",
			"download_url": "https://gitlab.com/api/v4/projects/72356192/attestations/1/download",
			"expire_at": "2027-10-07T20:59:26.967Z",
			"id": 1,
			"iid": 1,
			"predicate_kind": "provenance",
			"predicate_type": "https://slsa.dev/provenance/v1",
			"project_id": 1337,
			"status": "success",
			"subject_digest": "76c34666f719ef14bd2b124a7db51e9c05e4db2e12a84800296d559064eebe2c",
			"updated_at": "2025-10-07T20:59:27.085Z"
		}
		]`)
	})

	attestations, resp, err := client.Attestations.ListAttestations("1337", "76c34666f719ef14bd2b124a7db51e9c05e4db2e12a84800296d559064eebe2c")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Attestation{
		{
			ID:            1,
			IID:           1,
			BuildID:       1337,
			DownloadURL:   "https://gitlab.com/api/v4/projects/72356192/attestations/1/download",
			CreatedAt:     mustParseTime("2025-10-07T20:59:27.085Z"),
			UpdatedAt:     mustParseTime("2025-10-07T20:59:27.085Z"),
			ExpireAt:      mustParseTime("2027-10-07T20:59:26.967Z"),
			PredicateKind: "provenance",
			PredicateType: "https://slsa.dev/provenance/v1",
			ProjectID:     1337,
			Status:        "success",
			SubjectDigest: "76c34666f719ef14bd2b124a7db51e9c05e4db2e12a84800296d559064eebe2c",
		},
	}
	assert.Equal(t, want, attestations)
}

func TestDownloadAttestation(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	const expectedOut = "expected_output"

	mux.HandleFunc("/api/v4/projects/1337/attestations/1/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, expectedOut)
	})

	outBytes, resp, err := client.Attestations.DownloadAttestation("1337", 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, []byte(expectedOut), outBytes)
}
