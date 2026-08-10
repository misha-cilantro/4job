package main

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
)

// pickResult is the outcome of rolling a run: one job per slot, plus any notes
// about constraints that had to be relaxed to fill a slot.
type pickResult struct {
	Slots []assignedSlot
	Notes []string

	// Forbidden is the job the Void crosses out, when the Forbidden option is
	// set to roll it now. Empty when the option is off or left to the player.
	Forbidden string
}

// Jobs is the rolled job names in slot order.
func (r pickResult) Jobs() []string {
	out := make([]string, 0, len(r.Slots))
	for _, s := range r.Slots {
		out = append(out, s.Job)
	}
	return out
}

// pickJobs rolls one job for each of the run's slots - the four the run type
// defines, plus a fifth job and an Advance job when those options are on -
// honouring the selected job set, the excluded jobs and the duplicate rule.
func pickJobs(m run) (pickResult, error) {
	rt, ok := RunTypesByName[m.runType]
	if !ok {
		return pickResult{}, fmt.Errorf("unknown run type %q", m.runType)
	}

	restrictions, err := jobSetRestrictions(m, rt)
	if err != nil {
		return pickResult{}, err
	}

	var res pickResult
	roll := func(kind slotKind, pools PoolRef, restriction []string) error {
		position := len(res.Slots)
		job, note, err := pickJobForSlot(m, position, pools, restriction, res.Jobs())
		if err != nil {
			return err
		}
		if note != "" {
			res.Notes = append(res.Notes, note)
		}
		res.Slots = append(res.Slots, assignedSlot{Kind: kind, Job: job})
		return nil
	}

	for slot, pools := range rt.Pools {
		if err := roll(slotCrystal, pools, restrictions[slot]); err != nil {
			return pickResult{}, err
		}
	}

	if m.fifthJob {
		// The rules don't say which crystal a fifth job comes from, so it draws
		// from everything the run type can reach. A counted job set has already
		// spent its slots, so the fifth job may come from either side of it.
		if err := roll(slotFifth, runTypeReach(rt), jobSetReach(m.jobSet)); err != nil {
			return pickResult{}, err
		}
	}

	if m.extraJobs {
		// Advance jobs are in no crystal and no job set, so nothing restricts
		// this slot but the excluded jobs.
		if err := roll(slotAdvance, PoolRef{advancePoolName}, nil); err != nil {
			return pickResult{}, err
		}
	}

	if m.forbidden == forbiddenRolled {
		res.Forbidden = rollForbidden(res.Slots)
	}

	return res, nil
}

// rollForbidden picks which of the run's own jobs the Void crosses out. It
// chooses among the distinct job names rather than the slots, so a duplicated
// job isn't twice as likely to be the one lost - though when it is lost, every
// slot holding it goes with it.
func rollForbidden(slots []assignedSlot) string {
	var distinct []string
	for _, s := range slots {
		if !slices.Contains(distinct, s.Job) {
			distinct = append(distinct, s.Job)
		}
	}
	if len(distinct) == 0 {
		return ""
	}
	return distinct[rand.IntN(len(distinct))]
}

// runTypeReach is every pool the run type can draw from across all its slots.
func runTypeReach(rt RunType) PoolRef {
	var out PoolRef
	for _, ref := range rt.Pools {
		for _, name := range ref {
			if !slices.Contains(out, name) {
				out = append(out, name)
			}
		}
	}
	return out
}

// jobSetReach is every job the named job set permits anywhere, used for slots
// the set doesn't explicitly count. Nil when no job set is selected.
func jobSetReach(name string) []string {
	js, ok := JobSetsByName[name]
	if !ok {
		return nil
	}
	return jobSetUnion(js)
}

// jobSetRestrictions maps each of the run's slots to the list of jobs the
// selected job set permits there. A nil entry means "no restriction", which
// is what every slot gets when no job set is selected.
func jobSetRestrictions(m run, rt RunType) ([][]string, error) {
	out := make([][]string, len(rt.Pools))
	if m.jobSet == "" {
		return out, nil
	}

	js, ok := JobSetsByName[m.jobSet]
	if !ok {
		return nil, fmt.Errorf("unknown job set %q", m.jobSet)
	}

	if jobSetCounted(js) == 0 {
		union := jobSetUnion(js)
		for i := range out {
			out[i] = union
		}
		return out, nil
	}

	if err := checkCountedSlots(rt, js); err != nil {
		return nil, err
	}

	return assignCountedPools(rt, js), nil
}

// jobSetCounted is the total number of slots a job set pins to a specific
// pool. Zero means the job set applies to every slot equally.
func jobSetCounted(js JobSet) int {
	total := 0
	for _, ref := range js.Pools {
		total += ref.Count
	}
	return total
}

// jobSetUnion is every job any of the job set's pools allows.
func jobSetUnion(js JobSet) []string {
	var union []string
	for _, ref := range js.Pools {
		for _, job := range JobPoolsByName[ref.Pool].Jobs {
			if !slices.Contains(union, job) {
				union = append(union, job)
			}
		}
	}
	return union
}

// checkCountedSlots verifies a counted job set pins exactly as many slots as
// the run type has. Anything else is a data error, not something to roll with.
func checkCountedSlots(rt RunType, js JobSet) error {
	if counted := jobSetCounted(js); counted != len(rt.Pools) {
		return fmt.Errorf("job set %q fixes %d job slots but run type %q has %d",
			js.Name, counted, rt.Name, len(rt.Pools))
	}
	return nil
}

// assignCountedPools distributes a counted job set's pools across the run's
// slots - Team 375's two 750 slots and two non-750 slots - and returns the
// per-slot job lists.
//
// The order is random, but chosen only from assignments that leave every slot
// at least one job. A blind shuffle would sometimes pin a pool to a slot that
// can't satisfy it, forcing the ladder to abandon the job set even though a
// workable split existed. If no split works at all, an arbitrary one is
// returned and the ladder relaxes from there; validateCombinations rejects
// that case at startup for any pairing the wizard can actually offer.
func assignCountedPools(rt RunType, js JobSet) [][]string {
	var names []string
	for _, ref := range js.Pools {
		for range ref.Count {
			names = append(names, ref.Pool)
		}
	}

	chosen := names
	if workable := feasibleAssignments(rt, js); len(workable) > 0 {
		chosen = workable[rand.IntN(len(workable))]
	}

	out := make([][]string, len(chosen))
	for i, name := range chosen {
		out[i] = JobPoolsByName[name].Jobs
	}
	return out
}

// feasibleAssignments returns every distinct way to hand a counted job set's
// pools to rt's slots such that no slot ends up with an empty candidate list.
//
// Special jobs are ignored here even on run types that allow them. They're
// exempt from job set restrictions, so counting them would call every
// assignment workable and hide slots that no job set job can fill.
func feasibleAssignments(rt RunType, js JobSet) [][]string {
	var names []string
	for _, ref := range js.Pools {
		for range ref.Count {
			names = append(names, ref.Pool)
		}
	}
	if len(names) != len(rt.Pools) {
		return nil
	}

	var out [][]string
	seen := map[string]bool{}

	permutations(names, func(order []string) {
		key := strings.Join(order, "|")
		if seen[key] {
			return
		}
		seen[key] = true

		for slot, name := range order {
			if len(poolIntersection(rt.Pools[slot], JobPoolsByName[name].Jobs, false)) == 0 {
				return
			}
		}
		out = append(out, slices.Clone(order))
	})

	return out
}

// permutations calls fn with every ordering of items. Orderings repeat when
// items contains duplicate entries, so callers dedupe.
func permutations(items []string, fn func([]string)) {
	var walk func(prefix, rest []string)
	walk = func(prefix, rest []string) {
		if len(rest) == 0 {
			fn(prefix)
			return
		}
		for i := range rest {
			remaining := append(slices.Clone(rest[:i]), rest[i+1:]...)
			walk(append(slices.Clone(prefix), rest[i]), remaining)
		}
	}
	walk(nil, items)
}

// combinationFeasible reports whether every slot of rt can be filled under js
// without abandoning the job set.
//
// Special jobs don't count towards feasibility even on run types that allow
// them, for the reason given on feasibleAssignments: they ignore job sets, so
// including them would mask a slot that no job set job can fill.
//
// For counted job sets it asks whether *any* assignment of the counts to
// slots works, because assignCountedPools is free to choose one.
func combinationFeasible(rt RunType, js JobSet) error {
	if jobSetCounted(js) > 0 {
		if err := checkCountedSlots(rt, js); err != nil {
			return err
		}
		if len(feasibleAssignments(rt, js)) == 0 {
			return fmt.Errorf("no way to distribute its pools across the run's %d job slots leaves every slot a legal job",
				len(rt.Pools))
		}
		return nil
	}

	union := jobSetUnion(js)
	for slot, pools := range rt.Pools {
		if len(poolIntersection(pools, union, false)) == 0 {
			return fmt.Errorf("slot %d (%s) has no job in common with it",
				slot+1, strings.Join(pools, "/"))
		}
	}
	return nil
}

// relaxation is one rung of the fallback ladder used when a slot has no
// legal job left. Rungs are tried in order, loosest last, so a run stays as
// close to the player's chosen constraints as it can.
type relaxation struct {
	allowDuplicates bool
	ignoreExcludes  bool
	ignoreJobSet    bool
	note            string
}

// pickJobForSlot rolls a single slot. It walks the relaxation ladder and
// returns the first rung that yields any candidate, along with a note when
// that rung had to loosen something. Special jobs are never relaxed: if the
// run doesn't allow them, they stay out.
func pickJobForSlot(m run, slot int, pools PoolRef, jobSetJobs, picked []string) (job, note string, err error) {
	ladder := []relaxation{
		{allowDuplicates: m.duplicatesAllowed()},
		{allowDuplicates: true, note: "allowed a duplicate job"},
		{allowDuplicates: true, ignoreExcludes: true, note: "ignored the excluded jobs"},
		{allowDuplicates: true, ignoreExcludes: true, ignoreJobSet: true, note: "ignored the job set"},
	}

	for _, r := range ladder {
		candidates := slotCandidates(m, pools, jobSetJobs, picked, r)
		if len(candidates) == 0 {
			continue
		}
		if r.note != "" {
			note = fmt.Sprintf("slot %d (%s): %s, because nothing else was available",
				slot+1, strings.Join(pools, "/"), r.note)
		}
		return candidates[rand.IntN(len(candidates))], note, nil
	}

	return "", "", fmt.Errorf("no job available for slot %d (%s): every job in those pools is filtered out",
		slot+1, strings.Join(pools, "/"))
}

// poolIntersection returns the jobs a slot may draw considering only the
// constraints that no amount of rerolling can work around: the run type's
// pools, the job set restriction, and the special-job rule. A nil restriction
// means the job set doesn't limit this slot.
//
// Pools that overlap contribute each job once, so a job appearing in two of
// the slot's pools isn't twice as likely to be picked.
func poolIntersection(pools PoolRef, restriction []string, allowSpecial bool) []string {
	var out []string

	add := func(job string) {
		if !slices.Contains(out, job) {
			out = append(out, job)
		}
	}

	for _, poolName := range pools {
		for _, job := range JobPoolsByName[poolName].Jobs {
			// Special jobs belong to no crystal, so they're added below for
			// every slot rather than drawn from the slot's own pools.
			if IsSpecialJob(job) {
				continue
			}
			if restriction != nil && !slices.Contains(restriction, job) {
				continue
			}
			add(job)
		}
	}

	// Freelancer and Mime are available from any crystal when the run type
	// allows them, and the rules exempt them from job set restrictions - a
	// Team 750 Meteor run can still roll them.
	if allowSpecial {
		for _, job := range SpecialJobs() {
			add(job)
		}
	}

	return out
}

// slotCandidates returns every job the slot may legally roll under r, layering
// the per-roll constraints (excludes, duplicates) on top of poolIntersection.
func slotCandidates(m run, pools PoolRef, jobSetJobs, picked []string, r relaxation) []string {
	if r.ignoreJobSet {
		jobSetJobs = nil
	}

	var out []string
	for _, job := range poolIntersection(pools, jobSetJobs, m.specialAllowed()) {
		if !r.ignoreExcludes && slices.Contains(m.excludes, job) {
			continue
		}
		if !r.allowDuplicates && slices.Contains(picked, job) {
			continue
		}
		out = append(out, job)
	}

	return out
}
