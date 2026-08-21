# VOC-104 — Release Plan

## Release and deployment authorization

This package does not authorize production application deployment. Ready-for-review
reuse behavior changes take effect when the karsift-ai-infra task PR merges to that
repo's default branch; the caller already consumes `@main`. Calling-repo pipeline,
docs, and tests land on `develop` and then promote through the repository-controlled
release path. No founder `approved` comment is a merge/adopt/release gate under
active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 → T01 in
  order; CI, governance validation, and independent verification pass on each
  implementable task PR; T01 waits for controlled draft→ready proof after T00 is
  live.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No new or changed Kuma monitors/synthetics
  (`monitoring_impact.state: none`). Outcome signal is governance-operational:
  safe unchanged ready_for_review runs emit the required CI reuse marker, skip
  full validation/model review, and still evaluate merge-gate; unsafe cases take
  the full path; drafts never auto-merge; the
  exact-head proof verifier succeeds using metadata only.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #872:** closes with package roster completion after T01 verification.

## Rollback

- **Trigger:** Unsafe ready_for_review events incorrectly reuse prior evidence;
  drafts become mergeable; or safe unchanged events incorrectly fail to reuse and create
  operational incidents beyond acceptable redundant cost.
- **Mechanism:** Revert T00 infra (and calling-repo wiring if any) through normal
  PR/release paths.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm pipeline/reuse helpers match
  last-known-good; note that pre-T00 behavior reintroduces the #872 duplicate
  CI+review class.
- **Last-known-good:** commit before T00 merge.

## Independent verification, human approvals, and closure

- Each implementable task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow merge
  gate. Strengthened R4 evidence obligations: deterministic reuse-policy tests
  pass; scope confirmed; T01 live metadata recorded under the operator-owned
  contract; sanitization and no-secret invariants preserved; draft non-merge and
  App-only verdict trust preserved.
- Closure: AC results with evidence in `t00-evidence.md` and `t01-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion
  for any calling-repo wiring/docs. Infra merge is on the karsift-ai-infra release
  path for that repository.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
