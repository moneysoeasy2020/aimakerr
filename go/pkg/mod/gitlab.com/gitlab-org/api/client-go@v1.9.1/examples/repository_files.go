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

// This example demonstrates how to manage repository files in GitLab using the GitLab API client.
// It covers the following steps:
// 1. Initialize the GitLab client with a personal access token.
// 2. Create a new file in a repository by specifying its branch, content, and commit message.
// 3. Update an existing file in the repository with new content and a commit message.
// 4. Retrieve the contents of a file from the repository for a specific branch.
// 5. Retrieve the blame information for a file, including the number of blame ranges.

package main

import (
	"log"

	"gitlab.com/gitlab-org/api/client-go"
)

func repositoryFileExample() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	// Create a new repository file
	cf := &gitlab.CreateFileOptions{
		Branch:        gitlab.Ptr("master"),
		Content:       gitlab.Ptr("My file contents"),
		CommitMessage: gitlab.Ptr("Adding a test file"),
	}
	file, _, err := git.RepositoryFiles.CreateFile("myname/myproject", "file.go", cf)
	if err != nil {
		log.Fatal(err)
	}

	// Update a repository file
	uf := &gitlab.UpdateFileOptions{
		Branch:        gitlab.Ptr("master"),
		Content:       gitlab.Ptr("My file content"),
		CommitMessage: gitlab.Ptr("Fixing typo"),
	}
	_, _, err = git.RepositoryFiles.UpdateFile("myname/myproject", file.FilePath, uf)
	if err != nil {
		log.Fatal(err)
	}

	gf := &gitlab.GetFileOptions{
		Ref: gitlab.Ptr("master"),
	}
	f, _, err := git.RepositoryFiles.GetFile("myname/myproject", file.FilePath, gf)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("File contains: %s", f.Content)

	gfb := &gitlab.GetFileBlameOptions{
		Ref: gitlab.Ptr("master"),
	}
	fb, _, err := git.RepositoryFiles.GetFileBlame("myname/myproject", file.FilePath, gfb)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d blame ranges", len(fb))
}
