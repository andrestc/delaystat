//+build !linux

package main

import (
	"fmt"

	"github.com/mdlayher/taskstats"
)

func PrintStats(stats *taskstats.Stats) {
	fmt.Printf("%#+v", stats)
}
