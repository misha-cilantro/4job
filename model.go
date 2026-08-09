package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Step identifies which page of the wizard the user is currently on.
const (
	stepRunType = iota
	stepJobSet
	stepExcludes
	stepOptions
	stepSummary
)

const none = "(none)"

// newRun returns a fresh wizard in its initial state.
func newRun() run {
	return run{step: stepRunType}
}

// --- derived state ---
//
// These read the chosen run type and job set instead of caching their effects
// on the model, so re-deciding an earlier step always produces consistent
// locks. A modifier that is locked is forced on and can't be toggled off.

// runTypeDef returns the selected run type, or a zero RunType before one has
// been picked.
func (m run) runTypeDef() RunType {
	return RunTypesByName[m.runType]
}

// jobSetLocked reports whether the run type forbids applying a job set on
// top of it, in which case the job set step is skipped entirely.
func (m run) jobSetLocked() bool {
	return m.runTypeDef().NoJobSetSelect
}

// duplicatesLocked reports whether the run type or the selected job set
// forces Allow Duplicates on.
func (m run) duplicatesLocked() bool {
	if m.runTypeDef().ForcesAllowDuplicates {
		return true
	}
	js, ok := JobSetsByName[m.jobSet]
	return ok && js.ForcesAllowDuplicates
}

// duplicatesAllowed is the effective Allow Duplicates setting.
func (m run) duplicatesAllowed() bool {
	return m.allowDuplicates || m.duplicatesLocked()
}

// specialLocked reports whether the run type forces Allow Special Jobs on.
func (m run) specialLocked() bool {
	return m.runTypeDef().AllowSpecial
}

// specialAllowed is the effective Allow Special Jobs setting.
func (m run) specialAllowed() bool {
	return m.allowSpecial || m.specialLocked()
}

// stepLabel renders a "Step 2/4" header for step. Run types that skip the
// job set step have one fewer step, so the numbering is computed from the
// steps this particular run will actually visit.
func (m run) stepLabel(step int) string {
	steps := []int{stepRunType, stepJobSet, stepExcludes, stepOptions}
	if m.jobSetLocked() {
		steps = []int{stepRunType, stepExcludes, stepOptions}
	}
	for i, s := range steps {
		if s == step {
			return fmt.Sprintf("Step %d/%d", i+1, len(steps))
		}
	}
	return ""
}

func (m run) Init() tea.Cmd {
	return nil
}

func (m run) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "backspace":
		return m.back(), nil
	}

	switch m.step {
	case stepRunType:
		return m.updateRunType(keyMsg)
	case stepJobSet:
		return m.updateJobSet(keyMsg)
	case stepExcludes:
		return m.updateExcludes(keyMsg)
	case stepOptions:
		return m.updateOptions(keyMsg)
	case stepSummary:
		return m.updateSummary(keyMsg)
	}

	return m, nil
}

func (m run) View() tea.View {
	var s string
	switch m.step {
	case stepRunType:
		s = m.viewRunType()
	case stepJobSet:
		s = m.viewJobSet()
	case stepExcludes:
		s = m.viewExcludes()
	case stepOptions:
		s = m.viewOptions()
	case stepSummary:
		s = m.viewSummary()
	}
	return tea.NewView(s)
}

// back moves the wizard to the previous step. Nothing needs undoing: the
// locks are derived from runType and jobSet, so re-choosing either one
// updates them automatically. It's a no-op on stepRunType, since there's
// nowhere earlier to go.
func (m run) back() run {
	switch m.step {
	case stepJobSet:
		m.step = stepRunType
	case stepExcludes:
		if m.jobSetLocked() {
			m.step = stepRunType
		} else {
			m.step = stepJobSet
		}
	case stepOptions:
		m.step = stepExcludes
	case stepSummary:
		m.step = stepOptions
	default:
		return m
	}
	m.cursor = 0
	return m
}

// --- Step 1: run type ---

func (m run) updateRunType(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(RunTypes)-1 {
			m.cursor++
		}
	case "enter":
		m.runType = RunTypes[m.cursor].Name
		m.cursor = 0
		if m.jobSetLocked() {
			m.jobSet = ""
			m.step = stepExcludes
		} else {
			m.step = stepJobSet
		}
	}
	return m, nil
}

func (m run) viewRunType() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s: Pick a run type\n\n", m.stepLabel(stepRunType)))
	for i, rt := range RunTypes {
		b.WriteString(fmt.Sprintf("%s %s\n", cursorFor(i, m.cursor), rt.Name))
		if i == m.cursor {
			b.WriteString(fmt.Sprintf("    %s\n", rt.Description))
		}
	}
	b.WriteString("\n(up/down to move, enter to select, q to quit)")
	return b.String()
}

// --- Step 2: job set ---

func (m run) updateJobSet(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Index 0 is always "(none)"; job sets follow at index+1.
	max := len(JobSets)
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < max {
			m.cursor++
		}
	case "enter":
		if m.cursor == 0 {
			m.jobSet = ""
		} else {
			m.jobSet = JobSets[m.cursor-1].Name
		}
		m.cursor = 0
		m.step = stepExcludes
	}
	return m, nil
}

func (m run) viewJobSet() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s: Pick a job set (optional)\n\n", m.stepLabel(stepJobSet)))
	b.WriteString(fmt.Sprintf("%s (none - no job set restriction)\n", cursorFor(0, m.cursor)))
	for i, js := range JobSets {
		idx := i + 1
		b.WriteString(fmt.Sprintf("%s %s\n", cursorFor(idx, m.cursor), js.Name))
		if idx == m.cursor {
			b.WriteString(fmt.Sprintf("    %s\n", js.Description))
		}
	}
	b.WriteString("\n(up/down to move, enter to select, esc to go back, q to quit)")
	return b.String()
}

// --- Step 3: excluded jobs ---

func (m run) updateExcludes(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	names := AllJobNames()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(names)-1 {
			m.cursor++
		}
	case " ", "space":
		name := names[m.cursor]
		if containsString(m.excludes, name) {
			m.excludes = removeString(m.excludes, name)
		} else {
			m.excludes = append(append([]string{}, m.excludes...), name)
		}
	case "enter":
		m.cursor = 0
		m.step = stepOptions
	}
	return m, nil
}

func (m run) viewExcludes() string {
	names := AllJobNames()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s: Exclude any jobs? (e.g. to avoid recent repeats)\n\n", m.stepLabel(stepExcludes)))
	for i, name := range names {
		mark := " "
		if containsString(m.excludes, name) {
			mark = "x"
		}
		b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursorFor(i, m.cursor), mark, name))
	}
	b.WriteString("\n(up/down to move, space to toggle, enter to continue, esc to go back, q to quit)")
	return b.String()
}

// --- Step 4: options ---

const (
	optDuplicates = iota
	optSpecial
	optCount
)

func (m run) updateOptions(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < optCount-1 {
			m.cursor++
		}
	case " ", "space":
		switch m.cursor {
		case optDuplicates:
			if !m.duplicatesLocked() {
				m.allowDuplicates = !m.allowDuplicates
			}
		case optSpecial:
			if !m.specialLocked() {
				m.allowSpecial = !m.allowSpecial
			}
		}
	case "enter":
		m.name = generateName(m)
		m.step = stepSummary
	}
	return m, nil
}

func (m run) viewOptions() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s: Options\n\n", m.stepLabel(stepOptions)))
	b.WriteString(fmt.Sprintf("%s Allow Duplicates: %s\n",
		cursorFor(optDuplicates, m.cursor), toggleLabel(m.duplicatesAllowed(), m.duplicatesLocked())))
	b.WriteString(fmt.Sprintf("%s Allow Special Jobs (Freelancer/Mime): %s\n",
		cursorFor(optSpecial, m.cursor), toggleLabel(m.specialAllowed(), m.specialLocked())))
	b.WriteString("\n(up/down to move, space to toggle, enter to finish, esc to go back, q to quit)")
	return b.String()
}

// --- Step 5: summary ---

func (m run) updateSummary(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		return newRun(), nil
	case "enter":
		// The only path that authorises writing to disk.
		m.confirmed = true
		return m, tea.Quit
	}
	return m, nil
}

func (m run) viewSummary() string {
	var b strings.Builder

	b.WriteString("Summary\n\n")
	b.WriteString(fmt.Sprintf("Run type:        %s\n", m.runType))

	jobSet := m.jobSet
	if jobSet == "" {
		jobSet = none
	}

	b.WriteString(fmt.Sprintf("Job set:         %s\n", jobSet))
	b.WriteString(fmt.Sprintf("Allow Duplicates: %t\n", m.duplicatesAllowed()))
	b.WriteString(fmt.Sprintf("Allow Special:    %t\n", m.specialAllowed()))

	if len(m.excludes) == 0 {
		b.WriteString("Excluded jobs:    (none)\n")
	} else {
		b.WriteString(fmt.Sprintf("Excluded jobs:    %s\n", strings.Join(m.excludes, ", ")))
	}

	b.WriteString(fmt.Sprintf("\nRun folder: %s\n", m.name))

	b.WriteString("\n(enter to write run folder, r to start over, esc to go back, q to quit without writing)")

	return b.String()
}

// --- shared helpers ---

func cursorFor(i, cursor int) string {
	if i == cursor {
		return ">"
	}
	return " "
}

func toggleLabel(value, locked bool) string {
	state := "no"
	if value {
		state = "yes"
	}
	if locked {
		return state + " (locked)"
	}
	return state
}

func containsString(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

func removeString(list []string, target string) []string {
	out := make([]string, 0, len(list))
	for _, s := range list {
		if s != target {
			out = append(out, s)
		}
	}
	return out
}

func generateName(m run) string {
	var b strings.Builder
	b.WriteString(m.runType)

	if m.jobSet != "" && m.jobSet != none {
		b.WriteString("-")
		b.WriteString(m.jobSet)
	}

	if len(m.excludes) == 0 {
		b.WriteString("-all")
	} else {
		b.WriteString(fmt.Sprintf("-excl-%d", len(m.excludes)))
	}

	b.WriteString("-opt")
	if !m.duplicatesAllowed() && !m.specialAllowed() {
		b.WriteString("X")
	}

	if m.duplicatesAllowed() {
		b.WriteString("D")
	}

	if m.specialAllowed() {
		b.WriteString("S")
	}

	t := time.Now()
	b.WriteString(fmt.Sprintf("-%s", t.Format("20060102150405")))

	return b.String()
}
