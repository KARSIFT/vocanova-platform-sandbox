# VOC-077 — VOC-072-T00 Evidence File Regressed to Pending

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #578](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/578),
prepared for founder/steward review at adoption time.

## Identity and lifecycle

- Package ID: VOC-077
- Title: VOC-072-T00 Evidence File Regressed to Pending After Secret Was
  Already Provisioned
- Canonical path:
  `specs/changes/VOC-077-voc-072-t00-evidence-file-regressed-to-pending`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R1` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; path floor measured at drafting time
  is `R1`)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`
- Target branch: `develop`
- Linked GitHub issues:
  - [#578](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/578)
    (this package's requirement source)
  - [#543](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/543)
    (exhausted VOC-072-T00 task — both attempts spent)
- Related packages and PRs:
  - [VOC-072](specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05)
    — T00 evidence / DEP-01 are the correction targets; T01/T02 remain
    blocked until evidence matches reality
  - [PR #558](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/558)
    — stuck VOC-072-T00 PR; HEAD `4b021050` FAILed independent review after
    evidence regression from `f49ffc50`

## Why this exists

VOC-072-T00's real-world gate is already satisfied: the production GitHub
environment secret `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` exists
(added `2026-08-13T21:24:26Z`, confirmed via `gh secret list --env
production`). PR #558 commit `f49ffc50` documented that correctly and was
down to a single Medium finding (`change.yaml` DEP-01 still said
production-environment presence was outstanding).

The next implementer revision (`4b021050`) **regressed**
`t00-token-provisioning-evidence.md` back to
`gate_status: pending_operator_execution` with AC-00/TEST-00 marked "NOT
satisfied," as if the secret had never been provisioned. Independent review
correctly returned `VERDICT: FAIL`. Both automated attempts on issue #543
are exhausted; automation has stopped retrying. A founder `approved`
comment on PR #558 correctly did **not** force-merge — merge-gate requires
positive PASS proof and does not override a confirmed FAIL.

Without a fresh task to restore accurate evidence and fix the DEP-01
contradiction, VOC-072-T01/T02 stay blocked on a false pending gate.

## What this package does

1. **Correct VOC-072-T00 evidence** (`VOC-077-T00`): restore
   `t00-token-provisioning-evidence.md` (and related AC-00 / DEP status text)
   so they match the confirmed provisioned secret — building on the substance
   of `f49ffc50`, not the regressed pending template — then obtain independent
   review PASS on that exact revision.

## What this package deliberately does NOT do

- Not re-provisioning, rotating, or reading the secret value.
- Not implementing VOC-072-T01 (workflow wiring) or VOC-072-T02
  (`--verify-only`).
- Not editing `.github/workflows/deploy-production.yml`, cutover scripts, or
  Cloudflare dashboard configuration.
- Not redispatches of exhausted issue #543 (this is a new package / task).
- Not diagnosing or fixing the implementer pipeline regression unless
  adoption expands `VOC-077-DEP-01` (default: follow-up issue).
- Does not adopt itself. Every adoption/authorization field stays at the
  unadopted default.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Disposition of PR #558 / issue #543 (`VOC-077-DEP-00`).
2. Whether implementer-regression root cause is in-scope follow-up
   (`VOC-077-DEP-01`).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`. This
package carries no standing approval; adoption, implementation authorization,
independent verification, and any required human approval remain to be
recorded against the exact implemented revision, per AGENTS.md and CLAUDE.md.
