package main

import "charm.land/lipgloss/v2"

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
