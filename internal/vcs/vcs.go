package vcs

import (
	"runtime/debug"
	"strings"
)

// Version returns the version and commit hash of the current build.
func Version() (string, string) {
	var version, commit = "dev-build", "unknown"

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit
	}

	fullVersion := info.Main.Version
	if fullVersion != "" && fullVersion != "(devel)" && !strings.Contains(fullVersion, "-") {
		version = fullVersion
	} else {
		version = strings.SplitN(fullVersion, "-", 2)[0]
	}

	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			commit = s.Value
		}
	}

	if len(commit) > 7 {
		commit = commit[:7]
	}

	return version, commit
}
