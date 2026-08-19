---
evidence_id: VOC-088-EV-00
task_id: VOC-088-T00
acceptance_criteria:
  - VOC-088-AC-00
  - VOC-088-AC-01
  - VOC-088-AC-02
  - VOC-088-AC-03
tests:
  - VOC-088-TEST-00
  - VOC-088-TEST-01
  - VOC-088-TEST-02
  - VOC-088-TEST-03
  - VOC-088-TEST-04
date: 2026-08-19
related_change: VOC-088
accountable_owner: unassigned
gate_status: repository-complete-secret-bootstrap-confirmed-live-deploy-pending
live_deployment_claimed: false
---

# VOC-088-T00 — Persistent staging controlled-signup cohort

## Scope and outcome

This task replaces the ephemeral manual-dispatch allowlist with the repository
secret `STAGING_NEW_USER_SIGNUP_ALLOWLIST`. Both automatic and manual staging
deployments now use that one source. The former dispatch input is removed, so a
normal push deployment cannot silently erase the cohort.

When Google OAuth is enabled, a missing, empty, multiline, or malformed cohort
fails before the host configuration write. The remote write also refuses an
empty transmitted value before touching `api.env`. Successful validation emits
only `controlled signup ready: true`; it never emits addresses or cohort size.
Global signup remains disabled, and production deployment never references the
staging secret.

## Repository deliverables

| Artifact                    | Path                                                          |
| --------------------------- | ------------------------------------------------------------- |
| Secret-backed deployment    | `.github/workflows/deploy-staging.yml`                        |
| Fail-closed validator       | `infra/scripts/validate-staging-signup-allowlist.sh`          |
| Deterministic tests         | `scripts/foundation/voc088-deploy-staging-allowlist.test.mjs` |
| Prior OAuth contract update | `scripts/foundation/voc084-deploy-staging-oauth.test.mjs`     |

## Acceptance mapping

| Acceptance criterion | Repository result                                                                                                                       |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| AC-00                | Both trigger modes read the persistent repository secret and write it to staging `api.env`. Live preservation proof remains post-merge. |
| AC-01                | The dispatch allowlist input no longer exists.                                                                                          |
| AC-02                | OAuth-enabled empty/malformed configuration exits non-zero before convergence; remote empty transmission also exits before mutation.    |
| AC-03                | Tests assert fixed boolean/diagnostic output only; fixtures use the reserved invalid synthetic domain.                                  |

## Deterministic validation

```bash
node --test \
  scripts/foundation/voc088-deploy-staging-allowlist.test.mjs \
  scripts/foundation/voc084-deploy-staging-oauth.test.mjs
```

Result: 16 tests passed, 0 failed.

Additional required gates before merge:

- full repository validation and independent exact-SHA review;

Secret bootstrap confirmation: `STAGING_NEW_USER_SIGNUP_ALLOWLIST` was present
in the repository secret inventory on 2026-08-19 before merge. Only its name and
presence were checked; its value is not included in repository content, workflow
output, issue/PR text, or this evidence.

## Live deployment

Pending merge-triggered staging deployment after the encrypted secret bootstrap.
This evidence does not claim a live sign-in result; end-to-end cohort admission
and persistence evidence belongs to VOC-088-T03.
