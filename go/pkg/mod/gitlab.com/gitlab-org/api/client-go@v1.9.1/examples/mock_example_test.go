// This example demonstrates how to use testing client with mocks

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/api/client-go"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/testing"
)

func TestMockExample(t *testing.T) {
	client := gitlabtesting.NewTestClient(t)
	opts := &gitlab.ListAgentsOptions{}
	expectedResp := &gitlab.Response{}
	pid := 1
	// Setup expectations
	client.MockClusterAgents.EXPECT().
		ListAgents(pid, opts).
		Return([]*gitlab.Agent{{ID: 1}}, expectedResp, nil)

	// Use the client in your test
	// You'd probably call your own code here that gets the client injected.
	// You can also retrieve a `gitlab.Client` object from `client.Client`.
	agents, resp, err := client.ClusterAgents.ListAgents(pid, opts)
	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	assert.Len(t, agents, 1)
}
