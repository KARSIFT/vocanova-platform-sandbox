# VOC-050 — Add Hourly Sentry-Based Log/Error Monitoring Agent

**Status: proposed, not adopted.** Nothing in this package is
implementation-authorized. It is a draft response to
[issue #392](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/392),
prepared for founder/steward review at adoption time.

## Identity and lifecycle

| Field | Value |
|---|---|
| Package ID | VOC-050 |
| Title | Add Hourly Sentry-Based Log/Error Monitoring Agent That Files Issues on New Production/Staging Problems |
| Canonical path | `specs/changes/VOC-050-add-hourly-sentry-based-log-error-monitoring` |
| Lifecycle state | `draft` (not adopted) |
| Proposed risk | `R3` (see `change.yaml` for full reasoning; a draft proposal, not a determination) |
| Owner | unassigned |
| Approval evidence | none — `approval_status: not-approved` |
| Target branch | `develop` |
| Linked issue | [#392](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/392) |

## Why this exists

No monitoring/log-scanning workflow exists in this repository today. Sentry
is already wired into `apps/api` (per `docs/operations/11-devops-and-ci-cd.md`
(DOC-11) §1's "Error monitoring" row) but not into `apps/web` — browser-side
errors are currently invisible to any tooling. There is no scheduled process
that looks for new production/staging problems; issues only get filed when a
human or an oversight loop happens to notice something.

## What this package proposes

1. Wire `@sentry/nextjs` into `apps/web`, mirroring `apps/api`'s existing
   Sentry pattern (separate DSN per environment tier, disabled when unset).
2. Add a read-only, least-privilege Sentry API auth token as a new GitHub
   Actions secret.
3. Add a new hourly scheduled workflow that queries the Sentry API for new
   problems across both staging and production, with a duplicate-check
   guard (keyed on a stable Sentry issue identifier) before opening an
   **unlabeled** GitHub issue for each genuinely new problem, so
   `plan-from-issue` can pick it up automatically.
4. Deliberately withhold SSH access to either host from this new workflow —
   SSH-based raw log access stays the existing manual/on-demand deep-dive
   tool, not something an unattended hourly job authenticates with.
5. Amend DOC-11 §1's "Error monitoring" row to record the new mechanism, in
   the same pull request that adds it.

## What this package deliberately does NOT do

- It does not add uptime/liveness monitoring (Better Stack/UptimeRobot per
  DOC-11) — a separate concern, explicitly out of scope per issue #392.
- It does not change release-gating behavior. This is observability that
  feeds the existing issue → `plan-from-issue` pipeline, same as any other
  live-found bug.
- It does not add or reuse an SSH credential for the new scheduled workflow.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. No task in `tasks.md` may be dispatched
  until a real adoption decision is recorded.
- It does not obtain, enter, or handle any real secret value — the Sentry
  API auth token and per-environment DSNs are documented as required
  preconditions for a human to provision, not something an agent does (per
  `AGENTS.md`'s "Safety" section on agents not receiving production
  secrets).

## Open questions flagged for the reviewing human

`specification.md`'s "Decisions, contradictions, security, and privacy"
section flags four open questions this drafting pass could not resolve
unilaterally: (1) which GitHub API identity the scheduled workflow uses to
open issues; (2) the exact Sentry API auth token scope, given Sentry's own
scope granularity is coarser than "read-only for exactly these two
environments"; (3) whether the founder's actual Sentry plan/organization
supports the assumed additional `apps/web` project and token — see
`VOC-050-T00`, which exists specifically to confirm this before any code is
written; and (4) the exact "since last check" state-tracking approach for the
scheduled workflow.

## Structure

Mirrors recent packages' convention (e.g. VOC-049, VOC-048, VOC-047):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R3` classification, ideally by running
   `scripts/governance/classify-change-risk.sh` against the actual
   task-scoped file list once tasks are dispatched, per this package's own
   `blocking_reasons`.
2. Resolve (or explicitly delegate to the relevant task's implementer) the
   four open questions in `specification.md`.
3. Confirm whether `apps/web`'s new browser-to-Sentry data flow needs a
   privacy-policy or DPA review before adoption (`impact-analysis.md`'s
   security/privacy section flags this as unresolved by this drafting pass).
4. Adopt (or request changes to) this package, then dispatch `VOC-050-T00`
   first, before any other task, consistent with this package's own
   scoping.
