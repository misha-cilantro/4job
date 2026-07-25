package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

/*
 * Steps:
 * 1. Pick a run type
 * 2. Pick a job set, or skip if run type has noJobSetSelect == true
 * 3. Optionally add excluded jobs (so you can prevent recent repeats)
 * 4. Options: allowDuplicates (locked if forcesAllowDuplicates == true); allow rolling special jobs (Freelancer and Mime) (locked if allowSpecial == true))
 * 5. Confirm run; optionally name folder for run, otherwise named after the run type + date
 */

type run struct {
	step             int
	cursor           int // current highlighted item within the active step
	runType          string
	jobSetLocked     bool
	jobSet           string
	excludes         []string
	allowDuplicates  bool
	duplicatesLocked bool // true once a run type or job set forces Allow Duplicates on
	allowSpecial     bool
	specialLocked    bool // true once the run type forces Allow Special Jobs on
	name             string
}

func main() {
	p := tea.NewProgram(newRun())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(run); ok {
		fmt.Printf("Picking jobs...\n\n")
		jobs := pickJobs(m)

		fmt.Printf("Writing run folder...\n\n")
		writeFolder(m, jobs)

		fmt.Printf("Done! Your run is in ./%s", m.name)
	}
}
