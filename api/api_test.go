package api

import (
	"testing"
)

func TestInitialize(t *testing.T) {

	// Create app
	a := App{}

	// This test runs the initialize function and ensures
	// there are no errors returned.
	err := a.Initialize()
	defer a.DB.Close()

	if err != nil {
		t.Errorf("Expected error to be nil. Got '%s'", err)
	}
}
