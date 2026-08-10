package main

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// crystalSlots turns job names into the crystal slots a plain run rolls.
func crystalSlots(jobs ...string) []assignedSlot {
	out := make([]assignedSlot, 0, len(jobs))
	for _, job := range jobs {
		out = append(out, assignedSlot{Kind: slotCrystal, Job: job})
	}
	return out
}

// crystalRun is a pickResult holding nothing but crystal slots.
func crystalRun(jobs ...string) pickResult {
	return pickResult{Slots: crystalSlots(jobs...)}
}

// readRun reads back every file in a written run folder, keyed by file name.
func readRun(t *testing.T, dir, name string) map[string]string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("reading run folder: %v", err)
	}

	out := map[string]string{}
	for _, e := range entries {
		body, err := os.ReadFile(filepath.Join(dir, name, e.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		out[e.Name()] = string(body)
	}
	return out
}

func TestWriteFolderWritesReadableFilesInOrder(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	jobs := []string{"knight", "red mage", "ninja", "dragoon"}
	m := run{name: "test-run", runType: "Normal"}
	if err := writeFolder(m, crystalRun(jobs...)); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	for i, crystal := range crystalFiles {
		filename := slotFilename(assignedSlot{Kind: slotCrystal}, i) + ".txt"
		path := filepath.Join(dir, m.name, filename)

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", filename, err)
		}
		// A plain run has no rules to add, so the file is just the job name -
		// which keeps it usable as a one-line stream source.
		if string(got) != jobs[i] {
			t.Errorf("%s contains %q, want %q", filename, got, jobs[i])
		}
		if !strings.Contains(filename, crystal) {
			t.Errorf("%s should be named after the %s crystal", filename, crystal)
		}

		// The old mode of 02 produced write-only files that couldn't be
		// read back by anything (OBS, the player, this test).
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o400 == 0 {
			t.Errorf("%s has mode %v, which is not owner-readable", filename, info.Mode().Perm())
		}
	}
}

func TestWriteFolderRejectsBadInput(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := writeFolder(run{name: ""}, crystalRun("a", "b", "c", "d")); err == nil {
		t.Error("expected an error for an empty run name")
	}
	if err := writeFolder(run{name: "short"}, crystalRun("a", "b")); err == nil {
		t.Error("expected an error for too few jobs")
	}
}

func TestWriteFolderAlwaysWritesRules(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := run{name: "rules-run", runType: "Normal", jobSet: "Team 750", excludes: []string{"bard"}}
	if err := writeFolder(m, crystalRun("white mage", "red mage", "geomancer", "chemist")); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	files := readRun(t, dir, m.name)
	rules, ok := files[rulesFile+".txt"]
	if !ok {
		t.Fatalf("no %s.txt written; got %v", rulesFile, fileNames(files))
	}

	for _, want := range []string{"Normal run", "Team 750", "no restrictions", "bard"} {
		if !strings.Contains(rules, want) {
			t.Errorf("rules file does not mention %q:\n%s", want, rules)
		}
	}
}

// TestRulesFileSpoilsNothing is the point of the rules file: it's read before the
// run starts, so it must not name a single assigned job. It used to list the
// whole roll, and the rolled Forbidden rule named its job too.
func TestRulesFileSpoilsNothing(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// Assigned jobs and excluded jobs are kept disjoint, since the excludes do
	// legitimately appear in the file - the player chose them.
	assigned := []string{"knight", "red mage", "ninja", "dragoon"}
	m := run{name: "spoiler-run", runType: "Normal", jobSet: "Team 375",
		excludes:    []string{"bard", "chemist"},
		restriction: restrictNatural,
		fifthJob:    true,
		extraJobs:   true,
		forbidden:   forbiddenRolled,
	}
	slots := append(crystalSlots(assigned...),
		assignedSlot{Kind: slotFifth, Job: "summoner"},
		assignedSlot{Kind: slotAdvance, Job: "oracle"})
	res := pickResult{Slots: slots, Forbidden: "ninja"}

	if err := writeFolder(m, res); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	rules := readRun(t, dir, m.name)[rulesFile+".txt"]
	for _, job := range append(res.Jobs(), res.Forbidden) {
		if strings.Contains(rules, job) {
			t.Errorf("rules file names the assigned job %q:\n%s", job, rules)
		}
	}

	// It should still say what kind of run it is.
	for _, want := range []string{"Normal run", "Team 375", "natural jobs", "Forbidden"} {
		if !strings.Contains(rules, want) {
			t.Errorf("rules file no longer mentions %q:\n%s", want, rules)
		}
	}
}

func TestInstructionsListEveryFileInOrder(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := run{name: "instr-run", runType: "Normal", fifthJob: true, extraJobs: true,
		forbidden: forbiddenPlayer}
	slots := append(crystalSlots("knight", "red mage", "ninja", "dragoon"),
		assignedSlot{Kind: slotFifth, Job: "bard"},
		assignedSlot{Kind: slotAdvance, Job: "oracle"})

	if err := writeFolder(m, pickResult{Slots: slots}); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	files := readRun(t, dir, m.name)
	body, ok := files[instructionsFile+".txt"]
	if !ok {
		t.Fatalf("no %s.txt written; got %v", instructionsFile, fileNames(files))
	}

	// Every other file in the run should be listed, and nothing else should be.
	for name := range files {
		if name == instructionsFile+".txt" {
			continue
		}
		if !strings.Contains(body, name) {
			t.Errorf("instructions do not mention %s:\n%s", name, body)
		}
	}

	// Listed in the order they're opened, which is the order they sort in.
	var lastAt int
	for _, name := range fileNames(files) {
		if name == instructionsFile+".txt" {
			continue
		}
		at := strings.Index(body, name)
		if at < lastAt {
			t.Errorf("%s is listed out of order:\n%s", name, body)
		}
		lastAt = at
	}

	for _, want := range []string{"Wind Shrine", "Walse Tower", "Karnak", "Ronka",
		"Krile joins", "legendary weapons", "the Void"} {
		if !strings.Contains(body, want) {
			t.Errorf("instructions do not say when to open the %q file:\n%s", want, body)
		}
	}
}

func TestInstructionsOmitFilesTheRunDoesNotHave(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := run{name: "plain-run", runType: "Normal"}
	if err := writeFolder(m, crystalRun("knight", "red mage", "ninja", "dragoon")); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	body := readRun(t, dir, m.name)[instructionsFile+".txt"]
	for _, absent := range []string{"krile", "advance", "forbidden"} {
		if strings.Contains(body, absent) {
			t.Errorf("instructions mention %q, which this run has no file for:\n%s", absent, body)
		}
	}
}

func TestInstructionsExplainUpgradeTiming(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := run{name: "upgrade-run", runType: "Normal", restriction: restrictUpgrade}
	if err := writeFolder(m, crystalRun("knight", "red mage", "ninja", "dragoon")); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	body := readRun(t, dir, m.name)[instructionsFile+".txt"]
	if !strings.Contains(body, "Upgrade Jobs") {
		t.Errorf("upgrade runs unlock at the player's pace; the instructions should say so:\n%s", body)
	}

	// Other restriction modes read the files at the crystals, so they need no note.
	m2 := run{name: "natural-run", runType: "Normal", restriction: restrictNatural}
	if err := writeFolder(m2, crystalRun("knight", "red mage", "ninja", "dragoon")); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}
	if body := readRun(t, dir, m2.name)[instructionsFile+".txt"]; strings.Contains(body, "Upgrade Jobs") {
		t.Errorf("natural run should not carry the upgrade note:\n%s", body)
	}
}

func TestFifthJobWritesKrileFileWithItsRules(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := run{name: "fifth-run", runType: "Normal", fifthJob: true}
	slots := append(crystalSlots("knight", "red mage", "ninja", "dragoon"),
		assignedSlot{Kind: slotFifth, Job: "bard"})

	if err := writeFolder(m, pickResult{Slots: slots}); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	files := readRun(t, dir, m.name)
	body, ok := files["05_krile.txt"]
	if !ok {
		t.Fatalf("no 05_krile.txt written; got %v", fileNames(files))
	}

	lines := strings.Split(body, "\n")
	if lines[0] != "bard" {
		t.Errorf("first line is %q, want the job name %q", lines[0], "bard")
	}
	for _, want := range []string{
		"You must choose one of your previous Jobs to no longer use. This includes that Job's Abilities.",
		"You may not swap between all five Jobs. Only use four.",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("05_krile.txt is missing the rule %q:\n%s", want, body)
		}
	}
}

func TestNaturalJobsWritesCharacterRules(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := run{name: "natural-run", runType: "Normal", restriction: restrictNatural}
	if err := writeFolder(m, crystalRun("black mage", "red mage", "ninja", "dragoon")); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	files := readRun(t, dir, m.name)

	// The shape the request asked for: "black mage\nBartz must always be a black mage".
	if got, want := files["01_wind.txt"], "black mage\nBartz must always be a black mage."; !strings.HasPrefix(got, want) {
		t.Errorf("01_wind.txt = %q, want it to start %q", got, want)
	}
	for file, character := range map[string]string{
		"02_water.txt": "Lenna",
		"03_fire.txt":  "Faris",
		"04_earth.txt": "Galuf",
	} {
		if !strings.Contains(files[file], character) {
			t.Errorf("%s does not name %s:\n%s", file, character, files[file])
		}
	}

	// Krile inherits Galuf's job when there's no fifth job to give her.
	if !strings.Contains(files["04_earth.txt"], "Krile") {
		t.Errorf("04_earth.txt should say Krile shares Galuf's job:\n%s", files["04_earth.txt"])
	}
}

func TestNaturalJobsWithFifthJobGivesKrileHerOwn(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := run{name: "natural-fifth", runType: "Normal", restriction: restrictNatural, fifthJob: true}
	slots := append(crystalSlots("knight", "red mage", "ninja", "dragoon"),
		assignedSlot{Kind: slotFifth, Job: "bard"})

	if err := writeFolder(m, pickResult{Slots: slots}); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	files := readRun(t, dir, m.name)
	if strings.Contains(files["04_earth.txt"], "Krile") {
		t.Errorf("04_earth.txt should not claim Krile when she has her own job:\n%s", files["04_earth.txt"])
	}
	if !strings.Contains(files["05_krile.txt"], "Krile must always be a bard.") {
		t.Errorf("05_krile.txt should lock Krile to her own job:\n%s", files["05_krile.txt"])
	}
}

// TestAdvanceJobFileNumbering checks the Advance job takes slot 5 on its own and
// slot 6 alongside a fifth job, so the numbers never skip.
func TestAdvanceJobFileNumbering(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fifth    bool
		wantFile string
	}{
		{"advance only", false, "05_advance.txt"},
		{"fifth and advance", true, "06_advance.txt"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)

			m := run{name: "adv-run", runType: "Normal", extraJobs: true, fifthJob: tc.fifth}
			slots := crystalSlots("knight", "red mage", "ninja", "dragoon")
			if tc.fifth {
				slots = append(slots, assignedSlot{Kind: slotFifth, Job: "bard"})
			}
			slots = append(slots, assignedSlot{Kind: slotAdvance, Job: "oracle"})

			if err := writeFolder(m, pickResult{Slots: slots}); err != nil {
				t.Fatalf("writeFolder: %v", err)
			}

			files := readRun(t, dir, m.name)
			body, ok := files[tc.wantFile]
			if !ok {
				t.Fatalf("no %s written; got %v", tc.wantFile, fileNames(files))
			}
			if !strings.HasPrefix(body, "oracle\n") {
				t.Errorf("%s should start with the job name:\n%s", tc.wantFile, body)
			}
			if !strings.Contains(body, "Advance job") {
				t.Errorf("%s should explain it's an Advance job:\n%s", tc.wantFile, body)
			}
		})
	}
}

// fileNames is the sorted file list of a read-back run folder, for failure
// messages that need to say what was actually written.
func fileNames(files map[string]string) []string {
	return slices.Sorted(maps.Keys(files))
}
