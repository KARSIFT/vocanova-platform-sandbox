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

| Item               | Current state                                                                                                                                 |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Auto-advance       | `karsift-ai-infra/.github/workflows/auto-advance.yml` sets `should_dispatch=true` without ownership checks                                    |
| Ownership contract | `<package>/.karsift/live-evidence/<task_id>.yaml` with `ownership: operator` or `live-actions` (VOC-097 / `docs/operations/live-evidence.md`) |
| Incident           | VOC-098-T00 close → pipeline run 32462184971 → implement started for VOC-098-T01 (operator-owned); manual cancel                              |
| Reconcile path     | `live-evidence-reconcile.yml` / `pipeline.yml` action `reconcile-live-evidence` (operator path; must remain the dedicated path)               |
| Release            | Last-task close still belongs to release check-completion; auto-advance must not invent early release                                         |

## Scope and non-goals

In scope:

1. Before setting `should_dispatch=true`, load governed ownership for the next
   task from package data (primary: live-evidence contract; secondary: fail-closed
   consistency checks against task declaration where required).
2. When next-task ownership is `operator` or `live-actions` (or equivalent
   live-evidence-only mode), **do not** call `implement.yml`.
3. Leave the next task issue open and route it to a deterministic, non-LLM
   evidence-carrier publisher. That clean job creates or reuses the task branch/PR,
   writes only the contract-declared pending evidence file, and posts a sanitized
   stable waiting marker so the existing PR-centric reconciler has an attachment
   point.
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
- Changing VOC-097 reconcile matching, evidence allowlists, or remediation
  waiting semantics. This package supplies the waiting PR that the existing
  reconciler already requires; it does not teach the reconciler a second issue-only
  evidence model.
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
next task issue remains OPEN. Instead, a deterministic non-LLM clean publisher
creates or reuses the normal task branch and a draft evidence-carrier PR from the
integration branch. The only generated content is the contract-declared pending
evidence path and fixed privacy-safe waiting text. The existing PR-centric
`reconcile-live-evidence` path can then discover and update that PR exactly as it
does today.

`VOC-102-D02`: The read-only `advance` classifier emits a machine-readable
decision (`implement`, `prepare-live-evidence`, `fail-closed`, or `none`) and does
not mutate issues or branches. For `prepare-live-evidence`, a separate clean job
mints the existing GitHub App with only `contents: write`, `issues: write`, and
`pull-requests: write`; it receives no model key and no Actions-write permission.
It creates/reuses the carrier PR and posts one deduplicated sanitized waiting
marker on the task issue stating that no implementer run was started. The job
must never use `secrets: inherit` to call the general implementer.

`VOC-102-D03`: Ordinary implementation-owned next tasks (no live-evidence contract,
and no contradictory operator-owned declaration) continue to dispatch
`implement.yml` attempt 1 exactly as today after existing open-issue / existing-PR
guards.

`VOC-102-D04`: Fail closed on malformed YAML, missing required `ownership`,
unrecognized ownership values, unreadable contracts, or contradictory metadata
(task declared operator-owned live evidence in `tasks.md` / roster guidance without
a valid contract, or contract `task_id` mismatch). Fail-closed means do **not**
dispatch implementer. The same clean publisher posts one deduplicated sanitized
failure marker on the task issue, but it MUST NOT create an evidence carrier from
untrusted/malformed path metadata. The read-only classifier itself gains no write
permission.

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

## Adoption decisions recorded after independent plan review

1. **Waiting carrier selected:** use the dedicated clean, deterministic
   evidence-carrier job from `VOC-102-D01/D02`. An issue comment alone is not
   sufficient because the current reconciler is PR-centric. This resolves
   `VOC-102-DEP-03` without expanding the reconciler into an issue-only model.
2. **Least privilege selected:** classification stays read-only. Only the clean
   carrier/fail-closed publisher may mint the App, scoped to contents/issues/PR
   writes; it receives no Actions-write or model credentials. This explicitly
   records the mutation surface identified by plan review.
3. Confirmed fail-closed contradiction rule in `VOC-102-D04` when `tasks.md` prose
   marks operator-owned live evidence but the contract file is absent (proposed:
   no dispatch + sanitized escalation).
4. Confirmed proposed **R3**; exact implementation paths remain subject to the
   classifier and independent review.
5. Confirmed T01 may dogfood this package (T00 close → T01 operator-owned → observe
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
- The read-only classifier retains `issues`, `pull-requests`, and `contents` read.
  Only the separate clean publisher references App credentials and requests
  contents/issues/pull-requests write; neither job receives implementer/model
  credentials or Actions-write authority.
