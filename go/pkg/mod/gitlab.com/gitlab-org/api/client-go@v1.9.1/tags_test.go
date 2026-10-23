//
// Copyright 2021, Sander van Harmelen
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
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTagsService_ListTags(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `[
      {
        "name": "1.0.0",
        "message": "test",
        "target": "fffff",
        "protected": false
      },{
        "name": "1.0.1",
        "protected": true
      }
    ]`)
	})

	opt := &ListTagsOptions{ListOptions: ListOptions{Page: 2, PerPage: 3}}

	tags, resp, err := client.Tags.ListTags(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*Tag{
		{
			Name:      "1.0.0",
			Message:   "test",
			Target:    "fffff",
			Protected: false,
		},
		{
			Name:      "1.0.1",
			Protected: true,
		},
	}
	assert.Equal(t, want, tags)
}

func TestTagsService_GetTag(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/tags/v5.0.0", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
			{
				"name": "v5.0.0",
				"message": null,
				"target": "60a8ff033665e1207714d6670fcd7b65304ec02f",
				"commit": {
					"id": "60a8ff033665e1207714d6670fcd7b65304ec02f",
					"short_id": "60a8ff03",
					"title": "Initial commit",
					"created_at": "2015-02-01T21:56:31.000Z",
					"parent_ids": [
						"f61c062ff8bcbdb00e0a1b3317a91aed6ceee06b"
					],
					"message": "v5.0.0\n",
					"author_name": "Arthur Verschaeve",
					"author_email": "contact@arthurverschaeve.be",
					"authored_date": "2015-02-01T21:56:31.000Z",
					"committer_name": "Arthur Verschaeve",
					"committer_email": "contact@arthurverschaeve.be",
					"committed_date": "2015-02-01T21:56:31.000Z"
				},
				"release": null,
				"protected": false,
				"created_at": "2015-02-01T21:56:31.000Z"
			}
		`)
	})

	tag, resp, err := client.Tags.GetTag(1, "v5.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	date := time.Date(2015, time.February, 1, 21, 56, 31, 0, time.UTC)

	want := &Tag{
		Name:   "v5.0.0",
		Target: "60a8ff033665e1207714d6670fcd7b65304ec02f",
		Commit: &Commit{
			ID:             "60a8ff033665e1207714d6670fcd7b65304ec02f",
			ShortID:        "60a8ff03",
			Title:          "Initial commit",
			CreatedAt:      &date,
			ParentIDs:      []string{"f61c062ff8bcbdb00e0a1b3317a91aed6ceee06b"},
			Message:        "v5.0.0\n",
			AuthorName:     "Arthur Verschaeve",
			AuthorEmail:    "contact@arthurverschaeve.be",
			AuthoredDate:   &date,
			CommitterName:  "Arthur Verschaeve",
			CommitterEmail: "contact@arthurverschaeve.be",
			CommittedDate:  &date,
		},
		Release:   nil,
		Protected: false,
		CreatedAt: &date,
	}

	assert.Equal(t, want, tag)
}

func TestTagsService_CreateTag(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
		{
			"commit": {
				"id": "2695effb5807a22ff3d138d593fd856244e155e7",
				"short_id": "2695effb",
				"title": "Initial commit",
				"created_at": "2015-02-01T21:56:31.000Z",
				"parent_ids": [
					"2a4b78934375d7f53875269ffd4f45fd83a84ebe"
				],
				"message": "Initial commit",
				"author_name": "John Smith",
				"author_email": "john@example.com",
				"authored_date": "2015-02-01T21:56:31.000Z",
				"committer_name": "Jack Smith",
				"committer_email": "jack@example.com",
				"committed_date": "2015-02-01T21:56:31.000Z"
			},
			"release": null,
			"name": "v1.0.0",
			"target": "2695effb5807a22ff3d138d593fd856244e155e7",
			"message": null,
			"protected": false,
			"created_at": null
		}`)
	})

	opt := &CreateTagOptions{
		TagName: Ptr("v1.0.0"),
	}

	tag, resp, err := client.Tags.CreateTag(1, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	date := time.Date(2015, time.February, 1, 21, 56, 31, 0, time.UTC)

	want := &Tag{
		Name:   "v1.0.0",
		Target: "2695effb5807a22ff3d138d593fd856244e155e7",
		Commit: &Commit{
			ID:             "2695effb5807a22ff3d138d593fd856244e155e7",
			ShortID:        "2695effb",
			Title:          "Initial commit",
			CreatedAt:      &date,
			ParentIDs:      []string{"2a4b78934375d7f53875269ffd4f45fd83a84ebe"},
			Message:        "Initial commit",
			AuthorName:     "John Smith",
			AuthorEmail:    "john@example.com",
			AuthoredDate:   &date,
			CommitterName:  "Jack Smith",
			CommitterEmail: "jack@example.com",
			CommittedDate:  &date,
		},
		Release:   nil,
		Protected: false,
	}

	assert.Equal(t, want, tag)
}

func TestTagsService_DeleteTag(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/tags/v5.0.0", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.Tags.DeleteTag(1, "v5.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestTagsService_GetTagSignature(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/repository/tags/v1.0.0%2Frc-1/signature", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"signature_type": "X509",
			"verification_status": "unverified",
			"x509_certificate": {
				"id": 1,
				"subject": "CN=gitlab@example.org,OU=Example,O=World",
				"subject_key_identifier": "BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC",
				"email": "gitlab@example.org",
				"serial_number": 278969561018901340486471282831158785578,
				"certificate_status": "good",
				"x509_issuer": {
				"id": 1,
				"subject": "CN=PKI,OU=Example,O=World",
				"subject_key_identifier": "AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB",
				"crl_url": "http://example.com/pki.crl"
				}
			}
		}`)
	})

	signature, _, err := client.Tags.GetTagSignature(1, "v1.0.0/rc-1", nil)
	if err != nil {
		t.Errorf("Tags.GetTagSignature returned error: %v", err)
	}

	serialNumber, _ := big.NewInt(0).SetString("278969561018901340486471282831158785578", 10)
	want := &X509Signature{
		SignatureType:      "X509",
		VerificationStatus: "unverified",
		X509Certificate: X509Certificate{
			ID:                   1,
			Subject:              "CN=gitlab@example.org,OU=Example,O=World",
			SubjectKeyIdentifier: "BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC:BC",
			Email:                "gitlab@example.org",
			SerialNumber:         serialNumber,
			CertificateStatus:    "good",
			X509Issuer: X509Issuer{
				ID:                   1,
				Subject:              "CN=PKI,OU=Example,O=World",
				SubjectKeyIdentifier: "AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB:AB",
				CrlURL:               "http://example.com/pki.crl",
			},
		},
	}

	assert.Equal(t, want, signature)
}
