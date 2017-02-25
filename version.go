package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
)

var (
	// commit can be filled in by the compiler using ldflags:
	// go build -ldflags "-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.commit=`git rev-parse HEAD` -w -s"
	commit string

	// buildstamp can be filled in by the compiler using ldflags:
	// go build -ldflags "-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.commit=`git rev-parse HEAD` -w -s"
	buildstamp string
)

// formattedVersion returns a formatted version string which includes
// the git commit and development information.
func formattedVersion() string {

	path := strings.Split(os.Args[0], "/")
	name := strings.Title(path[len(path)-1])
	gver := runtime.Version()

	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "Running: %s\n", name)
	fmt.Fprintf(&versionString, "  - Built with %s\n", gver)

	if commit != "" {
		fmt.Fprintf(&versionString, "  - Commit %s\n", commit)
	}

	if buildstamp != "" {
		fmt.Fprintf(&versionString, "  - Built at %s\n", buildstamp)
	}

	return versionString.String()
}
