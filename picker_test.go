package main

import (
	"slices"
	"testing"
)

// iterations is how many times each randomised expectation is re-rolled.
// The picker is random, so a single pass proves very little.
const iterations = 500

// mustPick rolls a run and fails the test if the picker errors or returns
// the wrong number of jobs.
func mustPick(t *testing.T, m run) pickResult {
	t.Helper()
	res, err := pickJobs(m)
	if err != nil {
		t.Fatalf("pickJobs(%+v): %v", m, err)
	}
	if len(res.Jobs) != crystalCount {
		t.Fatalf("got %d jobs, want %d: %v", len(res.Jobs), crystalCount, res.Jobs)
	}
	return res
}

func TestSpecialJobsExcludedUnlessAllowed(t *testing.T) {
	m := run{runType: "Normal"}
	for range iterations {
		for _, job := range mustPick(t, m).Jobs {
			if IsSpecialJob(job) {
				t.Fatalf("rolled special job %q with allowSpecial off", job)
			}
		}
	}
}

func TestSpecialJobsRollableWhenRunTypeAllowsThem(t *testing.T) {
	m := run{runType: "Meteor"}
	if !m.specialAllowed() {
		t.Fatal("Meteor should allow special jobs")
	}

	for range iterations {
		for _, job := range mustPick(t, m).Jobs {
			if IsSpecialJob(job) {
				return
			}
		}
	}
	t.Fatalf("never rolled a special job in %d Meteor runs", iterations)
}

func TestJobSetRestrictsEveryJob(t *testing.T) {
	for _, tc := range []struct{ jobSet, pool string }{
		{"Team 750", "750"},
		{"Team No 750", "no750"},
	} {
		t.Run(tc.jobSet, func(t *testing.T) {
			allowed := JobPoolsByName[tc.pool].Jobs
			m := run{runType: "Normal", jobSet: tc.jobSet}
			for range iterations {
				res := mustPick(t, m)
				if len(res.Notes) != 0 {
					t.Fatalf("unexpected relaxation: %v", res.Notes)
				}
				for _, job := range res.Jobs {
					if !slices.Contains(allowed, job) {
						t.Fatalf("job %q is not in pool %q (run %v)", job, tc.pool, res.Jobs)
					}
				}
			}
		})
	}
}

func TestTeam375SplitsTwoAndTwo(t *testing.T) {
	m := run{runType: "Normal", jobSet: "Team 375"}
	if !m.duplicatesLocked() {
		t.Fatal("Team 375 should force Allow Duplicates on")
	}

	seenOrders := map[int]bool{}
	for range iterations {
		res := mustPick(t, m)
		if len(res.Notes) != 0 {
			t.Fatalf("unexpected relaxation: %v", res.Notes)
		}

		count750 := 0
		for i, job := range res.Jobs {
			in750 := slices.Contains(JobPoolsByName["750"].Jobs, job)
			inNo750 := slices.Contains(JobPoolsByName["no750"].Jobs, job)
			if in750 == inNo750 {
				t.Fatalf("job %q should be in exactly one of 750/no750", job)
			}
			if in750 {
				count750++
				seenOrders[i] = true
			}
		}
		if count750 != 2 {
			t.Fatalf("got %d jobs from the 750 pool, want 2: %v", count750, res.Jobs)
		}
	}

	// The 2/2 split should land in varying slots, not always the first two.
	if len(seenOrders) != crystalCount {
		t.Errorf("750 jobs only ever landed in slots %v; the split isn't being shuffled", seenOrders)
	}
}

func TestExcludedJobsAreAvoidedWhenPossible(t *testing.T) {
	excludes := []string{"knight", "monk", "thief"}
	m := run{runType: "Normal", excludes: excludes}
	for range iterations {
		res := mustPick(t, m)
		if len(res.Notes) != 0 {
			t.Fatalf("unexpected relaxation: %v", res.Notes)
		}
		for _, job := range res.Jobs {
			if slices.Contains(excludes, job) {
				t.Fatalf("rolled excluded job %q", job)
			}
		}
	}
}

func TestNoDuplicatesWhenNotAllowed(t *testing.T) {
	// Volcano's later slots draw from progressively narrower pools, so it's
	// the run type most likely to collide.
	m := run{runType: "Volcano"}
	for range iterations {
		res := mustPick(t, m)
		if len(res.Notes) != 0 {
			continue // a relaxation note means a duplicate was unavoidable
		}
		for i, job := range res.Jobs {
			if slices.Contains(res.Jobs[:i], job) {
				t.Fatalf("duplicate job %q without a relaxation note: %v", job, res.Jobs)
			}
		}
	}
}

// TestFullyExcludedPoolTerminates covers the case that used to recurse
// forever: every job a slot could draw is excluded, so no amount of
// duplicate-allowing helps.
func TestFullyExcludedPoolTerminates(t *testing.T) {
	m := run{runType: "Normal", excludes: JobPoolsByName["earth"].Jobs}

	res := mustPick(t, m)
	if len(res.Notes) == 0 {
		t.Fatal("expected a note explaining that excludes were ignored")
	}
	if !slices.Contains(JobPoolsByName["earth"].Jobs, res.Jobs[3]) {
		t.Errorf("earth slot should still hold an earth job, got %q", res.Jobs[3])
	}
}

// onionShaped is a stand-in for the Onion run type from before it was given
// noJobSetSelect. The wizard can no longer offer Onion alongside a job set,
// but the pools are still the clearest example of an infeasible pairing, so
// they're useful for testing the checks that keep it that way.
var onionShaped = RunType{
	Name:  "OnionShaped",
	Pools: []PoolRef{{"onion_wind"}, {"onion_water"}, {"onion_fire"}, {"onion_earth"}},
}

func TestOnionCannotTakeAJobSet(t *testing.T) {
	m := run{runType: "Onion"}
	if !m.jobSetLocked() {
		t.Error("Onion should skip job set selection; its pools can't satisfy Team 750 or Team No 750")
	}
}

func TestCombinationFeasibleRejectsEmptySlot(t *testing.T) {
	// Both single-pool sets need every slot to sit on one side of the 750
	// split, and the Onion pools straddle it.
	for _, name := range []string{"Team 750", "Team No 750"} {
		if err := combinationFeasible(onionShaped, JobSetsByName[name]); err == nil {
			t.Errorf("%s should be infeasible with the Onion pools", name)
		}
	}

	// Team 375 is the exception: it wants two slots per side, and the Onion
	// pools supply exactly two of each. See
	// TestFeasibleAssignmentsSkipsImpossibleSlots for the split it settles on.
	if err := combinationFeasible(onionShaped, JobSetsByName["Team 375"]); err != nil {
		t.Errorf("Team 375 should be feasible with the Onion pools: %v", err)
	}
}

func TestCombinationFeasibleAcceptsRealPairings(t *testing.T) {
	// This duplicates what init() already enforces, but it fails as a test
	// rather than a panic, which is a lot easier to read.
	for _, rt := range RunTypes {
		if rt.NoJobSetSelect {
			continue
		}
		for _, js := range JobSets {
			if err := combinationFeasible(rt, js); err != nil {
				t.Errorf("%s + %s: %v", rt.Name, js.Name, err)
			}
		}
	}
}

// TestFeasibleAssignmentsSkipsImpossibleSlots checks that a counted job set
// only ever considers splits that work. With the Onion pools, onion_fire has
// no non-750 job and onion_earth has no 750 job, so slots 3 and 4 are forced
// and only the first two slots may vary.
func TestFeasibleAssignmentsSkipsImpossibleSlots(t *testing.T) {
	got := feasibleAssignments(onionShaped, JobSetsByName["Team 375"], false)
	if len(got) == 0 {
		t.Fatal("expected at least one workable split")
	}

	for _, order := range got {
		if order[2] != "750" {
			t.Errorf("slot 3 assigned %q, but onion_fire holds only 750 jobs", order[2])
		}
		if order[3] != "no750" {
			t.Errorf("slot 4 assigned %q, but onion_earth holds no 750 jobs", order[3])
		}
	}

	// Slots 1 and 2 can take either pool, so both orders should be offered.
	if len(got) != 2 {
		t.Errorf("got %d workable splits, want 2: %v", len(got), got)
	}
}

// TestImpossibleJobSetCombinationTerminates is the backstop: the wizard can't
// produce this pairing any more, but if data changes ever reintroduce one, the
// picker must relax rather than hang.
func TestImpossibleJobSetCombinationTerminates(t *testing.T) {
	m := run{runType: "Onion", jobSet: "Team 750"}

	res := mustPick(t, m)
	if len(res.Notes) == 0 {
		t.Fatal("expected a note explaining that the job set was ignored")
	}
	if !slices.Contains(JobPoolsByName["onion_earth"].Jobs, res.Jobs[3]) {
		t.Errorf("slot 4 should fall back to its crystal pool, got %q", res.Jobs[3])
	}
}

func TestEveryRunTypeAndJobSetCombinationRolls(t *testing.T) {
	jobSets := []string{""}
	for _, js := range JobSets {
		jobSets = append(jobSets, js.Name)
	}

	for _, rt := range RunTypes {
		for _, js := range jobSets {
			m := run{runType: rt.Name, jobSet: js}
			if m.jobSetLocked() && js != "" {
				continue // the wizard can't produce this combination
			}
			for range 50 {
				mustPick(t, m)
			}
		}
	}
}

func TestUnknownRunTypeErrors(t *testing.T) {
	if _, err := pickJobs(run{}); err == nil {
		t.Fatal("expected an error for an empty run type")
	}
}
