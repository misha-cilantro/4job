package main

import (
	"embed"
	"encoding/json"
	"fmt"
)

// dataFS embeds the game data JSON files directly into the binary, so the
// app doesn't depend on being run from any particular working directory.
//
//go:embed data/jobPools.json data/runTypes.json data/jobSets.json
var dataFS embed.FS

var (
	// JobPools holds every named job pool, keyed by position; JobPoolsByName
	// is the same data keyed by name for fast lookups.
	JobPools       []JobPool
	JobPoolsByName map[string]JobPool

	// RunTypes holds every available run type (Normal, Typhoon, Meteor, ...).
	RunTypes []RunType

	// JobSets holds every available job set modifier (Team 750, ...).
	JobSets []JobSet
)

func init() {
	if err := loadData(); err != nil {
		panic(fmt.Sprintf("4job: failed to load embedded game data: %v", err))
	}
}

// loadData reads and parses the embedded JSON files into the package-level
// variables above, and builds the JobPoolsByName lookup index.
func loadData() error {
	if err := loadJSON("data/jobPools.json", &JobPools); err != nil {
		return err
	}
	if err := loadJSON("data/runTypes.json", &RunTypes); err != nil {
		return err
	}
	if err := loadJSON("data/jobSets.json", &JobSets); err != nil {
		return err
	}

	JobPoolsByName = make(map[string]JobPool, len(JobPools))
	for _, pool := range JobPools {
		JobPoolsByName[pool.Name] = pool
	}

	if err := validateData(); err != nil {
		return err
	}

	return nil
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

// validateData sanity-checks that every pool name referenced by run types
// and job sets actually exists in jobPools.json. This catches typos in the
// data files at startup instead of failing silently at random selection time.
func validateData() error {
	checkPool := func(context, name string) error {
		if _, ok := JobPoolsByName[name]; !ok {
			return fmt.Errorf("%s references unknown pool %q", context, name)
		}
		return nil
	}

	for _, rt := range RunTypes {
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

	return nil
}
