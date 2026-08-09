package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// filenames
const (
	wind  = "01_wind"
	water = "02_water"
	fire  = "03_fire"
	earth = "04_earth"
)

// fileOrder maps each crystal slot to the file its job is written to. Its
// length is the run's slot count, which validateData enforces for every run
// type at startup.
var fileOrder = []string{
	wind,
	water,
	fire,
	earth,
}

// writeFolder creates the run folder and writes one file per crystal,
// containing that crystal's job.
func writeFolder(m run, jobs []string) error {
	if m.name == "" {
		return errors.New("run has no folder name")
	}
	if len(jobs) != len(fileOrder) {
		return fmt.Errorf("expected %d jobs to write, got %d", len(fileOrder), len(jobs))
	}

	if err := os.Mkdir(m.name, 0o755); err != nil && !os.IsExist(err) {
		return err
	}

	for i, filename := range fileOrder {
		if err := writeFile(filename, m.name, jobs[i]); err != nil {
			return err
		}
	}

	return nil
}

func writeFile(filename string, dir string, job string) error {
	return os.WriteFile(filepath.Join(dir, filename+".txt"), []byte(job), 0o644)
}
