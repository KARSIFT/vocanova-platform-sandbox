# VOC-110 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment by itself. Application or
workflow fixes take effect on staging when T00 merges to `develop` and T01 records
a successful `deploy-staging` run. Production receives changes only through the
repository-controlled develop → main promotion path after roster completion. No
founder `approved` comment is a merge/adopt/release gate under active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 then T01
  merge in order; CI, governance validation, and independent verification pass on
  each task PR.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** Existing `kuma.availability.staging.web` and
  `synthetic.staging.authenticated-core-journey` cover the affected surface; no new
  monitor is needed. Outcome signals are the pre-merge production-image runtime gate
  and a post-fix `deploy-staging` success including the staging core-loop.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #911:** closes with package roster completion after T01 verification;
  fingerprint `deploy-staging:failure` must not accumulate duplicate open issues
  during successful remediation.

## Rollback

- **Trigger:** T00 fix fails T01 live proof; staging core-loop or health checks
  regress; new `deploy-staging:failure` issues open for the same root cause class.
- **Mechanism:** Pin the paired Next.js packages back to 16.3.0 through a governed
  PR if 16.3.2 cannot pass the image-runtime gate; confirm staging recovery through
  deploy-staging. Promote the rollback through the normal release path if already
  on `main`.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, record whether deploy-staging succeeds on
  `develop` without the T00 change; do not close issue #911 without a new governed
  plan if failure persists.
- **Last-known-good:** `develop` HEAD before T00 merge (exact SHA at task time).

## Independent verification, authority, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. R3 evidence obligations: log-derived root cause recorded, deterministic
  tests pass, live deploy success recorded, VOC-084/VOC-088 deploy regression tests
  green when touched.
- Closure: all AC results with evidence in `t00-evidence.md` and `t01-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
