# VOC-146 — Governance validation accepts an invalid base/head range: Specification

## Objective and requirement source

Stop `scripts/governance/validate-governance.sh` from reporting success when
`--base` or `--head` is unresolved or the requested diff range is invalid.
Malformed pull-request range metadata must fail closed. Apply the same
status-preserving range loading to `classify-change-risk.sh`, which uses the
identical process-substitution pattern.

**Requirement source:** [GitHub issue #1127](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1127).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1127)

| Item | Value |
|------|-------|
| Repository | `KARSIFT/vocanova-platform-sandbox` |
| Reproduction head | `79b2b3f1f4224235bdda3f77ee887c3004978deb` |
| Nonexistent `--base` | `376e00dd769afb0fe850052b3a5cb48f729e73ad` |
| Command | `bash scripts/governance/validate-governance.sh --base 376e00dd… --head 79b2b3f1…` |
| Observed Git error | `fatal: Invalid symmetric difference expression 376e00dd…...79b2b3f1…` |
| Observed success line | `Governance structure validation passed.` |
| Observed exit status | `0` |
| Nested loader | `scripts/governance/validate-monitoring-impact.sh` `mapfile -t files < <(git diff … "$base...$head")` |
| Twin loader | `scripts/governance/classify-change-risk.sh` line 30, same `mapfile` process substitution |
| Discovery context | local exact-range verification of emergency PR #1126; unrelated to VOC-112 |

Live contract at issue creation:

- `validate-governance.sh` forwards `--base`/`--head` to
  `validate-monitoring-impact.sh` and prints success after a zero nested
  exit.
- The nested `git diff` failure is swallowed by `mapfile` over process
  substitution even under `set -euo pipefail`.
- An empty changed-file list is treated as a valid range.
- `classify-change-risk.sh` with the same invalid range currently prints
  `No changed files to classify.` and exits `0`.
- A `pull_request` event with no resolved range already fails closed
  (VOC-086). That path is not this defect and must stay fail-closed.
- Working-tree fallback when neither `--files-from` nor both `--base` and
  `--head` are requested is existing local/push behavior.

## Scope and non-goals

### In scope

1. Resolve both `--base` and `--head` as commits before loading a changed-file
   list from `$base...$head`.
2. Capture `git diff` through a status-preserving path so a nonzero Git
   status is the script's status. Do not load an empty file list from a
   failed range.
3. `validate-governance.sh` must return nonzero and must not print
   `Governance structure validation passed.` when the requested range is
   unresolved or invalid.
4. `classify-change-risk.sh` must return nonzero on the same invalid-range
   class instead of classifying an empty list.
5. If either `--base` or `--head` is supplied without the other, fail closed
   rather than silently falling through to working-tree discovery.
6. Deterministic negative tests for nonexistent `--base`, nonexistent
   `--head`, and unrelated/no-merge-base revisions.
7. Preserve valid existing `--base`/`--head` ranges, `--files-from`,
   `--declarations-only`, VOC-086 `pull_request` missing-range fail-closed,
   and working-tree fallback when no range was requested.
8. Current-state docs found by exhaustive search, including `AGENTS.md`'s
   monitoring-impact range paragraph.
9. Evidence, validation after commit, and ordinary release handoff.

### Non-goals / explicitly excluded

- Changing `monitoring_impact` declaration semantics, monitor/synthetic
  inventory, or `infra/monitoring/validate-monitoring-impact.mjs` except as
  required to consume a correctly loaded file list.
- Recapturing VOC-112 JSON fixtures or hashing sources.
- Editing emergency PR #1126, rewriting VOC-086/VOC-112 package records, or
  mixing this repair into the VOC-112 provenance objective.
- Opening a `karsift-ai-infra` PR, advancing `PINNED_SHA.txt`, or mirroring
  fixtures.
- Changing `.github/workflows/` unless exhaustive search proves a current
  workflow claim would remain false; the live workflows already pass
  `--base`/`--head` on pull_request. Do not add `--base`/`--head` to the
  repository-governance push-path classifier invocation as this package's
  work.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- A VOC-097 operator-owned live-evidence second task.
- Snapshotting the current develop/main gap (`karsift-ai-infra#15`).
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: `scripts/governance/` (R4 path floor);
  `AGENTS.md` (R3 floor); foundation tests under `scripts/foundation/` if
  added; this package directory.
- Protected technical effect: whether governance validation and path-based
  risk classification may treat an unresolved or invalid `--base`/`--head`
  range as an empty valid diff. No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance-validation fail-closed changes
  still require exact-SHA independent verification.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-146-D00`: This is one outcome-sized fail-closed range repair. Use one
end-to-end implementation task covering both governance scripts that load
`$base...$head` via `mapfile` process substitution, tests, current-state
docs, evidence, and release handoff. Script count, tests-versus-docs, and
wrapper-versus-classifier are not split reasons. Including
`classify-change-risk.sh` is the same defect class: an invalid range must
not become an empty accepted file list.

`VOC-146-D01`: When both `--base` and `--head` are supplied, resolve each as
a commit object (`git rev-parse --verify --end-of-options "${rev}^{commit}"`
or live equivalent) before diffing. A well-formed 40-character hex that is
not in the object store is unresolved. Do not treat `git diff` stderr as
sufficient if the process still exits 0.

`VOC-146-D02`: Capture `git diff --no-renames --name-only
--diff-filter=ACDMRTUXB "$base...$head"` through a status-preserving path
(temporary file, `PIPESTATUS`, or equivalent). A nonzero Git status must
become the script's nonzero status. Do not use `mapfile < <(git diff …)` as
the load path for a requested range. A successful diff with zero names is a
valid empty change set, not an invalid range.

`VOC-146-D03`: Nonexistent `--base` against a resolvable `--head` must make
`validate-governance.sh` and `validate-monitoring-impact.sh` exit nonzero
and must not print `Governance structure validation passed.` The class of
issue #1127 (`376e00dd…` / `79b2b3f1…`, exit 0 after Git fatal) cannot
recur.

`VOC-146-D04`: Nonexistent `--head` against a resolvable `--base` must fail
the same way.

`VOC-146-D05`: Two resolvable commits with no merge base (unrelated
histories; `git diff A...B` fails) must fail closed. Resolving both
revisions is not sufficient without a status-preserving diff.

`VOC-146-D06`: If either `--base` or `--head` is supplied without the other,
fail closed. Do not fall through to working-tree discovery after a partial
range request. Working-tree fallback remains only when neither
`--files-from` nor any `--base`/`--head` argument was requested.

`VOC-146-D07`: Preserve:

- valid existing `--base`/`--head` ranges used by pull_request CI;
- `--files-from`, including under `GITHUB_EVENT_NAME=pull_request`;
- `--declarations-only` without a range;
- VOC-086 `pull_request` missing-range fail-closed when no `--base`/`--head`,
  `--files-from`, or parseable `GITHUB_EVENT_PATH` exists;
- existing diff-filter and `--no-renames` flags.

`VOC-146-D08`: `classify-change-risk.sh` must use the same fail-closed range
contract. An invalid range must not print `No changed files to classify.` or
exit 0. A valid empty range may still report no changed files.

`VOC-146-D09`: Tests must invoke the live scripts, not only assert that
`mapfile < <(` is absent. Cover D03–D08. A unit test that only greps the
wrapper for a comment is not sufficient coverage of #1127. Prefer
`scripts/foundation/` Node tests so `pnpm test` / `node --test
scripts/foundation/*.test.mjs` runs them. A disposable Git repository may
be used for the unrelated-histories case; the issue reproduction against
this repository's object store remains required for nonexistent `--base`.

`VOC-146-D10`: Docs in the same PR. Before editing, exhaustively search
tracked source and current documentation for claims that a pull_request
changed-file range is fail-closed only when missing, that governance
validation can succeed after Git reports an invalid symmetric difference, or
that `mapfile` over `git diff "$base...$head"` is the load path. Record the
searched patterns and path disposition in `t00-evidence.md`. Update
`AGENTS.md` so CI range fail-closed includes unresolved commits and invalid
diff ranges, not only a missing range. Preserve clearly labeled historical
VOC-086 records. Do not rewrite those package directories.

`VOC-146-D11`: No infrastructure pin, no workflow rewrite unless a live
false claim is found, no VOC-112 recapture, no emergency-PR #1126 edits, no
VOC-097 live-evidence task, and no snapshot of the develop/main gap.

`VOC-146-D12`: Validation after the repair is tracked and committed:

- the issue #1127 reproduction command against a nonexistent `--base` (expect
  nonzero; no success line);
- matching nonexistent `--head` and no-merge-base cases;
- a valid `--base`/`--head` pair from the implementation PR (expect the
  current success contract);
- `--files-from` still succeeds on a well-formed list;
- `bash scripts/governance/validate-governance.sh` with exact implementation
  PR base/head;
- `bash scripts/governance/classify-change-risk.sh` with exact base/head
  (expect R4 for this change);
- `node --test scripts/foundation/voc146-*.test.mjs` or the live equivalent
  added by T00;
- `node --test scripts/foundation/voc086-monitoring-impact.test.mjs`
  (VOC-086 range contracts still pass);
- `git diff --check`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

`VOC-146-D13`: Feasible exact-revision evidence. The App-authored
independent-review comment/check must bind the live PR head exactly and must
explicitly evaluate: nonexistent base, nonexistent head, no-merge-base,
status-preserving diff, preserved valid range and `--files-from`, classifier
parity, and current-state docs. Merge-gate must reject any mismatch.
Committed `t00-evidence.md` records the implementation PR base and the
contract that later exact-head binding is published as review/check
metadata. A tracked file must not be required to contain the SHA of the
same commit that contains it.

`VOC-146-D14`: Protected comparison versus implementation PR base.
Issue-creation reproduction commit is
`79b2b3f1f4224235bdda3f77ee887c3004978deb`. Implementation must resolve
current `develop` to a 40-character SHA before any in-scope edit and record
that SHA as the implementation PR base. Fail closed on unrelated/material
movement of `develop` (any tree change outside this package directory,
in-scope governance scripts, foundation tests, and named current-state
docs). This package's own plan/adoption/roster commits after `79b2b3f1…`
are governance-only and do not count as protected-file drift.

`VOC-146-D15`: Release handoff. Ordinary later promotion of this package
uses `release.yml` at the then-current `develop` tip once required checks
pass. `develop` is advanced to the exact promotion merge SHA before audit
close. Every promotion merge push to `main` triggers automatic production
deployment, whose exact-SHA result is verified. Do not snapshot the current
gap. Closed state alone is not completion proof. Root issue #1127 closes
only after allowlisted metadata from a successful implementation/promotion
path exists.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governance-validation
fail-closed work only. No database, schema, seed, analytics instrumentation,
or user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The repair prevents
governance CI from treating malformed range metadata as an empty valid diff.

Abuse/process risks:

1. Swallowing `git diff` failure through process substitution — forbidden by
   `VOC-146-D02`.
2. Treating a well-formed but missing SHA as an empty change set —
   forbidden by `VOC-146-D01`, `VOC-146-D03`, `VOC-146-D04`.
3. Resolving both revisions but ignoring a failed three-dot diff —
   forbidden by `VOC-146-D05`.
4. Falling through to working-tree discovery after a partial `--base` or
   `--head` request — forbidden by `VOC-146-D06`.
5. Changing monitoring-impact declaration semantics or inventory to make
   tests pass — forbidden by `VOC-146-DEP-04`.
6. Mixing VOC-112 / PR #1126 provenance work into this package — forbidden
   by `VOC-146-D11`.
7. Snapshotting the develop/main gap — forbidden by `VOC-146-D11` and
   `karsift-ai-infra#15`.
8. Requiring a commit to contain its own SHA — forbidden by `VOC-146-D13`.
9. Printing credentials or copying full CI logs into evidence — forbidden.

## Contradictions and open questions

1. **AGENTS.md currently fail-closes a missing range, not an invalid one.**
   VOC-086 required a `pull_request` run without a resolved changed-file
   range to fail closed. Live code still does that, but an *invalid*
   resolved pair is accepted. This package follows D10 and treats that gap
   as in-scope residual wording, not as a rewrite of VOC-086.
2. **`classify-change-risk.sh` is not named in the issue title** but uses
   the identical loader. This package includes it under D00/D08 as the same
   defect class. Leaving it fail-open would leave governance-policy.yml's
   `Enforce path-based risk floor` step accepting the same malformed range.
3. **repository-governance.yml push path still calls
   `classify-change-risk.sh` without `--base`/`--head`.** That is existing
   working-tree behavior and is out of scope. Do not expand this package
   into a workflow rewrite.
4. **Untracked local `karsift-ai-infra/` checkout:** if present in the
   workspace, it is not this repository's tracked tree and is not
   implementation evidence.
5. **Exact implementation PR base SHA is not available at drafting time.**
   Record it before the first in-scope edit per D14.
