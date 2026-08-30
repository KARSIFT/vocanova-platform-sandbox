# VOC-140-T00 — Evidence

Task: `VOC-140-T00` — Unblock release-convergence CI identity and
production-merge-guard App-token visibility.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

## Implementation PR base

Recorded before the first in-scope edit:

`c59548375764d938265910cd07f2c2a73e337c01` (`develop` at dispatch).

Issue-creation develop was `21eef75549226766fc4f78f62f232ee5fbdb8d6d`.
Plan/adoption/roster commits after that SHA are governance-only and do not count
as protected-file drift.

## Infrastructure merge

Independently reviewed `KARSIFT/karsift-ai-infra` merge consumed by this pin:

`0ee1daf1aecdb5039ecc0fc74f5c64b24cdd5f5d`

Issue-creation / VOC-139 pin (defective for this failure class; historical audit):
`599436835371f27fac52ec6b47a18b36257366ac`.

## Issue #1102 incident record

| Item | Value |
|------|-------|
| Promotion PR | #1090 |
| Release issue | #1089 |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head at issue creation | `21eef75549226766fc4f78f62f232ee5fbdb8d6d` |
| Circular-CI run / job | `33136633666` / `98738317266` |
| Circular-CI fail-closed | `untrusted_ci_recovery_identity` |
| Dedicated recovery success | run `33136865709` |
| Guard-fail release runs / jobs | `33136984634` / `98739074178`; `33137091931` / `98739420310` |
| Guard fail-closed | `production-merge-guard: production_merge_guard_missing` |
| Admin-visible ruleset | `20575146` with `bypass_actors: []` under administrator token |

## Recovery identity and token/API contract

| Case | Result |
|------|--------|
| Newest `ci / ci` SUCCESS on in-progress `pipeline.yml` release carrier | not attestable; recovery not complete without dedicated dispatch |
| PR-required `ci / ci` SUCCESS with attestable selection filtered from gate summary | recovery not complete (#1102 composition fix in `promotion_ci_context_is_attestable`) |
| Completed `promotion-pr-validation PR #<n>` exact-head/PR-bound | attestable; may suppress redispatch |
| Omitted/non-array `bypass_actors` in ruleset payload | `production_merge_guard_payload_incomplete` with operator action |
| `bypass_actors: []` under Administration token | `production-merge-guard: ok` |
| Mutation App token | exactly Contents/Issues/Pull requests write; sole token for merge/mutations |
| Guard App token | exactly Administration write; explicit owner/current-repository scope; only `verify-production-merge-guard.sh` |

## Exhaustive source-search disposition

| Pattern / topic | Disposition |
|-----------------|-------------|
| Pin `59943683…` in fixture README / historical package evidence | preserved as historical |
| Live `PINNED_SHA.txt` and pin-lock tests | updated to `0ee1daf…` |
| `mutation-only` App-token claims in DOC-11 | updated to mutation vs guard-only split |
| Active-A-003 / disabled production release in repository-settings, DOC-19, activation checklist | updated to active A-004 and enabled release/deploy path; RL1/RL2 remain disabled; `validate-governance.sh` reconciled |
| VOC-138 / VOC-139 package records under `specs/changes/` | unchanged (audit evidence) |

## Credential residual risk

Step-level token isolation does not split the underlying `karsift-ai-infra-bot`
registration/private key. Compromise can still mint up to installation
`148001476`'s combined permission ceiling across its current
`repository_selection: all` scope. A separately keyed single-repository
Administration-only guard App is optional future hardening and is not implemented
by VOC-140-T00.

## External activation (post-T00 release gate)

Before hosted promotion of #1090:

1. Configure `karsift-ai-infra-bot` Repository permissions Administration: Read and write.
2. Obtain installation-owner approval on KARSIFT organization installation `148001476`.
3. Retain explicit caller-repository guard-token scope at runtime.
4. Do not rotate App ID/private-key secrets.
5. Rerun failed guard / `reconcile-release` after hosted explicit `bypass_actors: []` proof.

Pending at implementation review is valid; this does not consume implementer retries.

## Validation commands (implementation working tree)

Remediation attempt 2 — results from the working tree after blocking-finding fixes:

```bash
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
# OK — 281 tests

python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
# OK — 490 tests

bash scripts/governance/validate-governance.sh --base c59548375764d938265910cd07f2c2a73e337c01 --head HEAD
# OK

bash scripts/governance/classify-change-risk.sh --base c59548375764d938265910cd07f2c2a73e337c01 --head HEAD
# OK — detected path-based risk floor R4

git diff --check
# OK
```

Targeted VOC-140 cases are covered by `test_voc140_release_convergence.py`
(recorded SHA-256 pin-lock literals), `test_voc140_release_carrier_attestation.py`
(filtered gate-summary composition, failed/queued/cancelled carriers), and
`test_voc140_production_merge_guard.py` (omitted/empty/non-empty bypass subprocess,
exhaustive Administration-mint allowlist across all fixture workflows).

Governance validation with exact base/head and independent exact-SHA review bind
the live implementation PR head after commit; this file does not require the
same commit to contain that head SHA.

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR must
bind the live head exactly. Merge-gate must reject any mismatch.
