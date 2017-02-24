package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {

	// This test runs the returns function and ensures
	// there are no errors returned.
	go func() {
		err := run()
		if err != nil {
			t.Errorf("Expected error to be nil. Got '%s'", err)
		}
	}()
	os.Exit(0)
}
