---
evidence_id: VOC-096-EV-00
task_id: VOC-096-T00
acceptance_criteria:
  - VOC-096-AC-00
  - VOC-096-AC-01
  - VOC-096-AC-02
  - VOC-096-AC-03
  - VOC-096-AC-07
  - VOC-096-AC-08
tests:
  - VOC-096-TEST-00
  - VOC-096-TEST-01
  - VOC-096-TEST-02
  - VOC-096-TEST-03
  - VOC-096-TEST-04
  - VOC-096-TEST-05
  - VOC-096-TEST-10
  - VOC-096-TEST-11
date: 2026-08-19
related_change: VOC-096
accountable_owner: unassigned
gate_status: repository-complete-secret-bootstrap-confirmed-live-deploy-pending
live_deployment_claimed: false
---

# VOC-096-T00 — Persistent production controlled-signup cohort

## Scope and outcome

This task replaces the ephemeral `workflow_dispatch` allowlist input with the
repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`. Both automatic push
and manual production deployments now use that one source. The former dispatch
input is removed, so a normal push deployment cannot silently erase the cohort.

When Google OAuth is enabled, a missing, empty, multiline, or malformed cohort
fails before the host configuration write. The remote write also refuses an
empty transmitted value before touching `api.env`. Successful validation emits
only `controlled signup ready: true`; it never emits addresses or cohort size.
Global signup remains disabled (`NEW_USER_SIGNUP_ENABLED=false`), and staging
deployment never references the production secret.

## Repository deliverables

| Artifact                 | Path                                                              |
| ------------------------ | ----------------------------------------------------------------- |
| Secret-backed deployment | `.github/workflows/deploy-production.yml`                         |
| Fail-closed validator    | `infra/scripts/validate-production-signup-allowlist.sh`           |
| Deterministic tests      | `scripts/foundation/voc096-deploy-production-allowlist.test.mjs`  |
| Isolation extension      | `scripts/foundation/voc088-deploy-staging-allowlist.test.mjs`   |

## Acceptance mapping

| Acceptance criterion | Repository result                                                                                                                         |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| AC-00                | Both trigger modes read the persistent repository secret and write it to production `api.env`. Live preservation proof remains post-merge. |
| AC-01                | The dispatch allowlist input no longer exists.                                                                                            |
| AC-02                | OAuth-enabled empty/malformed configuration exits non-zero before convergence; remote empty transmission also exits before mutation.      |
| AC-03                | Tests assert fixed boolean/diagnostic output only; fixtures use the reserved invalid synthetic domain.                                    |
| AC-07                | Production workflow references only `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`; staging workflow never references it.                       |
| AC-08                | Foundation tests lock secret precedence, dispatch removal, fail-closed validation, redaction, and tier isolation.                         |

## Deterministic validation

```bash
node --test \
  scripts/foundation/voc096-deploy-production-allowlist.test.mjs \
  scripts/foundation/voc088-deploy-staging-allowlist.test.mjs
```

Additional required gates before merge:

- full repository validation and independent exact-SHA review.

Secret bootstrap confirmation: issue #809 and `VOC-096-DEP-00` record that
`PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` already exists in the repository secret
inventory. This implementer run could not re-list repository secrets via the
GitHub API from the sandbox runner (`gh secret list` returned no matching entry).
Merge must not proceed until an operator confirms the secret remains populated
when production Google OAuth credentials are present. Only its name and
presence are checked; its value is not included in repository content, workflow
output, issue/PR text, or this evidence.

## Live deployment

Pending merge-triggered production deployment after the encrypted secret
bootstrap is re-confirmed at merge time. This evidence does not claim live
`/healthz.controlled_signup_ready=true`; that proof belongs to VOC-096-T02
after T01 extends the production OAuth/readiness synthetic.
