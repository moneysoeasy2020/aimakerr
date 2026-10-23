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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAllDeployKeys(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/deploy_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `[
		{
			"id": 1,
			"title": "Public key",
			"key": "ssh-rsa AAAA...",
			"fingerprint": "7f:72:08:7d:0e:47:48:ec:37:79:b2:76:68:b5:87:65",
			"created_at": "2013-10-02T10:12:29Z",
			"projects_with_write_access": [
			{
				"id": 73,
				"description": null,
				"name": "project2",
				"name_with_namespace": "Sidney Jones / project2",
				"path": "project2",
				"path_with_namespace": "sidney_jones/project2",
				"created_at": "2021-10-25T18:33:17.550Z"
			},
			{
				"id": 74,
				"description": null,
				"name": "project3",
				"name_with_namespace": "Sidney Jones / project3",
				"path": "project3",
				"path_with_namespace": "sidney_jones/project3",
				"created_at": "2021-10-25T18:33:17.666Z"
			}
			]
		},
			{
				"id": 3,
				"title": "Another Public key",
				"key": "ssh-rsa AAAA...",
				"fingerprint": "64:d3:73:d4:83:70:ab:41:96:68:d5:3d:a5:b0:34:ea",
				"created_at": "2013-10-02T11:12:29Z",
				"projects_with_write_access": []
			}
		  ]`)
	})

	deployKeys, resp, err := client.DeployKeys.ListAllDeployKeys(&ListInstanceDeployKeysOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*InstanceDeployKey{
		{
			ID:          1,
			Title:       "Public key",
			Key:         "ssh-rsa AAAA...",
			CreatedAt:   mustParseTime("2013-10-02T10:12:29Z"),
			Fingerprint: "7f:72:08:7d:0e:47:48:ec:37:79:b2:76:68:b5:87:65",
			ProjectsWithWriteAccess: []*DeployKeyProject{
				{
					ID:                73,
					Description:       "",
					Name:              "project2",
					NameWithNamespace: "Sidney Jones / project2",
					Path:              "project2",
					PathWithNamespace: "sidney_jones/project2",
					CreatedAt:         mustParseTime("2021-10-25T18:33:17.550Z"),
				},
				{
					ID:                74,
					Description:       "",
					Name:              "project3",
					NameWithNamespace: "Sidney Jones / project3",
					Path:              "project3",
					PathWithNamespace: "sidney_jones/project3",
					CreatedAt:         mustParseTime("2021-10-25T18:33:17.666Z"),
				},
			},
		},
		{
			ID:                      3,
			Title:                   "Another Public key",
			Key:                     "ssh-rsa AAAA...",
			Fingerprint:             "64:d3:73:d4:83:70:ab:41:96:68:d5:3d:a5:b0:34:ea",
			CreatedAt:               mustParseTime("2013-10-02T11:12:29Z"),
			ProjectsWithWriteAccess: []*DeployKeyProject{},
		},
	}
	assert.Equal(t, want, deployKeys)
}

func TestAddInstanceDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/deploy_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{
			"key" : "ssh-rsa AAAA...",
			"id" : 12,
			"title" : "My deploy key",
			"can_push": true,
			"created_at" : "2015-08-29T12:44:31.550Z",
			"expires_at": null
		 }`)
	})

	opt := &AddInstanceDeployKeyOptions{
		Key:   Ptr("ssh-rsa AAAA..."),
		Title: Ptr("My deploy key"),
	}
	deployKey, resp, err := client.DeployKeys.AddInstanceDeployKey(opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &InstanceDeployKey{
		Title:     "My deploy key",
		ID:        12,
		Key:       "ssh-rsa AAAA...",
		CreatedAt: mustParseTime("2015-08-29T12:44:31.550Z"),
	}
	assert.Equal(t, want, deployKey)
}

func TestListProjectDeployKeys(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/deploy_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `[
			{
			  "id": 1,
			  "title": "Public key",
			  "key": "ssh-rsa AAAA...",
			  "fingerprint": "4a:9d:64:15:ed:3a:e6:07:6e:89:36:b3:3b:03:05:d9",
			  "fingerprint_sha256": "SHA256:Jrs3LD1Ji30xNLtTVf9NDCj7kkBgPBb2pjvTZ3HfIgU",
			  "created_at": "2013-10-02T10:12:29Z",
			  "can_push": false
			},
			{
			  "id": 3,
			  "title": "Another Public key",
			  "key": "ssh-rsa AAAA...",
			  "fingerprint": "0b:cf:58:40:b9:23:96:c7:ba:44:df:0e:9e:87:5e:75",
			  "fingerprint_sha256": "SHA256:lGI/Ys/Wx7PfMhUO1iuBH92JQKYN+3mhJZvWO4Q5ims",
			  "created_at": "2013-10-02T11:12:29Z",
			  "can_push": false
			}
		  ]`)
	})

	deployKeys, resp, err := client.DeployKeys.ListProjectDeployKeys(5, &ListProjectDeployKeysOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*ProjectDeployKey{
		{
			ID:                1,
			Title:             "Public key",
			Key:               "ssh-rsa AAAA...",
			Fingerprint:       "4a:9d:64:15:ed:3a:e6:07:6e:89:36:b3:3b:03:05:d9",
			FingerprintSHA256: "SHA256:Jrs3LD1Ji30xNLtTVf9NDCj7kkBgPBb2pjvTZ3HfIgU",
			CreatedAt:         mustParseTime("2013-10-02T10:12:29Z"),
			CanPush:           false,
		},
		{
			ID:                3,
			Title:             "Another Public key",
			Key:               "ssh-rsa AAAA...",
			Fingerprint:       "0b:cf:58:40:b9:23:96:c7:ba:44:df:0e:9e:87:5e:75",
			FingerprintSHA256: "SHA256:lGI/Ys/Wx7PfMhUO1iuBH92JQKYN+3mhJZvWO4Q5ims",
			CreatedAt:         mustParseTime("2013-10-02T11:12:29Z"),
			CanPush:           false,
		},
	}
	assert.Equal(t, want, deployKeys)
}

func TestListUserProjectDeployKeys(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/users/5/project_deploy_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `[
			{
			  "id": 1,
			  "title": "Public key",
			  "key": "ssh-rsa AAAA...",
			  "fingerprint": "4a:9d:64:15:ed:3a:e6:07:6e:89:36:b3:3b:03:05:d9",
			  "fingerprint_sha256": "SHA256:Jrs3LD1Ji30xNLtTVf9NDCj7kkBgPBb2pjvTZ3HfIgU",
			  "created_at": "2013-10-02T10:12:29Z",
			  "can_push": false
			},
			{
			  "id": 3,
			  "title": "Another Public key",
			  "key": "ssh-rsa AAAA...",
			  "fingerprint": "0b:cf:58:40:b9:23:96:c7:ba:44:df:0e:9e:87:5e:75",
			  "fingerprint_sha256": "SHA256:lGI/Ys/Wx7PfMhUO1iuBH92JQKYN+3mhJZvWO4Q5ims",
			  "created_at": "2013-10-02T11:12:29Z",
			  "can_push": false
			}
		  ]`)
	})

	deployKeys, resp, err := client.DeployKeys.ListUserProjectDeployKeys(5, &ListUserProjectDeployKeysOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := []*ProjectDeployKey{
		{
			ID:                1,
			Title:             "Public key",
			Key:               "ssh-rsa AAAA...",
			Fingerprint:       "4a:9d:64:15:ed:3a:e6:07:6e:89:36:b3:3b:03:05:d9",
			FingerprintSHA256: "SHA256:Jrs3LD1Ji30xNLtTVf9NDCj7kkBgPBb2pjvTZ3HfIgU",
			CreatedAt:         mustParseTime("2013-10-02T10:12:29Z"),
			CanPush:           false,
		},
		{
			ID:                3,
			Title:             "Another Public key",
			Key:               "ssh-rsa AAAA...",
			Fingerprint:       "0b:cf:58:40:b9:23:96:c7:ba:44:df:0e:9e:87:5e:75",
			FingerprintSHA256: "SHA256:lGI/Ys/Wx7PfMhUO1iuBH92JQKYN+3mhJZvWO4Q5ims",
			CreatedAt:         mustParseTime("2013-10-02T11:12:29Z"),
			CanPush:           false,
		},
	}
	assert.Equal(t, want, deployKeys)
}

func TestGetDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/deploy_keys/11", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{
			"id": 1,
			"title": "Public key",
			"key": "ssh-rsa AAAA...",
			"fingerprint": "4a:9d:64:15:ed:3a:e6:07:6e:89:36:b3:3b:03:05:d9",
			"fingerprint_sha256": "SHA256:Jrs3LD1Ji30xNLtTVf9NDCj7kkBgPBb2pjvTZ3HfIgU",
			"created_at": "2013-10-02T10:12:29Z",
			"can_push": false
		  }`)
	})

	deployKey, resp, err := client.DeployKeys.GetDeployKey(5, 11)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectDeployKey{
		ID:                1,
		Title:             "Public key",
		Key:               "ssh-rsa AAAA...",
		Fingerprint:       "4a:9d:64:15:ed:3a:e6:07:6e:89:36:b3:3b:03:05:d9",
		FingerprintSHA256: "SHA256:Jrs3LD1Ji30xNLtTVf9NDCj7kkBgPBb2pjvTZ3HfIgU",
		CreatedAt:         mustParseTime("2013-10-02T10:12:29Z"),
		CanPush:           false,
	}
	assert.Equal(t, want, deployKey)
}

func TestAddDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/deploy_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{
			"key" : "ssh-rsa AAAA...",
			"id" : 12,
			"title" : "My deploy key",
			"can_push": true,
			"created_at" : "2015-08-29T12:44:31.550Z",
			"expires_at": null
		 }`)
	})

	opt := &AddDeployKeyOptions{
		Key:     Ptr("ssh-rsa AAAA..."),
		Title:   Ptr("My deploy key"),
		CanPush: Ptr(true),
	}
	deployKey, resp, err := client.DeployKeys.AddDeployKey(5, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectDeployKey{
		Title:     "My deploy key",
		ID:        12,
		Key:       "ssh-rsa AAAA...",
		CreatedAt: mustParseTime("2015-08-29T12:44:31.550Z"),
		CanPush:   true,
	}
	assert.Equal(t, want, deployKey)
}

func TestAddDeployKey_withExpiresAt(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/deploy_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{
			"key" : "ssh-rsa AAAA...",
			"id" : 12,
			"title" : "My deploy key",
			"can_push": true,
			"created_at" : "2015-08-29T12:44:31.550Z",
			"expires_at": "2999-03-01T00:00:00.000Z"
		 }`)
	})

	opt := &AddDeployKeyOptions{
		Key:       Ptr("ssh-rsa AAAA..."),
		Title:     Ptr("My deploy key"),
		CanPush:   Ptr(true),
		ExpiresAt: mustParseTime("2999-03-01T00:00:00.000Z"),
	}
	deployKey, resp, err := client.DeployKeys.AddDeployKey(5, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectDeployKey{
		Title:     "My deploy key",
		ID:        12,
		Key:       "ssh-rsa AAAA...",
		CreatedAt: mustParseTime("2015-08-29T12:44:31.550Z"),
		CanPush:   true,
		ExpiresAt: mustParseTime("2999-03-01T00:00:00.000Z"),
	}
	assert.Equal(t, want, deployKey)
}

func TestDeleteDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/deploy_keys/13", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.DeployKeys.DeleteDeployKey(5, 13)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestEnableDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/deploy_keys/13/enable", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{
			"key" : "ssh-rsa AAAA...",
			"id" : 12,
			"title" : "My deploy key",
			"created_at" : "2015-08-29T12:44:31.550Z"
		 }`)
	})

	deployKey, resp, err := client.DeployKeys.EnableDeployKey(5, 13)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectDeployKey{
		ID:        12,
		Title:     "My deploy key",
		Key:       "ssh-rsa AAAA...",
		CreatedAt: mustParseTime("2015-08-29T12:44:31.550Z"),
	}
	assert.Equal(t, want, deployKey)
}

func TestUpdateDeployKey(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/5/deploy_keys/11", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `{
			"id": 11,
			"title": "New deploy key",
			"key": "ssh-rsa AAAA...",
			"created_at": "2015-08-29T12:44:31.550Z",
			"can_push": true
		 }`)
	})

	opt := &UpdateDeployKeyOptions{
		Title:   Ptr("New deploy key"),
		CanPush: Ptr(true),
	}
	deployKey, resp, err := client.DeployKeys.UpdateDeployKey(5, 11, opt)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &ProjectDeployKey{
		ID:        11,
		Title:     "New deploy key",
		Key:       "ssh-rsa AAAA...",
		CreatedAt: mustParseTime("2015-08-29T12:44:31.550Z"),
		CanPush:   true,
	}
	assert.Equal(t, want, deployKey)
}
