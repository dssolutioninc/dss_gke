package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var jobName = flag.String("job-name", "", "specify job name")
var outputPath = flag.String("output-path", "", "specify output path for job")

func main() {

	flag.Parse()

	outlog("Dummy Job start ...")

	for i := 0; i < 10; i++ {
		unixtime := time.Now().Unix()
		fileName := fmt.Sprint(*jobName, "_", unixtime, ".txt")

		// make a file
		file, err := os.Create(*outputPath + "/" + fileName)
		if err != nil {
			log.Fatal(err)
		}
		outlog("created a file: ", fileName)

		// out some
		file.WriteString("hello from " + *jobName)
		file.Close()

		// sleep to delay process
		time.Sleep(2 * time.Second)
	}

	// list all file in path
	outlog("List all files:")
	files, err := ioutil.ReadDir(*outputPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		outlog(file.Name())
	}

	outlog("Dummy Job finished.")

}

func outlog(args ...string) {
	log.Println(*jobName+":", args)
}
