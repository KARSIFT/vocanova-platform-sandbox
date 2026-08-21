# VOC-106 — Stop remediation dispatching the general implementer for operator-owned live-evidence failures: Specification

## Objective and requirement source

Stop remediation from dispatching the general implementer when the exact-head
task is operator-owned or live-evidence-only, as recorded in
[GitHub issue #882](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/882).

Today `remediate.yml` / `decide-remediation.py` decide `RETRY` from CI failure
or review `FAIL` after parsing task identity from the PR body. They do not load
the VOC-097 live-evidence contract for that task. `WAITING FOR OPERATOR LIVE
EVIDENCE` already suppresses retry, but a genuine `FAIL` (or CI failure) on the
same operator-owned carrier can still start `implement.yml`.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Drafting-time grounding:

| Item               | Current state                                                                                                                         |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| Remediation decide | `decide-remediation.py` returns `RETRY` on `ci_failed` or review `FAIL`; no ownership argument                                        |
| WAITING path       | Review state `WAITING` returns `WAITING` and suppresses implementer retry                                                             |
| Ownership contract | `<package>/.karsift/live-evidence/<task_id>.yaml` with `ownership: operator` or `live-actions` (VOC-097)                              |
| Auto-advance       | VOC-102 already gates next-task implementer dispatch; remediation is a separate path                                                  |
| Incident           | VOC-104-T01 carrier PR #879 → exact-head FAIL → remediate selected implementer → cancelled workflow 32529860337; head unchanged       |
| Merge              | Must remain fail-closed; this package must not weaken merge-gate                                                                      |

## Scope and non-goals

In scope:

1. Before any remediation retry that would call `implement.yml`, resolve task
   ownership from the exact reviewed PR head and the adopted package's
   `.karsift/live-evidence/<task_id>.yaml` contract (with fail-closed package
   context checks consistent with VOC-102 ownership classification).
2. When a valid contract establishes ownership as `operator` or `live-actions`
   (or equivalent live-evidence-only mode), **never** dispatch the general
   implementer for review `FAIL` or CI failure.
3. Keep merge behavior fail-closed.
4. Route operator-owned review `FAIL` / CI failure to a sanitized, bounded
   operator/reconcile escalation. Preserve stale no-op and existing
   missing-evidence handling, and route malformed/mismatched ownership metadata
   to a separate fail-closed escalation. No path may leak logs, identity data,
   credentials, or evidence payloads.
5. Preserve normal bounded remediation (attempt 1 → attempt 2, then stop) for
   ordinary implementer-owned tasks with no live-evidence contract and no
   contradictory operator declaration.
6. Treat stale/missing evidence lifecycle states deliberately so they cannot
   deadlock the task or consume an implementation attempt.
7. Deterministic workflow-policy and decision-helper tests covering operator
   WAITING, FAIL, CI failure, stale result, malformed contract, and ordinary
   implementer FAIL and CI failure.
8. Controlled, non-destructive live proof after T00 is live, with a read-only
   exact-head verifier for reconcilable SHA lineage.
9. Update infra README and calling-repo operator docs with the ownership-gated
   FAIL/CI behavior, explicit stale/malformed paths, and retained ordinary
   bounded retry.
10. Calling-repo `pipeline.yml` consumes the fixed reusable workflow and exposes
    the narrow, manually dispatched, read-only exact-head verifier action.

Non-goals / explicitly excluded:

- Granting the implementer general Actions credentials.
- Changing VOC-097 reconcile matching, evidence allowlists, or waiting verdict
  semantics beyond what remediation must cooperate with.
- Reopening VOC-102 auto-advance behavior (already gated).
- Node runtime deprecations, action-input migration, Go cache configuration,
  dependency alerts.
- Application, migration, signup-policy, secrets, database, or
  `infra/monitoring/` inventory ID changes.
- Manufacturing live evidence for unrelated packages (including re-running
  VOC-104-T01 proof as part of this package).
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (CI/CD / remediation-dispatch lifecycle).
- **Measured path floor at drafting:** **R3** for `.github/workflows/` and
  related governance automation. Not proposed as R4; no authority-model or
  amendment docs. Calling-repo pipeline/doc touches may raise the path floor
  at implementation time.
- Protected areas: `remediate.yml` / `decide-remediation.py` dispatch gate;
  implementer least-privilege; merge-gate fail-closed behavior; VOC-097
  live-evidence ownership contract and reconcile path.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

The R3 value is a **draft proposal for the reviewing human at adoption time,
never a determination**. The path-based classifier and independent verifier
govern each task PR.

## Decisions

`VOC-106-D00`: Remediation MUST resolve task ownership from governed package
data at the exact reviewed PR head before selecting any implementer retry.
The authoritative machine-readable source is
`<package_path>/.karsift/live-evidence/<task_id>.yaml` when that file exists at
the reviewed head. Package-context checks MUST remain fail-closed and
consistent with the VOC-102 ownership classifier: canonical `tasks.md` MUST
remain readable where that classifier requires it; task identity MUST validate
against `.karsift/tasks.json` / PR-body task identity consistency rules already
used by remediation. A missing or unreadable required package context fails
closed. The only secondary expectation signal is an exact allowlisted marker
inside the matching `## <task_id>` stanza of canonical `tasks.md`:
`- Automation ownership: operator` or
`- Automation ownership: live-actions`. The parser MUST match the heading and
one marker structurally; it MUST NOT infer ownership from narrative prose.

`VOC-106-D01`: If a valid contract declares `ownership: operator` or
`ownership: live-actions`, remediation MUST NOT dispatch `implement.yml`,
regardless of review `FAIL`, CI failure, review-job infrastructure failure
already handled separately, stale caller run, or missing evidence lifecycle
state. Malformed or mismatched ownership metadata is governed by the separate
fail-closed rule in `VOC-106-D04`. Merge behavior remains fail-closed.

`VOC-106-D02`: For operator-owned / live-evidence-only conditions that would
previously have become `RETRY`, remediation MUST route to a sanitized, bounded
operator/reconcile escalation (machine-readable decision such as
`escalate-operator` / equivalent). Escalation comments or markers MUST be
allowlisted metadata only (task ID, package path, PR number, run IDs,
boolean/no-retry decision, scrubbed reason codes). Forbidden: logs, artifacts,
secrets, OAuth/session/cookie/token material, user identifiers, evidence
payloads, or raw review body replay.

`VOC-106-D03`: Ordinary implementer-owned tasks with no live-evidence contract
and no contradictory operator-owned declaration continue today's bounded
remediation: CI failure or genuine review `FAIL` may retry once
(`attempt + 1`, capped at 2), then stop and escalate to the authority issue.
`WAITING` continues to suppress retry without consuming an implementation
attempt. `STALE` and `REVIEW_INFRA_FAILURE` remain non-retrying.

`VOC-106-D04`: Stale caller runs, missing qualifying evidence, and
malformed/unreadable/contradictory ownership metadata MUST NOT dispatch the
implementer and MUST NOT consume an implementation attempt. A stale caller run
MUST retain the `STALE` no-op. Missing evidence lifecycle states MUST retain
their existing operator/reconcile handling. Malformed, unreadable, or
contradictory ownership metadata MUST use an explicit fail-closed escalation;
it MUST NOT be guessed as ordinary ownership. Absence of both a contract and an
automation-ownership marker means an ordinary implementer-owned task (same rule
as VOC-102).

`VOC-106-D05`: Deterministic tests cover at least: operator WAITING (no retry);
operator FAIL (no implementer; escalate); operator CI failure (no implementer;
escalate); stale result (no retry; no attempt consumed); malformed / mismatched
/ missing-required contract (no implementer; fail-closed escalate); ordinary
implementer FAIL (retry retained); ordinary CI failure (retry retained); and
permission/sanitization boundaries for escalation markers.

`VOC-106-D06`: Controlled proof uses a sanitized workflow event after T00 is
live showing that an operator-owned carrier experiencing FAIL or CI failure
produces zero executed `implement.yml` job from remediation, plus retained
bounded retry for an ordinary implementer-owned FAIL fixture. T00 MUST also
provide a read-only `pipeline.yml` workflow-dispatch proof action. After source
run allowlisted metadata is recorded on the deterministic T01 carrier, the
operator dispatches that verifier on the carrier branch; its own successful run
MUST use `exact_pr_head`. Evidence is metadata-only.

`VOC-106-D07`: Keep root scope focused. Node runtime deprecations, action-input
migration, Go cache configuration, dependency alerts, and application changes
are out of scope follow-ups.

## Open questions

- None that block drafting. Exact escalation marker wording and whether
  escalation reuses an existing reconcile dispatch hook versus a new sanitized
  issue/PR marker are implementer/design choices inside `VOC-106-D02`, provided
  they remain least-privileged, metadata-only, and never call `implement.yml`.

## Data, migrations, analytics, and accessibility

- Data / migrations: None — evidence-backed non-applicability.
- Analytics: None — evidence-backed non-applicability.
- Accessibility: None — evidence-backed non-applicability (no product UI).

## Security and privacy

- No new secrets. No broadening of implementer token scopes.
- Escalation signals and proof evidence are allowlisted metadata only.
- Decision helper and remediate decide job remain without model credentials and
  without a `secrets: inherit` path into the general implementer on operator /
  fail-closed paths.
- Ordinary implementer retry retains today's existing secret interface for
  implementer-owned tasks only.
