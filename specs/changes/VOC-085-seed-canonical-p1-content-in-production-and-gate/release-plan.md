# VOC-085 — Release Plan

## Release and deployment authorization

This package does **not** authorize production deployment by being merged as
a draft or even by being adopted alone. Adoption authorizes implementation
PRs only. Each task PR still requires independent verification against the
exact revision.

Proposed risk is **R3** (draft): path floor R3 for protected production
deploy workflow and production smoke/functional gates, with idempotent
canonical content upserts during deploy. Under **active A-004**,
engineering-workflow gates (plan adoption, merge, release promotion,
repository-controlled deploy) require **no** founder `approved` comment. R3
still requires strengthened evidence, independent verification, monitoring,
named rollback owner, and tested recovery. `automatic_merge_allowed: true`
is set per AGENTS.md (`VOC-080-DEP-02`); setting true does not bypass path
floors, CI, independent verification, unparseable-risk fail-closed, or EHR.

Production rollout uses the normal path: task merges to `develop` → package
roster completion → develop→main promotion via `release.yml` /
`pipeline.yml` → `deploy-production.yml` on push to `main` (or
`workflow_dispatch` fallback). Interrupted promotion retries via
`reconcile-release`; failed gates remain fail-closed.

Issue #702 records a founder remediation directive for the governed loop;
that does not let this draft self-adopt or bypass independent verification.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with R3 proposal accepted or amended in writing
  (including any elevation to R4).
- Stance on `VOC-085-DEP-02` (route-sweep harness shape) recorded at
  adoption or in T02 evidence as the accepted implementer choice.
- T00–T02 merged with independent-verification PASS (or PASS WITH
  NON-BLOCKING FINDINGS) on each exact SHA.
- Live production Cloudflare evidence recorded before claiming package
  closure against issue #702.

Monitoring after T00–T02:

- Production deploy seed step success/failure (fail closed before `up -d`).
- Production smoke: non-empty `/api/v1/journey-situations` and detail checks.
- Authenticated route sweep coverage including dynamic discover routes.
- Public production health endpoints and shared-edge health.
- Isolation invariants (no 8081/8443; staging/production boundaries).
- Do not invent naive unauthenticated page monitors here; later
  monitoring-inventory package adopts stable synthetic-check IDs.

Outcome owner: named in `VOC-085-EV-02` (unassigned at drafting).
Success = `VOC-085-AC-00` through `VOC-085-AC-08` with linked evidence.

## Rollback

Trigger: seed runs after convergence or with continue-on-error; empty
content still passes smoke; destructive/non-idempotent content behavior;
real-user mutation during verification; isolation/topology/health breakage;
route sweep false greens.

Mechanism:

1. Revert the responsible task commit(s) (primary).
2. Redeploy production via normal repository `deploy-production.yml` path.
3. Confirm gates match the rolled-back revision. Canonical seed rows are not
   destructively removed by rollback; if content must remain after a gate
   revert, record that explicitly in evidence (content preservation is
   expected).

Validation: smoke/route expectations match the reverted tree; no secrets in
git or workflow transcripts from the failed revision; isolation intact.

Accountable owner: T00/T01/T02 evidence authors. Last-known-good: tree
immediately preceding the first merged VOC-085 task (known-empty production
canonical content with green status-only smoke per issue #702).

## Independent verification, human approvals, and closure

Independent verifier (per `CLAUDE.md`) must:

- Bind each task report to the exact reviewed commit SHA.
- Confirm implementer did not approve/merge its own work.
- Identify active authority model **A-004** (`a004-active`).
- Confirm AC/test/evidence traceability and that seed writes stay within
  repository-owned canonical upsert semantics.
- Report remaining R3 evidence obligations; EHR not expected for this
  package.
- Confirm production live evidence through Cloudflare before closure.

Do not conflate repository merge, release promotion, activation, or closure.
Closing issue #702 requires AC results with evidence, not task-issue closure
alone.
