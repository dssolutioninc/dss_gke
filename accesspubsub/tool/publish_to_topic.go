package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

var (
	topicID   = flag.String("topic-id", "sample-topic", "Specify topic to publish message")
	projectID = flag.String("project-id", "sample-project", "Specify GCP project you want to work on")
)

func main() {
	flag.Parse()

	err := publishMsg(*projectID, *topicID,
		map[string]string{
			"user":    "Hashimoto",
			"message": "more than happy",
			"status":  "bonus day!",
		},
		nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

func publishMsg(projectID, topicID string, attr map[string]string, msg map[string]string) error {
	// projectID := "my-project-id"
	// topicID := "my-topic"
	// msg := message data publish to topic
	// attr := attribute of message
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}

	bMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Input msg error : %v", err)
	}

	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data:       bMsg,
		Attributes: attr,
	})

	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("Get: %v", err)
	}
	fmt.Printf("Published message with custom attributes; msg ID: %v\n", id)

	return nil
}
