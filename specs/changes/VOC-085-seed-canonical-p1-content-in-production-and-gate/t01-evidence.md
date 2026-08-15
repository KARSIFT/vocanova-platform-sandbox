# VOC-085-EV-01 — T01 content-aware production smoke evidence

Evidence for `VOC-085-T01` (`VOC-085-TEST-04`, `VOC-085-TEST-05`).

## Implemented smoke changes

File changed: `infra/scripts/smoke-test-production.sh`

- Authenticated `GET /api/v1/journey-situations` now requires HTTP 200 **and**
  a non-empty parsed `items` list (`VOC-085-D03`, `VOC-085-AC-03`).
- From the first returned situation slug, the suite performs read-only detail
  checks:
  - `GET /api/v1/journey-situations/{slug}` must return matching
    `situation.slug` and at least one `meanings[]` entry.
  - `GET /api/v1/canonical-words/{wordSlug}` must return matching `word.slug`
    from the first meaning's `wordSlug` (`VOC-085-D04`, `VOC-085-AC-04`).
- Reuses `SMOKE_TEST_SESSION_COOKIE` / `http_get_authed`; no magic links,
  OAuth completion, or mutating learning actions.
- Added `json_jq` helper with jq primary path and a python3 fallback for the
  small set of filters used by the content checks.

## Deterministic selftests

File changed: `infra/scripts/smoke-test-production.selftest.sh`

- Extended the local fake server with canonical list/detail/word payloads.
- `VOC-085-TEST-04`: `empty_situations` scenario returns HTTP 200 with
  `{"items":[]}` and asserts the suite fails hard (not SKIP).
- `VOC-085-TEST-05`: healthy scenario with session cookie exercises list
  parsing, situation detail, and canonical word detail success paths; additional
  fail-closed fixtures cover missing meanings and missing word detail.

## Acceptance mapping

| AC / decision | How this task satisfies it |
|---------------|----------------------------|
| `VOC-085-AC-03` (deterministic) | Empty list on HTTP 200 is a hard fail in smoke + selftest case 6 |
| `VOC-085-AC-04` | Non-mutating situation + word detail GETs from API-derived slugs |
| `VOC-085-AC-07` (smoke tests) | Selftest cases 5–8 cover positive parsing and failure behavior |

Live production non-empty content closure remains coupled to `VOC-085-T00`
deploy seed convergence and `VOC-085-T02` Cloudflare proof (`VOC-085-EV-02`).

## Local deterministic checks run for this task

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
bash infra/scripts/smoke-test-production.selftest.sh
```

Record results from the implementation run in the task PR (pass/fail per command).

Implementation run (2026-08-15):

| Command | Result |
|---------|--------|
| `bash scripts/governance/validate-governance.sh` | pass |
| `bash scripts/governance/classify-change-risk.sh --files-from …` | pass (floor R3 via smoke scripts) |
| `git diff --check` | pass |
| `bash infra/scripts/smoke-test-production.selftest.sh` | pass (8 cases) |

## Explicit deferrals

- Authenticated ten-route page sweep (`VOC-085-T02`).
- Live Cloudflare verification and topology/isolation proof (`VOC-085-T02`).
