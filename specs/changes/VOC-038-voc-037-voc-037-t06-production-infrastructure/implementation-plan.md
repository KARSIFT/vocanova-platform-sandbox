# VOC-038 — Implementation Plan

## Preconditions and protected areas

Do not begin any task until this package is adopted and the specific task is
separately implementation-authorized, per AGENTS.md's change workflow. All target
paths (`infra/`, `.github/workflows/deploy-production.yml`,
`docs/operations/11-devops-and-ci-cd.md`) are protected per
`docs/governance/protected-areas.md`'s "Production infrastructure" and
"Deployment and rollback" rows; each task's PR must run
`scripts/governance/classify-change-risk.sh` against its own actual diff before
merge, and must not lower the declared risk class below the detected floor.

Do not modify staging's existing `/opt/vocanova/infra/` tree, `docker-compose.yml`,
`deploy-staging.yml`, or `STAGING_SSH_*` secrets as part of any task in this
package — every task here adds a new, separate production artifact; none edits an
existing staging one. If a task's implementer finds that building the production
artifact requires touching a staging file (e.g. a genuinely shared root-level
compose include), stop and flag it as an open question rather than proceeding,
since that would contradict `VOC-037-D01`'s explicit isolation requirement.

## File reconciliation and implementation sequence

No existing file in this package's target areas currently defines a production
target (confirmed by `VOC-037-D01`'s own `INS-1` finding: "no matches — zero
workflows declare a GitHub Actions environment"), so there is no existing
production-tier work to preserve or reconcile — every artifact this package adds
is new.

Ordered, reversible implementation sequence:

1. `VOC-038-T00` — directory tree, deploy user, Compose project/file (host-side and
   repository-side; independently reversible by deleting the tree/user and
   reverting the compose file).
2. `VOC-038-T01` — `production` GitHub Actions environment (independently
   reversible by deleting the environment; may be done in parallel with `T00`).
3. `VOC-038-T02` — `deploy-production.yml` workflow, depends on `T00`+`T01`
   existing to target (reversible by git revert; the workflow does not run until
   explicitly triggered, so merging it alone has no production effect).
4. `VOC-038-T03` — negative-access rehearsal, depends on `T00`-`T02` (read-only;
   no artifact to roll back).
5. `VOC-038-T04` — documentation updates, depends on `T00`-`T03` producing the
   final actual values to document (reversible by git revert).

## Validation and independent verification

Deterministic commands to run at each task's PR, as applicable:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
rg -n "^on:" -A5 .github/workflows/deploy-production.yml
rg -n "environment:" .github/workflows/deploy-production.yml
rg -n "STAGING_SSH|infra/secrets" .github/workflows/deploy-production.yml
docker compose -f infra/docker-compose.production.yml -p vocanova-production config
```

Exact-SHA independent verification procedure: Claude Code reviews the exact merged
commit SHA of each task's PR per CLAUDE.md's required review steps, explicitly
checking (a) the `production` environment's required-reviewer configuration by
inspection, not assertion, (b) that `deploy-production.yml`'s trigger set excludes
`pull_request`/`pull_request_target`, (c) that no staging file was modified, and
(d) that the negative-access rehearsal evidence in `VOC-038-T03`'s PR is redacted
and does not contain a real secret value. Claude Code cannot approve its own
substantial correction to any of these findings; if it flags a Critical or High
finding requiring rework, the reworked revision requires a separate reviewer pass
per CLAUDE.md.

## Deployment and rollback

No task in this package deploys production traffic. `VOC-038-T02` merges a
workflow file that is capable of deploying, but nothing in this package triggers
it — that remains a distinct, later, explicitly authorized action (consistent with
AGENTS.md's "Agents do not receive production secrets and do not deploy directly
to production" and the repository's RL1/RL2 gate distinction: develop-merge
authority under A-003 §10 is separate from autonomous production release under
A-003 §11/12, which remains disabled).

Rollback trigger: any task found, by independent verification or by the negative-
access rehearsal in `VOC-038-T03`, to fail its own acceptance criterion. Rollback
mechanism: git revert of the specific task's PR (each task's artifact is additive
and independently removable, per the implementation sequence above); for
`VOC-038-T00`'s host-side directory/user, rollback also requires deleting the
created OS user and directory tree on the host itself, which the task's own PR
description must record as an explicit manual step alongside the git revert.
Rollback owner: the task's implementer, with independent verification confirming
the rollback restored the pre-task state. Last-known-good reference: the base
commit this package's `base_sha` is set to at adoption time.
