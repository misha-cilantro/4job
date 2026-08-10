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

// specialAllowed reports whether the run may roll Freelancer or Mime. It's
// decided entirely by the run type, not by the player: the rules only make
// them available on run types that declare it, so there's no toggle.
func (m run) specialAllowed() bool {
	return m.runTypeDef().AllowSpecial
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m run) handleKey(keyMsg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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

	start, end := visibleRange(len(RunTypes), m.cursor, m.listRows(runTypeChrome))

	if start > 0 {
		b.WriteString(fmt.Sprintf("  ... %d more above\n", start))
	}
	for i := start; i < end; i++ {
		b.WriteString(fmt.Sprintf("%s %s\n", cursorFor(i, m.cursor), RunTypes[i].Name))
		if i == m.cursor {
			b.WriteString(fmt.Sprintf("    %s\n", RunTypes[i].Description))
		}
	}
	if end < len(RunTypes) {
		b.WriteString(fmt.Sprintf("  ... %d more below\n", len(RunTypes)-end))
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

	b.WriteString(fmt.Sprintf("%s: Exclude any jobs? (e.g. to avoid recent repeats) [%d excluded]\n\n",
		m.stepLabel(stepExcludes), len(m.excludes)))

	// The full roster is 20-odd rows, which overflows a short terminal, so
	// only a window around the cursor is drawn.
	start, end := visibleRange(len(names), m.cursor, m.listRows(excludesChrome))

	if start > 0 {
		b.WriteString(fmt.Sprintf("     ... %d more above\n", start))
	}
	for i := start; i < end; i++ {
		mark := " "
		if containsString(m.excludes, names[i]) {
			mark = "x"
		}
		b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursorFor(i, m.cursor), mark, names[i]))
	}
	if end < len(names) {
		b.WriteString(fmt.Sprintf("     ... %d more below\n", len(names)-end))
	}

	b.WriteString("\n(up/down to move, space to toggle, enter to continue, esc to go back, q to quit)")
	return b.String()
}

// --- Step 4: options ---

// Allow Duplicates is the only option the player sets. Special jobs used to be
// a toggle here too, but they're fixed by the run type.
const (
	optDuplicates = iota
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
		if m.cursor == optDuplicates && !m.duplicatesLocked() {
			m.allowDuplicates = !m.allowDuplicates
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

	// Shown for information only - the run type decides it, so it gets no
	// cursor row.
	b.WriteString(fmt.Sprintf("\n  Special jobs (Freelancer/Mime): %s\n", yesNo(m.specialAllowed())))
	b.WriteString(fmt.Sprintf("  Fixed by the %s run type; only some run types make them available.\n", m.runType))

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

// --- list scrolling ---

// Rows each step spends on things other than list items, counting the header,
// the blank line under it, both scroll hints, and the blank line plus footer
// at the bottom.
const (
	excludesChrome = 6

	// The run type list spends one more row on the highlighted item's
	// description.
	runTypeChrome = 7

	// defaultListRows is used until the first tea.WindowSizeMsg arrives, and
	// minListRows keeps the list usable in a very short terminal. Below
	// chrome+minListRows rows the view does overflow, on the grounds that a
	// one-item list is worse than a little scrollback.
	defaultListRows = 18
	minListRows     = 3
)

// listRows is how many list items fit in the terminal, given that chrome rows
// are spent on everything around the list.
func (m run) listRows(chrome int) int {
	if m.height <= 0 {
		return defaultListRows
	}
	return max(m.height-chrome, minListRows)
}

// visibleRange returns the [start, end) bounds of a scrolling window of rows
// items over a list of total items. The window keeps cursor centred where it
// can, and stops sliding once it reaches either end so the last page stays
// full instead of trailing off.
func visibleRange(total, cursor, rows int) (start, end int) {
	if rows >= total {
		return 0, total
	}

	start = cursor - rows/2
	start = min(start, total-rows)
	start = max(start, 0)

	return start, start + rows
}

// --- shared helpers ---

func cursorFor(i, cursor int) string {
	if i == cursor {
		return ">"
	}
	return " "
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func toggleLabel(value, locked bool) string {
	if locked {
		return yesNo(value) + " (locked)"
	}
	return yesNo(value)
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
