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
}

func main() {
	p := tea.NewProgram(newRun())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
