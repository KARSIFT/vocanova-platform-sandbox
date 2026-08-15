# VOC-080-EV-06 — T06 non-production rehearsal proof

Evidence for `VOC-080-AC-00`, `VOC-080-AC-01`, `VOC-080-AC-04`,
`VOC-080-AC-05`, `VOC-080-AC-06`, and `VOC-080-AC-10` (readiness note).
Tests: `VOC-080-TEST-06`.

**Authority state:** A-003 remains effective until `VOC-080-T07`. This task
records rehearsal evidence only; it does **not** activate A-004 or flip
`authority_model`.

**Remediation (attempt 2):** Independent review of `29f13c57…` failed on an
out-of-scope binary change under protected `tooling/governance/__pycache__/`.
That churn is reverted to the pre-T06 tip (`81a961c`); this evidence file is
the only intentional T06 tree delta versus that tip. No `__pycache__` or other
`tooling/governance/` bytecode is part of this task’s deliverable.

## Task outcome

`VOC-080-T06` rehearses the post-T01/T02/T03/T04/T05 no-founder-gate loops on
the settled venue (`vocanova-platform-sandbox` + `karsift-ai-infra` self-ci /
policy harness). Live GitHub Actions runs on this sandbox demonstrate autonomous
adoption recovery, R4 task-PR auto-merge without founder `approved` comments,
reconcile-dispatch repair, release-path behavior without founder-comment gates,
and fail-closed merge behavior when verification or checks fail.

## Rehearsal venue (`VOC-080-DEP-04`)

| Venue | Role |
|-------|------|
| `KARSIFT/vocanova-platform-sandbox` | Primary live rehearsal target (`auto_merge_enabled=true`, `auto_release_enabled=true`) |
| `KARSIFT/karsift-ai-infra` `@main` | Reusable merge-gate / adopt / release contracts consumed by the sandbox |
| `tooling/governance/tests/test_voc080_*.py` | Caller policy harness (pinned infra fixtures) |
| `karsift-ai-infra/tests/test_*.py` | Infra policy harness (self-ci) |

Infra fixture pin in caller tests:
`489dd82b5403a36082e70c95185463f445d02c13` (post-T03 `main`).

## Rehearsal matrix

| Surface | Acceptance / test IDs | Live evidence | Harness evidence | Result |
|---------|----------------------|---------------|------------------|--------|
| Autonomous adopt + task roster | AC-00, AC-04; TEST-00 | Reconcile dispatch + roster PR (below) | `test_voc080_adoption_reconcile_policy.py` | **pass** |
| R4 merge without founder comment | AC-01; TEST-01 | Task PRs #640–#644 merge-gate auto-merge (below) | `test_voc080_merge_gate_policy.py` | **pass** |
| Reconcile dispatch recovery | AC-05; TEST-04 | `workflow_dispatch` run `31855980711` (below) | adopt.yml idempotency contract tests | **pass** (first live run; second live run — see limitations) |
| No founder override of failed gates | AC-02; TEST-02 | PR #644 fail-closed comments before PASS (below) | merge-gate + remediate policy tests | **pass** |
| Unparseable risk fail-closed | AC-03; TEST-03 | merge-gate contract (no live unknown-risk PR in this pass) | `test_unparseable_risk_fails_closed` | **pass** (contract + policy; live fixture PR not opened) |
| Release without founder interaction | AC-06; TEST-05 | T03 close release job (below) | `test_release_policy.py`, `test_voc080_workflow_policy.py` | **pass** (roster-closure promotion deferred) |
| VOC-079 resume readiness | AC-10 | Gates cleared for post-T07 path (below) | — | **ready pending T07 activation** |

## 1. Autonomous adopt and task roster (`VOC-080-TEST-00`)

### Historical contrast (pre-T01, not claimed as post-transition proof)

Plan PR [#628](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/628)
merged `2026-08-15T00:41:58Z` under the **pre-T01** merge-gate, which still
required a founder `approved` comment for R4. The merge-gate bot posted
`requires founder approval - R4 never auto-merges` and founder `m-e-h-r-d-a-a-d`
replied `approved`. This is preserved as historical evidence of the gap #627
closes; it is **not** post-transition behavior.

### Post-T02 autonomous adoption recovery

| Step | Evidence |
|------|----------|
| Recovery PR merged | [#629](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/629) *Recover VOC-080 autonomous adoption handoff* — merged `2026-08-15T01:14:19Z` |
| Reconcile dispatch | [pipeline run 31855980711](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31855980711) (`workflow_dispatch` on `develop`, success) — jobs `adopt / adopt` and `adopt / implement-first-task / implement` succeeded |
| Roster PR | [#638](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/638) *VOC-080: record task-issue roster and authority_issue* on branch `karsift/roster-voc-080` — merged `2026-08-15T01:17:24Z`; [run 31856002535](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31856002535) |
| Task issues opened | Issues [#630](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/630)–[#637](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/637) created `2026-08-15T01:15:03Z`–`01:15:11Z` (no founder comment on the critical path after reconcile) |
| Package adoption fields | `change.yaml` on `develop`: `status: adopted`, `implementation_authorized: true`, `approval_status: autonomously-adopted-after-independent-verification`, `adoption_pr: 628` |

**Conclusion:** merged plan package → reconcile dispatch → adopted roster +
task issues, without a founder `approved` comment on the recovery path.

## 2. R4 auto-merge without founder comment (`VOC-080-TEST-01`)

All VOC-080 implementation task PRs merged after T01 landed with merge-gate
`auto-merge` succeeding and **no** founder `approved` comment in the merge
path. Representative runs and bot status:

| PR | Risk | Merge run | merge-gate `auto-merge` | Founder `approved` on merge path |
|----|------|-----------|-------------------------|----------------------------------|
| [#640](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/640) T01 | R4 | [31856600505](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31856600505) | success | **none** — bot: `WOULD AUTO-MERGE (R4, auto_merge_enabled=true)` |
| [#641](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/641) T02 | R4 | [31857882910](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31857882910) | success | **none** |
| [#642](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/642) T03 | R4 | [31871459463](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31871459463) | success | **none** |
| [#643](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/643) T04 | R4 | [31872074497](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31872074497) | success | **none** |
| [#644](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/644) T05 | R4 | [31872866317](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31872866317) | success | **none** |

Post-T01 status text explicitly states:
`Founder approval comments are not part of the merge path; failed or missing
gates remain fail-closed.`

**Conclusion:** R4 task PRs auto-merged when CI, checks, and independent
verification passed — without founder-comment gating.

## 3. Reconcile dispatch (`VOC-080-TEST-04`)

### First live reconcile (event-independent recovery)

```bash
# Observed equivalent of:
gh workflow run pipeline.yml --ref develop \
  -f action=reconcile \
  -f plan_pr_number=628
```

| Item | Value |
|------|-------|
| Run | [31855980711](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31855980711) |
| Trigger | `workflow_dispatch` on `develop` |
| Outcome | `adopt` success → roster PR #638 → eight task issues #630–#637 |
| Founder comment required | **no** |

### Idempotency

Policy contract (adopt.yml + caller `pipeline.yml`) verified by
`test_voc080_adoption_reconcile_policy.py` (`existing task issues are reused`,
`unchanged roster is a no-op`, `Reconciliation is already complete`).

A **second** live `workflow_dispatch` reconcile for plan PR #628 was **not**
executed in this implementer run (no `GH_TOKEN` / `workflow_dispatch` credential
in the implementer sandbox — see Limitations). Idempotency of the second run
remains a **documented contract + policy-test** pass, not a second live run URL.

## 4. Fail-closed gates — no founder override (`VOC-080-TEST-02`)

Live observation on PR [#644](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/644)
(T05) before the final PASS revision:

| Commit phase | merge-gate status | Merged? |
|--------------|-------------------|---------|
| `0a93dfaf` | `would NOT merge yet (risk=R4, checks_ok=true, verdict=FAIL)` | **no** |
| `c4b161f6` | `would NOT merge yet (risk=R4, checks_ok=false, verdict=FAIL)` | **no** |
| `ff26c931` | `WOULD AUTO-MERGE (R4, auto_merge_enabled=true)` | **yes** (after PASS) |

Each fail-closed comment included:
`Founder approval comments are not part of the merge path`.

Infra PR #37 removed `approve-and-merge`; plan PR #628 run
[31854203601](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31854203601)
shows `merge-gate / approve-and-merge: skipped` on the post-T01 contract.

**Conclusion:** failed/missing verification blocks merge; no comment-based bypass.

## 5. Unparseable risk fail-closed (`VOC-080-TEST-03`)

### Policy harness (deterministic)

`test_voc080_merge_gate_policy.py::test_unparseable_risk_fails_closed` asserts
the pinned `merge-gate.yml` contract:

- missing parseable `Risk classification: R#` → `risk="unknown"`
- status `BLOCKED - no parseable 'Risk classification: R#' line found`
- `auto-merge` `if:` excludes `risk == 'unknown'`
- `never auto-mergeable` for unknown risk
- no `approve-and-merge` escape hatch

### Live sandbox

No dedicated rehearsal PR with a missing risk declaration was opened during
this T06 pass (would require a throwaway plan/task PR outside the VOC-080
roster). The live merge-gate on `@main` encodes the same fail-closed branch
(lines 93–101, 199–200, 220 in `karsift-ai-infra/.github/workflows/merge-gate.yml`
at pin `489dd82`).

**Conclusion:** contract enforced in production reusable workflow + policy
tests; dedicated live unknown-risk PR not run (limitation recorded).

## 6. Release path without founder interaction (`VOC-080-TEST-05`)

### Live sandbox (post-T03)

When task issue #633 (T03) closed after merge of PR #642:

| Run | Jobs observed | Founder comment gate |
|-----|---------------|---------------------|
| [31871607450](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31871607450) | `release / identify` success; `release / check-and-open` success; `auto-promote` skipped (roster not yet complete) | **none** |

`auto-promote` correctly skipped because the VOC-080 roster still has open
tasks (#636, #637). Full develop→main promotion for this package is expected
after T07 activation and roster closure — not claimed here.

### Infra self-ci

| Run | SHA | Outcome |
|-----|-----|---------|
| [31872588453](https://github.com/KARSIFT/karsift-ai-infra/actions/runs/31872588453) | `fcf2224` (post-T05 policy tests) | success |
| [31871414819](https://github.com/KARSIFT/karsift-ai-infra/actions/runs/31871414819) | `489dd82` (T03 release behavior) | success |

`test_release_policy.py` and caller `test_voc080_adoption_reconcile_policy.py`
assert: no founder `issue_comment` release authority; `reconcile-release`
dispatch for retry; production environment `reviewers: null` (recorded in
`VOC-080-EV-03`).

Push-driven production deploy (repository-controlled path, no founder approval
job) last observed on `main` push:
[deploy-production 31847058739](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31847058739)
(`2026-08-14T22:33:16Z`, success). That run predates T03; post-T03 deploy
contract is asserted by `test_voc080_workflow_policy.py` and unchanged
`deploy-production.yml` push trigger.

**Conclusion:** release identify/check-and-open runs without founder-comment
gates; full VOC-080 promotion rehearsal deferred until roster closes.

## 7. VOC-079 / issue #624 readiness (`VOC-080-AC-10`)

After T07 activation (not performed in this task):

- [VOC-079](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/624)
  plan package [#625](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/625)
  was the historical merged-as-draft failure class; recovery architecture now
  includes reconcile dispatch (proven for VOC-080 above).
- Post-T01 merge-gate, post-T02 adopt, and post-T03 release paths no longer
  require founder `approved` on engineering-workflow gates.
- VOC-079 implementation remains out of scope; only **gate clearance** for
  post-T07 progression is claimed here.

## Deterministic harness (this remediation run)

```bash
# Caller policy harness (VOC-080 modules)
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc080*.py' -v
# Observed: 32 tests, OK

# Full caller governance suite
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py' -v
# Observed: 94 tests, OK

# Infra policy harness (local checkout)
cd karsift-ai-infra && python3 -m unittest discover -s tests -p 'test_*.py' -v
# Observed: 26 tests, OK

bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
# Observed: foundation + governance structure validation passed; git diff --check clean
```

## Explicitly not done

- A-004 / `authority_model` activation (`VOC-080-T07`)
- Closing issue [#627](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627)
- VOC-079 technical cutover
- Full VOC-080 develop→main promotion (roster still open)

## Limitations

| Limitation | Impact |
|------------|--------|
| No `GH_TOKEN` in implementer sandbox | Could not dispatch a **second** live reconcile or a throwaway unknown-risk PR in this run; idempotency and unknown-risk live clauses rely on policy tests + first reconcile URL |
| Plan PR #628 merged under pre-T01 founder gate | Historical only; post-T02 recovery path is the authoritative adopt rehearsal |
| VOC-080 roster not complete | `release / auto-promote` not exercised end-to-end for this package; T03 release job behavior + policy tests cover the contract |
| A-003 still effective | Engineering workflows on this sandbox already follow post-T01/T02/T03 infra behavior, but canonical authority flip awaits T07 |

## Overall T06 result

**PASS with recorded limitations.** Live sandbox runs prove autonomous adopt
recovery, R4 auto-merge without founder comments, reconcile dispatch, and
fail-closed merge behavior. Release and unparseable-risk surfaces are proven
by live partial runs plus deterministic policy harness; second reconcile
dispatch and dedicated unknown-risk PR remain documented gaps in this attempt,
not silent passes.
