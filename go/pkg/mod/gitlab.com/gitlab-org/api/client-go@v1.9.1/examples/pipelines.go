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

// This example demonstrates how to manage pipelines in GitLab using the GitLab API client.
// It covers the following steps:
// 1. Initialize the GitLab client with a personal access token.
// 2. Specify filtering options to list project pipelines, such as scope, status, branch, and date range.
// 3. Retrieve a list of pipelines for a specific project based on the provided filters.
// 4. Iterate through the retrieved pipelines and log their details.

package main

import (
	"log"
	"time"

	"gitlab.com/gitlab-org/api/client-go"
)

func pipelineExample() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	opt := &gitlab.ListProjectPipelinesOptions{
		Scope:         gitlab.Ptr("branches"),
		Status:        gitlab.Ptr(gitlab.Running),
		Ref:           gitlab.Ptr("master"),
		YamlErrors:    gitlab.Ptr(true),
		Name:          gitlab.Ptr("name"),
		Username:      gitlab.Ptr("username"),
		UpdatedAfter:  gitlab.Ptr(time.Now().Add(-24 * 365 * time.Hour)),
		UpdatedBefore: gitlab.Ptr(time.Now().Add(-7 * 24 * time.Hour)),
		OrderBy:       gitlab.Ptr("status"),
		Sort:          gitlab.Ptr("asc"),
	}

	pipelines, _, err := git.Pipelines.ListProjectPipelines(2743054, opt)
	if err != nil {
		log.Fatal(err)
	}

	for _, pipeline := range pipelines {
		log.Printf("Found pipeline: %v", pipeline)
	}
}
