# VOC-107 — Release Plan

## Release and deployment authorization

This package does not authorize production application deployment. Implementer
bundle/publish behavior changes take effect when the karsift-ai-infra task PR
merges to that repo's default branch; the caller already consumes `@main`.
Calling-repo foundation tests or docs land on `develop` and then promote through
the repository-controlled release path. No founder `approved` comment is a
merge/adopt/release gate under active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 CI,
  governance validation, and independent verification pass on the implementable
  task PR; deterministic Git fixtures cover positive attempt-2 / rebase-derived
  import and negative incomplete lineage.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No new or changed Kuma monitors/synthetics
  (`monitoring_impact.state: none`). Outcome signal is governance-operational:
  remediation bundles verify/import in an integration-only bare repository;
  publisher guards remain enforced; attempt cap unchanged.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #891:** closes with package roster completion after T00 verification.

## Rollback

- **Trigger:** Attempt-2 remediation publications again fail `bundle verify` for
  missing prerequisites; soft-reset squashes prior task commits; publisher
  exact-SHA / ancestry / workflow-deny / force-with-lease weaken; or attempt
  policy changes without authority.
- **Mechanism:** Revert T00 infra (and calling-repo pin/tests if any) through
  normal PR/release paths.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm implement.yml bundle range and publish
  guards match last-known-good; note that pre-T00 behavior reintroduces the #891
  thin-bundle reject class.
- **Last-known-good:** commit before T00 merge.

## Independent verification, human approvals, and closure

- The implementable task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. Strengthened evidence obligations for this governance-automation
  change: deterministic Git fixtures pass; publisher guards preserved; soft-reset
  tip remains distinct from bundle/publish base; attempt cap unchanged; evidence
  remains metadata-only.
- Closure: AC results with evidence in `t00-evidence.md`. Package closure follows
  roster completion and normal develop → main promotion for any calling-repo
  tests/docs. Infra merge is on the karsift-ai-infra release path for that
  repository.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
