package main

import (
	"testing"
	"time"
)

func TestRun(t *testing.T) {

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// This test runs the returns function and ensures
	// there are no errors returned.
	go func() {
		err := run()
		if err != nil {
			t.Errorf("Expected error to be nil. Got '%s'", err)
		}
	}()

	time.Sleep(time.Second * 5)
	// c <- syscall.SIGTERM
	// <-c
	// os.Exit(0)
}
