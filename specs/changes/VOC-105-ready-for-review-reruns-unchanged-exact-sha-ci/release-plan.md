# VOC-105 — Release Plan

## Release and deployment authorization

This package does not authorize production application deployment. Reuse-gate
behavior takes effect when the karsift-ai-infra task PR merges to that repo's
default branch; the caller already consumes `@main`. Calling-repo wiring, docs,
and tests land on `develop` and then promote through the repository-controlled
release path. No founder `approved` comment is a merge/adopt/release gate under
active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 → T01 in
  order; CI, governance validation, and independent verification pass on each
  implementable task PR; T01 waits for controlled draft → ready proof after T00
  is live.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No new or changed Kuma monitors/synthetics
  (`monitoring_impact.state: none`). Outcome signal is governance-operational:
  safe unchanged ready_for_review skips full CI/model review; unsafe cases take
  the normal path; draft never auto-merges; exact-SHA guards remain.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #872:** closes with package roster completion after T01 verification.

## Rollback

- **Trigger:** Ready PRs merge without trusted exact-SHA evidence; draft PRs
  become mergeable; safe unchanged transitions still re-run full CI/review; or
  synchronize stops re-running CI/review.
- **Mechanism:** Revert T00 infra (and calling-repo wiring if any) through
  normal PR/release paths.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm ready_for_review again takes the
  full CI + review path; note that pre-T00 behavior reintroduces the #872
  duplicate-cost class.
- **Last-known-good:** commit before T00 merge.

## Independent verification, human approvals, and closure

- Each implementable task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. R4 evidence obligations: deterministic reuse-gate tests pass;
  scope confirmed; T01 live metadata recorded under the operator-owned
  contract; sanitization and no-secret invariants preserved; draft and
  exact-SHA protections preserved.
- Closure: AC results with evidence in `t00-evidence.md` and `t01-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion
  for any calling-repo wiring/docs. Infra merge is on the karsift-ai-infra
  release path for that repository.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
