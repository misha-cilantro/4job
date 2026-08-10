package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// maxNameRunes caps the folder name well short of any filesystem limit, since
// the name is only a label for the run.
const maxNameRunes = 80

// generateName builds the run folder name from the run's settings, ending in a
// timestamp so consecutive runs don't collide.
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
		fmt.Fprintf(&b, "-excl-%d", len(m.excludes))
	}

	// One letter per active option, or X for a run with none of them.
	var opts strings.Builder
	if m.duplicatesAllowed() {
		opts.WriteString("D")
	}
	if m.specialAllowed() {
		opts.WriteString("S")
	}
	switch m.restriction {
	case restrictNatural:
		opts.WriteString("N")
	case restrictUpgrade:
		opts.WriteString("U")
	}
	if m.fifthJob {
		opts.WriteString("5")
	}
	if m.extraJobs {
		opts.WriteString("A")
	}
	// FR for a forbidden job rolled up front, F for one the player picks later.
	// Distinguished by letters rather than case, since Windows folder names are
	// case-insensitive.
	switch m.forbidden {
	case forbiddenRolled:
		opts.WriteString("FR")
	case forbiddenPlayer:
		opts.WriteString("F")
	}

	b.WriteString("-opt")
	if opts.Len() == 0 {
		b.WriteString("X")
	} else {
		b.WriteString(opts.String())
	}

	fmt.Fprintf(&b, "-%s", time.Now().Format("20060102150405"))

	return sanitizeFolderName(b.String())
}

// sanitizeFolderName strips everything that can't safely be a single folder
// name: path separators, the characters Windows forbids, control characters,
// and leading or trailing dots and spaces. Windows device names get a suffix,
// since a folder called "CON" or "NUL" can't be created.
//
// The name is built from run type and job set names, so nothing here is
// user-typed - but those names come from the JSON data files, and a stray "/"
// in one would otherwise turn the run folder into a nested path.
func sanitizeFolderName(s string) string {
	const forbidden = `<>:"/\|?*`

	var b strings.Builder
	for _, r := range s {
		switch {
		case !unicode.IsPrint(r):
			// Drop control and non-printable characters entirely.
		case strings.ContainsRune(forbidden, r):
			b.WriteRune('-')
		default:
			b.WriteRune(r)
		}
	}

	name := b.String()
	if runes := []rune(name); len(runes) > maxNameRunes {
		name = string(runes[:maxNameRunes])
	}

	// Trimmed last: truncation above can expose a trailing dot or space, both
	// of which Windows silently strips from directory names.
	name = strings.Trim(name, " .")
	if name == "" {
		return ""
	}

	if isReservedName(name) {
		name += "-run"
	}
	return name
}

// reservedNames are the legacy DOS device names that Windows still refuses to
// use as a file or folder name, with or without an extension.
var reservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// isReservedName reports whether name is a Windows device name. The check
// applies to the portion before the first dot, so "NUL.txt" counts too.
func isReservedName(name string) bool {
	base, _, _ := strings.Cut(name, ".")
	return reservedNames[strings.ToUpper(base)]
}
