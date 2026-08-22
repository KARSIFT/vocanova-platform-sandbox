# VOC-108 — Tasks

No task below is implementation-authorized by this draft. Adoption and task
authorization are separate.

## VOC-108-T00 — Authoritative lifecycle evidence and idempotent advancement

- Requirement source: issue #903; `VOC-108-D00`–`D08`
- Acceptance criteria: `VOC-108-AC-00` through `VOC-108-AC-07`
- Tests: `VOC-108-TEST-00` through `VOC-108-TEST-09`
- Evidence: `VOC-108-EV-00` (`t00-evidence.md`)
- Status: implemented — exact caller merge and post-merge evidence pending

### Required work

1. Implement the shared latest-authoritative-check selector and integrate it
   into adoption, merge/reuse, and release consumers without weakening their
   individual required-check allowlists or exact-SHA binding.
2. Implement and publish the caller-merge completion marker only after the
   exact caller task PR merges; validate it in auto-advance and release.
3. Ensure cross-repository work uses non-closing references and cannot close the
   caller task through PR text.
4. Consolidate promotion evaluation/final merge so automatic, reconcile, and
   terminal-check triggers are serialized and idempotent.
5. Add lightweight re-evaluation on required external check completion that
   reuses exact-SHA CI/review evidence and does not dispatch full unchanged-SHA
   validation solely for observation.
6. Add all deterministic positive/negative/race fixtures in the test plan and
   update shared README/caller docs whose current claims would become false.
7. Preserve App auth, exact SHA/base, branch protection, fail-closed behavior,
   task ordering, publisher isolation, attempt limits, and A-004's no-founder-
   comment path.
8. Record metadata-only evidence and current shared/caller consumption SHAs.

### Cross-repository execution

Primary behavior lands in `KARSIFT/karsift-ai-infra`. Its PR must say `Relates
to KARSIFT/vocanova-platform-sandbox#<task>` and MUST NOT use a closing keyword.
Caller contract/evidence changes land in this repository on the task branch.
The caller task remains open until its exact caller evidence PR merges.

### Out of scope

Application/runtime changes; deployment or OAuth changes; monitor inventory;
scheduled-synthetic branch policy; operational-failure marker cleanup; Node
action upgrades; historical issue cleanup; extra model attempts.
