//go:build go1.23
// +build go1.23

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

package main

import (
	"log"

	"gitlab.com/gitlab-org/api/client-go"
)

func pagination() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
		Owned: gitlab.Ptr(true),
	}

	for project := range gitlab.Must(gitlab.Scan2(func(p gitlab.PaginationOptionFunc) ([]*gitlab.Project, *gitlab.Response, error) {
		return git.Projects.ListProjects(opt, p)
	})) {
		log.Printf("Found project: %s", project.Name)
	}
}

func keysetPagination() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			OrderBy:    "id",
			Pagination: "keyset",
			PerPage:    5,
			Sort:       "asc",
		},
		Owned: gitlab.Ptr(true),
	}

	for project := range gitlab.Must(gitlab.Scan2(func(p gitlab.PaginationOptionFunc) ([]*gitlab.Project, *gitlab.Response, error) {
		return git.Projects.ListProjects(opt, p)
	})) {
		log.Printf("Found project: %s", project.Name)
	}
}
