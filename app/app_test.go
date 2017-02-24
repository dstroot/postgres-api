package app

import (
	"testing"
)

func TestInitialize(t *testing.T) {

	// This test runs the initialize function and ensures
	// there are no errors returned.
	app, err := Initialize()
	defer app.DB.Close()

	if err != nil {
		t.Errorf("Expected error to be nil. Got '%s'", err)
	}
}
