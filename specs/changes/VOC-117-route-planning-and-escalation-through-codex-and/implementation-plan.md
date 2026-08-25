# VOC-117 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: AI model routing (`config/roles.yml`), reusable workflow
  invocation paths (`plan.yml`, `plan-review.yml`, `review.yml`, `implement.yml`),
  shared-infra tests, and caller `tooling/governance/` fixtures/tests.
- Prerequisites: confirm current resolve-model behavior, how each workflow strips
  or routes vendor prefixes today, and which comments still claim dormant
  OpenAI/Codex or obsolete Cursor bindings are active.
- Preserve one-retry limits, exact-SHA independent review, risk classification, and
  fail-closed credential checks.

## File reconciliation and implementation sequence

### T00 — Apply the Cursor role lineup, parameterized-model routing, tests, docs, and caller pin

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra/config/roles.yml` | modify | Persist the six exact bindings from `VOC-117-D00`; update current-state comments |
| `KARSIFT/karsift-ai-infra/config/resolve-model.sh` and/or shared model-parse helper | modify if needed | Only if parameterized parsing must be centralized rather than inlined in workflows |
| `KARSIFT/karsift-ai-infra/.github/workflows/plan.yml` | modify | Compatible invocation of `cursor/grok-4.6[effort=high,fast=false]` for planner |
| `KARSIFT/karsift-ai-infra/.github/workflows/plan-review.yml` | modify | Compatible invocation of `cursor/grok-4.6[effort=high,fast=false]` |
| `KARSIFT/karsift-ai-infra/.github/workflows/review.yml` | modify | Compatible invocation for reviewer / reviewer_fast_retry |
| `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` | modify | Keep composer-2.5 for implementer and implementer_escalation; preserve Cursor fail-closed auth; do not re-enable OpenAI/Codex as the active path |
| `KARSIFT/karsift-ai-infra/tests/*` | modify/extend | Exact six-binding assertions; parameterized routing; negative fail-closed cases |
| Infra workflow/roles current-state comments | modify | Historical OpenAI/Codex narrative may remain as history, but must not claim current active routing |
| Caller docs that assert an active lineup | modify if needed | Only where they would become false |
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact reviewed infra merge |
| `tooling/governance/tests/*` | modify/extend | Caller regressions for bindings and fail-closed behavior |
| `specs/changes/VOC-117-.../t00-evidence.md` | create/update | Record commands, results, infra SHA, and pin |

Ordered steps:

1. Land the authoritative `roles.yml` bindings and current-state comment updates in
   the primary infra repository.
2. Make plan/implement/review/plan-review workflows compatible with parameterized
   Cursor model strings; verify against the Cursor CLI that explicit-high
   Standard/non-fast is preserved without silent fallback and reject the
   effort-omitted unavailable form.
3. Add/extend deterministic tests for the six exact mappings, parameterized
   routing, and fail-closed credential/prefix cases.
4. Run infra self-CI / unit suite; merge the reviewed infra PR.
5. Sync and pin the caller mirrored fixture to that exact infra merge; update caller
   governance tests.
6. Run caller governance validation and fixture suites; record evidence in
   `t00-evidence.md`.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
# In the checked-out primary KARSIFT/karsift-ai-infra source:
python3 -m unittest discover -s tests -p 'test_*.py'

# In this caller repository:
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

If implementation adds a narrower targeted routing/role regression command, record
the exact command in `t00-evidence.md` and run it in addition to the suite above.

Independent verifier (exact reviewed task PR SHA) should confirm:

- the six stored bindings match `VOC-117-D00` exactly;
- OpenAI/Codex is not required or re-enabled as the active planner/escalation path;
- parameterized model strings are handled without silent vendor/model fallback;
- missing credentials / unsupported prefixes fail closed;
- current-state comments no longer describe dormant routes as active;
- exact-SHA review, risk floors, protected checks, and one-retry limits remain;
- the caller fixture pin equals the exact reviewed shared-infra merge.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** Future plan/implement/review/plan-review runs use the new
  Cursor bindings.
- **Rollback trigger:** Parameterized routing fails closed incorrectly, silent
  model substitution appears, review independence from implementer regresses, or
  docs/comments drift back to false current-state claims.
- **Rollback mechanism:** Revert the infra and caller fixture/test changes to the
  prior reviewed role bindings and routing behavior.
- **Last-known-good reference:** Current `roles.yml` / workflow routing on
  `main`/`develop` before VOC-117 implementation lands.
