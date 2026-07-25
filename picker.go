package main

import (
	"math/rand/v2"
	"slices"
)

func pickJobs(m run) []string {
	var jobs []string
	var rt RunType = RunTypesByName[m.runType]

	for _, pools := range rt.Pools {
		job := pickJobFromPools(m, pools, jobs, m.allowDuplicates)
		jobs = append(jobs, job)
	}

	return jobs
}

func pickJobFromPools(m run, pools PoolRef, pickedJobs []string, allowDupes bool) string {
	var possibleJobs []string

	for _, poolName := range pools {
		poolJobs := JobPoolsByName[poolName].Jobs
		for _, job := range poolJobs {
			if slices.Contains(m.excludes, job) {
				continue
			}

			if slices.Contains(pickedJobs, job) && !allowDupes {
				continue
			}

			possibleJobs = append(possibleJobs, job)
		}
	}

	if len(possibleJobs) == 0 {
		return pickJobFromPools(m, pools, pickedJobs, true)
	}

	randomIndex := rand.IntN(len(possibleJobs))
	return possibleJobs[randomIndex]
}
