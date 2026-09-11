// Package buildinfo exposes public, immutable image identity. It never reads
// runtime configuration, credentials, hostnames or account data.
package buildinfo

import (
	"regexp"
	"time"
)

// Values are set by the image build through Go linker flags.
var (
	Version     = "unknown"
	Commit      = "unknown"
	Environment = "unknown"
	BuiltAt     = "unknown"
)

type Info struct {
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	Environment string `json:"environment"`
	BuiltAt     string `json:"builtAt"`
}

var versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?$`)
var commitPattern = regexp.MustCompile(`^[a-f0-9]{40}$`)

func Current() Info {
	info := Info{Version: "unknown", Commit: "unknown", Environment: "unknown", BuiltAt: "unknown"}
	if len(Version) <= 64 && versionPattern.MatchString(Version) {
		info.Version = Version
	}
	if commitPattern.MatchString(Commit) {
		info.Commit = Commit
	}
	switch Environment {
	case "development", "test", "staging", "production":
		info.Environment = Environment
	}
	if parsed, err := time.Parse(time.RFC3339, BuiltAt); err == nil {
		info.BuiltAt = parsed.UTC().Format(time.RFC3339)
	}
	return info
}
