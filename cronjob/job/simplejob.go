package main

import (
	"context"
	"log"

	"github.com/dssolutioninc/dss_gke/cronjob/job/handler"
)

func main() {
	log.Println("Job Started.")

	ctx := context.Background()
	handler.SimpleJobHandler{}.Run(ctx)

	log.Println("Job Finished.")
}
