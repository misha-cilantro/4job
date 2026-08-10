package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// crystalFiles names the four crystal slots in unlock order.
var crystalFiles = []string{"wind", "water", "fire", "earth"}

// rulesFile holds the run's settings and the rules that apply to the whole run
// rather than to one job. Numbered 00 so it sorts above the jobs.
const rulesFile = "00_rules"

// slotFilename is the file a slot's job is written to, without its extension.
// Numbering is sequential over the slots actually rolled, so an Advance job
// lands on 05 when there's no fifth job and 06 when there is.
func slotFilename(s assignedSlot, position int) string {
	base := "job"
	switch s.Kind {
	case slotCrystal:
		if position < len(crystalFiles) {
			base = crystalFiles[position]
		}
	case slotFifth:
		base = "krile"
	case slotAdvance:
		base = "advance"
	}
	return fmt.Sprintf("%02d_%s", position+1, base)
}

// checkFolderName verifies the run folder name is usable. generateName
// sanitizes it already, but the folder is created relative to the working
// directory, so this confirms it really is a single path element before
// anything touches the disk.
func checkFolderName(name string) error {
	if name == "" {
		return errors.New("run has no folder name")
	}
	if name != filepath.Base(name) || name == "." || name == ".." {
		return fmt.Errorf("run folder name %q is not a plain folder name", name)
	}
	return nil
}

// writeFolder creates the run folder and writes one file per assigned job, plus
// the rules file and, when the Forbidden option is on, the forbidden file.
func writeFolder(m run, res pickResult) error {
	if err := checkFolderName(m.name); err != nil {
		return err
	}
	if len(res.Slots) < crystalCount {
		return fmt.Errorf("expected at least %d jobs to write, got %d", crystalCount, len(res.Slots))
	}

	if err := os.Mkdir(m.name, 0o755); err != nil && !os.IsExist(err) {
		return err
	}

	if err := writeFile(rulesFile, m.name, runRules(m, res)); err != nil {
		return err
	}

	for i, s := range res.Slots {
		// The job name stays on the first line so a one-line stream source
		// still reads correctly; any rules follow underneath.
		body := s.Job
		if notes := slotNotes(m, res, i); len(notes) > 0 {
			body += "\n" + strings.Join(notes, "\n")
		}
		if err := writeFile(slotFilename(s, i), m.name, body); err != nil {
			return err
		}
	}

	if m.forbidden != forbiddenOff {
		if err := writeFile(forbiddenFilename(res.Slots), m.name, forbiddenBody(m, res)); err != nil {
			return err
		}
	}

	return nil
}

// forbiddenFilename numbers the forbidden file after the job slots, so it lands
// on 05 for a plain run and 07 when both extra job options are on. Forbidden
// takes a job away rather than adding one, so it isn't a slot itself.
func forbiddenFilename(slots []assignedSlot) string {
	return fmt.Sprintf("%02d_forbidden", len(slots)+1)
}

func writeFile(filename string, dir string, body string) error {
	return os.WriteFile(filepath.Join(dir, filename+".txt"), []byte(body), 0o644)
}
