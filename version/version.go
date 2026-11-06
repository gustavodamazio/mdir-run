// Package version exposes the application version.
package version

import "runtime/debug"

// Version is injected at build time, e.g.:
//
//	go build -ldflags "-X=github.com/gustavodamazio/mdir-run/version.Version=v1.2.3"
//
// Builds made with fyne-cross can't use -ldflags reliably, so they rely on the
// fallback below.
var Version = ""

// Get returns the injected version, falling back to the main module version
// recorded by the Go toolchain (set by `go install module@vX.Y.Z`, or from the
// git tag via VCS stamping in Go 1.24+), and finally "dev".
func Get() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}
