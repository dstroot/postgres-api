package main

import "testing"

func TestFormattedVersion(t *testing.T) {

	version := formattedVersion()
	if version == "" {
		t.Errorf("Expected version not be empty. Got %s", version)
	}

}
