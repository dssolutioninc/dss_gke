package handler

import (
	"context"
	"flag"
	"log"
	"os/exec"
)

type SimpleJobHandler struct {
}

func (j SimpleJobHandler) Run(ctx context.Context) error {
	log.Println("Processing ...")

	runTime := getArguments()
	cmd := exec.Command("sleep", runTime)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error at command: %v", cmd)
		return err
	}

	log.Println("Process completed.")
	return nil
}

func getArguments() string {
	runTime := flag.String("run-time", "", "specify number of seconds to run job")
	flag.Parse()
	return *runTime
}
