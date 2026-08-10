package main

import (
	"fmt"
	"strings"
)

// Job restriction modes decide which character may use an assigned job. They're
// mutually exclusive, so the options step cycles between them rather than
// offering three toggles.
const (
	restrictNone = iota
	restrictNatural
	restrictUpgrade
	restrictCount
)

// restrictionNames label the modes in the wizard and in the rules file.
var restrictionNames = []string{
	restrictNone:    "no restrictions",
	restrictNatural: "natural jobs",
	restrictUpgrade: "upgrade jobs",
}

// restrictionRules is the one-line explanation of each mode for the rules file.
var restrictionRules = []string{
	restrictNone:    "Any character may use any assigned job, and you may swap who has which at any time.",
	restrictNatural: "Each character is locked to one job: Bartz job 1, Lenna job 2, Faris job 3, Galuf job 4.",
	restrictUpgrade: "All four characters share one job at a time. Unlock the next whenever you like; doing so retires every earlier job.",
}

// naturalCharacters maps each crystal slot to the character who owns it on a
// Natural run. Krile is handled separately: she inherits Galuf's job unless the
// Fifth Job option gives her one of her own.
var naturalCharacters = []string{"Bartz", "Lenna", "Faris", "Galuf"}

// Forbidden modes. The rule crosses out one job on entering the Void; the app
// can either decide which now or leave it to the player to pick in-game.
const (
	forbiddenOff = iota
	forbiddenRolled
	forbiddenPlayer
	forbiddenCount
)

var forbiddenNames = []string{
	forbiddenOff:    "no",
	forbiddenRolled: "yes, rolled now",
	forbiddenPlayer: "yes, set by player",
}

// forbiddenPlayerText stands in for the job name in the forbidden file when the
// player picks. It keeps the first line meaningful for a one-line stream source.
const forbiddenPlayerText = "set by player"

func forbiddenName(mode int) string {
	if mode < 0 || mode >= len(forbiddenNames) {
		return forbiddenNames[forbiddenOff]
	}
	return forbiddenNames[mode]
}

// forbiddenBody is the contents of the forbidden file: which job the Void takes,
// or a note that it's the player's call, followed by the rule.
func forbiddenBody(m run, res pickResult) string {
	var b strings.Builder

	if m.forbidden == forbiddenRolled && res.Forbidden != "" {
		fmt.Fprintf(&b, "%s\n", res.Forbidden)
		b.WriteString("Crossed out on entering the Void. You may no longer use it.\n")
		b.WriteString("Rolled when this run was created, so it's settled in advance.")
		return b.String()
	}

	fmt.Fprintf(&b, "%s\n", forbiddenPlayerText)
	b.WriteString("On entering the Void, choose one of your jobs and cross it out.\n")
	b.WriteString("You may no longer use it.")
	return b.String()
}

func restrictionName(mode int) string {
	if mode < 0 || mode >= len(restrictionNames) {
		return restrictionNames[restrictNone]
	}
	return restrictionNames[mode]
}

// slotNotes returns the rule lines written under the job in its own file.
//
// Only rules that actually constrain this slot are included: a plain
// No Restrictions run adds nothing, so its files stay a single job name and
// remain usable as a one-line stream source.
func slotNotes(m run, res pickResult, position int) []string {
	slots := res.Slots
	s := slots[position]
	var notes []string

	switch m.restriction {
	case restrictNatural:
		switch s.Kind {
		case slotCrystal:
			if position < len(naturalCharacters) {
				notes = append(notes, fmt.Sprintf("%s must always be a %s.", naturalCharacters[position], s.Job))
			}
			// Krile takes Galuf's job unless the Fifth Job option gives her
			// one of her own.
			if position == len(naturalCharacters)-1 && !hasKind(slots, slotFifth) {
				notes = append(notes, "Krile too, once she joins.")
			}
		case slotFifth:
			notes = append(notes, fmt.Sprintf("Krile must always be a %s.", s.Job))
		case slotAdvance:
			notes = append(notes, "Goes to whichever character retires a job for it.")
		}

	case restrictUpgrade:
		if s.Kind == slotCrystal {
			if position == 0 {
				notes = append(notes, "All four characters start with this job.")
			} else {
				notes = append(notes, "Unlock whenever you like. Doing so retires every earlier job.")
			}
		}
	}

	// The extra jobs carry their own rules whatever the restriction mode.
	switch s.Kind {
	case slotFifth:
		notes = append(notes,
			"Unlocked when Krile joins the party.",
			"You must choose one of your previous Jobs to no longer use. This includes that Job's Abilities.",
			"You may not swap between all five Jobs. Only use four.")
	case slotAdvance:
		notes = append(notes,
			"An Advance job (GBA and later), added once you have all twelve legendary weapons.",
			"You must choose one of your previous Jobs to no longer use. This includes that Job's Abilities.")
	}

	// A rolled Forbidden job is already known, so say so in its own file too.
	// Every slot holding it is affected, which matters when duplicates are on.
	if res.Forbidden != "" && s.Job == res.Forbidden {
		notes = append(notes, "Crossed out on entering the Void; unusable from then on.")
	}

	return notes
}

// crystalTriggers say when each crystal's job becomes available, so the player
// knows which file to open when.
var crystalTriggers = []string{
	"When the Wind Crystal shatters, at the end of the Wind Shrine.",
	"When the Water Crystal shatters, in Walse Tower.",
	"When the Fire Crystal shatters, in Karnak Castle.",
	"When the Earth Crystal shatters, in the Ronka Ruins.",
}

// slotTrigger says when a slot's file should be opened.
func slotTrigger(s assignedSlot, position int) string {
	switch s.Kind {
	case slotFifth:
		return "When Krile joins the party."
	case slotAdvance:
		return "Once you have all twelve legendary weapons."
	}
	if position < len(crystalTriggers) {
		return crystalTriggers[position]
	}
	return "When the next job unlocks."
}

// instructions is the contents of the instructions file: which file to open
// when. It lists only the files this run actually has.
func instructions(m run, res pickResult) string {
	var b strings.Builder

	b.WriteString("Open one file at a time, when its moment comes. No peeking ahead.\n\n")
	fmt.Fprintf(&b, "  %-18s %s\n", rulesFile+".txt", "Now, before you start.")

	for i, s := range res.Slots {
		fmt.Fprintf(&b, "  %-18s %s\n", slotFilename(s, i)+".txt", slotTrigger(s, i))
	}
	if m.forbidden != forbiddenOff {
		fmt.Fprintf(&b, "  %-18s %s\n", forbiddenFilename(res.Slots)+".txt", "On entering the Void.")
	}

	// Upgrade runs decouple learning the job from switching to it, so the file
	// still gets opened at the crystal but acting on it is the player's call.
	if m.restriction == restrictUpgrade {
		b.WriteString("\nUpgrade Jobs: opening a file tells you the job. When to switch to it is your choice,\n")
		b.WriteString("and switching retires everything before it.\n")
	}

	return b.String()
}

// runRules is the contents of the rules file: what the run is and the rules that
// apply to the whole run rather than to one slot.
//
// It names no assigned job. This file is read before the run starts, so listing
// the roll here would spoil every job file at once - which is what the
// instructions file is at pains to avoid.
func runRules(m run, res pickResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Four Job Fiesta - %s run\n\n", m.runType)

	jobSet := m.jobSet
	if jobSet == "" {
		jobSet = none
	}
	fmt.Fprintf(&b, "Job set:          %s\n", jobSet)
	fmt.Fprintf(&b, "Job restrictions: %s\n", restrictionName(m.restriction))
	fmt.Fprintf(&b, "Duplicates:       %s\n", allowedOrNot(m.duplicatesAllowed()))
	fmt.Fprintf(&b, "Forbidden:        %s\n", forbiddenName(m.forbidden))
	if m.specialAllowed() {
		fmt.Fprintf(&b, "Special jobs:     %s may be assigned\n", strings.Join(SpecialJobs(), " and "))
	}

	b.WriteString("\nRules\n")
	for _, rule := range runRuleLines(m, res) {
		fmt.Fprintf(&b, "- %s\n", rule)
	}

	if len(m.excludes) > 0 {
		fmt.Fprintf(&b, "\nKept out of the roll: %s\n", strings.Join(m.excludes, ", "))
	}

	return b.String()
}

// runRuleLines is the bulleted rule list for the rules file.
func runRuleLines(m run, res pickResult) []string {
	rules := []string{restrictionRules[m.restriction]}

	if hasKind(res.Slots, slotFifth) {
		rules = append(rules,
			"Fifth Job: Krile brings a fifth job. Retire one earlier job, its abilities included, and only ever use four.")
	}
	if hasKind(res.Slots, slotAdvance) {
		rules = append(rules,
			"Extra Jobs: one Advance job is assigned. Retire one earlier job, its abilities included.")
	}

	switch m.forbidden {
	case forbiddenRolled:
		// Deliberately doesn't name the job: this file is read first, and the
		// forbidden job is one of the assigned ones.
		rules = append(rules,
			fmt.Sprintf("Forbidden: on entering the Void, one of your jobs is crossed out and may no longer be used. Which one is already decided - open %s.txt when you get there.",
				forbiddenFilename(res.Slots)))
	case forbiddenPlayer:
		rules = append(rules,
			"Forbidden: on entering the Void, choose one of your jobs to cross out. It may no longer be used.")
	}

	if m.duplicatesAllowed() {
		rules = append(rules, "The same job may be assigned more than once.")
	}

	return rules
}

func hasKind(slots []assignedSlot, kind slotKind) bool {
	for _, s := range slots {
		if s.Kind == kind {
			return true
		}
	}
	return false
}

func allowedOrNot(value bool) string {
	if value {
		return "allowed"
	}
	return "not allowed"
}
