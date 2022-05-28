package main

import (
	"log"
	"os"

	"github.com/dssolutioninc/dss_gke/accesspubsub/app/handler"
)

func main() {
	log.Println("Application Started.")

	// projectID is identifier of project
	projectID := os.Getenv("PROJECT_ID")

	// pubsubSubscriptionName use to hear the comming request
	pubsubSubscriptionName := os.Getenv("PUBSUB_SUBSCRIPTION_NAME")

	err := handler.SampleHandler{}.StartWaitMessageOn(projectID, pubsubSubscriptionName)
	if err != nil {
		log.Println("Got Error.")
	}

	log.Println("Application Finished.")
}
