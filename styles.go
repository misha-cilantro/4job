package main

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// minTextWidth stops an absurdly narrow terminal from wrapping text to one
// character per line; below this the view overflows instead.
const minTextWidth = 24

// Styles are deliberately few, and use the terminal's own 16-colour palette by
// index rather than fixed RGB. That way the wizard sits comfortably in whatever
// theme the player already runs, light or dark, and degrades to plain text on a
// terminal without colour.
var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	activeStyle = lipgloss.NewStyle().Bold(true)
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	markStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	dimStyle    = lipgloss.NewStyle().Faint(true)
)

// title renders a step heading.
func title(text string) string {
	return titleStyle.Render(text) + "\n\n"
}

// help renders the key hints at the foot of a step.
func help(text string) string {
	return "\n" + dimStyle.Render(text)
}

// highlight emphasises the list item the cursor is on.
func highlight(text string, active bool) string {
	if active {
		return activeStyle.Render(text)
	}
	return text
}

// wrap word-wraps text to width, leaving it alone when width is zero or less.
// It's ANSI-aware, so it can be handed text that's already styled, and
// already-wrapped text passes through unchanged.
//
// It applies no minimum of its own: callers pass a width already run through
// run.effectiveWidth, so the floor is decided once rather than compounding when
// an indent is subtracted.
func wrap(text string, width int) string {
	if width <= 0 {
		return text
	}

	// Style.Width word-wraps and pads each line out to width. Trim the padding
	// back off so nothing writes into the last column needlessly, then clamp:
	// the wrap overshoots by a column when a line ends on a breakpoint such as
	// "-", and fitting the terminal has to be a hard guarantee rather than a
	// near miss. Anything the clamp removes is a character or two of trailing
	// punctuation the wrap should have carried over itself.
	clamp := lipgloss.NewStyle().MaxWidth(width)
	lines := strings.Split(lipgloss.NewStyle().Width(width).Render(text), "\n")
	for i, line := range lines {
		lines[i] = clamp.Render(strings.TrimRight(line, " "))
	}
	return strings.Join(lines, "\n")
}

// countLines is how many terminal rows a block occupies.
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// extraRows is how many rows a block takes beyond the single line the chrome
// constants assume it needs.
func extraRows(s string) int {
	return max(countLines(s)-1, 0)
}
