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
  Standard form via `prepare_cursor_model.py`.
- `.github/workflows/plan-review.yml` — plan_reviewer retains `effort=high` and
  `fast=false` through `prepare_cursor_model.py`.
- `.github/workflows/implement.yml` — implementer, escalation, and self-correct
  steps use `prepare_cursor_model.py` for `cursor/composer-2.5`.
- `tests/test_voc117_role_bindings.py` — VOC-117-TEST-00 through TEST-05 regressions.

### Caller mirrored fixture and tests

- `tooling/governance/fixtures/karsift-ai-infra/**` — synchronized copies of the
  infra changes above.
- `tooling/governance/fixtures/karsift-ai-infra/README.md` — VOC-117 contract note.
- `tooling/governance/tests/test_voc117_role_bindings.py` — caller fixture regressions.

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
| Authoritative infra PR | `KARSIFT/karsift-ai-infra#147` |
| Reviewed infra head | `d6ac23f70a10299e73629f257e275854525d15c8` |
| Infra merge commit | `27a44b298f1c234a94e02127eaeb55d66b28e30d` |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Prior pin | `c5d8bccfa8676bd367b53ad5f6f9a51a40c99405` |
| New pin | `27a44b298f1c234a94e02127eaeb55d66b28e30d` |

The first governed Cursor attempt produced the caller-side source projection but
correctly left the prior pin in place while no authoritative source merge existed.
Its exact-SHA review failed closed on that missing source/pin. PR #147 then landed
the authoritative source under the same T00 authority. This caller revision removes
the two unrelated VOC-112 fixture edits identified by that review, synchronizes the
seven mirrored source files byte-for-byte to reviewed infra head `d6ac23f…`, and pins
the exact GitHub merge commit `27a44b2…`; it does not infer or predeclare a SHA.

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `python3 -m unittest tests.test_voc117_role_bindings -v` in `karsift-ai-infra` | pass | 7 VOC-117 regressions on source head `d6ac23f…` |
| `python3 -m unittest discover -s tests -p 'test_*.py'` in `karsift-ai-infra` | pass | 273 tests on source head `d6ac23f…` |
| `bash scripts/governance/validate-governance.sh` | pass | Repository foundation + monitoring declarations |
| `bash scripts/governance/classify-change-risk.sh` | pass | Detected path floor `R4` |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass | 167 tests after exact-pin reconciliation |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc117*.py'` | pass | 7 tests |
| `git diff --check` | pass | No whitespace or patch-format errors |

## Acceptance mapping

- `VOC-117-AC-00` / `VOC-117-EV-00` — six exact bindings in `roles.yml` and fixture.
- `VOC-117-AC-01` / `VOC-117-EV-00` — `prepare_cursor_model.py` and workflow wiring.
- `VOC-117-AC-02` / `VOC-117-EV-00` — Grok 4.6 Standard + plan_reviewer `effort=high`.
- `VOC-117-AC-03` / `VOC-117-EV-00` — missing `CURSOR_API_KEY` and unsupported prefix tests.
- `VOC-117-AC-04` / `VOC-117-EV-00` — VOC-117 current-state header; dormant routes not active.
- `VOC-117-AC-05` / `VOC-117-EV-00` — suites + exact pin at infra merge handoff.
