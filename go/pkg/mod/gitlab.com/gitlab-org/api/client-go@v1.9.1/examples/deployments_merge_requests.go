//
// Copyright 2022, Daniela Filipe Bento
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

// This example demonstrates how to list merge requests associated with a deployment in GitLab using the GitLab API client.
// It covers the following steps:
// 1. Initialize the GitLab client with a personal access token.
// 2. Use the `ListDeploymentMergeRequests` method to retrieve merge requests for a specific deployment and environment.
// 3. Iterate through the retrieved merge requests and log their details.

package main

import (
	"log"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func deploymentExample() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	opt := &gitlab.ListMergeRequestsOptions{}
	mergeRequests, _, err := git.DeploymentMergeRequests.ListDeploymentMergeRequests(1, 1, opt)
	if err != nil {
		log.Fatal(err)
	}

	for _, mergeRequest := range mergeRequests {
		log.Printf("Found merge request: %v\n", mergeRequest)
	}
}
