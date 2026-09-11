package buildinfo

import "testing"

func TestBuildIdentityRejectsUntrustedOrIncompleteValues(t *testing.T) {
	oldVersion, oldCommit, oldEnvironment, oldBuiltAt := Version, Commit, Environment, BuiltAt
	t.Cleanup(func() { Version, Commit, Environment, BuiltAt = oldVersion, oldCommit, oldEnvironment, oldBuiltAt })
	Version, Commit, Environment, BuiltAt = "0.2.0", "0123456789012345678901234567890123456789", "staging", "2026-09-12T10:00:00+03:30"
	got := Current()
	if got.Version != Version || got.Commit != Commit || got.Environment != "staging" || got.BuiltAt != "2026-09-12T06:30:00Z" {
		t.Fatalf("unexpected build identity: %+v", got)
	}
	Version, Commit, Environment, BuiltAt = "private arbitrary value", "not a commit", "private-host", "not a date"
	if got := Current(); got != (Info{"unknown", "unknown", "unknown", "unknown"}) {
		t.Fatal("invalid build metadata should be omitted, not exposed")
	}
}
