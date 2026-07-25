package main

import (
	"fmt"
	"strings"

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

// newRun returns a fresh wizard in its initial state.
func newRun() run {
	return run{step: stepRunType}
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

// back moves the wizard to the previous step, undoing anything that step
// had locked or set. It's a no-op on stepRunType, since there's nowhere
// earlier to go.
func (m run) back() run {
	switch m.step {
	case stepJobSet:
		m.step = stepRunType
		m.cursor = 0
	case stepExcludes:
		if m.jobSetLocked {
			m.step = stepRunType
		} else {
			m.step = stepJobSet
		}
		m.cursor = 0
	case stepOptions:
		m.step = stepExcludes
		m.cursor = 0
	case stepSummary:
		m.step = stepOptions
		m.cursor = 0
	}
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
		rt := RunTypes[m.cursor]
		m.runType = rt.Name
		m.jobSetLocked = rt.NoJobSetSelect
		m.allowSpecial = rt.AllowSpecial
		m.specialLocked = rt.AllowSpecial
		m.allowDuplicates = rt.ForcesAllowDuplicates
		m.duplicatesLocked = rt.ForcesAllowDuplicates
		m.cursor = 0
		if m.jobSetLocked {
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
	b.WriteString("Step 1/4: Pick a run type\n\n")
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
			js := JobSets[m.cursor-1]
			m.jobSet = js.Name
			if js.ForcesAllowDuplicates {
				m.allowDuplicates = true
				m.duplicatesLocked = true
			}
		}
		m.cursor = 0
		m.step = stepExcludes
	}
	return m, nil
}

func (m run) viewJobSet() string {
	var b strings.Builder
	b.WriteString("Step 2/4: Pick a job set (optional)\n\n")
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
	b.WriteString("Step 3/4: Exclude any jobs? (e.g. to avoid recent repeats)\n\n")
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
			if !m.duplicatesLocked {
				m.allowDuplicates = !m.allowDuplicates
			}
		case optSpecial:
			if !m.specialLocked {
				m.allowSpecial = !m.allowSpecial
			}
		}
	case "enter":
		m.step = stepSummary
	}
	return m, nil
}

func (m run) viewOptions() string {
	var b strings.Builder
	b.WriteString("Step 4/4: Options\n\n")
	b.WriteString(fmt.Sprintf("%s Allow Duplicates: %s\n",
		cursorFor(optDuplicates, m.cursor), toggleLabel(m.allowDuplicates, m.duplicatesLocked)))
	b.WriteString(fmt.Sprintf("%s Allow Special Jobs (Freelancer/Mime): %s\n",
		cursorFor(optSpecial, m.cursor), toggleLabel(m.allowSpecial, m.specialLocked)))
	b.WriteString("\n(up/down to move, space to toggle, enter to finish, esc to go back, q to quit)")
	return b.String()
}

// --- Step 5: summary ---

func (m run) updateSummary(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		return newRun(), nil
	case "enter":
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
		jobSet = "(none)"
	}
	b.WriteString(fmt.Sprintf("Job set:         %s\n", jobSet))
	b.WriteString(fmt.Sprintf("Allow Duplicates: %t\n", m.allowDuplicates))
	b.WriteString(fmt.Sprintf("Allow Special:    %t\n", m.allowSpecial))
	if len(m.excludes) == 0 {
		b.WriteString("Excluded jobs:    (none)\n")
	} else {
		b.WriteString(fmt.Sprintf("Excluded jobs:    %s\n", strings.Join(m.excludes, ", ")))
	}
	b.WriteString("\n(enter to quit, r to start a new run, esc to go back)")
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
