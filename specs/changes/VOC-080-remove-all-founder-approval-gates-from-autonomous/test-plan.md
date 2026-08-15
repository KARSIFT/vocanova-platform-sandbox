# VOC-080 — Test Plan

## VOC-080-TEST-00 — Autonomous adopt path without founder comment

- Covers: `VOC-080-AC-00`, `VOC-080-AC-04`
- Preconditions: T02 mechanisms on rehearsal target; plan_reviewer (or
  equivalent) can PASS
- Procedure:
  1. Open or use a fixture change issue → plan PR with valid package.
  2. Ensure independent plan review PASS and governance/deterministic
     checks pass.
  3. Allow merge without any founder `approved` comment.
  4. Assert package `change.yaml` is adopted / implementation-authorized
     and task issues exist (or reconcile dispatch creates them).
- Expected result: adopted package + task roster; audit fields populated;
  no founder comment in the critical path.
- Evidence: `VOC-080-EV-02`, `VOC-080-EV-06`

## VOC-080-TEST-01 — R4 merges when non-founder gates pass

- Covers: `VOC-080-AC-01`
- Preconditions: T01 landed; fixture R4 package/PR
- Procedure:
  1. Create or simulate an R4 PR with CI green and independent
     verification PASS.
  2. Confirm merge-gate auto-merge (or equivalent) proceeds with
     `auto_merge_enabled=true`.
  3. Confirm no status text requires founder `approved` for R4.
- Expected result: R4 merges without founder comment; stronger R4
  evidence requirements still documented/enforced separately from the
  comment gate.
- Evidence: `VOC-080-EV-01`, `VOC-080-EV-06`

## VOC-080-TEST-02 — No founder override of failed/missing verification

- Covers: `VOC-080-AC-02`, `VOC-080-AC-07`
- Preconditions: T01 landed
- Procedure:
  1. With CI failing or verdict FAIL/PENDING/missing, attempt any residual
     `approved` comment or legacy approve-and-merge entrypoint.
  2. Assert merge does **not** occur.
  3. With CI green but verdict FAIL, assert merge does not occur.
- Expected result: fail-closed; no comment-based bypass.
- Evidence: `VOC-080-EV-01`, `VOC-080-EV-05`

## VOC-080-TEST-03 — Unparseable risk fails for correction

- Covers: `VOC-080-AC-03`
- Preconditions: T01 landed
- Procedure:
  1. Fixture PR body/package with missing or inconsistent risk.
  2. Observe merge-gate decision.
- Expected result: fails closed with actionable correction; no founder
  approval escape hatch; no auto-merge.
- Evidence: `VOC-080-EV-01`, `VOC-080-EV-05`

## VOC-080-TEST-04 — Reconcile dispatch is idempotent and event-independent

- Covers: `VOC-080-AC-04`, `VOC-080-AC-05`
- Preconditions: T02 landed
- Procedure:
  1. Create a merged-but-unadopted (or roster-missing) fixture matching
     the #625 failure class.
  2. Run reconcile `workflow_dispatch` (or documented entrypoint).
  3. Assert adoption + roster repaired.
  4. Run dispatch again; assert idempotent (no duplicate task issues).
- Expected result: recovery without `gh run rerun` of a garbage-collected
  event; second run is a no-op or safe restate.
- Evidence: `VOC-080-EV-02`, `VOC-080-EV-06`

## VOC-080-TEST-05 — Release/deploy path without founder interaction

- Covers: `VOC-080-AC-06`
- Preconditions: T03 landed; auto_release_enabled true on rehearsal
- Procedure:
  1. Complete a fixture package roster (or simulate release inputs).
  2. Confirm promotion occurs without founder `approved` on the Release
     issue.
  3. Confirm retry after a forced failure uses dispatch/remediation
     checks, not founder comment, and stays fail-closed until green.
  4. Confirm documented environment reviewer settings do not require
     founder click-approve on the repository-controlled path.
- Expected result: AC-06 holds on rehearsal; residual gates absent.
- Evidence: `VOC-080-EV-03`, `VOC-080-EV-06`

## VOC-080-TEST-06 — Integrated rehearsal on settled venue

- Covers: `VOC-080-AC-00`, `VOC-080-AC-01`, `VOC-080-AC-04`,
  `VOC-080-AC-05`, `VOC-080-AC-06`, `VOC-080-AC-10` (readiness)
- Preconditions: T01–T05 on rehearsal target; `VOC-080-DEP-04` venue
- Procedure:
  1. Execute end-to-end rehearsals listed in T06.
  2. Record redacted run URLs and conclusions in `t06-evidence.md`.
  3. Explicitly note VOC-079 resume readiness (gates cleared), without
     requiring VOC-079 implementation.
- Expected result: all rehearsals pass; limitations recorded if any
  external system is unavailable (not silently passed).
- Evidence: `VOC-080-EV-06`

## VOC-080-TEST-07 — Builder/verifier separation and control preservation

- Covers: `VOC-080-AC-07`
- Preconditions: T01–T05
- Procedure:
  1. Inspect workflows/prompts/role config: implementer cannot be the
     independent verifier for the same exact revision.
  2. Confirm protected-path classification, secrets isolation statements,
     and rollback/audit requirements remain in docs and gates.
  3. Spot-check that audit evidence is emitted for automatic adopt,
     merge, release, deploy, and rollback paths touched by this package.
- Expected result: separation and controls intact; no standing human
  approval role substituted for founder gates.
- Evidence: `VOC-080-EV-05`, `VOC-080-EV-07`

## VOC-080-TEST-08 — Documentation and settings agree with behavior

- Covers: `VOC-080-AC-09`
- Preconditions: T00, T04; T07 for post-activation final check
- Procedure:
  1. Grep AGENTS.md, CLAUDE.md, DOC-15/16 (as touched), approval-matrix,
     change-risk-classification, templates, pipeline comments for claims
     that founder `approved` is required for R4/adopt/merge/release.
  2. After T07, assert those claims are gone or clearly marked historical.
  3. Diff documented repository/environment settings against live settings
     for reviewer/approval gates this package claims to clear.
- Expected result: no contradictory live guidance; historical sections
  explicitly historical; settings match docs.
- Evidence: `VOC-080-EV-00`, `VOC-080-EV-04`, `VOC-080-EV-07`

## Cross-cutting validation (every caller-repo task PR)

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Infra PRs: run that repository's installed workflow/unit self-ci for
touched files. Missing rehearsal access is a recorded limitation — **never**
a pass for T06/T07 live clauses.
