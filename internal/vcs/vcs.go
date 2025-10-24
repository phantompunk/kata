package vcs

import (
	"runtime/debug"
	"strings"
)

// Version returns the version and commit hash of the current build.
func Version() (string, string) {
	var version, commit = "dev", "none"

	info, ok := debug.ReadBuildInfo()
	if ok {
		fullVersion := info.Main.Version

		if strings.Contains(fullVersion, "-") {
			parts := strings.Split(fullVersion, "-")
			version = parts[0] + "+dev"
		}

		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				commit = s.Value
			}
		}
	}

	return version, commit
}
