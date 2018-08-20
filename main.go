package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

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
		fmt.Printf("pid %v\n", pid)
		stats, err = client.PID(*pid)
	} else {
		fmt.Printf("tgid %v\n", *tgid)
		stats, err = client.TGID(*tgid)
	}
	if err != nil {
		log.Panic(err)
	}

	if err := PrintStats(stats, nil); err != nil {
		log.Panic(err)
	}
}

func PrintStats(stats *taskstats.Stats, lastStats *taskstats.Stats) error {
	diffStats := *stats
	if lastStats != nil {
		diffStats.CPUDelay -= lastStats.CPUDelay
		diffStats.CPUDelayCount -= lastStats.CPUDelayCount
		diffStats.BlockIODelay -= lastStats.BlockIODelay
		diffStats.BlockIODelayCount -= lastStats.BlockIODelayCount
		diffStats.FreePagesDelay -= lastStats.FreePagesDelay
		diffStats.FreePagesDelayCount -= lastStats.FreePagesDelayCount
		diffStats.SwapInDelay -= lastStats.SwapInDelay
		diffStats.SwapInDelayCount -= lastStats.SwapInDelayCount
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "CPU\tI/O\tSwap\tMemory Reclaim")
	fmt.Fprintf(w, "%v\t%v\t%v\t%v\n",
		diffStats.CPUDelay,
		diffStats.BlockIODelay,
		diffStats.SwapInDelay,
		diffStats.FreePagesDelay,
	)
	return w.Flush()
}
