package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	errorRate = 30
)

func main() {
	fmt.Printf("\nDummy Job start ...")

	var apiEndPoint, jobID string
	flag.StringVar(&apiEndPoint, "api-end-point", "", "update job process status url")
	flag.StringVar(&jobID, "job-id", "", "job id")
	flag.Parse()
	fmt.Printf("\napi-end-point : %s", apiEndPoint)
	fmt.Printf("\njob-id : %s", jobID)

	http.PostForm(apiEndPoint, url.Values{"job-id": {jobID}, "status": {"Start Processing"}})

	time.Sleep(2 * time.Second)
	fmt.Printf("\nDummy Job running ...")

	// update process status
	http.PostForm(apiEndPoint, url.Values{"job-id": {jobID}, "status": {"Processing..."}})

	time.Sleep(15 * time.Second)

	// raise a random error
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(100) < errorRate {
		fmt.Printf("\nJob ERROR!")

		// update process status
		http.PostForm(apiEndPoint, url.Values{"job-id": {jobID}, "status": {"ERROR"}})

		os.Exit(1)
	}

	// update process status
	http.PostForm(apiEndPoint, url.Values{"job-id": {jobID}, "status": {"Done"}})

	fmt.Printf("\nDummy Job finished.\n")
}
