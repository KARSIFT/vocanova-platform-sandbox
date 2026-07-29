// Package gate_readiness is the VOC-032-T12 cross-cutting
// evidence-presence check. It exists as its own subpackage
// (rather than living in apps/api/business/aifeedback alongside
// the existing T08 threshold-gate test, or in
// apps/api/cmd/api alongside the T00 server-wiring test) for
// the same reason the package's own task is the LAST in the
// R1 roster: it is intentionally not specific to any one
// module. T12's job is to confirm that every prior task's
// in-repo evidence is actually present at the paths
// `staging-evidence.md` claims, and to report the
// known-blocked live-evidence items honestly rather than
// silently mark them as passing.
//
// The test in this file (`gate_readiness_test.go`) is the
// load-bearing part of T12. It runs as part of
// `go test ./...` from apps/api, so any future regression
// that removes or renames a required evidence file fails
// CI before the change can land. The companion
// staging-evidence.md and mock-inventory.md updates are
// the human-readable half of the same gate; this file is
// the machine-readable half.
package gate_readiness
