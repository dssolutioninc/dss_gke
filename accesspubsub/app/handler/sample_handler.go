package handler

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

type SampleHandler struct {
}

// StartWaitMessageOn
// projectID := "my-project-id"
// subName := projectID + "-example-sub"
func (h SampleHandler) StartWaitMessageOn(projectID, subName string) error {
	log.Println(fmt.Sprintf("StartWaitMessageOn [Project: %s, Subscription Name: %s]", projectID, subName))

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	sub := client.Subscription(subName)
	err = sub.Receive(ctx, processMessage)
	if err != nil {
		return err
	}

	return nil
}

// processMessage implement callback function to process received message data
var processMessage = func(ctx context.Context, m *pubsub.Message) {
	log.Println(fmt.Sprintf("Message ID: %s\n", m.ID))
	log.Println(fmt.Sprintf("Message Time: %s\n", m.PublishTime.String()))

	log.Println(fmt.Sprintf("Message Attributes:\n %v\n", m.Attributes))

	log.Println(fmt.Sprintf("Message Data:\n %s\n", m.Data))

	m.Ack()
}
