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
	// Four crystal slots, plus one per advanced option that adds a job.
	want := crystalCount + len(m.extraSlotNames())
	if len(res.Slots) != want {
		t.Fatalf("got %d jobs, want %d: %v", len(res.Slots), want, res.Jobs())
	}
	return res
}

func TestSpecialJobsExcludedUnlessAllowed(t *testing.T) {
	m := run{runType: "Normal"}
	for range iterations {
		for _, job := range mustPick(t, m).Jobs() {
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
		for _, job := range mustPick(t, m).Jobs() {
			if IsSpecialJob(job) {
				return
			}
		}
	}
	t.Fatalf("never rolled a special job in %d Meteor runs", iterations)
}

// TestSpecialJobsReachAnyCrystal checks the rule that Freelancer and Mime are
// "available from any Crystal". They used to sit in the wind and water pools,
// which confined Freelancer to Wind slots and Mime to Water ones. Tested at the
// pool level because no run type draws a single crystal per slot *and* allows
// special jobs, so a whole-run roll can't distinguish this.
func TestSpecialJobsReachAnyCrystal(t *testing.T) {
	for _, crystal := range []string{"wind", "water", "fire", "earth"} {
		got := poolIntersection(PoolRef{crystal}, nil, true)
		for _, job := range SpecialJobs() {
			if !slices.Contains(got, job) {
				t.Errorf("%q is not available to a %s slot: %v", job, crystal, got)
			}
		}
	}

	// And absent when the run doesn't allow them.
	for _, crystal := range []string{"wind", "water", "fire", "earth"} {
		for _, job := range poolIntersection(PoolRef{crystal}, nil, false) {
			if IsSpecialJob(job) {
				t.Errorf("%q leaked into a %s slot with special jobs off", job, crystal)
			}
		}
	}
}

// TestOnlyDeclaredRunTypesRollSpecials is the restriction: special jobs are not
// a player option, so a run type that doesn't declare allowSpecial can never
// produce one.
func TestOnlyDeclaredRunTypesRollSpecials(t *testing.T) {
	for _, rt := range RunTypes {
		m := run{runType: rt.Name}

		if m.specialAllowed() != rt.AllowSpecial {
			t.Errorf("%s: specialAllowed() = %t, want %t (the run type decides)",
				rt.Name, m.specialAllowed(), rt.AllowSpecial)
		}
		if rt.AllowSpecial {
			continue
		}

		for range iterations {
			for _, job := range mustPick(t, m).Jobs() {
				if IsSpecialJob(job) {
					t.Fatalf("%s rolled special job %q but doesn't declare allowSpecial", rt.Name, job)
				}
			}
		}
	}
}

// TestDuplicatesForcedWhereRequired pins the two selections that require
// duplicates. Team 375 needs them because it fixes two jobs per side of the 750
// split; Classic because its six-job pool is treated the same way, which the
// wiki implies without stating.
func TestDuplicatesForcedWhereRequired(t *testing.T) {
	if !(run{runType: "Classic"}).duplicatesLocked() {
		t.Error("Classic must force Allow Duplicates on")
	}
	if !(run{runType: "Normal", jobSet: "Team 375"}).duplicatesLocked() {
		t.Error("Team 375 must force Allow Duplicates on")
	}

	// Forcing is a property of the selection, not of the step order, so it must
	// hold whichever run type Team 375 is applied to.
	for _, rt := range RunTypes {
		if rt.NoJobSetSelect {
			continue
		}
		m := run{runType: rt.Name, jobSet: "Team 375"}
		if !m.duplicatesLocked() || !m.duplicatesAllowed() {
			t.Errorf("%s + Team 375 does not force Allow Duplicates", rt.Name)
		}
	}
}

// TestSpecialJobsIgnoreJobSets checks that special jobs are "available
// regardless of Job Sets" - they're in neither the 750 nor the no750 pool, so
// filtering them through a job set would make them unreachable.
func TestSpecialJobsIgnoreJobSets(t *testing.T) {
	m := run{runType: "Meteor", jobSet: "Team 750"}
	if !m.specialAllowed() {
		t.Fatal("Meteor should allow special jobs")
	}

	for range iterations * 4 {
		for _, job := range mustPick(t, m).Jobs() {
			if IsSpecialJob(job) {
				return
			}
		}
	}
	t.Errorf("no special job appeared in %d Meteor + Team 750 runs", iterations*4)
}

func TestSpecialJobsCanStillBeExcluded(t *testing.T) {
	m := run{runType: "Meteor", excludes: SpecialJobs()}
	for range iterations {
		res := mustPick(t, m)
		if len(res.Notes) != 0 {
			t.Fatalf("unexpected relaxation: %v", res.Notes)
		}
		for _, job := range res.Jobs() {
			if IsSpecialJob(job) {
				t.Fatalf("rolled excluded special job %q", job)
			}
		}
	}
}

func TestFifthJobAddsASlotFromTheRunTypesReach(t *testing.T) {
	m := run{runType: "Normal", fifthJob: true}

	// A Normal run reaches every crystal across its four slots, so the fifth
	// job should eventually land on a job from each of them.
	seen := map[string]bool{}
	for range iterations * 2 {
		res := mustPick(t, m)

		last := res.Slots[len(res.Slots)-1]
		if last.Kind != slotFifth {
			t.Fatalf("last slot is kind %d, want slotFifth", last.Kind)
		}
		for _, crystal := range []string{"wind", "water", "fire", "earth"} {
			if slices.Contains(JobPoolsByName[crystal].Jobs, last.Job) {
				seen[crystal] = true
			}
		}
	}

	for _, crystal := range []string{"wind", "water", "fire", "earth"} {
		if !seen[crystal] {
			t.Errorf("the fifth job never came from the %s crystal", crystal)
		}
	}
}

func TestFifthJobRespectsExcludesAndJobSet(t *testing.T) {
	m := run{runType: "Normal", jobSet: "Team 750", fifthJob: true, excludes: []string{"bard"}}
	allowed := JobPoolsByName["750"].Jobs

	for range iterations {
		res := mustPick(t, m)
		if len(res.Notes) != 0 {
			t.Fatalf("unexpected relaxation: %v", res.Notes)
		}

		fifth := res.Slots[len(res.Slots)-1].Job
		if !slices.Contains(allowed, fifth) {
			t.Errorf("fifth job %q is not a Team 750 job", fifth)
		}
		if fifth == "bard" {
			t.Error("fifth job rolled an excluded job")
		}
	}
}

func TestExtraJobsRollsOnlyAdvanceJobs(t *testing.T) {
	advance := JobPoolsByName[advancePoolName].Jobs

	// Team 750 covers none of the Advance jobs, so a job set must not restrict
	// this slot or it would be unfillable.
	for _, jobSet := range []string{"", "Team 750", "Team No 750"} {
		m := run{runType: "Normal", jobSet: jobSet, extraJobs: true}

		seen := map[string]bool{}
		for range iterations {
			res := mustPick(t, m)
			if len(res.Notes) != 0 {
				t.Fatalf("job set %q: unexpected relaxation: %v", jobSet, res.Notes)
			}

			last := res.Slots[len(res.Slots)-1]
			if last.Kind != slotAdvance {
				t.Fatalf("last slot is kind %d, want slotAdvance", last.Kind)
			}
			if !slices.Contains(advance, last.Job) {
				t.Fatalf("job set %q: Advance slot rolled %q, which is not an Advance job", jobSet, last.Job)
			}
			seen[last.Job] = true
		}

		if len(seen) != len(advance) {
			t.Errorf("job set %q: only rolled %d of %d Advance jobs", jobSet, len(seen), len(advance))
		}
	}
}

func TestAdvanceJobsNeverFillCrystalSlots(t *testing.T) {
	advance := JobPoolsByName[advancePoolName].Jobs

	for _, rt := range RunTypes {
		m := run{runType: rt.Name, extraJobs: true}
		for range 50 {
			res := mustPick(t, m)
			for i, s := range res.Slots {
				if s.Kind == slotCrystal && slices.Contains(advance, s.Job) {
					t.Fatalf("%s: Advance job %q filled crystal slot %d", rt.Name, s.Job, i+1)
				}
			}
		}
	}
}

func TestBothExtraSlotsTogether(t *testing.T) {
	m := run{runType: "Normal", fifthJob: true, extraJobs: true}
	res := mustPick(t, m)

	if len(res.Slots) != crystalCount+2 {
		t.Fatalf("got %d slots, want %d", len(res.Slots), crystalCount+2)
	}
	if got := res.Slots[crystalCount].Kind; got != slotFifth {
		t.Errorf("slot %d is kind %d, want slotFifth", crystalCount+1, got)
	}
	if got := res.Slots[crystalCount+1].Kind; got != slotAdvance {
		t.Errorf("slot %d is kind %d, want slotAdvance", crystalCount+2, got)
	}
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
				for _, job := range res.Jobs() {
					if !slices.Contains(allowed, job) {
						t.Fatalf("job %q is not in pool %q (run %v)", job, tc.pool, res.Jobs())
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
		for i, job := range res.Jobs() {
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
			t.Fatalf("got %d jobs from the 750 pool, want 2: %v", count750, res.Jobs())
		}
	}

	// The 2/2 split should land in varying slots, not always the first two.
	if len(seenOrders) != crystalCount {
		t.Errorf("750 jobs only ever landed in slots %v; the split isn't being shuffled", seenOrders)
	}
}

// vibeRoles are the four tags Team Vibe Coded assigns one slot each.
var vibeRoles = []string{"physical", "magic", "support", "combat"}

func hasRole(jobs []string, role string) bool {
	return slices.ContainsFunc(jobs, func(j string) bool {
		return slices.Contains(JobsByName[j].Tags, role)
	})
}

// TestVibeCodedGuaranteesEveryRole is the whole point of the set: it removes the
// healerless and all-caster parties a plain run can hand you.
func TestVibeCodedGuaranteesEveryRole(t *testing.T) {
	for _, rt := range RunTypes {
		if rt.NoJobSetSelect {
			continue // the wizard can't pair a job set with these
		}
		if rt.AllowSpecial {
			continue // see TestVibeCodedYieldsToSpecialJobs
		}

		m := run{runType: rt.Name, jobSet: "Team Vibe Coded"}
		for range 200 {
			res := mustPick(t, m)
			if len(res.Notes) > 0 {
				// A relaxation gave up a constraint and said so in the note.
				continue
			}
			for _, role := range vibeRoles {
				if !hasRole(res.Jobs(), role) {
					t.Fatalf("%s: no %s job in %v", rt.Name, role, res.Jobs())
				}
			}
		}
	}
}

// TestPlainRunsCanLackARole is the baseline the set exists to fix. Without it a
// Normal run leaves you healerless around one time in nine.
func TestPlainRunsCanLackARole(t *testing.T) {
	m := run{runType: "Normal"}
	for range iterations * 2 {
		if !hasRole(mustPick(t, m).Jobs(), "support") {
			return
		}
	}
	t.Error("expected at least one Normal run with no support job; if this stops " +
		"happening, Team Vibe Coded no longer changes anything")
}

// TestVibeCodedYieldsToSpecialJobs records the one hole in the guarantee. Special
// jobs ignore job sets by rule, so on Meteor a Freelancer can land in the slot
// that was meant to cover support or magic.
func TestVibeCodedYieldsToSpecialJobs(t *testing.T) {
	m := run{runType: "Meteor", jobSet: "Team Vibe Coded"}
	if !m.specialAllowed() {
		t.Fatal("Meteor should allow special jobs")
	}

	for range iterations * 4 {
		res := mustPick(t, m)
		if !hasRole(res.Jobs(), "support") {
			return // a special job displaced the support slot, as documented
		}
	}
	t.Log("no special job displaced a role slot in this sample; the guarantee held")
}

func TestExcludedJobsAreAvoidedWhenPossible(t *testing.T) {
	excludes := []string{"knight", "monk", "thief"}
	m := run{runType: "Normal", excludes: excludes}
	for range iterations {
		res := mustPick(t, m)
		if len(res.Notes) != 0 {
			t.Fatalf("unexpected relaxation: %v", res.Notes)
		}
		for _, job := range res.Jobs() {
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
		for i, job := range res.Jobs() {
			if slices.Contains(res.Jobs()[:i], job) {
				t.Fatalf("duplicate job %q without a relaxation note: %v", job, res.Jobs())
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
	if !slices.Contains(JobPoolsByName["earth"].Jobs, res.Jobs()[3]) {
		t.Errorf("earth slot should still hold an earth job, got %q", res.Jobs()[3])
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
	got := feasibleAssignments(onionShaped, JobSetsByName["Team 375"])
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
	if !slices.Contains(JobPoolsByName["onion_earth"].Jobs, res.Jobs()[3]) {
		t.Errorf("slot 4 should fall back to its crystal pool, got %q", res.Jobs()[3])
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
