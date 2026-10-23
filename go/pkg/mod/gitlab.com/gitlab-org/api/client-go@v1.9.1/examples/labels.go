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

// This example demonstrates how to manage labels in GitLab using the GitLab API client.
// It covers the following steps:
// 1. Initialize the GitLab client with a personal access token.
// 2. Create a new label for a specific project by specifying its name and color.
// 3. List all labels associated with the specified project and log their details.

package main

import (
	"log"

	"gitlab.com/gitlab-org/api/client-go"
)

func labelExample() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	// Create new label
	opt := &gitlab.CreateLabelOptions{
		Name:  gitlab.Ptr("My Label"),
		Color: gitlab.Ptr("#11FF22"),
	}
	label, _, err := git.Labels.CreateLabel("myname/myproject", opt)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created label: %s\nWith color: %s\n", label.Name, label.Color)

	// List all labels
	labels, _, err := git.Labels.ListLabels("myname/myproject", nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, label := range labels {
		log.Printf("Found label: %s", label.Name)
	}
}
