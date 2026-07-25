package main

import (
	"encoding/json"
	"fmt"
)

// Job describes a single Final Fantasy 5 job and the descriptive tags used
// to group it (e.g. "physical", "magic", "rod_breaker").
type Job struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// JobPool is a named, reusable list of jobs. Everything that needs to
// restrict job selection (run types, job sets) references pools by name
// instead of inlining job lists directly.
type JobPool struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Jobs        []string `json:"jobs"`
}

// PoolRef is a single job slot within a RunType's Pools list. A slot may
// draw from one named pool ("wind") or from the union of several
// (["wind", "water"]). Both shapes unmarshal into the same []string.
type PoolRef []string

func (p *PoolRef) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*p = PoolRef{single}
		return nil
	}

	var multi []string
	if err := json.Unmarshal(data, &multi); err != nil {
		return fmt.Errorf("pool reference must be a string or an array of strings: %w", err)
	}
	*p = PoolRef(multi)
	return nil
}

// RunType describes a Four Job Fiesta run variant: an ordered list of job
// slots (one per crystal/event unlock), each of which draws from one or
// more named JobPools.
type RunType struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Pools       []PoolRef `json:"pools"`

	// AllowSpecial permits otherwise-nonstandard jobs (e.g. Freelancer,
	// Mime) to be rolled, as in Meteor runs.
	AllowSpecial bool `json:"allowSpecial,omitempty"`

	// ForcesAllowDuplicates means this run type always behaves as if the
	// Allow Duplicates modifier were enabled, regardless of user choice.
	ForcesAllowDuplicates bool `json:"forcesAllowDuplicates,omitempty"`

	// NoJobSetSelect means the player may not additionally apply a JobSet
	// on top of this run type (e.g. Classic already fixes its own pool).
	NoJobSetSelect bool `json:"noJobSetSelect,omitempty"`
}

// JobSetPoolRef references a named pool for a JobSet, optionally requiring
// an exact number of the run's job slots to be drawn from that pool.
type JobSetPoolRef struct {
	Pool string

	// Count is the exact number of job slots that must come from Pool.
	// Zero means "no fixed count" - i.e. every job slot may draw from
	// this pool freely (used when a JobSet has only one pool overall).
	Count int
}

func (r *JobSetPoolRef) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		r.Pool = name
		r.Count = 0
		return nil
	}

	var obj struct {
		Pool  string `json:"pool"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("job set pool reference must be a string or a {pool, count} object: %w", err)
	}
	r.Pool = obj.Pool
	r.Count = obj.Count
	return nil
}

// JobSet describes a job-selection filter (Team 750, Team No 750, ...)
// that further restricts which pools a run's jobs may be drawn from.
type JobSet struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Pools       []JobSetPoolRef `json:"pools"`

	// ForcesAllowDuplicates means selecting this job set always behaves
	// as if the Allow Duplicates modifier were enabled.
	ForcesAllowDuplicates bool `json:"forcesAllowDuplicates,omitempty"`
}
