# VOC-075 — Governance Still Violates "Approve Only R4": Specification

## Objective and requirement source

Make written drafting policy and live package fields match the founder's
confirmed rule: **only R4 requires founder approval** for merge into
`develop` via the `automatic_merge_allowed` / merge-gate path. Remove the
VOC-068 R3 case-by-case carve-out from `AGENTS.md`, stop non-R4 packages from
carrying `automatic_merge_allowed: false`, and optionally add a check that
prevents a third drift.

Requirement source:
[GitHub issue #573](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/573)
(follow-up to VOC-068 / issue #488).

Founder instruction (issue #573, confirmed 2026-08-14): only R4 should ever
require founder approval. No R3 carve-out, no "but this one is sensitive"
exception. The written policy's nuance must not win against that explicit
instruction.

Canonical policy already consistent with the founder rule at the *risk-class*
layer:

- `docs/governance/change-risk-classification.md` (active A-003): only R4
  requires founder approval; R0–R3 do not require standing founder or
  technical-steward approval solely because of risk class.
- `docs/governance/approval-matrix.md`: same R0–R3 / R4 split.
- `karsift-ai-infra` `merge-gate.yml`: R4 (or unparseable risk) always
  requires founder approval; separately, `automatic_merge_allowed: false`
  also requires founder approval even when risk would otherwise allow
  auto-merge.

The contradiction is in **drafting guidance and package field values**, not
in merge-gate's R4 hard block.

## Confirmed findings (from issue #573)

- `AGENTS.md` "Drafting `automatic_merge_allowed`" still says R3 is
  case-by-case and may warrant founder eyes for auth/secrets/production
  infrastructure — introduced by VOC-068 under `VOC-068-DEP-01`.
- Template `specs/templates/change-package/change.yaml` comment block still
  describes "R3: true or false with stated justification."
- Template `README.md` still says the literal `true` matches routine R0–R2
  and that "R4 packages and deliberate opt-outs must set `false`."
- Twenty-eight non-R4 packages on `develop` carry
  `automatic_merge_allowed: false` (table in issue #573). VOC-072 is called
  out as actively blocking.
- Drafting-time grep of `docs/governance/change-risk-classification.md` and
  `docs/governance/approval-matrix.md` found no `automatic_merge_allowed`
  carve-out language matching AGENTS.md; those docs already state R0–R3 do
  not require founder approval merely for risk class. DOC-15 §17.2/§17.3
  still describes a general per-package opt-out.

## Scope and non-goals

In scope:

- `VOC-075-T00`: Rewrite `AGENTS.md`'s drafting subsection so that:
  - **R0–R3** always draft `automatic_merge_allowed: true`.
  - **R4** sets `automatic_merge_allowed: false` (redundant with merge-gate's
    R4 hard block; keeps the package record self-describing).
  - Remove the R3 case-by-case carve-out and any "sensitive R3 warrants
    founder eyes on merge" language.
  - Remove the R0–R2 "unless the package records a specific reason to
    require founder eyes" opt-out encouragement — the founder rule is
    approve-only-R4, not approve-R4-plus-deliberate-opt-outs.
  - Preserve reminders that the field does not bypass risk classification,
    path floors, CI, independent verification, R4 founder authority, or EHR,
    and that a founder `approved` comment remains a valid override when the
    gate requires it (R4 / unparseable / residual false).
- `VOC-075-T01`: Update template `change.yaml` comment block and template
  `README.md` so they match T00 (no R3 justified-false path; no encouraged
  non-R4 opt-out).
- `VOC-075-T02`: Doc reconciliation — confirm or edit
  `docs/governance/change-risk-classification.md` and
  `docs/governance/approval-matrix.md`; apply `VOC-075-DEP-00` for DOC-15.
- `VOC-075-T03`: Set `automatic_merge_allowed: true` on VOC-072's
  `change.yaml` (required minimum). Expand to additional packages only if
  adoption settles `VOC-075-DEP-01` that way. When flipping a field, add a
  one-line comment citing VOC-075 / issue #573 so the change is auditable.
- `VOC-075-T04` (conditional on `VOC-075-DEP-02`): Add a deterministic check
  that fails when a package under `specs/changes/` has
  `automatic_merge_allowed: false` without `risk: R4` (exact placement and
  historical exemptions settled at adoption).

Non-goals / explicitly excluded:

- Not editing `karsift-ai-infra` workflows (`merge-gate.yml`, `plan.yml`,
  etc.). Gate R4 semantics stay.
- Not changing this repository's `auto_merge_enabled` switch, release
  autonomy, or deploy triggers.
- Not changing R4 founder authority, EHR, independent verification, or
  required CI.
- Not renaming or replacing the `automatic_merge_allowed` schema field.
- Not adopting, authorizing, implementing, or merging this package from
  within the draft itself.

## Risk and protected areas

Builder assessment / proposed class: **R3**, rising to **R4** if adoption
includes DOC-15 edits and/or a lint under `scripts/governance/` /
`tooling/governance/`.

Path floors (from `.github/approved-policy/protected-paths.yaml`):

| Path | Floor |
|---|---|
| `AGENTS.md` | R3 |
| `specs/templates/` | R3 |
| `docs/governance/` (non-amendment) | R3 |
| Ordinary `specs/changes/VOC-*/` (e.g. VOC-072) | default R1 unless a listed protected package |
| `docs/operations/15-…operating-model.md` (DOC-15) | R4 |
| `scripts/governance/` | R4 |
| `tooling/governance/` | R4 |

No application code, migrations, secrets, or production infrastructure
configuration are in default scope. EHR is not triggered. Under active
A-003, routine R3 does not by itself require standing technical-steward or
founder approval; if the effective class becomes R4, founder approval is
required for that consequence.

## Decisions, contradictions, security, and privacy

`VOC-075-D00` (recorded for traceability; formal decision numbering applies
after adoption): The VOC-068 R3 case-by-case carve-out in AGENTS.md, and any
non-R4 package carrying `automatic_merge_allowed: false`, contradict the
founder's explicit approve-only-R4 instruction. Written drafting policy and
active package fields must be corrected to that instruction; the written
carve-out must not override it.

Contradiction with VOC-068 adoption record: VOC-068 deliberately chose R3
case-by-case (`VOC-068-DEP-01`) and forward-only backfill
(`VOC-068-DEP-02`). This package supersedes that R3 drafting preference under
a later, more specific founder instruction (issue #573). It does not claim
VOC-068 was unauthorized at the time; it records that the carve-out is now
withdrawn.

Open questions for the reviewing human:

1. **`VOC-075-DEP-00` — DOC-15 reconciliation.** Once non-R4 packages must
   not set `automatic_merge_allowed: false`, DOC-15 §17.2/§17.3's general
   "package hasn't opted itself out" wording becomes incomplete or
   misleading. Options:
   - (a) Edit DOC-15 in the same PR to state that only R4 packages set
     `automatic_merge_allowed: false` (self-describing; redundant with the
     R4 hard block), and that R0–R3 packages must leave the field `true`.
     Raises this package to **R4**.
   - (b) Do not edit DOC-15; record evidence that merge-gate still
     *mechanically* honors a false opt-out, while AGENTS.md / lint forbid
     drafting one for non-R4 — accepting residual DOC-15 generality. Keeps
     path floor at R3 unless T04 is included.
   Adoption must pick (a) or (b). **Draft recommendation: (a)** so no
   canonical doc still describes a non-R4 founder-approval opt-out path.

2. **`VOC-075-DEP-01` — Backfill scope.** Options:
   - (a) **VOC-072 only** (issue #573 minimum; active blocker).
   - (b) VOC-072 plus any other non-R4 packages that still have open task
     issues / unmerged task PRs at adoption time (implementer enumerates in
     evidence).
   - (c) All 28 packages listed in issue #573 (and any newly discovered
     non-R4 `false` at implementation time).
   **Draft recommendation: (b)** — unblock live work without a large
   historical rewrite, unless adoption prefers (c) for one-shot cleanup.

3. **`VOC-075-DEP-02` — Lint.** Options:
   - (a) Include `VOC-075-T04` under `scripts/governance/` and/or
     `tooling/governance/` (raises package to **R4**). Decide whether the
     check scans all `specs/changes/**/change.yaml` immediately (forcing
     full backfill) or only fails for packages created/modified after this
     package / for files in the current PR diff.
   - (b) Defer lint to a follow-up package; this package ships policy +
     minimum backfill only.
   **Draft recommendation: (a) with scan-all after backfill scope is
   settled**, so a third drift cannot land. If adoption chooses backfill
   (a) or (b) without full historical cleanup, the lint must exempt
   pre-existing historical packages until a follow-up, or adoption must
   expand backfill to (c).

No new secret, credential, or personal-data handling.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** None.
- **Analytics:** None.
- **Accessibility:** None. Documentation / template / optional governance-lint
  / package-metadata package.
