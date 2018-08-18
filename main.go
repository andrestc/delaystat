package main

import (
	"flag"
	"log"
	"os"

	"github.com/mdlayher/taskstats"
)

func main() {
	var pid, tgid *int
	pid = flag.Int("p", -1, "PID to track delay stats")
	tgid = flag.Int("t", os.Getpid(), "TGID to track delay stats")
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

	var stats *taskstats.Stats
	if *pid != -1 {
		stats, err = client.PID(*pid)
	} else {
		stats, err = client.TGID(*tgid)
	}
	if err != nil {
		log.Panic(err)
	}

	err = PrintStats(stats)
	if err != nil {
		log.Panic(err)
	}
}
