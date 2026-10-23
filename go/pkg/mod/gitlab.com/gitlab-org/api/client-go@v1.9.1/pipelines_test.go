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
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListProjectPipelines(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteJSONResponse(t, w, []map[string]any{
			{"id": 1, "name": "test"},
			{"id": 2},
			{"id": 3, "status": "success", "ref": "master", "created_at": "2023-05-02T12:00:00Z"},
		})
	})

	opt := &ListProjectPipelinesOptions{Ref: Ptr("master"), CreatedAfter: mustParseTime("2023-05-01T15:00:00Z")}
	pipelines, _, err := client.Pipelines.ListProjectPipelines(1, opt)
	if err != nil {
		t.Errorf("Pipelines.ListProjectPipelines returned error: %v", err)
	}

	want := []*PipelineInfo{{ID: 1, Name: "test"}, {ID: 2}, {ID: 3, Status: "success", Ref: "master", CreatedAt: mustParseTime("2023-05-02T12:00:00Z")}}
	if !reflect.DeepEqual(want, pipelines) {
		t.Errorf("Pipelines.ListProjectPipelines returned %+v, want %+v", pipelines, want)
	}
}

func TestGetPipeline(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteJSONResponse(t, w, map[string]any{"id": 1, "status": "success", "source": "api"})
	})

	pipeline, _, err := client.Pipelines.GetPipeline(1, 5949167)
	if err != nil {
		t.Errorf("Pipelines.GetPipeline returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "success", Source: "api"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("Pipelines.GetPipeline returned %+v, want %+v", pipeline, want)
	}
}

func TestGetPipelineVariables(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167/variables", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteJSONResponse(t, w, []*PipelineVariable{
			{Key: "RUN_NIGHTLY_BUILD", Value: "true", VariableType: "env_var"},
			{Key: "foo", Value: "bar"},
		})
	})

	variables, _, err := client.Pipelines.GetPipelineVariables(1, 5949167)
	if err != nil {
		t.Errorf("Pipelines.GetPipelineVariables returned error: %v", err)
	}

	want := []*PipelineVariable{{Key: "RUN_NIGHTLY_BUILD", Value: "true", VariableType: "env_var"}, {Key: "foo", Value: "bar"}}
	if !reflect.DeepEqual(want, variables) {
		t.Errorf("Pipelines.GetPipelineVariables returned %+v, want %+v", variables, want)
	}
}

func TestGetPipelineTestReport(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/123456/test_report", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/get_pipeline_testreport.json")
	})

	testreport, _, err := client.Pipelines.GetPipelineTestReport(1, 123456)
	if err != nil {
		t.Errorf("Pipelines.GetPipelineTestReport returned error: %v", err)
	}

	want := &PipelineTestReport{
		TotalTime:    61.502,
		TotalCount:   9,
		SuccessCount: 5,
		ErrorCount:   4,
		TestSuites: []*PipelineTestSuites{
			{
				Name:         "Failing",
				TotalTime:    60.494,
				TotalCount:   8,
				SuccessCount: 4,
				ErrorCount:   4,
				TestCases: []*PipelineTestCases{
					{
						Status:        "error",
						Name:          "Error testcase 1",
						Classname:     "MyClassOne",
						File:          "/path/file.ext",
						ExecutionTime: 19.987,
						SystemOutput:  "Failed test",
						StackTrace:    "java.lang.Exception: Stack trace\nat java.base/java.lang.Thread.dumpStack(Thread.java:1383)",
						AttachmentURL: "http://foo.bar",
						RecentFailures: &RecentFailures{
							Count:      10,
							BaseBranch: "master",
						},
					},
					{
						Status:        "error",
						Name:          "Error testcase 2",
						Classname:     "MyClass",
						ExecutionTime: 19.984,
						SystemOutput: map[string]any{
							"message": "Failed test",
							"type":    "MultipleExceptionError",
						},
					},
					{
						Status:    "error",
						Name:      "Error testcase 3",
						Classname: "MyClass",
						SystemOutput: []any{
							"Failed test a",
							"Failed test b",
						},
					},
					{
						Status:        "success",
						Name:          "Success full testcase",
						Classname:     "MyClass",
						ExecutionTime: 19.7799999999999985,
					},
				},
			},
			{
				Name:         "Success suite",
				TotalTime:    1.008,
				TotalCount:   1,
				SuccessCount: 1,
				TestCases: []*PipelineTestCases{{
					Status:        "success",
					Name:          "Successful testcase",
					Classname:     "MyClass",
					ExecutionTime: 1.008,
				}},
			},
		},
	}
	if !reflect.DeepEqual(want, testreport) {
		t.Errorf("Pipelines.GetPipelineTestReport returned %+v, want %+v", testreport, want)
	}
}

func TestGetPipelineTestReportSummary(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/123456/test_report_summary", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/get_pipeline_testreport_summary.json")
	})

	testreport, _, err := client.Pipelines.GetPipelineTestReportSummary(1, 123456)
	if err != nil {
		t.Errorf("Pipelines.GetPipelineTestReportSummary returned error: %v", err)
	}

	want := &PipelineTestReportSummary{
		Total: PipelineTotalSummary{
			Time:       1904,
			Count:      3363,
			Success:    3351,
			Failed:     0,
			Skipped:    12,
			Error:      0,
			SuiteError: Ptr("JUnit XML parsing failed: 1:1: FATAL: Document is empty"),
		},
		TestSuites: []PipelineTestSuiteSummary{
			{
				Name:         "test",
				TotalTime:    1904,
				TotalCount:   3363,
				SuccessCount: 3351,
				FailedCount:  0,
				SkippedCount: 12,
				ErrorCount:   0,
				BuildIDs:     []int64{66004},
				SuiteError:   Ptr("JUnit XML parsing failed: 1:1: FATAL: Document is empty"),
			},
		},
	}

	if !reflect.DeepEqual(want, testreport) {
		t.Errorf("Pipelines.GetPipelineTestReportSummary returned %+v, want %+v", testreport, want)
	}
}

func TestGetLatestPipeline(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/latest", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Empty(t, r.URL.RawQuery)
		mustWriteJSONResponse(t, w, map[string]any{"id": 1, "status": "success"})
	})

	pipeline, _, err := client.Pipelines.GetLatestPipeline(1, nil)

	assert.NoError(t, err)
	assert.Equal(t, &Pipeline{ID: 1, Status: "success"}, pipeline)
}

func TestGetLatestPipeline_WithRef(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/latest", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testParam(t, r, "ref", "abc")
		mustWriteJSONResponse(t, w, map[string]any{"id": 1, "status": "success"})
	})

	pipeline, _, err := client.Pipelines.GetLatestPipeline(1, &GetLatestPipelineOptions{
		Ref: Ptr("abc"),
	})

	assert.NoError(t, err)
	assert.Equal(t, &Pipeline{ID: 1, Status: "success"}, pipeline)
}

// Assert that PipelineInputValue[T] implements the PipelineInputValueInterface.
var _ PipelineInputValueInterface = PipelineInputValue[string]{}

func TestCreatePipeline(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		opt     *CreatePipelineOptions
		want    map[string]any
		wantErr error
	}{
		{
			name: "base",
			opt:  &CreatePipelineOptions{Ref: Ptr("main")},
			want: map[string]any{
				"ref": "main",
			},
		},
		{
			name: "with variables",
			opt: &CreatePipelineOptions{
				Ref: Ptr("main"),
				Variables: Ptr([]*PipelineVariableOptions{
					{
						Key:          Ptr("UPLOAD_TO_S3"),
						Value:        Ptr("true"),
						VariableType: Ptr(FileVariableType),
					},
					{
						Key:          Ptr("TEST"),
						Value:        Ptr("test variable"),
						VariableType: Ptr(EnvVariableType),
					},
				}),
			},
			want: map[string]any{
				"ref": "main",
				"variables": []any{
					map[string]any{
						"key":           "UPLOAD_TO_S3",
						"value":         "true",
						"variable_type": "file",
					},
					map[string]any{
						"key":           "TEST",
						"value":         "test variable",
						"variable_type": "env_var",
					},
				},
			},
		},
		{
			name: "with inputs",
			opt: &CreatePipelineOptions{
				Ref: Ptr("main"),
				Inputs: map[string]PipelineInputValueInterface{
					"string_option":  NewPipelineInputValue("foo"),
					"integer_option": NewPipelineInputValue(42),
					"boolean_option": NewPipelineInputValue(true),
					"array_option":   NewPipelineInputValue([]string{"bar", "qux"}),
				},
			},
			want: map[string]any{
				"ref": "main",
				"inputs": map[string]any{
					"string_option":  "foo",
					"integer_option": float64(42),
					"boolean_option": true,
					"array_option":   []any{"bar", "qux"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mux, client := setup(t)

			mux.HandleFunc("/api/v4/projects/1/pipeline", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodPost)
				testBodyJSON(t, r, tc.want)
				mustWriteJSONResponse(t, w, map[string]any{"id": 1, "status": "pending"})
			})

			pipeline, _, err := client.Pipelines.CreatePipeline(1, tc.opt)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Pipelines.CreatePipeline() = error %v, want error %v", err, tc.wantErr)
			}

			if err != nil {
				return
			}

			want := &Pipeline{ID: 1, Status: "pending"}
			if !reflect.DeepEqual(want, pipeline) {
				t.Errorf("Pipelines.CreatePipeline returned %+v, want %+v", pipeline, want)
			}
		})
	}
}

func TestRetryPipelineBuild(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167/retry", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		mustWriteJSONResponse(t, w, map[string]any{"id": 1, "status": "pending"})
	})

	pipeline, _, err := client.Pipelines.RetryPipelineBuild(1, 5949167)
	if err != nil {
		t.Errorf("Pipelines.RetryPipelineBuild returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "pending"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("Pipelines.RetryPipelineBuild returned %+v, want %+v", pipeline, want)
	}
}

func TestCancelPipelineBuild(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167/cancel", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		mustWriteJSONResponse(t, w, map[string]any{"id": 1, "status": "canceled"})
	})

	pipeline, _, err := client.Pipelines.CancelPipelineBuild(1, 5949167)
	if err != nil {
		t.Errorf("Pipelines.CancelPipelineBuild returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "canceled"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("Pipelines.CancelPipelineBuild returned %+v, want %+v", pipeline, want)
	}
}

func TestDeletePipeline(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Pipelines.DeletePipeline("1", 5949167)
	if err != nil {
		t.Errorf("Pipelines.DeletePipeline returned error: %v", err)
	}
}

func TestUpdateMetadata(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/pipelines/234/metadata", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		mustWriteJSONResponse(t, w, map[string]any{"id": 1, "status": "running"})
	})

	opt := &UpdatePipelineMetadataOptions{Name: Ptr("new pipeline title")}
	pipeline, _, err := client.Pipelines.UpdatePipelineMetadata("1", 234, opt)
	if err != nil {
		t.Errorf("Pipelines.UpdatePipelineMetadata returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "running"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("Pipelines.UpdatePipelineMetadata returned %+v, want %+v", pipeline, want)
	}
}
