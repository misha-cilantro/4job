package main

import (
	"slices"
	"testing"
)

// wikiPools is the published run type table, transcribed from from-wiki.md.
// It's the authority for what the data files should say, so a data edit that
// drifts from the rules fails here.
var wikiPools = map[string][][]string{
	"Normal":  {{"wind"}, {"water"}, {"fire"}, {"earth"}},
	"Typhoon": {{"wind"}, {"wind", "water"}, {"wind", "water", "fire"}, {"wind", "water", "fire", "earth"}},
	"Geyser":  {{"wind", "water"}, {"wind", "water", "earth"}, {"water", "fire", "earth"}, {"water", "earth"}},
	"Sirocco": {{"wind", "fire"}, {"wind", "water", "fire"}, {"wind", "fire", "earth"}, {"fire", "earth"}},
	"Onsen":   {{"wind", "water"}, {"water", "fire"}, {"water", "fire", "earth"}, {"water", "fire", "earth"}},
	"Haboob":  {{"wind", "earth"}, {"wind", "water", "earth"}, {"wind", "fire", "earth"}, {"fire", "earth"}},
	"Volcano": {{"wind", "water", "fire", "earth"}, {"water", "fire", "earth"}, {"fire", "earth"}, {"earth"}},
	"Meteor": {
		{"wind", "water", "fire", "earth"},
		{"wind", "water", "fire", "earth"},
		{"wind", "water", "fire", "earth"},
		{"wind", "water", "fire", "earth"},
	},
}

func TestRunTypePoolsMatchPublishedRules(t *testing.T) {
	for name, want := range wikiPools {
		rt, ok := RunTypesByName[name]
		if !ok {
			t.Errorf("run type %q is missing from runTypes.json", name)
			continue
		}

		for slot := range want {
			// A slot's pools are a union, so order within it doesn't matter.
			got := slices.Clone([]string(rt.Pools[slot]))
			slices.Sort(got)
			expect := slices.Clone(want[slot])
			slices.Sort(expect)

			if !slices.Equal(got, expect) {
				t.Errorf("%s slot %d draws from %v, published rule is %v", name, slot+1, got, expect)
			}
		}
	}
}

// wikiCrystalJobs is the per-crystal roster from the wiki Jobs page. Freelancer
// and Mime are deliberately absent: neither belongs to a crystal.
var wikiCrystalJobs = map[string][]string{
	"wind":  {"knight", "monk", "thief", "white mage", "black mage", "blue mage"},
	"water": {"berserker", "mystic knight", "time mage", "summoner", "red mage"},
	"fire":  {"ninja", "ranger", "beastmaster", "geomancer", "bard"},
	"earth": {"dragoon", "samurai", "chemist", "dancer"},
}

func TestCrystalPoolsMatchPublishedRules(t *testing.T) {
	for name, want := range wikiCrystalJobs {
		got := slices.Clone(JobPoolsByName[name].Jobs)
		slices.Sort(got)
		expect := slices.Clone(want)
		slices.Sort(expect)

		if !slices.Equal(got, expect) {
			t.Errorf("pool %q holds %v, published roster is %v", name, got, expect)
		}
	}
}

func TestSpecialJobsAreNotCrystalJobs(t *testing.T) {
	for crystal := range wikiCrystalJobs {
		for _, job := range JobPoolsByName[crystal].Jobs {
			if IsSpecialJob(job) {
				t.Errorf("special job %q is in the %q crystal pool; it belongs to no crystal", job, crystal)
			}
		}
	}

	if want := []string{"freelancer", "mime"}; !slices.Equal(SpecialJobs(), want) {
		t.Errorf("SpecialJobs() = %v, want %v", SpecialJobs(), want)
	}
}

func TestJobSetPoolsMatchPublishedRules(t *testing.T) {
	want := map[string][]string{
		"750":   {"white mage", "black mage", "blue mage", "time mage", "summoner", "red mage", "geomancer", "bard", "chemist", "dancer"},
		"no750": {"thief", "monk", "knight", "mystic knight", "berserker", "ninja", "beastmaster", "ranger", "samurai", "dragoon"},
	}

	for name, expect := range want {
		got := slices.Clone(JobPoolsByName[name].Jobs)
		slices.Sort(got)
		slices.Sort(expect)

		if !slices.Equal(got, expect) {
			t.Errorf("pool %q holds %v, published roster is %v", name, got, expect)
		}
	}

	// Together the two sets should cover every crystal job exactly once, and
	// neither should contain a special job.
	var covered []string
	for _, name := range []string{"750", "no750"} {
		covered = append(covered, JobPoolsByName[name].Jobs...)
	}
	var crystal []string
	for _, jobs := range wikiCrystalJobs {
		crystal = append(crystal, jobs...)
	}
	slices.Sort(covered)
	slices.Sort(crystal)
	if !slices.Equal(covered, crystal) {
		t.Errorf("the 750/no750 split covers %v, want exactly the crystal jobs %v", covered, crystal)
	}
}
