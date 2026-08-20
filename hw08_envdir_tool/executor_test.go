package main

import "testing"

func TestRunCmd(t *testing.T) {
	code := RunCmd(
		[]string{"go", "version"},
		Environment{},
	)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRunCmd_CommandNotFound(t *testing.T) {
	code := RunCmd(
		[]string{"not-existing-command"},
		Environment{},
	)

	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
}
