//go:build integration

package gitlab_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// This test ensures that the datetime fields parses properly
// when reading epics from a group.
func TestEpicTimeFieldParsing(t *testing.T) {
	t.Parallel()

	client := SetupIntegrationClient(t)
	SkipIfNotLicensed(t, client)

	// GIVEN a test group
	group := CreateTestGroup(t, client)

	// AND date values set for the epic
	startDate := gitlab.ISOTime(time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC))
	dueDate := gitlab.ISOTime(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))

	// WHEN creating an epic with fixed start and due dates
	epic, err := CreateTestEpicWithOptions(t, client, group.ID, &gitlab.CreateEpicOptions{
		Title:            gitlab.Ptr("Epic with datetime fields"),
		Description:      gitlab.Ptr("Testing ISO 8601 datetime parsing"),
		StartDateIsFixed: gitlab.Ptr(true),
		StartDateFixed:   &startDate,
		DueDateIsFixed:   gitlab.Ptr(true),
		DueDateFixed:     &dueDate,
	})
	require.NoError(t, err, "Failed to create epic")

	// THEN the epic should be created successfully
	assert.NotNil(t, epic)
	assert.Equal(t, "Epic with datetime fields", epic.Title)

	fetchedEpic, _, err := client.Epics.GetEpic(group.ID, epic.IID)
	require.NoError(t, err, "Failed to fetch epic")

	// AND all datetime fields should be parsed successfully
	assert.NotNil(t, fetchedEpic.StartDate, "StartDate should not be nil")
	assert.NotNil(t, fetchedEpic.DueDate, "DueDate should not be nil")
	assert.NotNil(t, fetchedEpic.StartDateFixed, "StartDateFixed should not be nil")
	assert.NotNil(t, fetchedEpic.DueDateFixed, "DueDateFixed should not be nil")

	// AND the fixed dates should match what was set
	if fetchedEpic.StartDateFixed != nil {
		expectedStart := time.Time(startDate)
		actualStart := time.Time(*fetchedEpic.StartDateFixed)
		assert.Equal(t, expectedStart.Year(), actualStart.Year(), "StartDateFixed year should match")
		assert.Equal(t, expectedStart.Month(), actualStart.Month(), "StartDateFixed month should match")
		assert.Equal(t, expectedStart.Day(), actualStart.Day(), "StartDateFixed day should match")
	}

	if fetchedEpic.DueDateFixed != nil {
		expectedDue := time.Time(dueDate)
		actualDue := time.Time(*fetchedEpic.DueDateFixed)
		assert.Equal(t, expectedDue.Year(), actualDue.Year(), "DueDateFixed year should match")
		assert.Equal(t, expectedDue.Month(), actualDue.Month(), "DueDateFixed month should match")
		assert.Equal(t, expectedDue.Day(), actualDue.Day(), "DueDateFixed day should match")
	}
}
