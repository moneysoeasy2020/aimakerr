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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPipelineSchedules_ListPipelineSchedules(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		[
			{
				"id": 13,
				"description": "Test schedule pipeline",
				"ref": "refs/heads/main",
				"cron": "* * * * *",
				"cron_timezone": "Asia/Tokyo",
				"next_run_at": "2017-05-19T13:41:00.000Z",
				"active": true,
				"created_at": "2017-05-19T13:41:00.000Z",
				"updated_at": "2017-05-19T13:41:00.000Z",
				"owner": {
					"name": "Administrator",
					"username": "root",
					"id": 1,
					"state": "active",
					"avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
					"web_url": "https://gitlab.example.com/root"
				}
			}
		]
		`)
	})

	schedules, resp, err := client.PipelineSchedules.ListPipelineSchedules(1, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	testDate := time.Date(2017, time.May, 19, 13, 41, 0, 0, time.UTC)
	want := []*PipelineSchedule{
		{
			ID:           13,
			Description:  "Test schedule pipeline",
			Ref:          "refs/heads/main",
			Cron:         "* * * * *",
			CronTimezone: "Asia/Tokyo",
			NextRunAt:    &testDate,
			Active:       true,
			CreatedAt:    &testDate,
			UpdatedAt:    &testDate,
			Owner: &User{
				Name:      "Administrator",
				Username:  "root",
				ID:        1,
				State:     "active",
				AvatarURL: "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
				WebURL:    "https://gitlab.example.com/root",
			},
		},
	}
	assert.Equal(t, want, schedules)
}

func TestPipelineSchedules_GetPipelineSchedule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		{
			"id": 13,
			"description": "Test schedule pipeline",
			"ref": "refs/heads/main",
			"cron": "* * * * *",
			"cron_timezone": "Asia/Tokyo",
			"next_run_at": "2017-05-19T13:41:00.000Z",
			"active": true,
			"created_at": "2017-05-19T13:41:00.000Z",
			"updated_at": "2017-05-19T13:41:00.000Z",
			"owner": {
				"name": "Administrator",
				"username": "root",
				"id": 1,
				"state": "active",
				"avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
				"web_url": "https://gitlab.example.com/root"
			},
			"inputs": [
				{
					"name": "deploy_environment",
					"value": "production"
				}
			]
		}
		`)
	})

	schedule, resp, err := client.PipelineSchedules.GetPipelineSchedule(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	testDate := time.Date(2017, time.May, 19, 13, 41, 0, 0, time.UTC)
	want := &PipelineSchedule{
		ID:           13,
		Description:  "Test schedule pipeline",
		Ref:          "refs/heads/main",
		Cron:         "* * * * *",
		CronTimezone: "Asia/Tokyo",
		NextRunAt:    &testDate,
		Active:       true,
		CreatedAt:    &testDate,
		UpdatedAt:    &testDate,
		Owner: &User{
			Name:      "Administrator",
			Username:  "root",
			ID:        1,
			State:     "active",
			AvatarURL: "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
			WebURL:    "https://gitlab.example.com/root",
		},
		Inputs: []*PipelineInput{
			{
				Name:  "deploy_environment",
				Value: "production",
			},
		},
	}
	assert.Equal(t, want, schedule)
}

func TestPipelineSchedules_ListPipelinesTriggeredBySchedule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/2/pipelines", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		[
			{
				"id": 47,
				"iid": 12,
				"project_id": 29,
				"status": "pending",
				"source": "scheduled",
				"ref": "new-pipeline",
				"sha": "a91957a858320c0e17f3a0eca7cfacbff50ea29a",
				"web_url": "https://example.com/foo/bar/pipelines/47",
				"created_at": "2017-05-19T13:41:00.000Z",
				"updated_at": "2017-05-19T13:41:00.000Z"
			},
			{
				"id": 48,
				"iid": 13,
				"project_id": 29,
				"status": "pending",
				"source": "scheduled",
				"ref": "new-pipeline",
				"sha": "eb94b618fb5865b26e80fdd8ae531b7a63ad851a",
				"web_url": "https://example.com/foo/bar/pipelines/48",
				"created_at": "2017-05-19T13:41:00.000Z",
				"updated_at": "2017-05-19T13:41:00.000Z"
			}
		]
		`)
	})

	pipelines, resp, err := client.PipelineSchedules.ListPipelinesTriggeredBySchedule(1, 2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	testDate := time.Date(2017, time.May, 19, 13, 41, 0, 0, time.UTC)
	want := []*Pipeline{
		{
			ID:        47,
			IID:       12,
			ProjectID: 29,
			Status:    "pending",
			Source:    "scheduled",
			Ref:       "new-pipeline",
			SHA:       "a91957a858320c0e17f3a0eca7cfacbff50ea29a",
			WebURL:    "https://example.com/foo/bar/pipelines/47",
			CreatedAt: &testDate,
			UpdatedAt: &testDate,
		},
		{
			ID:        48,
			IID:       13,
			ProjectID: 29,
			Status:    "pending",
			Source:    "scheduled",
			Ref:       "new-pipeline",
			SHA:       "eb94b618fb5865b26e80fdd8ae531b7a63ad851a",
			WebURL:    "https://example.com/foo/bar/pipelines/48",
			CreatedAt: &testDate,
			UpdatedAt: &testDate,
		},
	}
	assert.Equal(t, want, pipelines)
}

func TestPipelineSchedules_CreatePipelineSchedule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
		{
			"id": 13,
			"description": "Test schedule pipeline",
			"ref": "refs/heads/main",
			"cron": "* * * * *",
			"cron_timezone": "Asia/Tokyo",
			"next_run_at": "2017-05-19T13:41:00.000Z",
			"active": true,
			"created_at": "2017-05-19T13:41:00.000Z",
			"updated_at": "2017-05-19T13:41:00.000Z",
			"owner": {
				"name": "Administrator",
				"username": "root",
				"id": 1,
				"state": "active",
				"avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
				"web_url": "https://gitlab.example.com/root"
			}
		}
		`)
	})

	schedule, resp, err := client.PipelineSchedules.CreatePipelineSchedule(1, &CreatePipelineScheduleOptions{
		Description:  Ptr("Test schedule pipeline"),
		Ref:          Ptr("refs/heads/main"),
		Cron:         Ptr("* * * * *"),
		CronTimezone: Ptr("Asia/Tokyo"),
		Active:       Ptr(true),
		Inputs: []*PipelineInput{
			{
				Name:  "my_input_name",
				Value: "my_ci_value",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	testDate := time.Date(2017, time.May, 19, 13, 41, 0, 0, time.UTC)
	want := &PipelineSchedule{
		ID:           13,
		Description:  "Test schedule pipeline",
		Ref:          "refs/heads/main",
		Cron:         "* * * * *",
		CronTimezone: "Asia/Tokyo",
		NextRunAt:    &testDate,
		Active:       true,
		CreatedAt:    &testDate,
		UpdatedAt:    &testDate,
		Owner: &User{
			Name:      "Administrator",
			Username:  "root",
			ID:        1,
			State:     "active",
			AvatarURL: "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
			WebURL:    "https://gitlab.example.com/root",
		},
	}
	assert.Equal(t, want, schedule)
}

func TestPipelineSchedules_CreatePipelineScheduleWithMultipleInputTypes(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		// Validate that inputs support different value types
		testBodyJSON(t, r, map[string]any{
			"description": "Test schedule with various input types",
			"ref":         "refs/heads/main",
			"cron":        "0 1 * * *",
			"inputs": []any{
				map[string]any{
					"name":  "string_input",
					"value": "production",
				},
				map[string]any{
					"name":  "number_input",
					"value": float64(42),
				},
				map[string]any{
					"name":  "boolean_input",
					"value": true,
				},
				map[string]any{
					"name":  "array_input",
					"value": []any{"us-east", "eu-west"},
				},
			},
		})
		fmt.Fprint(w, `{"id": 13}`)
	})

	_, resp, err := client.PipelineSchedules.CreatePipelineSchedule(1, &CreatePipelineScheduleOptions{
		Description: Ptr("Test schedule with various input types"),
		Ref:         Ptr("refs/heads/main"),
		Cron:        Ptr("0 1 * * *"),
		Inputs: []*PipelineInput{
			{Name: "string_input", Value: "production"},
			{Name: "number_input", Value: float64(42)},
			{Name: "boolean_input", Value: true},
			{Name: "array_input", Value: []string{"us-east", "eu-west"}},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestPipelineSchedules_EditPipelineSchedule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
		{
			"id": 13,
			"description": "Test schedule pipeline",
			"ref": "refs/heads/main",
			"cron": "* * * * *",
			"cron_timezone": "Asia/Tokyo",
			"next_run_at": "2017-05-19T13:41:00.000Z",
			"active": true,
			"created_at": "2017-05-19T13:41:00.000Z",
			"updated_at": "2017-05-19T13:41:00.000Z",
			"owner": {
				"name": "Administrator",
				"username": "root",
				"id": 1,
				"state": "active",
				"avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
				"web_url": "https://gitlab.example.com/root"
			}
		}
		`)
	})

	schedule, resp, err := client.PipelineSchedules.EditPipelineSchedule(1, 2, &EditPipelineScheduleOptions{
		Description:  Ptr("Test schedule pipeline"),
		Ref:          Ptr("refs/heads/main"),
		Cron:         Ptr("* * * * *"),
		CronTimezone: Ptr("Asia/Tokyo"),
		Active:       Ptr(true),
		Inputs: []*PipelineInput{
			{
				Name:  "my_input_name",
				Value: "my_ci_value",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	testDate := time.Date(2017, time.May, 19, 13, 41, 0, 0, time.UTC)
	want := &PipelineSchedule{
		ID:           13,
		Description:  "Test schedule pipeline",
		Ref:          "refs/heads/main",
		Cron:         "* * * * *",
		CronTimezone: "Asia/Tokyo",
		NextRunAt:    &testDate,
		Active:       true,
		CreatedAt:    &testDate,
		UpdatedAt:    &testDate,
		Owner: &User{
			Name:      "Administrator",
			Username:  "root",
			ID:        1,
			State:     "active",
			AvatarURL: "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
			WebURL:    "https://gitlab.example.com/root",
		},
	}
	assert.Equal(t, want, schedule)
}

func TestPipelineSchedules_TakeOwnershipOfPipelineSchedule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/2/take_ownership", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
		{
			"id": 13,
			"description": "Test schedule pipeline",
			"ref": "refs/heads/main",
			"cron": "* * * * *",
			"cron_timezone": "Asia/Tokyo",
			"next_run_at": "2017-05-19T13:41:00.000Z",
			"active": true,
			"created_at": "2017-05-19T13:41:00.000Z",
			"updated_at": "2017-05-19T13:41:00.000Z",
			"owner": {
				"name": "Administrator",
				"username": "root",
				"id": 1,
				"state": "active",
				"avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
				"web_url": "https://gitlab.example.com/root"
			}
		}
		`)
	})

	schedule, resp, err := client.PipelineSchedules.TakeOwnershipOfPipelineSchedule(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	testDate := time.Date(2017, time.May, 19, 13, 41, 0, 0, time.UTC)
	want := &PipelineSchedule{
		ID:           13,
		Description:  "Test schedule pipeline",
		Ref:          "refs/heads/main",
		Cron:         "* * * * *",
		CronTimezone: "Asia/Tokyo",
		NextRunAt:    &testDate,
		Active:       true,
		CreatedAt:    &testDate,
		UpdatedAt:    &testDate,
		Owner: &User{
			Name:      "Administrator",
			Username:  "root",
			ID:        1,
			State:     "active",
			AvatarURL: "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon",
			WebURL:    "https://gitlab.example.com/root",
		},
	}
	assert.Equal(t, want, schedule)
}

func TestPipelineSchedules_DeletePipelineSchedule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.PipelineSchedules.DeletePipelineSchedule(1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestPipelineSchedules_RunPipelineSchedule(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/1/play", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusCreated)
	})

	resp, err := client.PipelineSchedules.RunPipelineSchedule(1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestPipelineSchedules_CreatePipelineScheduleVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/2/variables", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
		{
			"key": "NEW_VARIABLE",
			"variable_type": "env_var",
			"value": "new value"
		}
		`)
	})

	variable, resp, err := client.PipelineSchedules.CreatePipelineScheduleVariable(1, 2, &CreatePipelineScheduleVariableOptions{
		Key:   Ptr("NEW_VARIABLE"),
		Value: Ptr("new value"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &PipelineVariable{
		Key:          "NEW_VARIABLE",
		Value:        "new value",
		VariableType: "env_var",
	}
	assert.Equal(t, want, variable)
}

func TestPipelineSchedules_EditPipelineScheduleVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/2/variables/NEW_VARIABLE", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
		{
			"key": "NEW_VARIABLE",
			"variable_type": "env_var",
			"value": "new value"
		}
		`)
	})

	variable, resp, err := client.PipelineSchedules.EditPipelineScheduleVariable(1, 2, "NEW_VARIABLE", &EditPipelineScheduleVariableOptions{
		Value: Ptr("new value"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &PipelineVariable{
		Key:          "NEW_VARIABLE",
		Value:        "new value",
		VariableType: "env_var",
	}
	assert.Equal(t, want, variable)
}

func TestPipelineSchedules_DeletePipelineScheduleVariable(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipeline_schedules/2/variables/NEW_VARIABLE", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `
		{
			"key": "NEW_VARIABLE",
			"variable_type": "env_var",
			"value": "new value"
		}
		`)
	})

	variable, resp, err := client.PipelineSchedules.DeletePipelineScheduleVariable(1, 2, "NEW_VARIABLE")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	want := &PipelineVariable{
		Key:          "NEW_VARIABLE",
		Value:        "new value",
		VariableType: "env_var",
	}
	assert.Equal(t, want, variable)
}
