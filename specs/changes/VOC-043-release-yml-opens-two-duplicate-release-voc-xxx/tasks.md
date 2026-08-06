# VOC-043 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation-authorization are separate,
mirroring VOC-039/VOC-040/VOC-041/VOC-042's convention.

## VOC-043-T00 — Make release-issue creation idempotent per package

- Requirement source: issue #328's confirmed root cause and suggested fix
- Acceptance criteria: `VOC-043-AC-00`, `VOC-043-AC-01`, `VOC-043-AC-02`
- Status: pending
- Summary: in `karsift-ai-infra/.github/workflows/release.yml`, restructure
  `check-completion` so that near-simultaneous `issues: closed` events for
  tasks in the same package's roster cannot each independently create a
  duplicate "Release: `<change_id>`" issue, exactly as specified in
  `implementation-plan.md`. Preserve `check-completion`'s existing
  early-exit logic, roster-completeness check, label creation, and created-
  issue title/body content unchanged; leave `promote` untouched. Verify by
  reproducing the concurrent-events scenario `VOC-043-TEST-00` describes and
  confirming exactly one release issue results, plus the single-event
  regression scenarios in `VOC-043-TEST-01`, plus confirming `promote` is
  byte-for-byte unchanged (`VOC-043-TEST-02`).
