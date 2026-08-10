package main

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestVisibleRangeKeepsCursorOnScreen(t *testing.T) {
	const total, rows = 22, 5

	for cursor := range total {
		start, end := visibleRange(total, cursor, rows)

		if end-start != rows {
			t.Errorf("cursor %d: window is %d rows, want %d", cursor, end-start, rows)
		}
		if cursor < start || cursor >= end {
			t.Errorf("cursor %d falls outside the window [%d,%d)", cursor, start, end)
		}
		if start < 0 || end > total {
			t.Errorf("cursor %d: window [%d,%d) is out of bounds", cursor, start, end)
		}
	}
}

func TestVisibleRangeShowsEverythingWhenItFits(t *testing.T) {
	start, end := visibleRange(4, 2, 10)
	if start != 0 || end != 4 {
		t.Errorf("got [%d,%d), want [0,4)", start, end)
	}
}

func TestVisibleRangeFillsTheLastPage(t *testing.T) {
	// At the bottom of the list the window should stop sliding rather than
	// run past the end and render a half-empty page.
	start, end := visibleRange(22, 21, 5)
	if start != 17 || end != 22 {
		t.Errorf("got [%d,%d), want [17,22)", start, end)
	}
}

func TestListRowsFallsBackBeforeFirstResize(t *testing.T) {
	if got := (run{}).listRows(excludesChrome); got != defaultListRows {
		t.Errorf("got %d rows with no window size, want %d", got, defaultListRows)
	}
	if got := (run{height: 4}).listRows(excludesChrome); got != minListRows {
		t.Errorf("got %d rows in a 4-row terminal, want %d", got, minListRows)
	}
}

// TestExcludesViewFitsTerminal is the point of the viewport: the step used to
// print all 22 jobs unconditionally and overflow a short window.
func TestExcludesViewFitsTerminal(t *testing.T) {
	for _, height := range []int{10, 20, 24, 60} {
		m := run{step: stepExcludes, runType: "Normal", height: height}
		for _, cursor := range []int{0, 11, len(AllJobNames()) - 1} {
			m.cursor = cursor
			lines := strings.Count(m.viewExcludes(), "\n") + 1
			if lines > height {
				t.Errorf("height %d, cursor %d: view is %d lines", height, cursor, lines)
			}
		}
	}
}

func TestExcludesViewShowsScrollHints(t *testing.T) {
	m := run{step: stepExcludes, runType: "Normal", height: 12, cursor: 11}
	view := m.viewExcludes()

	if !strings.Contains(view, "more above") {
		t.Error("expected a hint that items are scrolled off the top")
	}
	if !strings.Contains(view, "more below") {
		t.Error("expected a hint that items are scrolled off the bottom")
	}
}

func TestRunTypeViewFitsTerminal(t *testing.T) {
	for _, height := range []int{10, 24} {
		m := run{step: stepRunType, height: height}
		for cursor := range RunTypes {
			m.cursor = cursor
			lines := strings.Count(m.viewRunType(), "\n") + 1
			if lines > height {
				t.Errorf("height %d, cursor %d: view is %d lines", height, cursor, lines)
			}
		}
	}
}

// everyView renders each step of a fully-populated run, so a width or height
// check can sweep all of them.
func everyView(t *testing.T, width, height int) map[string]string {
	t.Helper()

	m := run{
		width: width, height: height,
		runType: "Onion", // its description is one of the longest
		excludes: []string{"bard", "beastmaster", "berserker", "black mage", "blue mage",
			"cannoneer", "chemist", "dancer", "dragoon", "freelancer"},
		restriction: restrictUpgrade,
		fifthJob:    true,
		extraJobs:   true,
		forbidden:   forbiddenRolled,
	}
	m.name = generateName(m)

	out := map[string]string{}
	for _, step := range []struct {
		name string
		step int
		max  int
	}{
		{"runType", stepRunType, len(RunTypes) - 1},
		{"jobSet", stepJobSet, len(JobSets)},
		{"excludes", stepExcludes, len(AllJobNames()) - 1},
		{"options", stepOptions, optCount - 1},
		{"summary", stepSummary, 0},
	} {
		// Render at both ends of the list and in the middle: the cursor decides
		// which description is shown, and descriptions are the longest text.
		for _, cursor := range []int{0, step.max / 2, step.max} {
			at := m
			at.step, at.cursor = step.step, cursor
			out[fmt.Sprintf("%s@%d", step.name, cursor)] = at.render()
		}
	}
	return out
}

// TestViewsFitTerminalWidth is the width counterpart to the height tests: no
// rendered line may be wider than the terminal, on any step, at any width down
// to the minTextWidth floor.
func TestViewsFitTerminalWidth(t *testing.T) {
	for _, width := range []int{minTextWidth, 30, 40, 55, 80, 120, 200} {
		for name, view := range everyView(t, width, 40) {
			for i, line := range strings.Split(view, "\n") {
				// lipgloss.Width measures display columns, ignoring style codes.
				if got := lipgloss.Width(line); got > width {
					t.Errorf("width %d, %s line %d is %d columns: %q",
						width, name, i+1, got, line)
				}
			}
		}
	}
}

// TestVeryNarrowTerminalsClampRatherThanShred records what happens below the
// floor: lines stay at minTextWidth instead of wrapping to a couple of
// characters each, so the view overflows sideways rather than becoming a column
// of syllables.
func TestVeryNarrowTerminalsClampRatherThanShred(t *testing.T) {
	for _, width := range []int{1, 8, minTextWidth - 1} {
		for name, view := range everyView(t, width, 40) {
			for i, line := range strings.Split(view, "\n") {
				if got := lipgloss.Width(line); got > minTextWidth {
					t.Errorf("width %d, %s line %d is %d columns, above the %d floor: %q",
						width, name, i+1, got, minTextWidth, line)
				}
			}
		}
	}
}

// TestViewsFitTerminalHeight re-checks the height budget now that wrapping can
// make the header, footer and description taller than one line each. A narrow
// terminal spends more rows on that chrome, so it needs more height before the
// list can fit — hence the floor below, which is minListRows plus the most
// chrome any step draws once wrapped.
func TestViewsFitTerminalHeight(t *testing.T) {
	const heightFloor = 18

	for _, width := range []int{40, 80, 120} {
		for _, height := range []int{heightFloor, 24, 50} {
			for name, view := range everyView(t, width, height) {
				// Only the scrolling steps promise to fit a given height.
				if !strings.HasPrefix(name, "runType") && !strings.HasPrefix(name, "excludes") {
					continue
				}
				if got := countLines(view); got > height {
					t.Errorf("width %d height %d: %s is %d lines\n%s", width, height, name, got, view)
				}
			}
		}
	}
}

func TestGeneratedNameIsAPlainFolderName(t *testing.T) {
	// Every run type / job set pairing the wizard can reach, since both names
	// end up in the folder name.
	for _, rt := range RunTypes {
		jobSets := []string{""}
		if !rt.NoJobSetSelect {
			for _, js := range JobSets {
				jobSets = append(jobSets, js.Name)
			}
		}

		for _, js := range jobSets {
			m := run{runType: rt.Name, jobSet: js}
			m.name = generateName(m)

			if m.name == "" {
				t.Errorf("%s + %q produced an empty folder name", rt.Name, js)
				continue
			}
			if err := checkFolderName(m.name); err != nil {
				t.Errorf("%s + %q: %v", rt.Name, js, err)
			}
		}
	}
}

func TestSanitizeFolderName(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Normal-all-optX-20260809", "Normal-all-optX-20260809"},
		{"Team 750 run", "Team 750 run"},
		{`a/b\c`, "a-b-c"},
		{`re<>:"|?*mo`, "re-------mo"},
		{"..", ""},
		{".", ""},
		{"   ", ""},
		{"", ""},
		{"  trimmed.  ", "trimmed"},
		{"nul", "nul-run"},
		{"COM1.txt", "COM1.txt-run"},
		{"tab\there", "tabhere"},
		{"nul-but-not-really", "nul-but-not-really"},
	} {
		if got := sanitizeFolderName(tc.in); got != tc.want {
			t.Errorf("sanitizeFolderName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSanitizeFolderNameTruncatesWithoutSplittingRunes(t *testing.T) {
	got := sanitizeFolderName(strings.Repeat("é", maxNameRunes+20))
	if n := len([]rune(got)); n != maxNameRunes {
		t.Errorf("got %d runes, want %d", n, maxNameRunes)
	}
	if !strings.HasPrefix(got, "é") || strings.ContainsRune(got, '�') {
		t.Errorf("truncation split a multi-byte rune: %q", got)
	}
}
