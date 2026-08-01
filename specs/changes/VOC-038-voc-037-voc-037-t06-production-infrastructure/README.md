# VOC-038 — VOC-037-T06: Production Infrastructure Provisioning

This is a **draft, unapproved change package**. It is not implementation authority.
Nothing in this directory may be treated as adopted, approved, or authorized until a
human founder decision is recorded in `change.yaml`'s `approval_status`,
`implementation_authorized`, and `repository_adoption_status` fields.

## Identity and lifecycle

- Package ID: `VOC-038`
- Title: VOC-037-T06: Production Infrastructure Provisioning
- Canonical path: `specs/changes/VOC-038-voc-037-voc-037-t06-production-infrastructure`
- Lifecycle state: `draft` (see `change.yaml`)
- Proposed risk: `R3` — a draft proposal only, not a determination. See
  `change.yaml`'s `planned_implementation_risk_floor` and `specification.md`'s "Risk
  and protected areas" for the reasoning.
- Owner (decision): founder
- Owner (package preparation): planner
- Approval evidence: none yet — `approval_status: not-approved`
- Target branch: `develop`
- Linked GitHub issue: [#269](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/269)

## Objective and requirement source

This package drafts the task set to execute task `VOC-037-T06` from the already-
adopted package
[`VOC-037`](../VOC-037-begin-milestone-r2-production-readiness-docs/), per issue
[#269](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/269). `VOC-037`'s
own `T00`/`T01` decision-record tasks only produced designs (per each task's own
explicit non-goals); nothing in `VOC-037`'s original task roster actually built the
production target that `VOC-037-T03`/`VOC-037-T04` need to verify against. This
package executes `VOC-037-D00` (accepted: Option A-modified — production co-located
on the same physical host as staging, logically isolated, portable) and
`VOC-037-D01` (accepted: corrected mechanism 4A — separate production directory
tree, separate Compose project, separate least-privilege deploy user). It does not
make any new hosting, secrets, or architecture decision.

See `specification.md` for the full objective, requirement grounding, and the two
open questions this package could not resolve on its own (production hostname/DNS
placeholder and exact shared-host resource-limit values).

## Scope, non-goals, risk, and protected areas

See `specification.md`'s "Scope and non-goals" and "Risk and protected areas"
sections, and `impact-analysis.md` for the full security/privacy/data/rollback
analysis. In summary: this package builds the `/opt/vocanova/production/` directory
tree, the `vocanova-production` Compose project and its compose file, the
production-only deploy user, the `production` GitHub Actions environment, the
`.github/workflows/deploy-production.yml` workflow, and the negative-access
rehearsal proving staging's deploy path cannot read production's secrets. It
explicitly excludes: choosing a final production hostname/DNS name (uses a
founder-confirmable placeholder), publishing any legal/privacy document, and
reopening `VOC-037-D00`/`VOC-037-D01` themselves.

## Verification, approvals, release, and closure

No verification, approval, or release has occurred. This package proposes (in
`tasks.md`, `test-plan.md`, and `release-plan.md`) how each future task should be
verified and by whom, but none of that is authorized to run against production
until the package is adopted and each task is separately implementation-authorized
per the active A-003 model (see `AGENTS.md` and `CLAUDE.md`).
