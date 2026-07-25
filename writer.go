package main

import (
	"log"
	"os"
)

// filenames
const (
	wind  = "01_wind"
	water = "02_water"
	fire  = "03_fire"
	earth = "04_earth"
)

func writeFolder(m run, jobs []string) {
	var fileOrder = []string{
		wind,
		water,
		fire,
		earth,
	}

	err := os.Mkdir(m.name, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	for i, filename := range fileOrder {
		err := writeFile(filename, m.name, jobs[i])
		if err != nil {
			log.Fatal(err)
		}
	}
}

func writeFile(filename string, dir string, job string) error {
	return os.WriteFile(dir+"/"+filename+".txt", []byte(job), 02)
}
