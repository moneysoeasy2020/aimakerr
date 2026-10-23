package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestClient_SmokeTestMetadataService(t *testing.T) {
	t.Parallel()

	// GIVEN
	tc := NewTestClient(t)

	// setup mock expectations
	tc.MockMetadata.EXPECT().GetMetadata().Times(1)

	// WHEN
	err := runMyApp(tc.Client)

	// THEN
	assert.NoError(t, err)
}

func runMyApp(client *gitlab.Client) error {
	// this is just an example for some kind of app that internally uses the gitlab.Client
	metadata, _, err := client.Metadata.GetMetadata()
	if err != nil {
		return err
	}

	fmt.Printf("Metadata: %+v\n", metadata)

	return nil
}
