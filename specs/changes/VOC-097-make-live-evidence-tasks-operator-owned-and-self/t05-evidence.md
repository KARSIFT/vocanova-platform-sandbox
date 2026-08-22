---
evidence_id: VOC-097-EV-05
task_id: VOC-097-T05
acceptance_criteria:
  - VOC-097-AC-06
  - VOC-097-AC-10
tests:
  - VOC-097-TEST-11
  - VOC-097-TEST-16
date: 2026-08-22
related_change: VOC-097
gate_status: live-complete-awaiting-independent-review
live_fixture_claimed: true
live_reconcile_claimed: true
observer_health_claimed: true
waiting_proof_run: 32456592473
reconcile_proof_run: 32493488150
reviewed_sha: bind-at-independent-review
rollback_owner: revert karsift-ai-infra live-evidence-reconcile and caller pipeline wiring; restore prior remediate behavior; leave audit evidence files
last_known_good_sha: 4fc78ff66ba8e0b681302191921f46107a706d01
---

# VOC-097-T05 — Controlled live proof and observer/Sentry separation health

This task closes the package with hosted sandbox proof that (1) operator-owned
live-evidence **waiting** does not invoke remediation retries, (2) a qualifying
reconcile wake records allowlisted metadata and leaves merge gated on a **new**
exact-SHA independent review, and (3) deploy/synthetic failure observation and
Sentry monitoring remain separate from live-evidence lifecycle signals.

No logs, artifacts, secrets, OAuth/session/cookie/token material, environment
values, user identifiers, or personal data appear in this evidence.

## 1. Waiting does not invoke remediation (VOC-097-AC-02 live; TEST-03)

Controlled fixture: stranded-task migration PR **#789** (`VOC-093-T01`) on the
governed waiting path after VOC-097-T04.

| Item | Value |
| --- | --- |
| Task PR | [#789](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/789) |
| Pipeline run | [32456592473](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32456592473) |
| Reviewed head SHA | `c83625909125418e12e9695d0df158e5b0f62dc1` |
| Independent review comment | `5366335489` |
| Verdict | `VERDICT: WAITING FOR OPERATOR LIVE EVIDENCE` |
| CI | `ci / ci` success |
| `remediate / decide` | success |
| `remediate / retry` | **skipped** (no attempt 2; `should_retry=false`) |
| `implement` | skipped |
| `merge-gate / auto-merge` | skipped (fail-closed until fresh exact-SHA review after evidence) |

The review narrative explicitly classifies the head as waiting — not a code-defect
`FAIL` that must be remediated by editing workflows. Remediation consumed no
implementation attempt and did not re-dispatch the implementer.

Deterministic matrix coverage for the same policy remains in
`tooling/governance/tests/test_voc097_live_evidence_lifecycle.py` (T03); this
section claims the hosted positive control (`live_fixture_claimed: true`).

## 2. Qualifying reconcile wake and fresh exact-SHA review gate (VOC-097-AC-06 live; TEST-11)

Controlled fixture: operator-owned carrier on the **VOC-102-T01** path, which
consumes the same `live-evidence-reconcile` reusable workflow delivered in
VOC-097-T02.

| Step | Run / artifact | Result |
| --- | --- | --- |
| Reconcile dispatch | [32493488150](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32493488150) on `agent/voc-102-voc-102-t01` | `live-evidence-reconcile / reconcile` **success** |
| Allowlisted qualification record | `specs/changes/VOC-102-auto-advance-dispatches-implementer-for-operator/.karsift/live-evidence/VOC-102-T01.result.json` | `state: qualified`, `run_id: 32493135121`, sanitized metadata only |
| Post-carrier verifier | [32493135121](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32493135121) | `verify-auto-advance-live-evidence / verify` **success** |
| Duplicate reconcile on same carrier | [32542667963](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32542667963) (VOC-106 path) | reconcile success; prior `VOC-106-T01.result.json` shows idempotent `qualified` state without double-wake |

Wake semantics (one commit/comment/ref effect per qualifying run identity) and
the requirement that merge-gate bind to the **post-reconcile** head — not an
earlier PASS — are also locked deterministically in
`tooling/governance/tests/test_voc097_live_evidence_reconcile.py` TEST-11/14 and
were exercised on the VOC-097-T02 implementation draft at run
[32443838227](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32443838227):
independent review and remediation decision succeeded, `remediate / retry` and
`merge-gate / auto-merge` were skipped while the PR remained draft pending a new
exact-SHA verdict.

## 3. Observer / Sentry separation remains healthy (VOC-097-AC-10; TEST-16)

Live-evidence lifecycle automation does not open operational-failure issues or
couple into Sentry for expected waiting/reconcile outcomes.

| Check | Evidence |
| --- | --- |
| Code separation | `test_voc097_live_evidence_reconcile.py::test_live_evidence_stays_separate_from_operational_failure_observer` — reconcile workflow/runner never references `operational-failure-monitoring.yml` or `open-failure-issue.sh`; observer workflow contains no `karsift-live-evidence` marker |
| Observer posture during proof window | `operational-failure-monitoring` scheduled runs [32561214088](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32561214088), [32560039594](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32560039594) **skipped** — no false failure-to-issue emission while reconcile and waiting runs completed successfully |
| Hourly reconcile poll | [32559573897](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32559573897) on `main` — `live-evidence-reconcile / reconcile` success; observer not triggered (success conclusions only) |
| Sentry monitor (separate path) | [error-monitoring 32555555744](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32555555744) **success** on `main` |
| Deploy / synthetic gates (observed workflows) | [scheduled-synthetics 32555436255](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32555436255) **success**; [deploy-staging 32554992337](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32554992337) **success** on `develop` |
| Waiting comments | PR #789 waiting review (`5366335489`) contains only allowlisted contract/SHA metadata — no operational-failure fingerprint markers |

## 4. Stranded #779 / #785 status (T04 follow-through)

| Task | Issue | PR | Live state at T05 |
| --- | --- | --- | --- |
| `VOC-093-T01` | [#779](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/779) open | [#789](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/789) closed | Governed **waiting** proven live (section 1); reconcile qualification still operator-owned per contract |
| `VOC-094-T01` | [#785](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/785) open | [#791](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/791) closed | On governed waiting path per `t04-evidence.md`; green `deploy-staging` observe pending |

Package mechanism closure does not require merging #789/#791; T04 recorded the
safe migration and T05 proves the lifecycle behaviors those tasks depend on.
Issue closure after qualifying reconcile + fresh exact-SHA review remains on
the operator path documented in `t04-evidence.md` and
`docs/operations/live-evidence.md`.

## Acceptance mapping

| Criterion / test | Result |
| --- | --- |
| VOC-097-AC-06 / TEST-11 | **Met (live)** — reconcile run `32493488150` qualified carrier metadata; verifier run `32493135121` success; merge remains gated on fresh exact-SHA review after wake |
| VOC-097-AC-10 / TEST-16 | **Met (live)** — observer skipped on success-only window; Sentry monitor green; no coupling markers in waiting/reconcile paths |
| VOC-097-AC-02 (live supplement) | **Met (live)** — waiting run `32456592473` skipped remediation retry |

## Deterministic validation

```bash
node --test scripts/foundation/voc097-live-verification.test.mjs
node --test scripts/foundation/voc097-*.test.mjs
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc097_*.py' -v
bash scripts/governance/validate-governance.sh
git diff --check
```

Local results on the implementation working tree:

- `voc097-live-verification.test.mjs`: evidence structure and claim gates passed.
- All `voc097-*.test.mjs` foundation tests: 24 passed.
- VOC-097 Python matrix: 18 passed.
- `validate-governance.sh`: passed.
- `git diff --check`: passed.

Independent verification must bind to this evidence commit's exact SHA. Hosted
PR checks remain the canonical gate for the task PR head.

## Rollback

1. Revert VOC-097 infra and caller wiring commits through the normal PR path.
2. Restore prior `remediate.yml` behavior (waiting no longer suppresses retry).
3. Confirm operational-failure observer and Sentry workflows remain green on
   `main` / `develop`.
4. Leave `t00`–`t05` evidence files as audit history.
