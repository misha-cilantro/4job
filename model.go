package main

import (
	"fmt"
	"slices"
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

// extraSlotNames labels the job slots the advanced options add beyond the four
// crystals, for the summary.
func (m run) extraSlotNames() []string {
	var out []string
	if m.fifthJob {
		out = append(out, "Fifth Job")
	}
	if m.extraJobs {
		out = append(out, "Advance job")
	}
	return out
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
		m.height = msg.Height
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
	b.WriteString(title(m.stepLabel(stepRunType) + ": Pick a run type"))

	start, end := visibleRange(len(RunTypes), m.cursor, m.listRows(runTypeChrome))

	if start > 0 {
		fmt.Fprintf(&b, "%s\n", dimStyle.Render(fmt.Sprintf("  ... %d more above", start)))
	}
	for i := start; i < end; i++ {
		fmt.Fprintf(&b, "%s %s\n", cursorFor(i, m.cursor), highlight(RunTypes[i].Name, i == m.cursor))
		if i == m.cursor {
			fmt.Fprintf(&b, "    %s\n", dimStyle.Render(RunTypes[i].Description))
		}
	}
	if end < len(RunTypes) {
		fmt.Fprintf(&b, "%s\n", dimStyle.Render(fmt.Sprintf("  ... %d more below", len(RunTypes)-end)))
	}

	b.WriteString(help("(up/down to move, enter to select, q to quit)"))
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
	b.WriteString(title(m.stepLabel(stepJobSet) + ": Pick a job set (optional)"))
	fmt.Fprintf(&b, "%s %s\n", cursorFor(0, m.cursor),
		highlight("(none - no job set restriction)", m.cursor == 0))
	for i, js := range JobSets {
		idx := i + 1
		fmt.Fprintf(&b, "%s %s\n", cursorFor(idx, m.cursor), highlight(js.Name, idx == m.cursor))
		if idx == m.cursor {
			fmt.Fprintf(&b, "    %s\n", dimStyle.Render(js.Description))
		}
	}
	b.WriteString(help("(up/down to move, enter to select, esc to go back, q to quit)"))
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
		// Cloned rather than mutated in place: run is passed by value, so the
		// previous model must keep its own excludes intact.
		name := names[m.cursor]
		if slices.Contains(m.excludes, name) {
			m.excludes = slices.DeleteFunc(slices.Clone(m.excludes), func(s string) bool {
				return s == name
			})
		} else {
			m.excludes = append(slices.Clone(m.excludes), name)
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

	b.WriteString(title(fmt.Sprintf("%s: Exclude any jobs? (e.g. to avoid recent repeats) [%d excluded]",
		m.stepLabel(stepExcludes), len(m.excludes))))

	// The full roster is 20-odd rows, which overflows a short terminal, so
	// only a window around the cursor is drawn.
	start, end := visibleRange(len(names), m.cursor, m.listRows(excludesChrome))

	if start > 0 {
		fmt.Fprintf(&b, "%s\n", dimStyle.Render(fmt.Sprintf("     ... %d more above", start)))
	}
	for i := start; i < end; i++ {
		mark := "[ ]"
		if slices.Contains(m.excludes, names[i]) {
			mark = markStyle.Render("[x]")
		}
		fmt.Fprintf(&b, "%s %s %s\n", cursorFor(i, m.cursor), mark, highlight(names[i], i == m.cursor))
	}
	if end < len(names) {
		fmt.Fprintf(&b, "%s\n", dimStyle.Render(fmt.Sprintf("     ... %d more below", len(names)-end)))
	}

	b.WriteString(help("(up/down to move, space to toggle, enter to continue, esc to go back, q to quit)"))
	return b.String()
}

// --- Step 4: options ---

// Rows on the options step. Special jobs used to be a toggle here too, but
// they're fixed by the run type and now show as information only.
const (
	optDuplicates = iota
	optRestriction
	optFifthJob
	optExtraJobs
	optForbidden
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
		case optRestriction:
			// Cycles rather than toggles: the three modes are exclusive.
			m.restriction = (m.restriction + 1) % restrictCount
		case optFifthJob:
			m.fifthJob = !m.fifthJob
		case optExtraJobs:
			m.extraJobs = !m.extraJobs
		case optForbidden:
			// Cycles: off, rolled now, left to the player.
			m.forbidden = (m.forbidden + 1) % forbiddenCount
		}
	case "enter":
		m.name = generateName(m)
		m.step = stepSummary
	}
	return m, nil
}

func (m run) viewOptions() string {
	var b strings.Builder
	b.WriteString(title(m.stepLabel(stepOptions) + ": Options"))

	option := func(row int, label, value string) {
		fmt.Fprintf(&b, "%s %s %s\n", cursorFor(row, m.cursor),
			highlight(label+":", row == m.cursor), value)
	}
	option(optDuplicates, "Allow Duplicates", toggleLabel(m.duplicatesAllowed(), m.duplicatesLocked()))
	option(optRestriction, "Job restrictions", valueStyle.Render(restrictionName(m.restriction)))
	option(optFifthJob, "Fifth Job (Krile)", valueStyle.Render(yesNo(m.fifthJob)))
	option(optExtraJobs, "Extra Jobs (one Advance job)", valueStyle.Render(yesNo(m.extraJobs)))
	option(optForbidden, "Forbidden (a job is crossed out in the Void)", valueStyle.Render(forbiddenName(m.forbidden)))

	fmt.Fprintf(&b, "\n  %s\n", dimStyle.Render(restrictionRules[m.restriction]))

	// Shown for information only - the run type decides it, so it gets no
	// cursor row.
	fmt.Fprintf(&b, "  %s\n", dimStyle.Render(fmt.Sprintf(
		"Special jobs (Freelancer/Mime): %s, fixed by the %s run type.", yesNo(m.specialAllowed()), m.runType)))

	b.WriteString(help("(up/down to move, space to toggle, enter to finish, esc to go back, q to quit)"))
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

	b.WriteString(title("Summary"))

	// Labels padded before styling, since style codes would throw the width off.
	field := func(label, value string) {
		fmt.Fprintf(&b, "%s %s\n", dimStyle.Render(fmt.Sprintf("%-18s", label+":")), valueStyle.Render(value))
	}

	jobSet := m.jobSet
	if jobSet == "" {
		jobSet = none
	}

	field("Run type", m.runType)
	field("Job set", jobSet)
	field("Job restrictions", restrictionName(m.restriction))
	field("Allow Duplicates", yesNo(m.duplicatesAllowed()))
	field("Allow Special", yesNo(m.specialAllowed()))

	if extras := m.extraSlotNames(); len(extras) > 0 {
		field("Extra jobs", strings.Join(extras, ", "))
	}
	if m.forbidden != forbiddenOff {
		field("Forbidden", forbiddenName(m.forbidden))
	}

	if len(m.excludes) == 0 {
		field("Excluded jobs", none)
	} else {
		field("Excluded jobs", strings.Join(m.excludes, ", "))
	}

	fmt.Fprintf(&b, "\n%s %s\n", dimStyle.Render("Run folder:"), activeStyle.Render(m.name))

	b.WriteString(help("(enter to write run folder, r to start over, esc to go back, q to quit without writing)"))

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
		return cursorStyle.Render(">")
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
	label := valueStyle.Render(yesNo(value))
	if locked {
		return label + " " + dimStyle.Render("(locked)")
	}
	return label
}
