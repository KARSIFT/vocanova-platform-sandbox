# VOC-145 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable `config/roles.yml` and VOC-117 regression expectations; named
  current-state docs including infra `README.md` and `CHANGELOG.md`.
- Prerequisites: confirm live caller `PINNED_SHA.txt` still equals
  `8993e867640dfb604dec0466c4e0787e68d8e258`, or record a later pin if
  another package advanced it, and still confirm that pin's `roles.yml`
  stores the VOC-142 DEP-06 bindings. Confirm live infra `main` still
  differs at `reviewer`, `reviewer_fast_retry`, `plan_reviewer`, and
  `tests/test_voc117_role_bindings.py`, or record if that drift was
  already reverted without docs/pin reconciliation. Confirm self-CI run
  `33443684483` remains incident evidence, not authority. Confirm adoption
  either left Path A as default or recorded `VOC-145-DEP-07` Path B.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits do not count as protected-file drift. Do
  not edit `specs/changes/VOC-142-…/`, `VOC-141-…/`, `VOC-140-…/`, or
  `VOC-117-…/`.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-145
  carrier from current `develop`. Do not treat the untracked local
  `karsift-ai-infra/` checkout as this repo's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not rotate `KARSIFT_BOT_*` secrets.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.
- Do not pin `d8720829…`. Do not recapture VOC-112 fixtures. Do not resume
  issue #1120.

## File reconciliation and implementation sequence

### T00 — Governed reconciliation of the unauthorized role-binding and VOC-117 expectation drift

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra` `config/roles.yml` | restore (Path A) or current-state update (Path B) | Default restore `8993e867…` exact bindings and header comments |
| `KARSIFT/karsift-ai-infra` `tests/test_voc117_role_bindings.py` | restore (Path A) or split historical/current (Path B) | Must not rewrite VOC-117 history to bless `xhigh` |
| `KARSIFT/karsift-ai-infra` `README.md` | modify | Describe the authorized current lineup; Path A records that the ungoverned `xhigh` drift is not current |
| `KARSIFT/karsift-ai-infra` `CHANGELOG.md` | modify | Record the governed reconciliation; do not leave the 2026-08-25 high-effort note as the only current role story if Path B is adopted |
| Caller `tooling/governance/fixtures/karsift-ai-infra/**` | replace from new infra merge | Pin `PINNED_SHA.txt` to that exact merge; mirror every changed authoritative file |
| Caller `tooling/governance/tests/` | extend/reconcile | Advance every current-pin, mirrored-hash, and live reviewer-binding assertion (`test_voc117_role_bindings.py`, `test_voc136_caller_replacement.py`, `test_voc137_pr_sha_scan.py`, and any additional match) while preserving historical pin constants |
| `specs/changes/VOC-142-…/`, `VOC-141-…/`, `VOC-140-…/`, `VOC-117-…/` | **do not modify** | Audit evidence |
| `specs/changes/VOC-145-.../t00-evidence.md` | update | Record implementation PR base, new infra merge, authorized path, binding table, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Read adoption records. If `VOC-145-DEP-07` Path B is not named, implement
   Path A. Do not infer Path B from issue #1124's "either/or" wording.
3. Exhaustively search tracked source/docs for the six role literals,
   `xhigh`, `VOC117_BINDINGS`, pin hashes, and current-state
   `effort=high,fast=false` review claims; record each match as update,
   historical, or irrelevant.
4. Open the coordinated `KARSIFT/karsift-ai-infra` PR from current infra
   `main`. Implement D01–D08 there. Do not treat an untracked nested
   checkout as already-merged work. Do not pin or merge `d8720829…` as the
   reconciliation.
5. Obtain independent exact-revision review of that infra PR and merge it.
   Record the exact merge SHA. Confirm it is not `d8720829…`.
6. From current caller `develop`, create a new VOC-145 implementation
   branch. Set `PINNED_SHA.txt` and mirror every changed authoritative
   fixture file from that exact merge. Update the named current-state docs
   and caller tests. Reconcile all current-pin and mirrored-hash
   assertions; preserve historical authoritative-pin constants and package
   records.
7. Confirm no VOC-142/VOC-141/VOC-140/VOC-117 package file is staged.
   Confirm no OpenAI route, retry-cap expansion, exact-SHA skip,
   App-token mint change, or #1120/VOC-112 recapture helper was added.
8. Track and commit the caller repair. Re-run suites against the committed
   tree. A pass obtained only while untracked is not acceptance.
9. Record evidence in `t00-evidence.md`, including authorized path,
   binding table, source-search disposition, and the exact-head binding
   contract. This package's caller PR `Closes` only its own VOC-145 task
   issue.
10. After the exact reviewed caller PR merges, ordinary later promotion
    uses existing release evaluation. Do not add a snapshot-gap task.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set, after
the repair is tracked and committed:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
git diff --check
```

Also record the exact targeted infra and caller commands that prove:

- live `roles.yml` contains exactly the authorized six bindings;
- Path A restored VOC-117 current-state expectations, or Path B preserved
  historical VOC-117 constants separately from current-state constants;
- effort-omitted Grok 4.6 and missing `CURSOR_API_KEY` still fail closed;
- `PINNED_SHA.txt` equals the new infra merge and is not `d8720829…`;
- README/CHANGELOG/fixture README describe the authorized current set;
- retry cap, exact-SHA review wiring, and no-OpenAI constraints remain;
- VOC-142/VOC-117 package directories are absent from the implementation
  diff.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA, and the infra PR SHA)
should confirm:

- authorized path (A by default, B only if adoption recorded it);
- six exact current bindings;
- historical VOC-117 assertions were not rewritten to bless a later lineup;
- `PINNED_SHA.txt` equals the new infra merge and is not `d8720829…`;
- current-state source search is exhaustive; current docs match the
  authorized lineup;
- VOC-142, VOC-141, VOC-140, and VOC-117 package records are unchanged;
- no OpenAI route, retry weakening, or exact-SHA skip;
- `t00-evidence.md` names the implementation PR base, new infra merge, and
  authorized path, states that the live head is bound by the App-authored
  independent-review comment/check, and does not require a commit to
  contain its own SHA;
- the independent-review comment binds this exact live PR head; merge-gate
  would reject a mismatch;
- no snapshot-gap task and no #1120 carrier were used;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected and must run only for a real tree change, not for
  tree-equivalent post-promotion sync.
- **Operational effect:** After this repair is live, governed roles invoke
  the authorized current bindings, current-state tests and docs match that
  set, and unauthorized `d8720829…` is not the caller pin.
- **Rollback trigger:** live bindings again disagree with the authorized
  set; VOC-117 tests again rewritten to bless an ungoverned lineup; pin
  equals `d8720829…`; retry/exact-SHA/fail-closed controls weakened;
  OpenAI route added; snapshot-gap commit; #1120 carrier edits.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` merge and revert the coordinated infrastructure
  PR. Last-known-good caller pin is `8993e867…` unless a later reviewed pin
  superseded it before T00 started. That last-known-good still has the
  #1124 unreconciliation against drifted infra `main`; rollback restores a
  known reviewed caller state, not a claim that infra `main` was already
  governed.
- **Last-known-good reference:** caller `develop` before this package's
  merge and pin `8993e867640dfb604dec0466c4e0787e68d8e258` unless
  superseded before T00.
