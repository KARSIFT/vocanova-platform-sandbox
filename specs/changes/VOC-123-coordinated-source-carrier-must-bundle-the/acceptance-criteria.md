# VOC-123 — Acceptance Criteria

## VOC-123-AC-00 — Nested source bundle advertises the exact committed head

- Requirement source: `VOC-123-D01`
- Tasks: `VOC-123-T00`
- Tests: `VOC-123-TEST-00`, `VOC-123-TEST-01`
- Evidence: `VOC-123-EV-00`
- Result: pending

On the exact reviewed infrastructure revision, after an authorized nested
`karsift-ai-infra` commit, source-bundle creation binds that commit's exact
40-character object ID to a named positive revision, produces a non-empty
`/tmp/implementer-source.bundle` (or the equivalent path used by the job),
and `git bundle list-heads` advertises exactly that committed head. A range
whose positive tip is only the raw `SOURCE_HEAD_SHA` is a failing result.
A later independent implementer rerun succeeding does not satisfy this
criterion for the first invocation that already created the nested commit.

## VOC-123-AC-01 — Raw-SHA positive tip remains a proven failing class

- Requirement source: `VOC-123-D01`, `VOC-123-D04`
- Tasks: `VOC-123-T00`
- Tests: `VOC-123-TEST-00`
- Evidence: `VOC-123-EV-00`
- Result: pending

Deterministic tests that create a real temporary Git repository with a base
commit and one child commit reproduce `git bundle create … "$base_sha..$head_sha"`
failing with empty-bundle (exit 128). That reproduction is kept as a
regression of the #1003 / job `98017696468` class, not deleted once the
production path is fixed.

## VOC-123-AC-02 — Advertised heads, base, SHA, cleanup, and publish stay fail-closed

- Requirement source: `VOC-123-D01`, `VOC-123-D03`, `VOC-123-D04`
- Tasks: `VOC-123-T00`
- Tests: `VOC-123-TEST-02`, `VOC-123-TEST-03`
- Evidence: `VOC-123-EV-00`
- Result: pending

Wrong, missing, or multiple advertised bundle heads fail closed. Wrong
prerequisite/base, malformed SHA, and cleanup/publish mismatches (advertised
object ≠ recorded `SOURCE_HEAD_SHA`, or the publisher would not fetch that
object onto the publish branch) fail closed. The temporary named ref is
removed after bundle creation. No unrelated refs or objects become
publishable.

## VOC-123-AC-03 — Caller and planner recovery bundles are proven, not assumed

- Requirement source: `VOC-123-D02`
- Tasks: `VOC-123-T00`
- Tests: `VOC-123-TEST-04`
- Evidence: `VOC-123-EV-00`
- Result: pending

Tests prove whether `implement.yml`'s `integration_sha..HEAD` caller recovery
bundle and `plan.yml`'s `base_sha..HEAD` planner recovery bundle advertise a
safe named head matching the recorded `head_sha`. Those paths are changed
only if they reproduce empty-bundle or unsafe advertised-head behavior. If
they remain unchanged, evidence records that proof.

## VOC-123-AC-04 — VOC-121 publication, isolation, tokens, and retry limits remain

- Requirement source: `VOC-123-D03`, `VOC-123-D06`
- Tasks: `VOC-123-T00`
- Tests: `VOC-123-TEST-05`
- Evidence: `VOC-123-EV-00`
- Result: pending

The change does not weaken nested-repository isolation, gitlink refusal,
credential-free source-bundle handoff, clean `publish-source` App-token
separation (no caller-token fallback), `bundle verify` plus authorized-SHA
fetch, force-with-lease against `EXPECTED_SOURCE_HEAD_SHA`, two-attempt
implementer bound, non-closing source PR reference, or caller `Closes #N`.
It does not mix the GitHub App token onto the model-controlled runner or
bundle secrets.

## VOC-123-AC-05 — Deterministic real-repository tests cover the live failure class

- Requirement source: `VOC-123-D04`
- Tasks: `VOC-123-T00`
- Tests: `VOC-123-TEST-00` through `VOC-123-TEST-04`
- Evidence: `VOC-123-EV-00`
- Result: pending

Infrastructure tests create real temporary Git repositories and cover
raw-SHA empty-bundle, named-ref success, fail-closed advertised-head/base/SHA
cases, no unrelated refs, and caller/planner `..HEAD` proof or equivalent
fix. Existing VOC-121 tests that bundle `..HEAD` or `..$branch` are not by
themselves sufficient coverage of AC-00 or AC-01. Tests do not use secrets
or production data.

## VOC-123-AC-06 — Current-state docs and caller pin follow the reviewed infra merge

- Requirement source: `VOC-123-D05`, `VOC-123-D06`, `VOC-123-D07`, `VOC-123-D08`
- Tasks: `VOC-123-T00`
- Tests: `VOC-123-TEST-06`
- Evidence: `VOC-123-EV-00`
- Result: pending

Current-state source-carrier comments and infra README no longer present a
raw-SHA positive tip as a working bundle contract. Both repositories are
validated on their exact reviewed revisions. If the caller fixture consumes
the infrastructure change, `PINNED_SHA.txt` equals that exact infra merge
SHA and matching caller pin assertions are advanced. If not, the pin is
unchanged and non-consumption is recorded in `t00-evidence.md`. Evidence
also records that #1003 remains a distinct VOC-122 task to re-dispatch
against that exact revision, not work implemented in this package.
Evidence records the bounded supervised bootstrap infra PR, its exact reviewed
head and merge SHA, separate merger, and exhaustion before the normal caller
carrier resumed; no direct `main` push or runner-environment interception was
used.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
