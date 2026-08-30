# VOC-137 — Fail closed on PR base/head SHA overrides in every scannable caller executable: Specification

## Objective and requirement source

Correct the VOC-136 exhaustive caller-diff bypass scanner so PR base/head SHA
override detection is content-semantic and filename-independent. Every added
or modified scannable caller executable outside the exact infra fixture mirror
must fail closed when executable behavior assigns `PR_BASE_SHA` or
`PR_HEAD_SHA` in shell, Node, or Python forms relevant to validation/test
invocation. Source-safe scanner construction, arbitrary-filename negative
cases, benign positive controls, the already-pinned infrastructure #167
fixture, VOC-136 audit records, current KARSIFT role bindings, and ordinary
post-merge release must be preserved.

**Requirement source:** [GitHub issue #1083](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1083).
VOC-136 (`specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive`)
already merged caller PR [#1080](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080)
at exact reviewed head `5d0c2350ab9a20ace586eaadd1169203140ffad0` as
`develop` merge `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8`. This package is
the governed correction of the remaining filename-heuristic gap. It is not a
second VOC-136 implementation, not a retry of PR #1080, and not a rewrite of
VOC-136 package records.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1083)

| Item | Value |
|------|-------|
| Exhausted-gap carrier | VOC-136 PR [#1080](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080), merged |
| Exact reviewed VOC-136 head | `5d0c2350ab9a20ace586eaadd1169203140ffad0` |
| VOC-136 `develop` merge | `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` |
| Defective heuristic | `voc136_bypass_scan.py` applies `PR_SHA_SET_PATTERN` only when the relative filename contains `validate-workspace` or ends in `.test.mjs` |
| Counterexample | `scripts/arbitrary-wrapper.sh` with `export PR_BASE_SHA=deadbeef` then `pnpm test` is accepted |
| Pre-merge supervisor evidence | [comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444745877) |
| Reviewer verdict that missed the gap | [comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444825638) |
| Post-merge audit | [comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444861382) |
| Canceled release runs | `33113425829`, `33113547909` |
| Stopped release PR | [#1082](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1082) converted to draft |
| Issue-creation `develop` | `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` |
| Issue-creation `main` | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| Live pin (already #167) | `b263c0c110591cc798b89277dfc35542abb1597b` |
| Protected comparison anchor (VOC-112 eight-path) | `b9e74fc2db4691c48c637639b265d527de9f4505` |
| Why VOC-136 is not retried on #1080 | already merged; its records stay audit evidence; this is a correction package |

## Scope and non-goals

### In scope

1. Remove the filename dependency from PR base/head SHA override detection in
   the existing exhaustive caller-diff scanner
   (`tooling/governance/tests/voc136_bypass_scan.py` or a same-suite successor
   used by the live complete-diff scan). Every added or modified scannable
   caller executable outside `tooling/governance/fixtures/karsift-ai-infra/**`
   must fail closed when executable behavior assigns `PR_BASE_SHA` or
   `PR_HEAD_SHA` in shell, Node, or Python forms relevant to
   validation/test invocation.
2. Reconstruct `PR_SHA_SET_PATTERN` from non-contiguous source-safe parts.
   The scanner module already defines `_PR_BASE_ENV` and `_PR_HEAD_ENV` and
   does not currently use them in that pattern.
3. Make existing contiguous assertion literals that would match the widened
   scan source-safe, including
   `tooling/governance/tests/test_voc136_caller_replacement.py` line 289's
   `export PR_BASE_SHA=` fixture assertion. That edit is in-scope causal
   remediation of the live test module, not a rewrite of VOC-136 package
   records.
4. Add deterministic negative regression coverage using arbitrary filenames
   that do not contain `validate-workspace` and do not end in `.test.mjs`,
   including the four required classes in `VOC-137-D03`.
5. Add positive controls for benign textual discussion, pattern construction,
   and the exact excluded infra fixture mirror.
6. Keep `SCAN_EXCLUDE_PREFIXES` limited to the fixture mirror. Do not restore
   a wholesale exclusion for caller tooling, governance, or tests.
7. Record implementation PR base, scan-scope change, negative-case results,
   pin freeze, validation, and the feasible exact-head binding contract in
   `t00-evidence.md`.
8. After exact-SHA review and merge, ordinary develop-to-main promotion for
   outstanding completed packages including already-merged VOC-136 and this
   package.

### Non-goals / explicitly excluded

- Re-implementing VOC-136, re-mirroring infrastructure #167, or retargeting
  `PINNED_SHA.txt`.
- Opening a `KARSIFT/karsift-ai-infra` PR.
- Rewriting `specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive/`
  or manufacturing VOC-136 completion, review, or merge records.
- Reusing, cherry-picking, or restating PR #1080 review verdicts as this
  package's independent review.
- Changing any of the eight VOC-112 no-change paths.
- Adding a capture-commit fetch helper, hydrate/materialize helper,
  provenance-mode wrapper, import side effect, evidence-stamping helper, skip,
  or equivalent under any filename.
- Excluding `tooling/governance/tests/**`, this regression's own module,
  `scripts/`, or another executable directory wholesale from the scan.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- Fast-forwarding `main` instead of preserving `--merge` promotion commits.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, retry caps, or fail-closed missing-binding behavior.
- Splitting scanner logic, tests, docs, or release reconciliation into
  separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. VOC-136 / VOC-137 promotion/closure are ordinary
  release-path evidence, not a VOC-097 live-evidence gate.
- Cleaning unrelated VOC-128 work or user worktrees.
- Editing live `.github/workflows/pipeline.yml` or other live caller
  workflows. Expected: no live workflow edit.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/tests/` scanner and regressions.
  The VOC-136 mirrored fixture tree is protected *against* change. The eight
  VOC-112 no-change paths are protected *against* change relative to
  `b9e74fc2db4691c48c637639b265d527de9f4505`.
- Protected technical effect: whether an added caller executable can assign
  `PR_BASE_SHA` or `PR_HEAD_SHA` around validation/test execution merely by
  choosing a filename that is not `validate-workspace*` and does not end in
  `.test.mjs`. No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance-test changes still require
  exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-137-D00`: This is one outcome-sized caller scanner correction. Use one
end-to-end implementation task covering scanner behavior, source-safe
literals, required negative and positive cases, pin/fixture freeze, tests,
docs, evidence, and release handoff. File count, test-versus-scanner, and
VOC-136-versus-VOC-137 documentation are not split reasons. VOC-136
promotion after this correction merges is evidence of this outcome, not an
additional VOC-137 task and not a further attempt on PR #1080.

`VOC-137-D01`: Remove the filename dependency from PR base/head SHA override
detection. In live `voc136_bypass_scan.py`, `PR_SHA_SET_PATTERN` is compiled
as a content-semantic assignment matcher but is invoked only when:

```python
relative != "package.json" and (
    "validate-workspace" in relative or relative.endswith(".test.mjs")
)
```

Delete that filename gate. Apply the pattern to every path that
`should_scan_path` already selects, except the fixture-mirror exclusion
already enforced by `is_scan_excluded`. `package.json` remains scannable for
other bypass classes; PR SHA assignment in `package.json` must also fail
closed if present. Do not introduce a replacement heuristic based on
suffix, directory, or helper name.

`VOC-137-D02`: Preserve source-safe scanner construction. Rebuild
`PR_SHA_SET_PATTERN` from the already-defined `_PR_BASE_ENV` and
`_PR_HEAD_ENV` parts (or equivalent non-contiguous concatenation), matching
how `PROVENANCE_MODE_SET_PATTERN` uses `_PROV_ENV`. Test and assertion
literals that would otherwise match the scanner must remain non-contiguous.
Contiguous forbidden payloads may be reconstructed only in memory (or in
non-executable, non-scanned test data). The tracked committed scanner and
test modules must be scan-clean after commit.

`VOC-137-D03`: Required negative unit cases, each under an arbitrary filename
that does not contain `validate-workspace` and does not end in `.test.mjs`:

1. Shell: relative `scripts/arbitrary-wrapper.sh` (the issue example) with
   `export PR_BASE_SHA=deadbeef` then `pnpm test`.
2. Node: an arbitrary `*.mjs` / `*.js` path such as
   `scripts/arbitrary-head-wrapper.mjs` assigning `process.env.PR_HEAD_SHA`
   before validation/test execution.
3. Python: an arbitrary `*.py` path such as `scripts/arbitrary_wrapper.py`
   assigning `os.environ["PR_BASE_SHA"]` before validation/test execution.
4. An added or modified `*.py` file outside the fixture mirror: the live
   complete-diff scan must include every added/changed `*.py` under the
   caller tree except the fixture mirror, and a negative in-memory payload
   whose relative path is a `*.py` file outside
   `tooling/governance/fixtures/karsift-ai-infra/` must fail closed.

The existing VOC-136 case `validate-workspace-wrapper.mjs` may remain as a
non-sufficient extra, but it does not satisfy this decision because its
filename still matches the defective heuristic.

`VOC-137-D04`: Positive controls. Prove the scanner does not fail on:

1. benign textual discussion that names the environment variables without
   assigning them (comment / assertion prose reconstructed from source-safe
   parts);
2. the scanner's own source-safe pattern construction;
3. the exact excluded infra fixture mirror, including
   `tooling/governance/fixtures/karsift-ai-infra/config/run-app-checks.sh`,
   which legitimately contains `export PR_BASE_SHA=` as the #167 contract.

`should_scan_path` must continue to return false for that fixture path.
Do not restore a wholesale exclusion for caller tooling, governance, or
tests.

`VOC-137-D05`: Pin freeze. `PINNED_SHA.txt` remains
`b263c0c110591cc798b89277dfc35542abb1597b`. No file under
`tooling/governance/fixtures/karsift-ai-infra/` may appear in the
implementation diff. Issue #1083's phrase "eight caller workflow/fixture
paths" is bound to the complete VOC-136-D02/D11 mirrored set plus
`PINNED_SHA.txt`, not a subset:

| Fixture path | VOC-136-D11 SHA-256 |
|--------------|---------------------|
| `.github/workflows/ci.yml` | `0e0d485359d31325bf8b4c41b2047752ac42c6a5139251bd46b03cf7d671a9bb` |
| `.github/workflows/implement.yml` | `e0612aa46dff58d3c06ff338864af3fa32cc725f151235cbe8b6789a80995d2a` |
| `.github/workflows/release.yml` | `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` |
| `config/run-app-checks.sh` | `2ee4eaa25788af72146eee1ef9adb8cc9f42f2c4077d24e6aed25b55deddd1b2` |
| `config/implementer_nested_checkout.py` | `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` |
| `tests/test_app_check_context.py` | `74b8a0c1bcc00a137801e28888b3e6b78371934c14a99ea8eb34f4e0793bb5e0` |
| `tests/test_release_policy.py` | `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` |
| `tests/test_voc121_implement_policy.py` | `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` |
| `tests/test_voc123_source_bundle.py` | `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` |
| `CHANGELOG.md` | `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` |

If any of those bytes already drifted from #167 at implementation dispatch,
fail closed rather than composing with an unexpected fixture tree.

`VOC-137-D06`: Do not add `tooling/governance/tests/` (or this module, or
`scripts/`, or another executable directory) to `SCAN_EXCLUDE_PREFIXES`.
That is the VOC-135 attempt-2 class. The only allowed executable-tree
exclusion remains `tooling/governance/fixtures/karsift-ai-infra/**`.

`VOC-137-D07`: Eight VOC-112 no-change paths remain byte-identical to
protected comparison anchor `b9e74fc2db4691c48c637639b265d527de9f4505` and
must be absent from the implementation diff against that SHA:

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
- `scripts/foundation/voc112-navigation-benchmark.test.mjs`
- `scripts/foundation/voc112-navigation-benchmark-run.mjs`
- `scripts/foundation/validate-workspace.mjs`
- `AGENTS.md`
- `.agents/skills/vocanova-repo-navigator/SKILL.md`
- `package.json`

JSON `subject_revision` remains `f9d11e232a07c7d7a9c433d02c9267912543ba10`.
Those paths currently contain legitimate `process.env.PR_BASE_SHA`
assignments inside the provenance test; they stay out of the scan because
they are not in the implementation diff. Do not edit them to satisfy this
package. Do not retarget the protected comparison anchor.

`VOC-137-D08`: Roles and credentials. Fixture `config/roles.yml` remains
implementer / `implementer_escalation` `cursor/composer-2.5` and planner /
reviewer / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route. No
`OPENAI_API_KEY` request. Do not print credential values.

`VOC-137-D09`: Validation after the regression is tracked and committed:

- `bash scripts/governance/validate-governance.sh` with exact base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4);
- `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
- `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
- targeted VOC-137 negative/positive scanner cases;
- `git diff --check`;
- a direct reproduction that
  `scan_changed_path_for_bypasses("scripts/arbitrary-wrapper.sh", <issue payload>)`
  raises.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-137-D10`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate the arbitrary-filename shell/Node/Python negative cases
and benign controls. Merge-gate must reject any mismatch. Committed
`t00-evidence.md` records the implementation PR base, scan-scope change,
negative-case results, pin freeze, validation commands, and the contract
that later exact-head binding is published as review/check metadata. A
tracked file must not be required to contain the SHA of the same commit that
contains it. Tests must not rewrite evidence at runtime.

`VOC-137-D11`: Preserve VOC-136 audit evidence. Do not rewrite
`specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive/`. Do
not manufacture a VOC-136 completion marker. Do not treat PR #1080's PASS
as this package's independent review. Extending or source-safe-editing live
modules under `tooling/governance/tests/` is allowed and required; those
modules are executable caller tests, not VOC-136 package records. This
package's implementation PR `Closes` only its own VOC-137 task issue. Root
issue #1083 closes only after promotion/audit evidence exists.

`VOC-137-D12`: After this correction merges into `develop`, ordinary
governed release reconciles outstanding completed packages (including
already-merged VOC-136 and this package), promotes `develop` to `main`,
deploys where applicable, and converges `develop` to the exact resulting
main merge SHA without a tree-equivalent staging loop. Do not snapshot the
current develop/main gap (`karsift-ai-infra#15`). Closed state alone is not
completion proof. Canceled release runs `33113425829` and `33113547909` and
draft release PR #1082 are audit context only.

`VOC-137-D13`: Protected comparison versus implementation PR base.
Protected comparison anchor for the eight VOC-112 paths remains
`b9e74fc2db4691c48c637639b265d527de9f4505`. Issue-creation `develop`
`0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` already contains the VOC-136
scanner and tests. Implementation must resolve current `develop` to a
40-character SHA before any in-scope edit and record that SHA as the
implementation PR base. Fail closed on unrelated/material movement of
`develop` (any tree change outside this package directory and the in-scope
`tooling/governance/tests/` scanner correction). This package's own
plan/adoption/roster commits after `0cee20c…` are governance-only and do not
count as protected-file drift. The live complete-diff scan must continue to
cover added/modified scannable executables, including this correction's own
modules, without re-opening VOC-112 no-change paths.

`VOC-137-D14`: Existing VOC-136 complete-diff scan against
`b9e74fc2…` remains in the suite. Widening PR SHA detection therefore
requires the VOC-136 test module's contiguous `export PR_BASE_SHA=`
assertion to become source-safe in the same task; otherwise that already
tracked module would false-positive. Do not disable, skip, or xfail
VOC-136 scanner tests to make this package pass.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The correction prevents
a rename bypass that could wrap validation or tests with attacker-chosen
`PR_BASE_SHA` / `PR_HEAD_SHA` values and mask default `local` fail-closed
provenance. The fixture mirror remains the only executable-tree exclusion so
that legitimate #167 `export PR_BASE_SHA=` in mirrored `run-app-checks.sh` is
not treated as a caller-side override.

Abuse/process risks:

1. Leaving the `validate-workspace` / `.test.mjs` filename gate in place —
   forbidden by `VOC-137-D01`.
2. Replacing that gate with another filename, directory, or helper-name
   heuristic — forbidden by `VOC-137-D01`.
3. Adding `tooling/governance/tests/` to `SCAN_EXCLUDE_PREFIXES` — the
   VOC-135 attempt-2 class, forbidden by `VOC-137-D06`.
4. Committing contiguous forbidden assertion literals that make the tracked
   regression self-fail — the VOC-135 attempt-1 class, forbidden by
   `VOC-137-D02`.
5. Treating the existing `validate-workspace-wrapper.mjs` case as sufficient
   — forbidden by `VOC-137-D03`.
6. Editing the fixture mirror or `PINNED_SHA.txt` — forbidden by
   `VOC-137-D05`.
7. Rewriting VOC-136 package records or restating PR #1080's review as this
   package's review — forbidden by `VOC-137-D11`.
8. Snapshotting the develop/main gap — forbidden by `VOC-137-D12` and
   `karsift-ai-infra#15`.
9. Changing `roles.yml` or adding an OpenAI route — forbidden by
   `VOC-137-D08`.
10. Printing credentials or copying full CI logs into evidence — forbidden.
11. Skipping or xfailing VOC-136 scanner tests to avoid source-safe edits —
    forbidden by `VOC-137-D14`.

## Contradictions and open questions

1. **Issue wording "eight caller workflow/fixture paths":** issue #1083 does
   not enumerate the eight paths. This package binds that phrase to the
   complete VOC-136-D02/D11 mirrored set plus `PINNED_SHA.txt`
   (`VOC-137-D05`) so incidental fixture drift cannot hide behind a guessed
   subset. That is a tightening of "keep pinned," not a new pin.
2. **Live complete-diff base:** VOC-136's `scan_implementation_diff` still
   diffs against `b9e74fc2…`. Widening PR SHA detection therefore collides
   with the contiguous `export PR_BASE_SHA=` literal in
   `test_voc136_caller_replacement.py`. The required resolution is
   `VOC-137-D14` (source-safe that literal), not retargeting the VOC-136
   scan base and not excluding the test module.
3. **`os.putenv`:** `PROVENANCE_MODE_SET_PATTERN` already matches
   `os.putenv("VOC112_CAPTURE_PROVENANCE_MODE")`. `PR_SHA_SET_PATTERN` does
   not currently match `os.putenv("PR_BASE_SHA")`. The required negative
   cases in issue #1083 are `export`, `process.env`, and `os.environ[...]`.
   Adding `os.putenv` for the PR SHA names is compatible fail-closed
   strengthening and is allowed; it is not a substitute for the three named
   cases.
4. **Live caller workflows:** no live `.github/workflows/*.yml` edit is
   expected or in scope.
5. **VOC-136 PASS on #1080:** the exact-head reviewer returned a verdict that
   missed this blocking gap. That PASS remains historical audit evidence for
   VOC-136. It is not sufficient for this correction and must not be reused
   as VOC-137 independent review.
6. **Release PR #1082:** converted to draft when the automatic release was
   stopped. This package does not implement, close, or restage that PR. Ordinary
   release after this correction merges is the recovery path.
