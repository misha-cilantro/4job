package main

import "fmt"

/*
 * Steps:
 * 1. Pick a run type
 * 2. Pick a job set, or skip if run type has noJobSetSelect == true
 * 3. Optionally add excluded jobs (so you can prevent recent repeats)
 * 4. Options: allowDuplicates (locked if forcesAllowDuplicates == true); allow rolling special jobs (Freelancer and Mime) (locked if allowSpecial == true))
 */

type run struct {
	step            int
	runType         string
	jobSetLocked    bool
	jobSet          string
	excludes        []string
	allowDuplicates bool
}

func main() {
	fmt.Printf("Loaded %d job pools, %d run types, %d job sets\n",
		len(JobPools), len(RunTypes), len(JobSets))

	for _, rt := range RunTypes {
		fmt.Printf("  - %s: %d job slots\n", rt.Name, len(rt.Pools))
	}
}
