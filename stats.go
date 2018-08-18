package main

import (
	"os"
	"text/template"

	"github.com/mdlayher/taskstats"
)

func PrintStats(stats *taskstats.Stats) error {
	tmpl := `
CPU
	delay total:	{{.CPUDelay}}
	times:	{{.CPUDelayCount}}
I/O
	delay total:	{{.BlockIODelay}}
	times:	{{.BlockIODelayCount}}
Swap
	delay total:	{{.SwapInDelay}}
	times:	{{.SwapInDelayCount}}
FreePages
	delay total:	{{.FreePagesDelay}}
	times:	{{.FreePagesDelayCount}}
`
	t, err := template.New("delay").Parse(tmpl)
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, stats)
}
