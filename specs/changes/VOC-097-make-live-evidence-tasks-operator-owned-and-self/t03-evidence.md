---
evidence_id: VOC-097-EV-03
task_id: VOC-097-T03
acceptance_criteria:
  - VOC-097-AC-09
tests:
  - VOC-097-TEST-02
  - VOC-097-TEST-03
  - VOC-097-TEST-04
  - VOC-097-TEST-05
  - VOC-097-TEST-06
  - VOC-097-TEST-07
  - VOC-097-TEST-08
  - VOC-097-TEST-09
  - VOC-097-TEST-10
  - VOC-097-TEST-11
  - VOC-097-TEST-12
  - VOC-097-TEST-13
  - VOC-097-TEST-14
date: 2026-08-21
related_change: VOC-097
gate_status: complete
live_fixture_claimed: false
post_merge_source_run_claimed: false
---

# VOC-097-T03 — Deterministic fixture matrix

## Outcome

The repository now carries a complete deterministic matrix for
`VOC-097-TEST-02` through `VOC-097-TEST-14`. Waiting lifecycle,
remediation suppression, least-privilege separation, reconcile
qualification, sanitization, timeout, and deduplication are exercised
against vendored `karsift-ai-infra` policy fixtures plus the live caller
`pipeline.yml`. This task does not claim controlled live proof
(`live_fixture_claimed: false`; T05 owns that).

## Deliverables

| Artifact | Path |
| --- | --- |
| Vendored infra pin + policy fixtures | `tooling/governance/fixtures/karsift-ai-infra/` (`PINNED_SHA.txt` → `0ca6de2b5965237b310a64df1f2588a384e16d2c`) |
| Waiting/remediation matrix | `tooling/governance/tests/test_voc097_live_evidence_lifecycle.py` |
| Reconcile/sanitizer matrix | `tooling/governance/tests/test_voc097_live_evidence_reconcile.py` |
| Shared loader | `tooling/governance/tests/voc097_fixtures.py` |
| Caller matrix orchestrator | `scripts/foundation/voc097-fixture-matrix.test.mjs` |
| Prior caller/doc locks (unchanged scope) | `scripts/foundation/voc097-{operator-docs,waiting-lifecycle,reconcile}.test.mjs` |

## Matrix mapping

| Test | Coverage |
| --- | --- |
| TEST-02 | Machine-readable `WAITING` marker, fail-dominant classifier, trusted review prompt contract |
| TEST-03 | `decide-remediation.py` returns `WAITING`; remediate emits `should_retry=false` before retry path |
| TEST-04 | Genuine `FAIL` and CI failure still enter bounded `RETRY` |
| TEST-05 | Implementer workflow has no `actions` permission; operator reconcile keeps read-only caller floor |
| TEST-06 | Wrong workflow identity fail-closed |
| TEST-07 | Missing/failed required job fail-closed |
| TEST-08 | Wrong event, branch, or SHA lineage fail-closed |
| TEST-09 | Allowlisted metadata only in evidence JSON |
| TEST-10 | Reconcile workflow/runner never references log or artifact APIs |
| TEST-11 | One result commit then fresh exact-SHA PR review on caller pipeline |
| TEST-12 | Stale age and non-success conclusions rejected |
| TEST-13 | 72-hour timeout marker is single-use |
| TEST-14 | Duplicate qualified result short-circuits reconciliation |
| (supporting) | Caller hourly reconcile without `workflow_run` recursion; live-evidence paths stay separate from operational-failure observer |

Shared infra self-CI on `KARSIFT/karsift-ai-infra` remains the authoritative
runtime source. These vendored copies replay the merged policy at the pinned
SHA for caller-repo CI without cloning infra on every run.

## Validation

```bash
node --test scripts/foundation/voc097-fixture-matrix.test.mjs
node --test scripts/foundation/voc097-*.test.mjs
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc097_*.py' -v
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py' -v
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Local results on the reviewed working tree:

- `voc097-fixture-matrix.test.mjs`: 5 tests passed (matrix registry, pin, Python suite, method map, evidence structure).
- All `voc097-*.test.mjs` foundation tests: 14 passed.
- VOC-097 Python matrix: Ran 18 tests, all passed.
- Full governance Python suite: Ran 113 tests, all passed (includes preserved VOC-080 regressions after fixture refresh).
- `validate-governance.sh`: passed.
- `classify-change-risk.sh`: reported R3 for this task diff.
- `git diff --check`: passed.

No secrets, logs, OAuth material, personal identifiers, tokens, or credentials
are recorded. T04 owns stranded #779/#785 migration; T05 owns controlled live
proof.
