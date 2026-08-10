package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// The wizard is a Bubble Tea program, so it can't be driven through stdin
// without a real terminal. Feeding messages to Update directly exercises the
// same wiring: key routing, step transitions and the confirmation flag.

func special(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func char(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }

// press feeds messages to the model the way the tea runtime would, and returns
// the final model plus whether any of them asked the program to quit.
func press(t *testing.T, m run, msgs ...tea.Msg) (run, bool) {
	t.Helper()

	quit := false
	for i, msg := range msgs {
		next, cmd := m.Update(msg)
		got, ok := next.(run)
		if !ok {
			t.Fatalf("message %d: Update returned %T, want run", i, next)
		}
		m = got
		if cmd != nil {
			// tea.Quit is the only command the wizard issues.
			quit = true
		}
	}
	return m, quit
}

// indexOfRunType and indexOfJobSet turn a name into the number of "down"
// presses needed to reach it, so the tests don't depend on data file ordering.
func indexOfRunType(t *testing.T, name string) int {
	t.Helper()
	i := slices.IndexFunc(RunTypes, func(rt RunType) bool { return rt.Name == name })
	if i < 0 {
		t.Fatalf("no run type named %q", name)
	}
	return i
}

func indexOfJobSet(t *testing.T, name string) int {
	t.Helper()
	i := slices.IndexFunc(JobSets, func(js JobSet) bool { return js.Name == name })
	if i < 0 {
		t.Fatalf("no job set named %q", name)
	}
	return i + 1 // index 0 in the list is "(none)"
}

func downTo(index int) []tea.Msg {
	msgs := make([]tea.Msg, 0, index)
	for range index {
		msgs = append(msgs, special(tea.KeyDown))
	}
	return msgs
}

func TestWizardFullFlowWritesARun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := newRun()
	m, _ = press(t, m, tea.WindowSizeMsg{Width: 80, Height: 30})

	// Step 1: Normal.
	m, _ = press(t, m, downTo(indexOfRunType(t, "Normal"))...)
	m, _ = press(t, m, special(tea.KeyEnter))
	if m.step != stepJobSet {
		t.Fatalf("after picking a run type, step is %d, want stepJobSet", m.step)
	}

	// Step 2: Team 750.
	m, _ = press(t, m, downTo(indexOfJobSet(t, "Team 750"))...)
	m, _ = press(t, m, special(tea.KeyEnter))
	if m.jobSet != "Team 750" {
		t.Fatalf("job set is %q, want Team 750", m.jobSet)
	}

	// Step 3: exclude whichever job the cursor starts on.
	firstJob := AllJobNames()[0]
	m, _ = press(t, m, special(tea.KeySpace), special(tea.KeyEnter))
	if !slices.Equal(m.excludes, []string{firstJob}) {
		t.Fatalf("excludes is %v, want [%s]", m.excludes, firstJob)
	}

	// Step 4: accept the options as they are.
	m, _ = press(t, m, special(tea.KeyEnter))
	if m.step != stepSummary {
		t.Fatalf("after options, step is %d, want stepSummary", m.step)
	}
	if m.name == "" {
		t.Fatal("no run folder name was generated")
	}

	// Step 5: confirm.
	m, quit := press(t, m, special(tea.KeyEnter))
	if !quit {
		t.Error("confirming should quit the program")
	}
	if !m.confirmed {
		t.Fatal("confirmed should be set after accepting the summary")
	}

	// Then the part main() does.
	res, err := pickJobs(m)
	if err != nil {
		t.Fatalf("pickJobs: %v", err)
	}
	if err := writeFolder(m, res.Jobs); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	allowed := JobPoolsByName["750"].Jobs
	for i, filename := range fileOrder {
		got, err := os.ReadFile(filepath.Join(dir, m.name, filename+".txt"))
		if err != nil {
			t.Fatalf("reading %s: %v", filename, err)
		}
		job := string(got)
		if job != res.Jobs[i] {
			t.Errorf("%s holds %q, want %q", filename, job, res.Jobs[i])
		}
		if !slices.Contains(allowed, job) {
			t.Errorf("%s holds %q, which is not a Team 750 job", filename, job)
		}
		if job == firstJob {
			t.Errorf("%s holds %q, which was excluded", filename, job)
		}
	}
}

func TestWizardQuitDoesNotConfirm(t *testing.T) {
	// Quitting at each step used to still write a folder, and at the earlier
	// steps that meant a fatal error because no name existed yet.
	for _, step := range []int{stepRunType, stepJobSet, stepExcludes, stepOptions, stepSummary} {
		m, quit := press(t, run{step: step, runType: "Normal"}, char('q'))
		if !quit {
			t.Errorf("step %d: q should quit", step)
		}
		if m.confirmed {
			t.Errorf("step %d: q should not confirm the run", step)
		}
	}
}

// TestWizardBackClearsJobSetLock is the regression test for the stale lock:
// choosing Team 375 forced Allow Duplicates on, and backing out to pick a
// different job set used to leave it forced.
func TestWizardBackClearsJobSetLock(t *testing.T) {
	m := newRun()
	m, _ = press(t, m, downTo(indexOfRunType(t, "Normal"))...)
	m, _ = press(t, m, special(tea.KeyEnter))

	m, _ = press(t, m, downTo(indexOfJobSet(t, "Team 375"))...)
	m, _ = press(t, m, special(tea.KeyEnter))
	if !m.duplicatesLocked() {
		t.Fatal("Team 375 should lock Allow Duplicates on")
	}

	// Back to the job set step, then choose (none) at cursor 0.
	m, _ = press(t, m, special(tea.KeyEscape))
	if m.step != stepJobSet {
		t.Fatalf("esc from excludes went to step %d, want stepJobSet", m.step)
	}
	m, _ = press(t, m, special(tea.KeyEnter))

	if m.jobSet != "" {
		t.Fatalf("job set is %q, want none", m.jobSet)
	}
	if m.duplicatesLocked() {
		t.Error("Allow Duplicates is still locked after dropping Team 375")
	}
	if m.duplicatesAllowed() {
		t.Error("Allow Duplicates is still on after dropping Team 375")
	}
}

func TestWizardSkipsJobSetStepWhenLocked(t *testing.T) {
	m := newRun()
	m, _ = press(t, m, downTo(indexOfRunType(t, "Classic"))...)
	m, _ = press(t, m, special(tea.KeyEnter))

	if m.step != stepExcludes {
		t.Fatalf("Classic went to step %d, want stepExcludes", m.step)
	}
	if !m.duplicatesLocked() {
		t.Error("Classic should force Allow Duplicates on")
	}

	// esc should skip back over the job set step too, not land on it.
	m, _ = press(t, m, special(tea.KeyEscape))
	if m.step != stepRunType {
		t.Errorf("esc from excludes went to step %d, want stepRunType", m.step)
	}
}

// TestWizardCannotToggleSpecialJobs checks the options step offers no way to
// turn special jobs on. It used to be a togglable row, which let a Normal run
// roll Freelancer or Mime.
func TestWizardCannotToggleSpecialJobs(t *testing.T) {
	m := newRun()
	m, _ = press(t, m, downTo(indexOfRunType(t, "Normal"))...)
	m, _ = press(t, m, special(tea.KeyEnter)) // run type
	m, _ = press(t, m, special(tea.KeyEnter)) // job set: (none)
	m, _ = press(t, m, special(tea.KeyEnter)) // excludes: none
	if m.step != stepOptions {
		t.Fatalf("step is %d, want stepOptions", m.step)
	}

	// Toggle every reachable row, from every cursor position, several times.
	for range 3 {
		for cursor := 0; cursor < optCount+2; cursor++ {
			m, _ = press(t, m, special(tea.KeySpace), special(tea.KeyDown))
		}
	}

	if m.specialAllowed() {
		t.Error("special jobs became allowed on a Normal run")
	}

	res, err := pickJobs(m)
	if err != nil {
		t.Fatalf("pickJobs: %v", err)
	}
	for _, job := range res.Jobs {
		if IsSpecialJob(job) {
			t.Errorf("rolled special job %q on a Normal run", job)
		}
	}
}

func TestWizardCursorStaysInBounds(t *testing.T) {
	// Hold "down" well past the end of each list; the views index by cursor,
	// so an unclamped cursor would panic. View() is called each time to prove
	// the cursor is a legal index, not just a legal number.
	overshoot := len(AllJobNames()) + 10

	m := newRun()
	m, _ = press(t, m, downTo(overshoot)...)
	if m.cursor != len(RunTypes)-1 {
		t.Errorf("run type cursor is %d, want %d", m.cursor, len(RunTypes)-1)
	}
	m.View()

	// Walk to a run type that does allow a job set, since the last one in the
	// list happens to skip that step.
	m = newRun()
	m, _ = press(t, m, downTo(indexOfRunType(t, "Normal"))...)
	m, _ = press(t, m, special(tea.KeyEnter))

	m, _ = press(t, m, downTo(overshoot)...)
	if m.cursor != len(JobSets) { // "(none)" occupies index 0
		t.Errorf("job set cursor is %d, want %d", m.cursor, len(JobSets))
	}
	m.View()

	m, _ = press(t, m, special(tea.KeyEnter))
	m, _ = press(t, m, downTo(overshoot)...)
	if m.cursor != len(AllJobNames())-1 {
		t.Errorf("excludes cursor is %d, want %d", m.cursor, len(AllJobNames())-1)
	}
	m.View()

	m, _ = press(t, m, special(tea.KeyEnter))
	m, _ = press(t, m, downTo(overshoot)...)
	if m.cursor != optCount-1 {
		t.Errorf("options cursor is %d, want %d", m.cursor, optCount-1)
	}
	m.View()
}
