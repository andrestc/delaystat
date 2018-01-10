package main

import (
	"flag"
	"log"
	"os"

	"github.com/mdlayher/taskstats"
)

func main() {
	var pid *int
	pid = flag.Int("p", os.Getpid(), "Process ID to track delay stats")
	flag.Parse()

	client, err := taskstats.New()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	stats, err := client.PID(*pid)
	if err != nil {
		log.Panic(err)
	}
	PrintStats(stats)
}
