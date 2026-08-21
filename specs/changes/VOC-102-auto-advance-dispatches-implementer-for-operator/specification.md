# VOC-102 — Stop auto-advance dispatching implementer for operator-owned live-evidence tasks: Specification

## Objective and requirement source

Stop `auto-advance.yml` from dispatching the general implementer when the next
roster task is operator-owned or live-evidence-only, as recorded in
[GitHub issue #863](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/863).

Today auto-advance only checks that a next task exists, its issue is open, and no
PR already exists on the deterministic agent branch. It does not read the VOC-097
live-evidence contract. Closing an ordinary implementation task therefore starts
`implement.yml` for an evidence-only successor that has no legitimate code work.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Drafting-time grounding:

| Item | Current state |
|------|---------------|
| Auto-advance | `karsift-ai-infra/.github/workflows/auto-advance.yml` sets `should_dispatch=true` without ownership checks |
| Ownership contract | `<package>/.karsift/live-evidence/<task_id>.yaml` with `ownership: operator` or `live-actions` (VOC-097 / `docs/operations/live-evidence.md`) |
| Incident | VOC-098-T00 close → pipeline run 32462184971 → implement started for VOC-098-T01 (operator-owned); manual cancel |
| Reconcile path | `live-evidence-reconcile.yml` / `pipeline.yml` action `reconcile-live-evidence` (operator path; must remain the dedicated path) |
| Release | Last-task close still belongs to release check-completion; auto-advance must not invent early release |

## Scope and non-goals

In scope:

1. Before setting `should_dispatch=true`, load governed ownership for the next
   task from package data (primary: live-evidence contract; secondary: fail-closed
   consistency checks against task declaration where required).
2. When next-task ownership is `operator` or `live-actions` (or equivalent
   live-evidence-only mode), **do not** call `implement.yml`.
3. Leave the next task issue open and mark it clearly as waiting for the dedicated
   reconcile/evidence path (sanitized comment and/or machine-readable marker —
   see open question 1 / `VOC-102-D02`).
4. Preserve automatic implementer dispatch for ordinary implementation-owned next
   tasks (no contract, or ownership that is implementation-owned).
5. Preserve final-roster release behavior: skipping implementer for a non-final
   operator task must not open release; closing the final task (including after
   operator evidence completes) still drives release as today.
6. Fail closed when ownership metadata is missing where required, malformed,
   contradictory (for example tasks.md declares operator-owned live evidence but
   no valid contract exists, or contract ownership is unrecognized), or unreadable.
7. Deterministic positive, negative, malformed-metadata, and regression tests.
8. Controlled sanitized workflow proof that an operator-owned next task yields
   zero implementer dispatch, while an ordinary implementation next task still
   dispatches.
9. Update infra README and calling-repo operator docs only where current text would
   otherwise claim auto-advance always dispatches implement for every next task.
10. Calling-repo `pipeline.yml` pin bump only if required to consume the fixed
    reusable workflow.

Non-goals / explicitly excluded:

- Granting the implementer general Actions credentials.
- Changing VOC-097 reconcile algorithm, evidence allowlists, or remediation
  waiting semantics beyond what auto-advance must cooperate with.
- Duplicate exact-SHA review fixes, action-runtime upgrades, cache-path warnings.
- Application, migration, signup-policy, secrets, database, or
  `infra/monitoring/` inventory ID changes.
- Manufacturing live evidence for unrelated packages (including re-running
  VOC-098-T01 proof as part of this package).
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (CI/CD / agent-dispatch lifecycle).
- **Measured path floor at drafting:** **R3** for `.github/workflows/` and related
  governance automation. Not proposed as R4; no authority-model or amendment docs.
- Protected areas: `auto-advance.yml` dispatch gate; implementer least-privilege;
  release check-completion boundary; VOC-097 live-evidence ownership contract.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

The `risk: R3` value in `change.yaml` is a **draft proposal for the reviewing human
at adoption time, never a determination**. The path-based classifier and independent
verifier govern each task PR.

## Decisions

`VOC-102-D00`: Before auto-advance sets `should_dispatch=true`, it MUST determine
next-task ownership from governed package data. The authoritative machine-readable
source is
`<package_path>/.karsift/live-evidence/<next_task_id>.yaml` when that file exists.

`VOC-102-D01`: If the contract declares `ownership: operator` or
`ownership: live-actions`, auto-advance MUST NOT dispatch `implement.yml`. The
next task issue remains OPEN and is left for the dedicated
`reconcile-live-evidence` / live-evidence waiting path from VOC-097.

`VOC-102-D02` (proposed default; confirm at adoption — DEP-03): When skipping
dispatch, auto-advance records a **sanitized** waiting signal on the next-task
issue (App-authenticated comment and/or stable machine-readable marker) stating
that the task is waiting for operator-owned live evidence / reconcile and that no
implementer run was started. No logs, secrets, tokens, or personal data.

`VOC-102-D03`: Ordinary implementation-owned next tasks (no live-evidence contract,
and no contradictory operator-owned declaration) continue to dispatch
`implement.yml` attempt 1 exactly as today after existing open-issue / existing-PR
guards.

`VOC-102-D04`: Fail closed on malformed YAML, missing required `ownership`,
unrecognized ownership values, unreadable contracts, or contradictory metadata
(task declared operator-owned live evidence in `tasks.md` / roster guidance without
a valid contract, or contract `task_id` mismatch). Fail-closed means do **not**
dispatch implementer and emit a sanitized failure/escalation signal rather than
guessing ownership.

`VOC-102-D05`: Skipping implementer for an operator-owned next task must not open
or advance release. Release remains driven by final-roster completion / existing
release check-completion behavior after the operator task actually closes.

`VOC-102-D06`: Deterministic tests cover at least: (positive) ordinary next task
dispatches; (negative) operator/live-actions next task does not dispatch; (malformed)
bad/missing/contradictory metadata does not dispatch; (regression) last-task /
no-next-task still no-ops toward release rather than inventing implement dispatch.

`VOC-102-D07`: Controlled proof uses a sanitized workflow event (for example this
package's own T00→T01 advance after T00 merges, and/or a fixture) showing zero
implementer jobs for an operator-owned next task, plus retained dispatch for an
ordinary implementation next task. Evidence is metadata-only. Do not manufacture
unrelated package live evidence or copy secrets.

`VOC-102-D08`: Keep root scope focused. Duplicate exact-SHA reviews, action-runtime
upgrades, and cache-path warnings are out of scope follow-ups.

## Open questions for the reviewing human

1. Confirm `VOC-102-D02` skip signaling (sanitized issue comment vs machine-readable
   marker vs both). If `live-evidence-reconcile.yml` currently assumes a waiting PR
   already exists, confirm whether skip-without-implementer is sufficient because
   reconcile can attach evidence to the open task issue / later bookkeeping PR, or
   whether a separate minimal waiting-PR path is required (must still not use the
   general implementer).
2. Confirm fail-closed contradiction rule in `VOC-102-D04` when `tasks.md` prose
   marks operator-owned live evidence but the contract file is absent (proposed:
   no dispatch + sanitized escalation).
3. Confirm proposed **R3**, or raise in writing if auto-advance ownership gating is
   treated as R4.
4. Confirm T01 may dogfood this package (T00 close → T01 operator-owned → observe
   zero implementer dispatch) plus a deterministic/fixture proof that ordinary
   next tasks still dispatch, without re-running VOC-098 live proof.

## Data, migrations, analytics, and accessibility

- Data / migrations: None — evidence-backed non-applicability.
- Analytics: None — evidence-backed non-applicability.
- Accessibility: None — evidence-backed non-applicability (no product UI).

## Security and privacy

- No new secrets. No broadening of implementer token scopes.
- Skip/fail signals and proof evidence are allowlisted metadata only (workflow/run
  IDs, conclusions, task IDs, boolean dispatch decisions). Forbidden: logs,
  artifacts, secrets, OAuth/session/cookie/token material, user identifiers.
- Preserve App-authenticated mutation patterns already used by pipeline automation
  where comments are posted.
