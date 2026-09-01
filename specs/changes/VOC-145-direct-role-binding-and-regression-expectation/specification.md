# VOC-145 — Direct role-binding and regression-expectation drift lacks governed reconciliation: Specification

## Objective and requirement source

Stop unauthorized `KARSIFT/karsift-ai-infra` `main` role, effort, and speed
changes, and the VOC-117 regression-expectation rewrite that blessed them,
from remaining the undeclared live contract. Land one governed
reconciliation so live `config/roles.yml`, current-state tests, caller
fixtures/pin, README, CHANGELOG, and every other current-state document
describe the same authorized binding set.

**Requirement source:** [GitHub issue #1124](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1124).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1124)

| Item | Value |
|------|-------|
| Repository | `KARSIFT/karsift-ai-infra` drift, caller `KARSIFT/vocanova-platform-sandbox` unreconciliation |
| Last governed VOC-142 infrastructure base | `8993e867640dfb604dec0466c4e0787e68d8e258` |
| Unauthorized infra head | `d8720829b176cf1287e633f9382989fc8f258105` |
| Infra self-CI run claimed green | `33443684483` |
| Caller pin at issue creation | `8993e867640dfb604dec0466c4e0787e68d8e258` |
| Reproduction | `git -C ../karsift-ai-infra diff 8993e867640dfb604dec0466c4e0787e68d8e258..origin/main -- config/roles.yml tests/test_voc117_role_bindings.py README.md CHANGELOG.md` |

Drift from the last governed VOC-142 infrastructure base:

| Role | Last adopted | Ungoverned `main` |
|------|--------------|-------------------|
| `implementer` | `cursor/composer-2.5` | unchanged |
| `implementer_escalation` | `cursor/composer-2.5` | unchanged |
| `planner` | `cursor/grok-4.6[effort=high,fast=false]` | unchanged |
| `reviewer` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=false]` |
| `reviewer_fast_retry` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=true]` |
| `plan_reviewer` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=false]` |

Live caller fixture at pin `8993e867…` still stores the last adopted
bindings. Live infra `tests/test_voc117_role_bindings.py` rewrote
`VOC117_BINDINGS` to the drifted values and generalized
`prepare_cursor_model` assertions so they no longer require
`effort=high,fast=false` for review roles. Live infra `README.md` and
`CHANGELOG.md` were not updated to describe the new lineup. Adopted VOC-142
DEP-06 / AC-09 remain the last governed current-state claim and must stay
historical package records.

A green self-CI run at `33443684483` does not confer adoption, pin, or
current-state authority.

## Scope and non-goals

### In scope

1. One authorized current binding set on live infra `config/roles.yml` and
   the caller mirrored fixture, with header comments that describe that set
   and do not describe a stale or ungoverned lineup as current.
2. Default Path A: restore the last adopted exact bindings from `8993e867…`
   / VOC-142 DEP-06, including restoration of VOC-117 current-state
   regression expectations that were rewritten to bless the ungoverned
   values.
3. Adoption-only Path B: if adoption records `VOC-145-DEP-07` Path B,
   formally adopt the drifted `xhigh` review lineup as the new current-state
   contract, keep historical VOC-117 expected bindings as historical
   constants, and update current-state tests/docs/fixtures without claiming
   VOC-117 originally required `xhigh` or `fast=true` retry.
4. `README.md`, `CHANGELOG.md`, fixture README, caller pin, mirrored
   fixture files, and every other current-state match found by exhaustive
   tracked-source search.
5. Deterministic regressions for the authorized six bindings, historical-
   versus-current VOC-117 assertions, fail-closed effort-omitted and missing
   credential paths, pin/mirror identity, and unchanged retry / exact-SHA /
   provider-isolation controls.
6. Pin advance to a new independently reviewed infrastructure merge. Do not
   pin `d8720829…` as-is.

### Non-goals / explicitly excluded

- Using this package as a carrier for the VOC-112 provenance EHR
  escalation in issue [#1120](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1120).
- Weakening retry caps, exact-SHA independent review, provider isolation,
  or fail-closed model resolution.
- Adding an OpenAI execution route or requesting `OPENAI_API_KEY`.
- Pinning unauthorized head `d8720829…` because self-CI run `33443684483`
  was green.
- Rewriting VOC-117, VOC-142, VOC-141, VOC-140, or earlier package records
  under `specs/changes/`.
- Fetching, hydrating, or recapturing VOC-112 JSON fixtures or hashed sources.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- A VOC-097 operator-owned live-evidence second task.
- Application runtime, deployment topology, credential-value, production
  merge-guard, or monitor-inventory changes.
- Expanding implementer retry beyond the existing two-attempt bound, or
  making `reviewer_fast_retry` skip exact-SHA review.
- Self-adoption or self-authorization of this package.
- Treating an untracked local `karsift-ai-infra/` checkout as this
  repository's tracked tree.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable `config/roles.yml` and VOC-117 regression expectations; named
  current-state docs found by the required exhaustive search.
- Protected technical effect: which model, effort, and speed each governed
  role invokes. No application runtime effect is intended.
- EHR: not triggered by this package. Issue #1120 remains a separate
  stopped operation and is not this package's trigger or carrier.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but agent-authority and governance-test changes
  still require exact-SHA independent verification and fail-closed
  controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-145-D00`: This is one outcome-sized authority reconciliation. Use one
end-to-end implementation task covering the coordinated infrastructure PR,
caller pin/fixture mirror, authorized current binding set,
historical-versus-current VOC-117 assertions, tests, docs, evidence, and
release handoff. Repository count, restore versus documentation, and
tests-versus-docs are not split reasons. Later promotion of this repair is
evidence of the outcome, not a second task and not a snapshot of the
develop/main gap.

`VOC-145-D01`: Default Path A, unless adoption records Path B. Restore the
last adopted exact bindings from VOC-142 DEP-06 / pin `8993e867…`:

- `implementer: cursor/composer-2.5`
- `implementer_escalation: cursor/composer-2.5`
- `planner: cursor/grok-4.6[effort=high,fast=false]`
- `reviewer: cursor/grok-4.6[effort=high,fast=false]`
- `reviewer_fast_retry: cursor/grok-4.6[effort=high,fast=false]`
- `plan_reviewer: cursor/grok-4.6[effort=high,fast=false]`

Restore VOC-117 current-state regression expectations that were rewritten
to bless the ungoverned values. Header comments on `roles.yml` must again
describe explicit high effort on the Standard tier (`fast=false`) for
planner and review roles.

`VOC-145-D02`: Path B is an adoption-time alternative only. If
`adoption_resolved_decisions` names `VOC-145-DEP-07` Path B, the authorized
current set becomes:

- `implementer: cursor/composer-2.5`
- `implementer_escalation: cursor/composer-2.5`
- `planner: cursor/grok-4.6[effort=high,fast=false]`
- `reviewer: cursor/grok-4.6[effort=xhigh,fast=false]`
- `reviewer_fast_retry: cursor/grok-4.6[effort=xhigh,fast=true]`
- `plan_reviewer: cursor/grok-4.6[effort=xhigh,fast=false]`

The implementer must not select Path B from issue #1124's "either restore
or formally adopt" wording. Path B still requires a new independently
reviewed infra merge; `d8720829…` remains unauthorized because it weakened
historical assertions in the same ungoverned sequence.

`VOC-145-D03`: Historical VOC-117 assertions must not be weakened. VOC-117
AC-00 recorded the six `effort=high,fast=false` review/planner bindings as
that package's observable outcome. Path A restores those as current-state
assertions. Path B must keep a named historical constant equal to the
VOC-117 / `8993e867…` bindings and a separate current-state constant equal
to the Path B lineup. Tests must not rewrite `VOC117_BINDINGS` so that
VOC-117 appears to have always required `xhigh` or `fast=true`.

`VOC-145-D04`: Pin advance. Issue-creation pin `8993e867…` is the last
governed role-binding contract. Unauthorized head `d8720829…` must not
become `PINNED_SHA.txt`. T00 opens one new `KARSIFT/karsift-ai-infra` PR,
obtains independent exact-revision review, and after that merge sets
`PINNED_SHA.txt` and every changed mirrored fixture file to that exact
merge. Mirror at least `config/roles.yml`, `tests/test_voc117_role_bindings.py`,
`README.md`, and `CHANGELOG.md` when those files change. If exact
comparison proves another authoritative fixture file also changed, mirror
it too. Do not treat the untracked local `karsift-ai-infra/` checkout as
this repository's tracked tree. Reconcile all live caller pin-lock tests.
Preserve historical `AUTHORITATIVE_PIN` / issue-era pin constants and
package evidence, including VOC-142 pin `8993e867…` as historical.

`VOC-145-D05`: Docs in the same PR. Before editing, exhaustively search
tracked source and current documentation for the six role literals, `xhigh`,
`VOC117_BINDINGS`, pin hashes, and claims that all review roles use
`effort=high,fast=false` as current. Record the searched patterns and
resulting path disposition in `t00-evidence.md`. Update every current-state
document that would otherwise remain false, including infra `README.md`,
infra `CHANGELOG.md`, fixture README, and caller tests that hard-code the
reviewer binding (`test_voc117_role_bindings.py`,
`test_voc136_caller_replacement.py`, `test_voc137_pr_sha_scan.py`, and any
additional match). Preserve clearly labeled historical VOC-117 / VOC-142
records. Do not rewrite those package directories.

`VOC-145-D06`: Safety controls remain. Two-attempt implementer bound,
`reviewer_fast_retry` as the bounded review retry rather than an extra
implementer attempt, exact-SHA independent review, App-token isolation,
fail-closed missing `CURSOR_API_KEY`, fail-closed unsupported prefixes,
and fail-closed effort-omitted Grok 4.6 identifiers remain. Path B's
`fast=true` on `reviewer_fast_retry` must not expand retry count, skip
exact-SHA review, or silently select another provider or model. No OpenAI
route. No `OPENAI_API_KEY` request. Do not print credential values.

`VOC-145-D07`: This package is not a carrier for issue #1120 or VOC-112
fixture recapture. Do not edit VOC-112 capture fixtures, hashed sources,
the navigator skill, or `package.json` to manufacture #1120 progress. Do
not recapture VOC-112 JSON fixtures.

`VOC-145-D08`: Tests must exercise live `roles.yml` and
`prepare_cursor_model`, not only comments. Include at least: six exact
authorized current bindings; historical VOC-117 constants preserved;
effort-omitted and missing-credential fail-closed paths; pin/mirror
identity with the new infra merge; exhaustive current-state search
disposition; unchanged retry cap and exact-SHA review wiring. A green
self-CI run against rewritten VOC-117 expectations is not coverage of
#1124.

`VOC-145-D09`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and
must explicitly evaluate: authorized path (A or B); six exact current
bindings; historical-versus-current VOC-117 split; README/CHANGELOG/fixture
reconciliation; pin advance not equal to `d8720829…`; unchanged retry /
exact-SHA / fail-closed controls; and exclusion of #1120. Merge-gate must
reject any mismatch. Committed `t00-evidence.md` records the implementation
PR base, new infra merge, authorized path, and the contract that later
exact-head binding is published as review/check metadata. A tracked file
must not be required to contain the SHA of the same commit that contains it.

`VOC-145-D10`: Protected comparison versus implementation PR base.
Issue-creation pin is `8993e867…`. Implementation must resolve current
`develop` to a 40-character SHA before any in-scope edit and record that
SHA as the implementation PR base. Fail closed on unrelated/material
movement of `develop` (any tree change outside this package directory,
in-scope fixture/pin/tests, and the named current-state docs). This
package's own plan/adoption/roster commits are governance-only and do not
count as protected-file drift. VOC-142 / VOC-117 / VOC-141 / VOC-140
package files under `specs/changes/` are out of scope and must not be
edited.

`VOC-145-D11`: Validation after the repair is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- targeted VOC-145 binding, historical-assertion, pin, and safety cases;
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-145-D12`: Release handoff. After the exact reviewed caller merge,
ordinary later promotion uses `release.yml` at the then-current `develop`
tip once required checks pass. `develop` is advanced to the exact promotion
merge SHA before audit close. Every promotion merge push to `main` triggers
automatic production deployment, whose exact-SHA result is verified. Do not
snapshot the current gap. Closed state alone is not completion proof.
Preserve `d8720829…`, `8993e867…`, and run `33443684483` as audit evidence.
Root issue #1124 closes only after allowlisted metadata from a successful
implement/release path exists. Do not create a duplicate promotion PR or
release audit. Do not resume #1120.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation
authority work only. No database, schema, seed, analytics instrumentation,
or user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents
an ungoverned agent-authority transition from remaining live without
adopting a package, pinning a reviewed merge, or reconciling current-state
docs and tests.

Abuse/process risks:

1. Pinning `d8720829…` because self-CI was green — forbidden by
   `VOC-145-D04`.
2. Rewriting VOC-117 tests so historical assertions bless a later lineup —
   forbidden by `VOC-145-D03`.
3. Implementer selecting Path B without adoption recording it — forbidden
   by `VOC-145-D02`.
4. Weakening retry caps, exact-SHA review, provider isolation, or
   fail-closed model resolution — forbidden by `VOC-145-D06`.
5. Adding an OpenAI route or printing credentials — forbidden by
   `VOC-145-D06`.
6. Using this package to bypass issue #1120 — forbidden by `VOC-145-D07`.
7. Rewriting VOC-142 / VOC-117 package directories — forbidden by
   `VOC-145-D05` and `VOC-145-DEP-10`.
8. Snapshotting the develop/main gap — forbidden by `VOC-145-D12` and
   `karsift-ai-infra#15`.
9. Requiring a commit to contain its own SHA — forbidden by `VOC-145-D09`.
10. Treating the untracked nested `karsift-ai-infra/` checkout as tracked
    implementation evidence — forbidden by `VOC-145-D04`.

## Contradictions and open questions

1. **Issue #1124 "either restore or formally adopt":** this draft proposes
   Path A (restore last adopted exact bindings) as the fail-closed default
   because the drifted lineup lacks an adopted package, and a rewritten
   green test is not authority. Path B remains `VOC-145-DEP-07`, deferred
   to adoption. Implementer must not guess Path B.
2. **VOC-142 DEP-06 / AC-09 "roles.yml is unchanged":** those statements
   are VOC-142-era constraints and remain historical in the VOC-142
   package directory. This package does not rewrite them. Current-state
   caller tests and fixtures are reconciled to the VOC-145 authorized set.
3. **VOC-117 AC-00 / AC-02 non-fast review lineup versus Path B
   `fast=true`:** Path B would change the current-state contract relative
   to VOC-117. That is why Path B must keep historical VOC-117 constants
   and why it requires an explicit adoption record rather than an
   implementer choice.
4. **Live infra README cheaper-retry prose versus Path A
   `fast=false` retry:** VOC-117 flattened `reviewer_fast_retry` to the
   same high-effort Standard identifier as `reviewer`. Path A restores
   that governed flattening. Path B would make retry `xhigh,fast=true`
   and must then update README so "cheaper/faster retry" matches the live
   binding without expanding retry count.
5. **Unauthorized head versus last governed pin:** caller pin `8993e867…`
   and live infra `d8720829…` currently disagree. T00 must not leave that
   disagreement in place.
6. **Untracked local `karsift-ai-infra/` checkout:** if present, it may
   already contain the drifted files. It is not this repository's tracked
   tree and is not a substitute for the coordinated infra PR.
7. **Issue #1120:** explicitly out of scope. Remaining VOC-112 provenance
   EHR work is not authorized here and must not ride along.
