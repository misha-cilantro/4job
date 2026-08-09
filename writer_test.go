package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFolderWritesReadableFilesInOrder(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	jobs := []string{"knight", "red mage", "ninja", "dragoon"}
	m := run{name: "test-run"}
	if err := writeFolder(m, jobs); err != nil {
		t.Fatalf("writeFolder: %v", err)
	}

	for i, filename := range fileOrder {
		path := filepath.Join(dir, m.name, filename+".txt")

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading back %s: %v", filename, err)
		}
		if string(got) != jobs[i] {
			t.Errorf("%s contains %q, want %q", filename, got, jobs[i])
		}

		// The old mode of 02 produced write-only files that couldn't be
		// read back by anything (OBS, the player, this test).
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o400 == 0 {
			t.Errorf("%s has mode %v, which is not owner-readable", filename, info.Mode().Perm())
		}
	}
}

func TestWriteFolderRejectsBadInput(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := writeFolder(run{name: ""}, []string{"a", "b", "c", "d"}); err == nil {
		t.Error("expected an error for an empty run name")
	}
	if err := writeFolder(run{name: "short"}, []string{"a", "b"}); err == nil {
		t.Error("expected an error for too few jobs")
	}
}
