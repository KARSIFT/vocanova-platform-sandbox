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
  printing raw provider responses or credentials.
- `tests/test_voc117_role_bindings.py` — VOC-117-TEST-00 through TEST-05 regressions.

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
| Authoritative infra PRs | `KARSIFT/karsift-ai-infra#147`, diagnostic hardening `#148` |
| Independently reviewed infra heads | `d6ac23f70a10299e73629f257e275854525d15c8`, `9457233e5b2e2fb03674ef963d89ac4767596a4f` |
| Infra merge commits | initial `27a44b298f1c234a94e02127eaeb55d66b28e30d`; authoritative current `42aa66757a521b1187193fba17b74e440964c27f` |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Prior pin | `c5d8bccfa8676bd367b53ad5f6f9a51a40c99405` |
| New pin | `42aa66757a521b1187193fba17b74e440964c27f` |

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
revision mirrors that authoritative current source and advances the pin to the exact
follow-up merge.

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `python3 -m unittest tests.test_voc117_role_bindings -v` in `karsift-ai-infra` | pass | 7 VOC-117 regressions on source head `d6ac23f…` |
| `python3 -m unittest discover -s tests -p 'test_*.py'` in `karsift-ai-infra` | pass | 275 tests on follow-up source head `9457233…` |
| `bash scripts/governance/validate-governance.sh` | pass | Repository foundation + monitoring declarations |
| `bash scripts/governance/classify-change-risk.sh` | pass | Detected path floor `R4` |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass | 168 tests after exact-pin and diagnostic reconciliation |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc117*.py'` | pass | 8 deterministic tests (also included in the full 168-test run) |
| `node --test scripts/foundation/voc097-fixture-matrix.test.mjs scripts/foundation/voc104-ready-for-review-reuse.test.mjs scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | pass | 16 tests after pin advance to `42aa667…` |
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
