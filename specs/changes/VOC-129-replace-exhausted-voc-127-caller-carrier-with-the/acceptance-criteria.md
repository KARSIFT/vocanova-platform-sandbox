# VOC-129 — Acceptance Criteria

## VOC-129-AC-00 — Caller fixture and pin equal exact infrastructure #164 merge

- Requirement source: `VOC-129-D02`
- Tasks: `VOC-129-T00`
- Tests: `VOC-129-TEST-00`, `VOC-129-TEST-01`
- Evidence: `VOC-129-EV-00`
- Result: pending

`tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` equals
`863fc1f35b1d35e4981a59166b0e939be1a2b681`. Every live caller pin assertion
equals that same SHA. The pin does not equal
`a9df74a63976d5239b84151fd01310835c999e7c` or
`60afda3a44fd06b8c00b219771de7112f1aded6e`. The mirrored fixture includes the
in-scope #164 files needed by the caller.

## VOC-129-AC-01 — Checkout-ref ordering and missing-develop fallback are regression-covered

- Requirement source: `VOC-129-D05`
- Tasks: `VOC-129-T00`
- Tests: `VOC-129-TEST-02`
- Evidence: `VOC-129-EV-00`
- Result: pending

The caller fixture contains `config/release-checkout-ref-runner.py` and its
#164 tests. Deterministic coverage proves: existing `develop` is selected
without reading `main`; absent `develop` falls back to live `main`;
ambiguous or malformed refs fail closed; checkout-ref resolution runs before
the caller release-state checkout. The #163 absent-`develop` preflight defect
cannot recur unnoticed in the pinned fixture.

## VOC-129-AC-02 — Live caller dispatches reconcile-production-change under the 25-input limit

- Requirement source: `VOC-129-D03`
- Tasks: `VOC-129-T00`
- Tests: `VOC-129-TEST-03`, `VOC-129-TEST-04`
- Evidence: `VOC-129-EV-00`
- Result: pending

Live `.github/workflows/pipeline.yml` exposes action
`reconcile-production-change`, forwards `issue_number` as task authority, and
does not declare operator-typed SHA inputs. `existing_pr_number` remains
implement-only. Every live `workflow_dispatch` block has at most 25 inputs.
`release` waits on `reconcile-production-change`. `auto-advance` does not
proceed when that reconcile fails.

## VOC-129-AC-03 — Mirrored #164 release contract no longer restores CHECKED_HEAD_SHA

- Requirement source: `VOC-129-D04`
- Tasks: `VOC-129-T00`
- Tests: `VOC-129-TEST-05`
- Evidence: `VOC-129-EV-00`
- Result: pending

The pinned fixture `release.yml` binds and advances `develop` to
`mergeCommit.oid` before audit close, recreates a missing integration ref at
the merge SHA, and no longer encodes restoration at `CHECKED_HEAD_SHA` as
success. Unique develop commits, moved `main`, malformed merges, and unbound
refs fail closed in the mirrored helpers.

## VOC-129-AC-04 — Tree-equivalent sync does not schedule unnecessary staging; path selection remains

- Requirement source: `VOC-129-D04`, `VOC-129-DEP-07`
- Tasks: `VOC-129-T00`
- Tests: `VOC-129-TEST-06`
- Evidence: `VOC-129-EV-00`
- Result: pending

A `develop` ref update whose tree is identical to the previous `develop` tip,
or whose changed paths are outside the VOC-111 runtime/deploy allowlist, does
not perform a staging deployment. Pushes that change allowlisted
runtime/deploy paths still select `deploy-staging.yml`. The allowlist is not
broadened.

## VOC-129-AC-05 — Unpublishable VOC-127 carrier is superseded without attempt 3

- Requirement source: `VOC-129-D01`, `VOC-129-D08`
- Tasks: `VOC-129-T00`
- Tests: `VOC-129-TEST-07`
- Evidence: `VOC-129-EV-00`
- Result: pending

This package's implementation PR is a new VOC-129 carrier from current
`develop`. PR #1041 is not merged, not published, and not composed into the
replacement. VOC-127-T00 is not dispatched as attempt `3`. After the
replacement merges and is promoted, #1041 is closed as superseded, then #1039
and #1035 are closed with audit comments naming the VOC-129 merge. No VOC-127
completion marker is manufactured from #1041.

## VOC-129-AC-06 — Existing controls, docs, and roles remain

- Requirement source: `VOC-129-D06`, `VOC-129-D07`
- Tasks: `VOC-129-T00`
- Tests: `VOC-129-TEST-08`, `VOC-129-TEST-09`
- Evidence: `VOC-129-EV-00`
- Result: pending

Roster markers, required-check recovery, independent review, retry caps, risk
floors, secret redaction, App-token isolation, and rollback controls remain.
`config/roles.yml` is unchanged. No OpenAI route is added. Current-state docs
that describe release scope or branch behavior state that `develop` advances
to the promotion merge SHA before audit close, that `reconcile-release`
retries that sync, and that exceptional reconciliation is
`reconcile-production-change`. Historical A-003 / VOC-075 / VOC-127 records
are not rewritten.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
