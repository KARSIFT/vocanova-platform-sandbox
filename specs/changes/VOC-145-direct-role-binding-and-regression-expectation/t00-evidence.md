# VOC-145-T00 — Evidence

Task: `VOC-145-T00` — Governed reconciliation of the unauthorized
role-binding and VOC-117 expectation drift.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

This file is the T00 evidence contract. Implementation fills the recorded
fields below after the repair is tracked and committed. Do not require this
file to contain the SHA of the same commit that contains it.

## Implementation PR base

Recorded before the first in-scope edit:

`pending — resolve current develop to a 40-character SHA before any in-scope
edit and record it here.`

Issue-creation infrastructure pin was
`8993e867640dfb604dec0466c4e0787e68d8e258`.
Plan/adoption/roster commits after that SHA are governance-only and do not
count as protected-file drift.

## Infrastructure merge

Independently reviewed `KARSIFT/karsift-ai-infra` merge consumed by this pin:

`pending — record the exact new infra merge SHA after independent review.
Must not be d8720829b176cf1287e633f9382989fc8f258105.`

Last governed VOC-142 infrastructure base (historical audit):
`8993e867640dfb604dec0466c4e0787e68d8e258`.

Unauthorized infra head at issue creation (audit only, not a pin target):
`d8720829b176cf1287e633f9382989fc8f258105`.

## Authorized path

Record Path A (default restore) or Path B (only if adoption recorded
`VOC-145-DEP-07`):

`pending — Path A unless adoption_resolved_decisions names Path B.`

## Issue #1124 incident record

| Item | Value |
|------|-------|
| Last governed VOC-142 infrastructure base | `8993e867640dfb604dec0466c4e0787e68d8e258` |
| Unauthorized infra head | `d8720829b176cf1287e633f9382989fc8f258105` |
| Infra self-CI run claimed green | `33443684483` |
| Caller pin at issue creation | `8993e867640dfb604dec0466c4e0787e68d8e258` |
| `reviewer` drift | `effort=high,fast=false` → `effort=xhigh,fast=false` |
| `reviewer_fast_retry` drift | `effort=high,fast=false` → `effort=xhigh,fast=true` |
| `plan_reviewer` drift | `effort=high,fast=false` → `effort=xhigh,fast=false` |
| VOC-117 tests | rewritten on infra `main` to bless the new values |
| Unreconciled | `README.md`, `CHANGELOG.md`, caller fixtures/pin, VOC-142 DEP-06/AC-09 current-state evidence |
| Explicitly excluded | issue #1120 VOC-112 provenance EHR escalation |

A green test changed in the same direct sequence does not establish
governed authority.

## Binding contract

| Role | Path A (default) | Path B (adoption only) |
|------|------------------|------------------------|
| `implementer` | `cursor/composer-2.5` | `cursor/composer-2.5` |
| `implementer_escalation` | `cursor/composer-2.5` | `cursor/composer-2.5` |
| `planner` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=high,fast=false]` |
| `reviewer` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=false]` |
| `reviewer_fast_retry` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=true]` |
| `plan_reviewer` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=false]` |

Historical VOC-117 expected bindings (must remain named/historical on Path
B; restored as current-state on Path A): the Path A column.

## Exhaustive source-search disposition

Record searched patterns and path disposition after implementation. Pattern
families must include:

| Pattern family | Examples searched |
|----------------|-------------------|
| Role literals | `reviewer:`, `reviewer_fast_retry:`, `plan_reviewer:`, `effort=high`, `effort=xhigh`, `fast=true` |
| VOC-117 expectations | `VOC117_BINDINGS`, `test_voc117_role_bindings.py` |
| Pin and hash assertions | `8993e867`, `d8720829`, `CURRENT_PIN`, `AUTHORITATIVE_PIN`, `PINNED_SHA` |
| Caller hard-codes | `test_voc136_caller_replacement.py`, `test_voc137_pr_sha_scan.py` |
| Docs | infra `README.md`, infra `CHANGELOG.md`, fixture README |
| Historical packages | `specs/changes/VOC-142-…`, `VOC-141-…`, `VOC-140-…`, `VOC-117-…` must remain unmodified |
| Exclusions | VOC-112 fixtures; issue #1120 resume helpers |

## Validation commands (after the repair is tracked and committed)

Record exact commands and results. Required:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
git diff --check
```

Do not treat a missing suite or an untracked-only pass as acceptance.

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR
must bind the live head exactly. Merge-gate must reject any mismatch. This
file must not be required to contain that live head SHA in the same commit.

Independent review must explicitly evaluate: authorized path; six exact
current bindings; historical-versus-current VOC-117 split;
README/CHANGELOG/fixture reconciliation; pin not equal to `d8720829…`;
unchanged retry/exact-SHA/fail-closed controls; exclusion of issue #1120.

## Release handoff

After the exact reviewed caller merge, ordinary later promotion uses
existing release evaluation. Do not snapshot the develop/main gap. Do not
resume issue #1120. Do not create a duplicate promotion PR or release
audit. Root issue #1124 closes only after allowlisted metadata from a
successful implement/release path exists. Unauthorized head `d8720829…`,
governed base `8993e867…`, and run `33443684483` remain incident evidence
only.
