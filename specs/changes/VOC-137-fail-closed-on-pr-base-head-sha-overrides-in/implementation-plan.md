# VOC-137 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/tests/` scanner and regressions.
  The VOC-136 mirrored fixture tree is protected against change. The eight
  VOC-112 no-change paths are protected against change relative to
  `b9e74fc2db4691c48c637639b265d527de9f4505`.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `b263c0c110591cc798b89277dfc35542abb1597b`. Confirm
  `voc136_bypass_scan.py` still gates `PR_SHA_SET_PATTERN` on
  `"validate-workspace" in relative or relative.endswith(".test.mjs")`.
  Confirm `test_voc136_caller_replacement.py` still uses
  `validate-workspace-wrapper.mjs` as its PR SHA negative case and still
  contains a contiguous `export PR_BASE_SHA=` fixture assertion. Confirm
  VOC-136 PR #1080 remains merged as
  `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` from reviewed head
  `5d0c2350ab9a20ace586eaadd1169203140ffad0`. Confirm issue-creation `main`
  remains `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` or that any later main
  movement is unrelated and this package still does not snapshot the gap.
  Confirm the eight VOC-112 paths still match `b9e74fc2…`. Confirm fixture
  `config/roles.yml` still binds implementer/escalation `cursor/composer-2.5`
  and planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits after `0cee20c…` do not count as
  protected-file drift. If any fixture file already drifted from #167, or if
  any of the eight no-change paths differs from the anchor, fail closed.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-137
  carrier from current `develop`. Do not reuse PR #1080. Do not rewrite
  VOC-136 package records.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions,
  rotate `KARSIFT_BOT_*` secrets, or edit `config/roles.yml`.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.

## File reconciliation and implementation sequence

### T00 — Filename-independent PR SHA override fail-closed

| Target | Action | Notes |
|--------|--------|-------|
| `tooling/governance/tests/voc136_bypass_scan.py` | modify | Remove the `validate-workspace` / `.test.mjs` filename gate around `PR_SHA_SET_PATTERN`. Reconstruct that pattern from `_PR_BASE_ENV` / `_PR_HEAD_ENV`. Keep `SCAN_EXCLUDE_PREFIXES` as the fixture mirror only. Optional compatible strengthening: match `os.putenv` for the PR SHA names if it stays source-safe |
| `tooling/governance/tests/test_voc136_caller_replacement.py` | modify | Make the contiguous `export PR_BASE_SHA=` fixture assertion source-safe. Do not skip or xfail existing VOC-136 tests. The existing `validate-workspace-wrapper.mjs` case may remain but is not sufficient |
| `tooling/governance/tests/test_voc137_*.py` (or equivalent addition to the existing suite) | create/extend | Arbitrary-filename shell/Node/Python negative cases; `*.py` outside fixture mirror; benign/pattern/fixture positive controls; pin freeze; no wholesale tests/ exclusion; direct reproduction of `scripts/arbitrary-wrapper.sh` |
| `tooling/governance/fixtures/karsift-ai-infra/**` | **do not modify** | Pin freeze; `PINNED_SHA.txt` stays `b263c0c…`; VOC-136-D11 hashes remain |
| Eight VOC-112 no-change paths | **do not modify** | Must remain byte-identical to `b9e74fc2…` |
| `specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive/` | **do not modify** | Audit evidence |
| `.github/workflows/pipeline.yml` | **do not modify** | No live workflow edit expected |
| `specs/changes/VOC-137-.../t00-evidence.md` | update | Record implementation PR base, scan-scope change, negative/positive results, pin freeze, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. From current `develop`, create a new VOC-137 implementation branch. Do not
   reuse PR #1080. Do not rewrite the VOC-136 package directory.
3. Remove the filename gate and reconstruct `PR_SHA_SET_PATTERN` from
   source-safe parts. Keep fixture-mirror exclusion only.
4. Source-safe the existing VOC-136 contiguous `export PR_BASE_SHA=`
   assertion. Add the required arbitrary-filename negative cases and benign
   / fixture positive controls, including a direct in-memory reproduction of
   the issue payload at `scripts/arbitrary-wrapper.sh`.
5. Confirm no fixture path, no eight-path VOC-112 file, and no VOC-136
   package file is staged. Confirm `roles.yml` is untouched.
6. Track and commit the regression. Re-run the caller governance suite and
   the complete-diff scan against the committed tree. A pass obtained only
   while untracked is not acceptance.
7. Record evidence in `t00-evidence.md`. This package's caller PR `Closes`
   only its own VOC-137 task issue.
8. After the exact reviewed caller PR merges, ordinary release evaluation (or
   `reconcile-release`) completes promotion of outstanding completed
   packages including VOC-136 and this package. Do not add a snapshot-gap
   task.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set, after
the regression is tracked and committed:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc136*.py' -v
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc137*.py' -v
git diff --check
```

If implementation keeps VOC-137 cases inside `test_voc136_caller_replacement.py`
instead of a `test_voc137_*.py` module, record the exact targeted command
actually run. Also record the exact in-process reproduction that
`scan_changed_path_for_bypasses("scripts/arbitrary-wrapper.sh", payload)`
raises. Do not treat a missing suite as a pass. Do not treat an
untracked-only pass as acceptance.

Independent verifier (exact reviewed caller SHA) should confirm:

- the filename gate around `PR_SHA_SET_PATTERN` is gone;
- `scripts/arbitrary-wrapper.sh` with the issue payload fails closed;
- Node `process.env.PR_HEAD_SHA` and Python `os.environ["PR_BASE_SHA"]`
  wrappers fail closed under arbitrary names that do not contain
  `validate-workspace` and do not end in `.test.mjs`;
- added/modified `*.py` outside the fixture mirror is scanned;
- benign discussion, source-safe pattern construction, and the excluded
  fixture `run-app-checks.sh` do not false-positive;
- `SCAN_EXCLUDE_PREFIXES` does not list `tooling/governance/tests/`;
- VOC-136 scanner tests were not skipped or xfailed;
- `PINNED_SHA.txt` equals `b263c0c…` and no fixture-mirror path is in the
  diff;
- all eight VOC-112 no-change paths are absent from the diff against
  `b9e74fc2…`;
- `roles.yml` is unchanged and no OpenAI route was added;
- `t00-evidence.md` names the implementation PR base, states that the live
  head is bound by the App-authored independent-review comment/check, and
  does not require a commit to contain its own SHA;
- the independent-review comment binds this exact live PR head and
  explicitly evaluates the arbitrary-filename shell/Node/Python cases and
  benign controls; merge-gate would reject a mismatch;
- VOC-136 package records were not rewritten; PR #1080's review was not
  reused as this verdict;
- no snapshot-gap task and no bootstrap exception were used;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected and must run only for a real tree change, not for
  tree-equivalent post-promotion sync.
- **Operational effect:** After this correction is live, an added caller
  executable that assigns `PR_BASE_SHA` or `PR_HEAD_SHA` around
  validation/tests fails the exhaustive scan regardless of filename.
  Already-merged VOC-136 pin/fixture bytes stay at #167. Ordinary release
  can then promote outstanding completed packages.
- **Rollback trigger:** filename gate still present; arbitrary-wrapper
  example accepted; `tooling/governance/tests/` excluded from the scan;
  fixture pin retargeted; eight VOC-112 paths rewritten; VOC-136 records
  rewritten; evidence mutated at test time; self-referential exact-head SHA
  required; PR #1080 review reused as this verdict; snapshot-gap commit;
  `roles.yml` / OpenAI route changed.
- **Rollback mechanism:** Revert the caller scanner/test/doc changes to the
  last reviewed `develop` merge
  `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8`. That last-known-good still has
  the filename heuristic; rollback restores a known reviewed state, not a
  complete fail-closed PR SHA scan. Do not roll back by reverting
  infrastructure #167 or by restoring VOC-136 package rewrites that this
  package forbids.
- **Last-known-good reference:** caller `develop` before this package's merge
  (`0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` at issue creation), pin
  `b263c0c110591cc798b89277dfc35542abb1597b`, VOC-112 `subject_revision`
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`, and unmodified eight no-change
  paths relative to `b9e74fc2db4691c48c637639b265d527de9f4505`.
