# VOC-117-T00 — Evidence

Task: `VOC-117-T00` — Apply the Cursor role lineup, parameterized-model routing,
tests, docs, and caller pin.

Do not record secrets, credentials, session values, OAuth material, personal data,
or complete CI logs.

## Changed surfaces

### Shared infra primary source (`KARSIFT/karsift-ai-infra`)

- `config/roles.yml` — six VOC-117-D00 bindings; current-state header; historical
  OpenAI/Codex/cursor/auto commentary removed from active-state section.
- `config/prepare_cursor_model.py` — fail-closed Cursor model preparation: strip
  `cursor/` prefix while preserving bracket parameters; optional `CURSOR_API_KEY`
  requirement; unsupported-prefix rejection.
- `.github/workflows/plan.yml` — planner resolves `cursor/grok-4.6[fast=false]`
  through `prepare_cursor_model.py`.
- `.github/workflows/review.yml` — reviewer paths use parameterized Grok 4.6
  Standard form via `prepare_cursor_model.py`; terminal failures expose only a
  bounded classification and withhold raw provider output.
- `.github/workflows/plan-review.yml` — plan_reviewer retains `effort=high` and
  `fast=false` through `prepare_cursor_model.py`; terminal failures use the same
  sanitized classification path.
- `.github/workflows/implement.yml` — implementer, escalation, and self-correct
  steps use `prepare_cursor_model.py` for `cursor/composer-2.5`.
- `config/extract-cursor-result.py` — bounded Cursor failure reason codes without
  printing raw provider responses or credentials; safe step-output transport.
- `config/build-review-failure-comment.py` — validates bounded classification,
  exact PR base/head identity, and non-verdict failure-comment content.
- `tests/test_voc117_role_bindings.py` — VOC-117-TEST-00 through TEST-05 regressions.
- `tests/test_cursor_result.py` and `tests/test_review_failure_comment.py` —
  deterministic output-channel, vocabulary, identity, and CLI regressions.

### Caller mirrored fixture and tests

- `tooling/governance/fixtures/karsift-ai-infra/**` — synchronized copies of the
  infra changes above.
- `tooling/governance/fixtures/karsift-ai-infra/README.md` — VOC-117 contract note.
- `tooling/governance/tests/test_voc117_role_bindings.py` — caller fixture regressions.
- `scripts/foundation/voc097-fixture-matrix.test.mjs`,
  `scripts/foundation/voc104-ready-for-review-reuse.test.mjs`, and
  `scripts/foundation/voc108-authoritative-lifecycle.test.mjs` — pin assertions
  advanced to the VOC-117 fixture merge so `pnpm test` / pre-push validation stay
  aligned with `PINNED_SHA.txt`.

## Authoritative stored role mapping (VOC-117-D00)

| Role | Stored value |
|------|----------------|
| `implementer` | `cursor/composer-2.5` |
| `implementer_escalation` | `cursor/composer-2.5` |
| `planner` | `cursor/grok-4.6[fast=false]` |
| `reviewer` | `cursor/grok-4.6[fast=false]` |
| `reviewer_fast_retry` | `cursor/grok-4.6[fast=false]` |
| `plan_reviewer` | `cursor/grok-4.6[effort=high,fast=false]` |

CLI `--model` values after prefix handling:

| Role | Cursor CLI `--model` |
|------|----------------------|
| `planner` | `grok-4.6[fast=false]` |
| `reviewer` | `grok-4.6[fast=false]` |
| `reviewer_fast_retry` | `grok-4.6[fast=false]` |
| `plan_reviewer` | `grok-4.6[effort=high,fast=false]` |
| `implementer` | `composer-2.5` |
| `implementer_escalation` | `composer-2.5` |

Bracket form is passed through unchanged per VOC-117-D02; verified compatible with
the stored binding shape without silent vendor/model substitution.

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Governed Cursor implementation | caller run `32783787908`, job `97612143209` (`cursor-agent`, attempt 1) |
| Authoritative infra PRs | `KARSIFT/karsift-ai-infra#147`, diagnostic classification `#148`, structured diagnostic publication `#149`, annotation channel correction `#150`, isolated bounded failure publisher `#151` |
| Independently reviewed infra heads | `d6ac23f70a10299e73629f257e275854525d15c8`, `9457233e5b2e2fb03674ef963d89ac4767596a4f`, `e7f3804e41658bf45acc5f580134dd76a4a6ea3c`, `6526cf63f2e7ef35750b9eab0ddeb74fdc071af9`, `364a8996f45b4a39298ca7d42f298edca35d773b` |
| Infra merge commits | initial `27a44b298f1c234a94e02127eaeb55d66b28e30d`; classification `42aa66757a521b1187193fba17b74e440964c27f`; annotation publication `12e5cd65159b5315b7e618facb251e0324dcfbb5`; stdout correction `2f2569cb03ef3dbfee8beb956ec125e81c94a785`; authoritative current `21a24db03703b693a363737cbd6e479d50801107` |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Prior pin | `2f2569cb03ef3dbfee8beb956ec125e81c94a785` |
| New pin | `21a24db03703b693a363737cbd6e479d50801107` |

The first governed Cursor attempt produced the caller-side source projection but
correctly left the prior pin in place while no authoritative source merge existed.
Its exact-SHA review failed closed on that missing source/pin. PR #147 then landed
the authoritative source under the same T00 authority. This caller revision removes
the two unrelated VOC-112 fixture edits identified by that review, synchronizes the
seven mirrored source files byte-for-byte to reviewed infra head `d6ac23f…`, and pins
the exact GitHub merge commit rather than inferring or predeclaring a SHA. A live
reviewer invocation then exposed a diagnostic defect: terminal Cursor application
errors were withheld without a safe reason code. The same T00 authority produced
independently reviewed infra PR #148. Its exact merge `42aa667…` preserves the six
bindings and retry controls while adding bounded error classification. This caller
revision initially mirrored that source. Live runs then proved the classification
remained inside the failed job and was absent from structured check metadata. The
same T00 authority produced infra PR #149, independently reviewed at exact head
`e7f3804…` and merged as `12e5cd6…`. It publishes only the bounded allowlisted
diagnostic as a check annotation, keeps raw provider output withheld, and preserves
the role mappings, bracket parameters, exact-SHA binding, protected checks, and
retry limits. The caller mirror and pin now use that exact authoritative merge.
Live runs then proved the workflow command was written to stderr, which GitHub
Actions does not parse into structured annotations. Infra PR #150 corrected only
that channel, retained ordinary diagnostics on stderr, and added subprocess channel
assertions. Its independently reviewed head `6526cf6…` merged as `2f2569c…`.
Fresh caller runs resolved that exact reusable-workflow SHA but still exposed no
bounded annotation in GitHub's structured run metadata. Infra PR #151 therefore
moved the already-bounded values through step/job outputs to a dedicated clean
publisher. That publisher validates the live base/head pair, checks out its own
exact workflow SHA, mints a narrowly scoped App token only after validation,
rechecks identity, and posts a non-verdict failure comment without model output.
Its independently reviewed head `364a899…` merged as exact authoritative source
`21a24db…`; the caller mirror and pin now use that merge.

The exact fixture advance also carries authoritative infrastructure changes that
landed after the caller's prior `c5d8bcc…` pin but before VOC-117 source work. In
particular, `implement.yml`'s repository-explicit REST update for an existing PR
was independently merged by infrastructure PR #119 as `2188c72…` and is already
present in the pre-VOC-117 source base `d4a8f5c…`. Its appearance in the caller
base-to-head diff is stale-fixture reconciliation required by `VOC-117-D05`, not a
new VOC-117 publication-behavior decision.

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `python3 -m unittest tests.test_voc117_role_bindings -v` in `karsift-ai-infra` | pass | 7 VOC-117 regressions on source head `d6ac23f…` |
| `python3 -m unittest discover -s tests -p 'test_*.py'` in `karsift-ai-infra` | pass | 280 tests on authoritative merge `21a24db…` |
| `bash scripts/governance/validate-governance.sh` | pass | Repository foundation + monitoring declarations |
| `bash scripts/governance/classify-change-risk.sh` | pass | Detected path floor `R4` |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass | 170 tests after exact pin and isolated failure-publisher reconciliation |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc117*.py'` | pass | 10 deterministic tests (also included in the full 170-test run) |
| `node --test scripts/foundation/voc097-fixture-matrix.test.mjs scripts/foundation/voc104-ready-for-review-reuse.test.mjs scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | pass | 16 tests after pin advance to `21a24db…` |
| Independent Claude Code exact-revision review of infra PR #149 | pass | `VERDICT: PASS` on `e7f3804…`; no findings after error-response, I/O-failure, and combined-flag coverage |
| Independent Claude Code exact-revision review of infra PR #150 | pass | `VERDICT: PASS` on `6526cf6…`; no findings on sanitized stdout/stderr channel boundary |
| Independent Claude Code exact-revision review of infra PR #151 | pass | `VERDICT: PASS` on `364a899…`; no blocking findings on failed-step outputs, exact-SHA identity, App-token isolation, bounded content, or merge-gate separation |
| `bash karsift-ai-infra/config/run-app-checks.sh` on caller head `e75c5ce…` | pass | Governed attempt-2 pre-push deterministic CI |
| `bash karsift-ai-infra/config/run-app-checks.sh` after pinning `42aa667…` | environment-limited locally | Format, lint, typecheck, 338 foundation tests, API-client tests, middleware tests, web build, and API build passed; two controlled-signup API tests could not start because Docker is unavailable in this WSL environment. Hosted CI remains authoritative for those Docker-backed tests. |
| `git diff --check` | pass | No whitespace or patch-format errors |

## Acceptance mapping

- `VOC-117-AC-00` / `VOC-117-EV-00` — six exact bindings in `roles.yml` and fixture.
- `VOC-117-AC-01` / `VOC-117-EV-00` — `prepare_cursor_model.py` and workflow wiring.
- `VOC-117-AC-02` / `VOC-117-EV-00` — Grok 4.6 Standard + plan_reviewer `effort=high`.
- `VOC-117-AC-03` / `VOC-117-EV-00` — missing `CURSOR_API_KEY` and unsupported prefix tests.
- `VOC-117-AC-04` / `VOC-117-EV-00` — VOC-117 current-state header; dormant routes not active.
- `VOC-117-AC-05` / `VOC-117-EV-00` — suites + exact pin at infra merge handoff.
