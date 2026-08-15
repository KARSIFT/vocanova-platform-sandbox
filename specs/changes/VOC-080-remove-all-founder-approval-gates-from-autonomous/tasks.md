# VOC-080 — Tasks

The tasks below are implementation-authorized by the adopted package and run
in the recorded order subject to deterministic checks and independent review.
Recommended order: **T00 → T01 → T02 → T03 → T04 → T05 → T06 → T07**.
T05 may land incrementally with T01–T04 when that keeps each PR reviewable;
T06 must not claim pass before T01–T04 mechanisms exist; T07 is last.

Cross-repo note: T01–T03 primarily change `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for those behaviors; this package is the
authorizing change package for the required outcome. Do not treat the
untracked local `karsift-ai-infra/` checkout (if present) as this repo's
tracked tree.

## VOC-080-T00 — Author successor amendment and transition scaffolding

- Requirement source: issue #627 transition authority; `VOC-080-D00`,
  `VOC-080-D03`; `VOC-080-DEP-00`
- Acceptance criteria: `VOC-080-AC-09`
- Tests: `VOC-080-TEST-08`
- Evidence: `VOC-080-EV-00` (`t00-evidence.md`)
- Status: pending — package adopted

### Required work

1. Per `VOC-080-DEP-00` (recommended A-004): add the successor amendment
   under `docs/governance/amendments/` stating that after effective
   activation, autonomous engineering workflows require no founder
   `approved` comment at any risk class; R4 remains a strengthened
   evidence class; EHR stays exceptional; builder/verifier separation
   remains; this amendment cannot authorize its own adoption.
2. Add or extend transition-state scaffolding (new file and/or fields
   beside `a003-transition-state.yaml`) with **not-yet-effective** markers
   until T07. Do not flip `authority_model` to the post-transition value
   in this task.
3. Preserve A-003 / VOC-075 historical text; cite supersession by issue
   #627 / this package rather than rewriting old records.

### Explicitly out of scope for this task

- Changing `merge-gate.yml` / `adopt.yml` behavior (T01/T02).
- Declaring the transition effective (T07).
- Editing application code.

## VOC-080-T01 — Remove founder merge gates in karsift-ai-infra merge-gate

- Requirement source: issue #627 policy outcomes 1, 3, 4; `VOC-080-D00`,
  `VOC-080-D02`; `VOC-080-DEP-02`
- Acceptance criteria: `VOC-080-AC-01`, `VOC-080-AC-02`, `VOC-080-AC-03`,
  `VOC-080-AC-07`
- Tests: `VOC-080-TEST-01`, `VOC-080-TEST-02`, `VOC-080-TEST-03`
- Evidence: `VOC-080-EV-01` (`t01-evidence.md`)
- Status: pending — after T00 text exists or in parallel if amendment ID
  is already settled at adoption

### Required work

1. In `KARSIFT/karsift-ai-infra` `merge-gate.yml`:
   - Allow auto-merge for **R0–R4** when `auto_merge_enabled` is true, CI
     is green, and independent verification is PASS / PASS WITH
     NON-BLOCKING FINDINGS (plan_reviewer or review as applicable).
   - Stop treating R4 as a hard founder-approval block.
   - Per `VOC-080-DEP-02`: stop treating `automatic_merge_allowed: false`
     as a founder-attention gate (ignore, fail-closed for correction, or
     retire — match adoption choice).
   - Unparseable / inconsistent risk: **fail for correction**, never
     founder override, never auto-merge.
   - Remove or neutralize Path 2 (`approve-and-merge` on founder
     `approved`) as a merge authority. If a residual comment handler
     remains for logging, it must **not** merge and must **not** override
     failed/missing gates (`VOC-080-D02`).
2. Update merge-gate status comments and README/prompts that tell humans
   to reply `approved` to merge.
3. Keep builder/verifier separation and CI green as hard gates.

### Explicitly out of scope for this task

- Autonomous adoption field flips (T02).
- Calling-repo AGENTS.md full rewrite (T04) — may note follow-up.

## VOC-080-T02 — Autonomous adoption and reconcile dispatch

- Requirement source: issue #627 policy outcomes 2, 6; `VOC-080-D01`;
  VOC-040
- Acceptance criteria: `VOC-080-AC-00`, `VOC-080-AC-04`, `VOC-080-AC-05`
- Tests: `VOC-080-TEST-00`, `VOC-080-TEST-04`
- Evidence: `VOC-080-EV-02` (`t02-evidence.md`)
- Status: pending — after or with T01 so merge of plan PRs is not still
  founder-gated

### Required work

1. Change `adopt.yml` / plan-merge path so an independently reviewed plan
   package that passes governance + deterministic validation transitions
   to adopted / implementation-authorized **automatically**, recording:
   exact revision, review evidence, resolved/explicitly deferred
   decisions, risk, and authority provenance.
2. Ensure a bot-mediated merge cannot leave a package silently merged as
   `draft` (atomic adopt-on-merge, or immediate follow-up that cannot be
   skipped by missing events).
3. Add an idempotent **`workflow_dispatch`** (or equivalent observable
   reconcile entrypoint) that repairs merged-but-unadopted packages and
   missing task rosters without replaying an old GitHub event.
4. Update VOC-040 / AGENTS.md recovery docs in T04 to point at this path;
   this task must at least document the dispatch inputs in infra workflow
   headers.

### Explicitly out of scope for this task

- Full caller-doc sweep (T04).
- Activation (T07).

## VOC-080-T03 — Release, remediate, and deploy-path founder-gate removal

- Requirement source: issue #627 policy outcomes 1, 5; `VOC-080-D02`
- Acceptance criteria: `VOC-080-AC-02`, `VOC-080-AC-06`, `VOC-080-AC-07`
- Tests: `VOC-080-TEST-05`
- Evidence: `VOC-080-EV-03` (`t03-evidence.md`)
- Status: pending — may parallel T02 once merge-gate direction is clear

### Required work

1. In `release.yml`: preserve `auto_release_enabled` promotion; remove
   residual **required** founder `approved` comment as a gate (including
   retry-as-primary-gate semantics). Retry after failed auto-promotion
   must be possible via dispatch/remediation checks without founder
   comment.
2. Ensure remediate / deploy retry paths stay fail-closed until checks
   pass; no founder override of failed remediation.
3. Identify and neutralize repository-controlled production environment
   **reviewer** requirements that still demand founder click-approve
   (`VOC-080-R05` / `VOC-080-AC-06` / open question 7); document ops steps if settings
   are changed outside git, with immediate doc follow-up in T04.

### Explicitly out of scope for this task

- Changing RL1/RL2 technical activation (remain separate unless adoption
  explicitly expands scope).
- Application deploy script rewrites unrelated to approval gates.

## VOC-080-T04 — Caller-repo wiring and canonical documentation reconciliation

- Requirement source: issue #627 policy outcome 7; AGENTS.md
  doc-reconciliation rule; `VOC-080-DEP-01`, `VOC-080-DEP-02`
- Acceptance criteria: `VOC-080-AC-09`, `VOC-080-AC-07`
- Tests: `VOC-080-TEST-08`
- Evidence: `VOC-080-EV-04` (`t04-evidence.md`)
- Status: pending — after T01–T03 inputs/behavior stabilize

### Required work

1. Update `.github/workflows/pipeline.yml`: remove or neutralize
   `founder_username` inputs and comments that describe founder
   `approved` as required; align with post-T01/T03 infra contracts.
2. Rewrite `AGENTS.md` / `CLAUDE.md` sections that claim R4 or residual
   gates require founder approval; replace with post-transition model;
   update `automatic_merge_allowed` drafting per `VOC-080-DEP-02`; replace
   merged-but-unadopted recovery with T02 reconcile dispatch.
3. Reconcile DOC-15, DOC-16 (as applicable), `approval-matrix.md`,
   `change-risk-classification.md`, `repository-settings.md`,
   `protected-areas.md`, post-merge checklist, PR template, and
   `specs/templates/change-package/*` so none claim founder approval is
   required for engineering-workflow gates after activation.
4. Keep historical A-003 / VOC-075 narratives as historical; cite #627
   supersession.

### Explicitly out of scope for this task

- Flipping transition-state to effective (T07).
- Infra workflow logic (already T01–T03).

## VOC-080-T05 — Deterministic regression coverage

- Requirement source: issue #627 AC (tests)
- Acceptance criteria: `VOC-080-AC-08`, `VOC-080-AC-02`, `VOC-080-AC-03`,
  `VOC-080-AC-07`
- Tests: `VOC-080-TEST-01`–`VOC-080-TEST-05`, `VOC-080-TEST-07`
- Evidence: `VOC-080-EV-05` (`t05-evidence.md`)
- Status: pending — may land incrementally with T01–T04; complete before
  T06 claims full harness pass

### Required work

1. Add or extend tests in `karsift-ai-infra` and/or this repo covering:
   R0–R4 auto-merge eligibility; plan PR path; task PR path; unparseable
   risk fail-closed; no founder-override path; reconcile dispatch
   idempotency; release without founder comment; deploy/remediate
   fail-closed.
2. Wire tests into existing CI / self-ci entrypoints.
3. Do not use production secrets or production data.

## VOC-080-T06 — Non-production rehearsal proof

- Requirement source: issue #627 risk/rollback; `VOC-080-DEP-04`
- Acceptance criteria: `VOC-080-AC-00`, `VOC-080-AC-01`, `VOC-080-AC-04`,
  `VOC-080-AC-05`, `VOC-080-AC-06`
- Tests: `VOC-080-TEST-06`
- Evidence: `VOC-080-EV-06` (`t06-evidence.md`)
- Status: pending — after T01–T05 land on the rehearsal target

### Required work

1. Run recorded rehearsals on the settled venue (recommended: this sandbox
   and/or infra smoke harness) proving: autonomous adopt, R4 merge without
   founder comment, reconcile-dispatch recovery, release without founder
   comment, unparseable-risk fail-closed.
2. Redact secrets; store run URLs and conclusions in `t06-evidence.md`.
3. Do not declare the transition effective in this task.

## VOC-080-T07 — Transition activation and post-activation unblock

- Requirement source: issue #627 transition authority; `VOC-080-D03`;
  `VOC-080-AC-10`
- Acceptance criteria: `VOC-080-AC-00`–`VOC-080-AC-10` as applicable
- Tests: `VOC-080-TEST-06`, `VOC-080-TEST-08`
- Evidence: `VOC-080-EV-07` (`t07-evidence.md`)
- Status: pending — **last**; requires T06 rehearsal pass, recorded superseding
  no-approval requirement provenance, and independent verification bound to the
  exact activation revision

### Required work

1. Record issue #627 comment #5301333790, which explicitly revokes the earlier
   one-time transition-approval clause, and obtain an independent PASS verdict bound
   to the **exact** activation revision after deterministic checks and rehearsals pass.
2. Flip transition-state / amendment lifecycle to effective; set
   `authority_model` (or successor marker) consistently across
   `a003-transition-state.yaml` / successor, protected-paths lockstep
   fields if required, and docs that key off the active model.
3. Record activation evidence (PR, comment, run URLs).
4. Confirm VOC-079 / issue #624 can proceed on the no-founder-approval
   path (note in evidence; do not implement VOC-079 here).
5. Independent verification of the activation revision must report that
   post-activation, no later founder approval gate remains for engineering
   workflows, while non-founder controls remain.

### Explicitly out of scope for this task

- Re-opening historical packages to rewrite their past approval records.
- VOC-079 technical cutover work.

## Task ordering notes

- T00 before activation semantics are referenced elsewhere.
- T01 before T02 if plan-PR merge would otherwise still wait on founder.
- T04 after infra contracts stabilize to avoid doc/code drift.
- T06 before T07; T07 is the only task that makes the new authority
  effective.
- Tasks may dispatch because this package is adopted and implementation-authorized.
- Closing issue #627 is gated on AC results with evidence, not task-issue
  closure alone.
