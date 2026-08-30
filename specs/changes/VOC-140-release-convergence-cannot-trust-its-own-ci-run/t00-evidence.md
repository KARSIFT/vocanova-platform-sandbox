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

`9fdff24cd387cc2cdc468c84a3012b0c34b6c8e8`

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

The tracked-source scan used `git grep -n -I -E` over Markdown, YAML, workflow,
Python, and shell sources. Pattern families were:

| Pattern family | Examples searched |
|----------------|-------------------|
| Pin and hash assertions | old/current full pin, `CURRENT_PIN`, `AUTHORITATIVE_PIN`, `PINNED_SHA`, SHA-256 tables |
| Token permissions and use | `mutation-only`, Contents/Issues/Pull requests grants, `Administration`, guard token scope/use |
| Authority state | `active-A-003`, `active A-003`, `A003`, active/effective A-004 |
| Human gates | R3/R4 with founder approval, founder `approved`, technical-steward, EHR |
| Automation state | automatic/autonomous merge or release and production deployment described as disabled, blocked, inactive, or unimplemented |

Every resulting path was classified as follows:

| Paths | Disposition |
|-------|-------------|
| `.github/CODEOWNERS`; `.github/README.md`; `CLAUDE.md`; `CONTRIBUTING.md` | Live repository instructions/routing corrected to active A-004, no standing founder-comment/steward workflow gate, and enabled gated merge/promotion/deploy state. |
| `docs/README.md`; `docs/engineering/04-technical-architecture.md`; `docs/decisions/README.md` | Current indexes/architecture/decision guidance corrected to reference A-004 and current activation sources. |
| `docs/governance/README.md`; `docs/governance/change-risk-classification.md`; `docs/governance/16-autonomous-development-operating-model.md`; `docs/governance/protected-areas.md` | Current governance text corrected; bootstrap and pre-A-004 statements are now explicitly historical. DOC-16's obsolete R0-R2 release ceiling is replaced with gated R0-R4 eligibility. |
| `docs/operations/10-development-workflow.md`; `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | Current branch/deploy and waiver claims corrected to enabled gated release/deploy and no founder-comment override. |
| `docs/templates/change-specification.md`; `docs/templates/release-record.md`; `docs/templates/technical-approval-request.md` | Current authoring templates corrected to active A-004, stronger R4 evidence, requirement clarification, and exceptional-only human review. |
| `tooling/governance/fixtures/karsift-ai-infra/README.md`; `docs/operations/11-devops-and-ci-cd.md`; `docs/governance/repository-settings.md`; `docs/governance/post-merge-activation-checklist.md`; `docs/operations/19-governance-reconciliation-notes.md`; `scripts/governance/validate-governance.sh` | Current recovery, two-token, active-A-004, release/deploy, RL1/RL2, and validator claims corrected and revalidated. |
| `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`; `tooling/governance/tests/test_voc129_caller_replacement.py`; `tooling/governance/tests/test_voc136_caller_replacement.py`; `tooling/governance/tests/test_voc137_pr_sha_scan.py`; `tooling/governance/tests/test_voc138_promotion_pr_provenance.py`; `tooling/governance/tests/test_voc139_promotion_recovery_metadata.py`; `tooling/governance/tests/test_voc140_release_convergence.py` | All six changed live `CURRENT_PIN` assertions were updated to `9fdff24…`, with the literal 18-file hash table in the VOC-140 test; issue-era `AUTHORITATIVE_PIN` / `VOC139_INFRA_PIN` values remain `59943683…`. The VOC-140 regression also enumerates every corrected live-authority path and rejects its prior stale literal. |
| `docs/governance/amendments/A-002-governed-autonomous-releases.md`; `docs/governance/amendments/A-003-governed-autonomous-engineering-authority.md`; `docs/governance/amendments/A-004-remove-founder-approval-gates-from-autonomous-engineering-workflows.md`; `docs/governance/a003-transition-state.yaml`; `docs/governance/a004-transition-state.yaml`; `docs/governance/technical-steward-appointment.md` | Frozen amendment, transition, and appointment evidence preserved. Current overlay notices already distinguish active A-004 from historical authority. |
| `docs/architecture/17-autonomous-development-architecture.md`; `docs/planning/18-autonomous-development-implementation-roadmap.md` (DOC-17 / DOC-18) | Preserved adopted historical design/roadmap exactly as required; `docs/README.md` and DOC-19 explicitly identify their non-current disposition. |
| `docs/archive/`; historical `specs/changes/` packages | Preserved audit evidence. VOC-138 / VOC-139 were explicitly excluded from edits. |
| `tooling/governance/validate_repository_foundation.py`; `tooling/governance/tests/test_validate_repository_foundation.py`; `tooling/governance/tests/test_voc080_workflow_policy.py`; historical amendment assertions in `scripts/governance/validate-governance.sh` | Preserved validator/test historical strings required to validate frozen A-003/A-004 records; no live gate depends on them. |
| The 18 paths under `tooling/governance/fixtures/karsift-ai-infra/` listed by the literal hash table | Exact byte/mode mirrors of authoritative infra merge `9fdff24…`; old token/recovery strings inside older package evidence remain historical rather than current fixture contract. |

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

## Validation commands (final implementation revision)

Manual reconciliation after the independently reviewed infrastructure merge. The
evidence-bearing commit was revalidated without further file changes; these results
therefore describe the exact pushed implementation revision:

```bash
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
# OK — 284 tests

python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
# OK — 451 tests

python3 -m unittest discover -s tooling/governance/tests -p 'test_voc140_release_convergence.py'
# OK — 9 tests

python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_voc140_*.py'
# OK — 27 tests

python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_production_merge_guard.py'
# OK — 7 tests

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
`test_voc138_promotion_pr_provenance.py` also forces the historical capture object
to be unavailable through a deterministic Git wrapper, reproducing GitHub's shallow
checkout while retaining the fail-closed `pr-ancestry` assertion in full local clones.

Governance validation with exact base/head and independent exact-SHA review bind
the live implementation PR head after commit; this file does not require the
same commit to contain that head SHA.

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR must
bind the live head exactly. Merge-gate must reject any mismatch.
