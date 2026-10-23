// Package testing provides test utilities for the GitLab API client
package testing

import (
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"go.uber.org/mock/gomock"
)

// TestClient wraps the GitLab client with mock implementations
// for testing purposes. It embeds the standard GitLab client and
// provides access to mock interfaces for the various services.
type TestClient struct {
	*gitlab.Client
	*testClientMocks
}

// NewTestClient creates a new TestClient with mocked services using the provided
// testing.T instance. It internally creates a new gomock.Controller automatically.
//
// Example:
//
//	func TestClusterAgentList(t *testing.T) {
//	    client := testing.NewTestClient(t)
//
//	    // Setup expectations
//	    client.MockClusterAgents.EXPECT().
//	        List(gomock.Any(), 123, nil).
//	        Return([]*gitlab.ClusterAgent{{ID: 1}}, nil)
//
//	    // Use the client in your test
//	    agents, err := client.ClusterAgents.List(ctx, 123, nil)
//	    assert.NoError(t, err)
//	    assert.Len(t, agents, 1)
//	}
func NewTestClient(t *testing.T, options ...gitlab.ClientOptionFunc) *TestClient {
	ctrl := gomock.NewController(t)
	return NewTestClientWithCtrl(ctrl, options...)
}

// NewTestClientWithCtrl creates a new TestClient with mocked services using
// a provided gomock.Controller. This is useful when you need more control over
// the mock controller's lifecycle or when sharing a controller across multiple
// test clients.
//
// Example:
//
//	func TestMultipleClients(t *testing.T) {
//	    ctrl := gomock.NewController(t)
//
//	    client1 := testing.NewTestClientWithCtrl(ctrl)
//	    client2 := testing.NewTestClientWithCtrl(ctrl)
//
//	    // Setup expectations for both clients
//	    client1.MockClusterAgents.EXPECT().
//	        Get(gomock.Any(), 123, 1).
//	        Return(&gitlab.ClusterAgent{ID: 1}, nil)
//
//	    client2.MockClusterAgents.EXPECT().
//	        Get(gomock.Any(), 123, 2).
//	        Return(&gitlab.ClusterAgent{ID: 2}, nil)
//
//	    // Use both clients in your test
//	}
func NewTestClientWithCtrl(ctrl *gomock.Controller, options ...gitlab.ClientOptionFunc) *TestClient {
	return newTestClientWithCtrl(ctrl, options...)
}
