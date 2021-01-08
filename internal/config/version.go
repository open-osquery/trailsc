package config

import "fmt"

var (
	// Version of the binary
	Version = "0.1.0"

	// Build number
	Build = "0"

	// Commit id of the current build
	Commit = "000000"

	// Release type of the build
	Release = "dirty"
)

// GetVersion returns the current version of the binary.
func GetVersion() string {
	return fmt.Sprintf("%s-%s+%s.%s", Version, Release, Build, Commit)
}
