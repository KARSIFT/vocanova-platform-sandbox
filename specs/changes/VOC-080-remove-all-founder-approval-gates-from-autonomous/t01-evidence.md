# VOC-080-EV-01 — T01 merge-gate founder-gate removal

Evidence for `VOC-080-AC-01`, `VOC-080-AC-02`, `VOC-080-AC-03`, and
`VOC-080-AC-07` (merge-gate slice). Tests: `VOC-080-TEST-01`–`VOC-080-TEST-03`
(policy regressions; live rehearsal is `VOC-080-T06`).

## Task outcome

`VOC-080-T01` removes founder-comment merge authority from
`KARSIFT/karsift-ai-infra` `merge-gate.yml` and reconciles merge-gate-adjacent
prompts/comments. R0–R4 PRs may auto-merge when the calling project passes
`auto_merge_enabled=true` and CI plus independent verification (review or
plan-review) pass. Unparseable risk fails closed for correction. No path uses a
founder `approved` comment to override failed or missing gates.

## Infra delivery

| Item | Location | Notes |
|------|----------|-------|
| Core merge-gate behavior | `karsift-ai-infra` PR [#37](https://github.com/KARSIFT/karsift-ai-infra/pull/37) (`da91f64`) | Removed `approve-and-merge`, R4 hard block, and `automatic_merge_allowed` gate |
| T01 polish + policy tests | This task's infra working-tree delta | Header/comments, `plan-review.md`, expanded `tests/test_merge_gate_policy.py` |

Caller repos pin `@main`; merge-gate behavior is effective once the infra delta
lands on `karsift-ai-infra` `main`.

## Behavioral checklist (VOC-080-D00 / D02)

| Requirement | Implementation |
|-------------|----------------|
| R0–R4 auto-merge when gates pass | `auto-merge` job: `risk != 'unknown'`, `checks_ok`, verdict not FAIL/PENDING, `auto_merge_enabled=true` |
| R4 not a founder block | No `risk = R4` branch in status or auto-merge `if:` |
| `automatic_merge_allowed: false` neutralized | Field no longer read in merge-gate (DEP-02) |
| Unparseable risk fail-closed | `risk=unknown` → BLOCKED status; auto-merge `if:` excludes unknown |
| No founder override of failed gates | `issue_comment` trigger and `approve-and-merge` job removed |
| Builder/verifier + CI remain hard gates | `checks_ok` and verdict parsing unchanged; App token required for merge |

## Deterministic verification

```bash
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py' -v
```

Expected: all policy tests pass, including `test_merge_gate_policy.py` and
`test_adoption_handoff.py::test_founder_comment_is_not_a_merge_path`.

## Explicitly not done (other tasks)

- Autonomous adoption field flips (`VOC-080-T02`)
- Release/deploy founder-gate removal (`VOC-080-T03`)
- Caller `pipeline.yml` / AGENTS.md / DOC-15 reconciliation (`VOC-080-T04`)
- Live sandbox rehearsal (`VOC-080-T06`)
- Authority activation (`VOC-080-T07`)

## Limitations

- Live R4 auto-merge rehearsal requires `auto_merge_enabled=true` on the caller
  and a passing plan/task PR (`VOC-080-T06`).
- `founder_username` input remains as deprecated compatibility until caller
  wiring is removed in `VOC-080-T04`.
