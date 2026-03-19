package cli

import (
	"errors"
	"flag"
	"strings"
)

type Input struct {
	Task   string
	Debug  bool
	DryRun bool
}

func ParseInput() (Input, error) {
	taskPtr := flag.String("task", "", "Task to execute")
	debugPtr := flag.Bool("debug", false, "Enable debug output")
	dryRunPtr := flag.Bool("dry-run", false, "Show generated script without executing")

	flag.Parse()

	task := strings.TrimSpace(*taskPtr)
	if task == "" {
		return Input{}, errors.New("you must provide a task with --task")
	}

	return Input{
		Task:   task,
		Debug:  *debugPtr,
		DryRun: *dryRunPtr,
	}, nil
}
