# VOC-098 — Release Plan

## Release and deployment authorization

This package does not authorize production application deployment. Observer
workflow and test changes take effect when task PRs merge to `develop` and, for
live observer behavior, when promoted to the default branch (`main`) via the
repository-controlled release path. No founder `approved` comment is a
merge/adopt/release gate under active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 → T01 in
  order; CI, governance validation, and independent verification pass on each
  task PR; T01 waits until T00 is live on the observer execution branch.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No new or changed Kuma monitors/synthetics
  (`monitoring_impact.state: none`). Outcome signal is operational: observer mint
  succeeds; watched non-success produces exactly one sanitized App-authored issue
  or dedupe hit.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #840:** closes with package roster completion after T01 verification.

## Rollback

- **Trigger:** Observer mint still fails; issues not created for real watched
  failures; plan-from-issue broken because issues are `GITHUB_TOKEN`-authored; or
  benign-cancel classifier systematically misbehaves after token split.
- **Mechanism:** Revert T00 on `develop` and promote revert through normal release
  if already on `main`.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm workflow YAML matches last-known-good;
  note that pre-T00 mint still requests App Actions and remains incompatible with
  the current installation — a follow-up governed issue is required if rollback is
  used without an alternate fix.
- **Last-known-good:** commit before T00 merge.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow merge
  gate. R3 evidence obligations: deterministic permission/token tests pass; scope
  confirmed; T01 live metadata recorded under the operator-owned contract;
  sanitization and App-only issue invariants preserved.
- Closure: AC results with evidence in `t00-evidence.md` and `t01-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
