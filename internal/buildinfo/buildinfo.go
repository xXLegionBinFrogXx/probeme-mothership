// Package buildinfo carries version metadata injected at link time.
package buildinfo

import "runtime"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func GoVersion() string { return runtime.Version() }

func String() string {
	return "probeme-mothership " + Version +
		" (commit " + Commit + ", built " + Date + ", " + GoVersion() + ")"
}
