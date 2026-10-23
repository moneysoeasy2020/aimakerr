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

// This example demonstrates how to download a repository archive in GitLab using the GitLab API client.
// It covers the following steps:
// 1. Initialize the GitLab client with a personal access token.
// 2. Specify the archive options, such as format (e.g., tar.gz) and path.
// 3. Download the repository archive for a specific project and log the size of the downloaded content.

package main

import (
	"log"

	"gitlab.com/gitlab-org/api/client-go"
)

func repositoryArchiveExample() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	// Get repository archive
	opt := &gitlab.ArchiveOptions{
		Format: gitlab.Ptr("tar.gz"),
		Path:   gitlab.Ptr("mydir"),
	}
	content, _, err := git.Repositories.Archive("mygroup/myproject", opt, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Repository archive contains %d byte(s)", len(content))
}
