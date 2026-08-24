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
| Coordinated infra source | `KARSIFT/karsift-ai-infra` (same T00 carrier) |
| Caller fixture directory | `tooling/governance/fixtures/karsift-ai-infra/` |
| Prior pin | `c5d8bccfa8676bd367b53ad5f6f9a51a40c99405` |
| New pin | Set to the exact reviewed karsift-ai-infra merge SHA at workflow handoff |

Fixture file content is synchronized to the infra changes in this working tree.
`PINNED_SHA.txt` is updated to the exact reviewed infra merge commit when the
coordinated infra PR merges (handoff step; not pre-merge guess).

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` in `karsift-ai-infra` | pass | 274 tests (includes 8 VOC-117 regressions) |
| `bash scripts/governance/validate-governance.sh` | pass | Repository foundation + monitoring declarations |
| `bash scripts/governance/classify-change-risk.sh` | pass | Detected path floor `R4` |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass | 167 tests (includes 7 VOC-117 fixture regressions) |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc117*.py'` | pass | 7 tests |
| `git diff --check` | pass | No whitespace or patch-format errors |

## Acceptance mapping

- `VOC-117-AC-00` / `VOC-117-EV-00` — six exact bindings in `roles.yml` and fixture.
- `VOC-117-AC-01` / `VOC-117-EV-00` — `prepare_cursor_model.py` and workflow wiring.
- `VOC-117-AC-02` / `VOC-117-EV-00` — Grok 4.6 Standard + plan_reviewer `effort=high`.
- `VOC-117-AC-03` / `VOC-117-EV-00` — missing `CURSOR_API_KEY` and unsupported prefix tests.
- `VOC-117-AC-04` / `VOC-117-EV-00` — VOC-117 current-state header; dormant routes not active.
- `VOC-117-AC-05` / `VOC-117-EV-00` — suites + exact pin at infra merge handoff.
