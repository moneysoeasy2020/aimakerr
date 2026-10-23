//
// Copyright 2021, Pavel Kostohrys
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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAvatar(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	const url = "https://www.gravatar.com/avatar/10e6bf7bcf22c2f00a3ef684b4ada178"

	mux.HandleFunc("/api/v4/avatar", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusAccepted)
		avatar := Avatar{AvatarURL: url}
		resp, _ := json.Marshal(avatar)
		_, _ = w.Write(resp)
	})

	opt := &GetAvatarOptions{Email: Ptr("sander@vanharmelen.nnl")}
	avatar, resp, err := client.Avatar.GetAvatar(opt)

	require.NoError(t, err, "Avatar.GetAvatar should not return an error")
	require.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, "202 Accepted", resp.Status, "Expected HTTP status 202 Accepted")

	require.NotNil(t, avatar, "Avatar should not be nil")
	assert.Equal(t, url, avatar.AvatarURL, "Avatar.GetAvatar returned unexpected URL")
}
