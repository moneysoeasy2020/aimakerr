//
// Copyright 2021, Timo Furrer <tuxtimo@gmail.com>
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

// This example demonstrates how to manage topics in GitLab using the GitLab API client.
// It covers the following steps:
// 1. Initialize the GitLab client with a personal access token.
// 2. Create a new topic by specifying its name and description.
// 3. Set an avatar for the topic by uploading an image file.
// 4. Remove the avatar from the topic.
// 5. Log the details of the topic at each step to verify the changes.

package main

import (
	"log"
	"os"

	"gitlab.com/gitlab-org/api/client-go"
)

func topicExample() {
	git, err := gitlab.NewClient("yourtokengoeshere")
	if err != nil {
		log.Fatal(err)
	}

	// New topic
	topic, _, err := git.Topics.CreateTopic(&gitlab.CreateTopicOptions{
		Name:        gitlab.Ptr("My Topic 2"),
		Description: gitlab.Ptr("Some description"),
	})
	if err != nil {
		panic(err)
	}

	log.Printf("Topic: %+v\n", topic)

	// Set topic avatar
	avatarFile, err := os.Open("5746961_detect_direction_gps_location_map_icon.png")
	if err != nil {
		panic(err)
	}
	topic, _, err = git.Topics.UpdateTopic(topic.ID, &gitlab.UpdateTopicOptions{
		Avatar: &gitlab.TopicAvatar{
			Filename: "5746961_detect_direction_gps_location_map_icon.png",
			Image:    avatarFile,
		},
	})
	if err != nil {
		panic(err)
	}
	log.Printf("Topic with Avatar: %+v\n", topic)

	// Remove topic avatar
	topic, _, err = git.Topics.UpdateTopic(topic.ID, &gitlab.UpdateTopicOptions{
		Avatar: &gitlab.TopicAvatar{},
	})
	if err != nil {
		panic(err)
	}

	log.Printf("Topic without Avatar: %+v\n", topic)
}
