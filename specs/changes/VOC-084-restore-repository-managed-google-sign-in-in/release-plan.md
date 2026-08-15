# VOC-084 — Release Plan

## Release and deployment authorization

This package does **not** authorize production deployment by being merged as
a draft or even by being adopted alone. Adoption authorizes implementation
PRs only. Each task PR still requires independent verification against the
exact revision.

Proposed risk is **R3** (draft): path floor R3 for protected staging deploy
workflow and authentication/secret synchronization. Under **active
A-004**, engineering-workflow gates (plan adoption, merge, release
promotion, repository-controlled deploy) require **no** founder
`approved` comment. R3 still requires strengthened evidence, independent
verification, monitoring, named rollback owner, and tested recovery.
`automatic_merge_allowed: true` is set per AGENTS.md (`VOC-080-DEP-02`);
setting true does not bypass path floors, CI, independent verification,
unparseable-risk fail-closed, or EHR.

Staging rollout: merge to `develop` triggers `deploy-staging.yml`.
Production promotion remains the normal develop→main path after the
roster completes; this package intends **no** production OAuth behavior
change. Interrupted promotion retries via `reconcile-release`; failed
gates remain fail-closed.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with R3 proposal accepted or amended in writing
  (including any elevation to R4).
- Stance on `VOC-084-DEP-01` (Google callback authorization) and
  `VOC-084-DEP-02` (allowlist control surface) recorded at adoption.
- T00–T02 merged with independent-verification PASS (or PASS WITH
  NON-BLOCKING FINDINGS) on each exact SHA.
- Live staging OAuth-start evidence recorded before claiming package
  closure against issue #691.

Monitoring after T00–T02:

- Staging `POST /api/v1/auth/oauth/google/start` status and callback URI.
- Staging `/healthz` `kill_switches.oauth_enabled` vs UI advertise behavior.
- `NEW_USER_SIGNUP_ENABLED` remains false; allowlist stays empty unless
  explicitly expanded.
- Public staging/production health endpoints and shared-edge health.
- Workflow logs must not contain Google client secrets.

Outcome owner: named in `VOC-084-EV-02` (unassigned at drafting).
Success = `VOC-084-AC-00` through `VOC-084-AC-07` with linked evidence.

## Rollback

Trigger: partial-pair writes; credential exposure in logs; wrong staging
callback URI; signup kill switch enabled; UI advertising disabled Google;
production OAuth regression; shared-edge/isolation/health breakage.

Mechanism:

1. Revert the responsible task commit(s) (primary).
2. Redeploy staging via normal repository `deploy-staging.yml` path.
3. Confirm absent-pair / reverted tree converges to coherent disabled OAuth
   (no half-config) and public health remains OK.

Validation: staging OAuth-start and UI capability match the rolled-back
expectation; no secrets remain in git or workflow transcripts from the
failed revision.

Accountable owner: T00/T01/T02 evidence authors. Last-known-good: tree
immediately preceding the first merged VOC-084 task (known-broken staging
OAuth start / advertise-while-disabled per issue #691).

## Independent verification, human approvals, and closure

Independent verifier (per `CLAUDE.md`) must:

- Bind each task report to the exact reviewed commit SHA.
- Confirm implementer did not approve/merge its own work.
- Identify active authority model **A-004** (`a004-active`).
- Confirm AC/test/evidence traceability and production-scope isolation.
- Report remaining R3 evidence obligations; EHR not expected for this
  package unless separately triggered.
- Confirm Google callback disposition is either evidenced or precisely
  recorded as external (`VOC-084-AC-07`).

Closure of issue #691 requires AC results with evidence, including the
live OAuth-start check and either verified Google callback authorization or
the exact external remaining action. Do not conflate repository merge,
staging deploy, production promotion, or package closure.
