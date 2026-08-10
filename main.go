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
 * 4. Options: allowDuplicates (locked on if the run type or job set has
 *    forcesAllowDuplicates). Special jobs (Freelancer and Mime) are not an
 *    option - only run types declaring allowSpecial can roll them.
 * 5. Confirm run; folder named after the run type + date
 */

// run is the whole wizard state. Whether a modifier is locked is derived
// from the chosen run type and job set rather than stored, so backing up and
// choosing again can't leave a stale lock behind - see duplicatesLocked and
// friends in model.go.
type run struct {
	step            int
	cursor          int // current highlighted item within the active step
	runType         string
	jobSet          string
	excludes        []string
	allowDuplicates bool // the user's choice; a run type or job set may force it on
	name            string

	// Terminal height, from the last tea.WindowSizeMsg. Zero until the first
	// one arrives; the list views fall back to a default. Width isn't tracked
	// because nothing wraps to it yet.
	height int

	// confirmed is set only when the user accepts the summary. Quitting at
	// any point leaves it false and writes nothing.
	confirmed bool
}

func main() {
	p := tea.NewProgram(newRun())
	finalModel, err := p.Run()
	if err != nil {
		fail(err)
	}

	m, ok := finalModel.(run)
	if !ok || !m.confirmed {
		fmt.Println("Cancelled - no run folder written.")
		return
	}

	fmt.Printf("Picking jobs...\n\n")
	res, err := pickJobs(m)
	if err != nil {
		fail(err)
	}

	for i, job := range res.Jobs {
		fmt.Printf("  %d. %s\n", i+1, job)
	}
	fmt.Println()

	for _, note := range res.Notes {
		fmt.Printf("Note: %s\n", note)
	}
	if len(res.Notes) > 0 {
		fmt.Println()
	}

	fmt.Printf("Writing run folder...\n\n")
	if err := writeFolder(m, res.Jobs); err != nil {
		fail(err)
	}

	fmt.Printf("Done! Your run is in ./%s\n", m.name)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
