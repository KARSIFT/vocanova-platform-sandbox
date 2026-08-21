# VOC-106 — Release Plan

## Release and deployment authorization

This package does not authorize production application deployment. Remediation
behavior changes take effect when the karsift-ai-infra task PR merges to that
repo's default branch; the caller already consumes `@main`. Calling-repo
verifier, docs, and tests land on `develop` and then promote through the
repository-controlled release path. No founder `approved` comment is a
merge/adopt/release gate under active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 → T01 in
  order; CI, governance validation, and independent verification pass on each
  implementable task PR; T01 waits for controlled workflow proof after T00 is
  live.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No new or changed Kuma monitors/synthetics
  (`monitoring_impact.state: none`). Outcome signal is governance-operational:
  operator-owned FAIL/CI failure yields zero implementer dispatch plus sanitized
  escalation; ordinary FAIL still retries; the exact-head proof verifier succeeds
  using metadata only; merge remains fail-closed.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #882:** closes with package roster completion after T01 verification.

## Rollback

- **Trigger:** Ordinary tasks no longer auto-remediate; operator-owned FAIL/CI
  failure again gets implementer runs; merge becomes non-fail-closed; or
  fail-closed escalation storms on valid packages.
- **Mechanism:** Revert T00 infra (and calling-repo pin if any) through normal
  PR/release paths.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm remediate YAML / decide helper match
  last-known-good; note that pre-T00 behavior reintroduces the #882
  spurious-dispatch class.
- **Last-known-good:** commit before T00 merge.

## Independent verification, human approvals, and closure

- Each implementable task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. Strengthened evidence obligations for this governance-automation
  change: deterministic ownership-gate tests pass; scope confirmed; T01 live
  metadata recorded under the operator-owned contract; sanitization and no-secret
  invariants preserved; merge fail-closed preserved.
- Closure: AC results with evidence in `t00-evidence.md` and `t01-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion
  for any calling-repo pin/docs. Infra merge is on the karsift-ai-infra release
  path for that repository.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
