package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/mdlayher/taskstats"
)

func main() {
	var pid, tgid *int
	var interval *time.Duration
	pid = flag.Int("p", -1, "PID to track delay stats")
	tgid = flag.Int("t", os.Getpid(), "TGID to track delay stats")
	interval = flag.Duration("i", time.Duration(0), "Interval between collection")
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

	var stats, prevStats *taskstats.Stats
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "CPU\tI/O\tSwap\tMemory Reclaim")
	if *pid != -1 {
		fmt.Printf("PID [%v]\n", *pid)
	} else {
		fmt.Printf("TGID [%v]\n", *tgid)
	}
	for {
		if *pid != -1 {
			stats, err = client.PID(*pid)
		} else {
			stats, err = client.TGID(*tgid)
		}
		if err != nil {
			log.Panic(err)
		}

		PrintStats(w, stats, prevStats)
		if err := w.Flush(); err != nil {
			log.Panic(err)
		}
		prevStats = stats

		if *interval == time.Duration(0) {
			return
		}
		time.Sleep(*interval)
	}

}

func PrintStats(w io.Writer, stats *taskstats.Stats, lastStats *taskstats.Stats) {
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
	fmt.Fprintf(w, "%v (avg %v)\t%v (avg %v)\t%v (avg %v)\t%v (avg %v)\n",
		diffStats.CPUDelay,
		avgDuration(stats.CPUDelay.Nanoseconds(), int64(stats.CPUDelayCount)),
		diffStats.BlockIODelay,
		avgDuration(stats.BlockIODelay.Nanoseconds(), int64(stats.BlockIODelayCount)),
		diffStats.SwapInDelay,
		avgDuration(stats.SwapInDelay.Nanoseconds(), int64(stats.SwapInDelayCount)),
		diffStats.FreePagesDelay,
		avgDuration(stats.FreePagesDelay.Nanoseconds(), int64(stats.FreePagesDelayCount)),
	)
}

func avgDuration(total, count int64) time.Duration {
	if count == 0 {
		return time.Duration(0)
	}
	return time.Duration(total / count)
}
