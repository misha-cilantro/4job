package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
)

// dataFS embeds the game data JSON files directly into the binary, so the
// app doesn't depend on being run from any particular working directory.
//
//go:embed data/jobPools.json data/runTypes.json data/jobSets.json data/jobs.json
var dataFS embed.FS

var (
	// Jobs holds every job in the game; JobsByName is the same data keyed
	// by name for fast lookups.
	Jobs       []Job
	JobsByName map[string]Job

	// JobPools holds every named job pool, keyed by position; JobPoolsByName
	// is the same data keyed by name for fast lookups.
	JobPools       []JobPool
	JobPoolsByName map[string]JobPool

	// RunTypes holds every available run type (Normal, Typhoon, Meteor, ...).
	RunTypes       []RunType
	RunTypesByName map[string]RunType

	// JobSets holds every available job set modifier (Team 750, ...);
	// JobSetsByName is the same data keyed by name for fast lookups.
	JobSets       []JobSet
	JobSetsByName map[string]JobSet
)

func init() {
	if err := loadData(); err != nil {
		panic(fmt.Sprintf("4job: failed to load embedded game data: %v", err))
	}
}

// loadData reads and parses the embedded JSON files into the package-level
// variables above, and builds the JobPoolsByName lookup index.
func loadData() error {
	if err := loadJSON("data/jobs.json", &Jobs); err != nil {
		return err
	}
	if err := loadJSON("data/jobPools.json", &JobPools); err != nil {
		return err
	}
	if err := loadJSON("data/runTypes.json", &RunTypes); err != nil {
		return err
	}
	if err := loadJSON("data/jobSets.json", &JobSets); err != nil {
		return err
	}

	RunTypesByName = make(map[string]RunType, len(RunTypes))
	for _, rt := range RunTypes {
		RunTypesByName[rt.Name] = rt
	}

	JobsByName = make(map[string]Job, len(Jobs))
	for _, job := range Jobs {
		JobsByName[job.Name] = job
	}

	JobPoolsByName = make(map[string]JobPool, len(JobPools))
	for _, pool := range JobPools {
		JobPoolsByName[pool.Name] = pool
	}

	JobSetsByName = make(map[string]JobSet, len(JobSets))
	for _, js := range JobSets {
		JobSetsByName[js.Name] = js
	}

	if err := validateData(); err != nil {
		return err
	}

	return nil
}

// AllJobNames returns every job name in the game, sorted alphabetically.
// It's used by steps (like picking excluded jobs) that need the full
// roster rather than a specific pool.
func AllJobNames() []string {
	names := make([]string, 0, len(Jobs))
	for _, job := range Jobs {
		names = append(names, job.Name)
	}
	sort.Strings(names)
	return names
}

// loadJSON reads name from the embedded filesystem and unmarshals it into out.
func loadJSON(name string, out interface{}) error {
	raw, err := dataFS.ReadFile(name)
	if err != nil {
		return fmt.Errorf("reading %s: %w", name, err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("parsing %s: %w", name, err)
	}
	return nil
}

// SpecialJobs returns the nonstandard jobs (Freelancer, Mime). They belong to
// no crystal, so they live in their own pool rather than in one of the four,
// and run types that allow them offer them for every job slot.
func SpecialJobs() []string {
	return JobPoolsByName[specialPoolName].Jobs
}

// IsSpecialJob reports whether a job is one of the nonstandard picks
// (Freelancer, Mime) that only some run types allow.
func IsSpecialJob(name string) bool {
	return slices.Contains(SpecialJobs(), name)
}

// validateData sanity-checks the data files against each other at startup,
// so typos fail loudly here instead of silently narrowing a pool - or
// crashing - at random selection time. It checks that every pool name
// referenced by a run type or job set exists, that every job named inside a
// pool exists, and that each run type defines exactly crystalCount slots.
func validateData() error {
	checkPool := func(context, name string) error {
		if _, ok := JobPoolsByName[name]; !ok {
			return fmt.Errorf("%s references unknown pool %q", context, name)
		}
		return nil
	}

	for _, pool := range JobPools {
		for _, job := range pool.Jobs {
			if _, ok := JobsByName[job]; !ok {
				return fmt.Errorf("pool %q references unknown job %q", pool.Name, job)
			}
		}
	}

	if _, ok := JobPoolsByName[specialPoolName]; !ok {
		return fmt.Errorf("jobPools.json is missing the %q pool", specialPoolName)
	}

	// The special tag and the special pool must agree. The pool decides which
	// jobs the picker treats as nonstandard, so a job tagged special but left
	// out of it would be rolled as an ordinary crystal job.
	for _, job := range Jobs {
		tagged := slices.Contains(job.Tags, tagSpecial)
		pooled := slices.Contains(SpecialJobs(), job.Name)
		if tagged != pooled {
			return fmt.Errorf("job %q has %s tag = %t but membership of the %q pool = %t; the two must agree",
				job.Name, tagSpecial, tagged, specialPoolName, pooled)
		}
	}

	for _, rt := range RunTypes {
		if len(rt.Pools) != crystalCount {
			return fmt.Errorf("run type %q defines %d job slots, expected %d", rt.Name, len(rt.Pools), crystalCount)
		}
		for slot, ref := range rt.Pools {
			for _, name := range ref {
				if err := checkPool(fmt.Sprintf("run type %q, slot %d", rt.Name, slot+1), name); err != nil {
					return err
				}
			}
		}
	}

	for _, js := range JobSets {
		for _, ref := range js.Pools {
			if err := checkPool(fmt.Sprintf("job set %q", js.Name), ref.Pool); err != nil {
				return err
			}
		}
	}

	return validateCombinations()
}

// validateCombinations rejects any run type / job set pairing the wizard can
// offer where some slot has no legal job at all - Onion's earth slot holds
// dragoon, ninja and ranger, none of which are Team 750 jobs. Such a pairing
// still produces a run, because the picker relaxes its way out, but it
// produces one that quietly violates the job set the player chose. Better to
// fail at startup than to ship a data file that can't mean what it says.
//
// Run types that skip job set selection are exempt: the wizard never offers
// them a job set, so an empty intersection there is unreachable.
func validateCombinations() error {
	for _, rt := range RunTypes {
		if rt.NoJobSetSelect {
			continue
		}
		for _, js := range JobSets {
			if err := combinationFeasible(rt, js); err != nil {
				return fmt.Errorf("run type %q cannot be combined with job set %q: %w", rt.Name, js.Name, err)
			}
		}
	}
	return nil
}
