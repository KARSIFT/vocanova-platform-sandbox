# VOC-106-T00 — Evidence

## evidence_id

VOC-106-EV-00

## gate_status

complete

## drafting-time diagnosis

Confirmed `karsift-ai-infra/config/decide-remediation.py` returned `RETRY` on
`ci_failed` or review `FAIL` without reading
`<package_path>/.karsift/live-evidence/<task_id>.yaml` ownership.
`WAITING` already suppressed retry. Issue #882 records cancelled remediate→
implement selection for operator-owned VOC-104-T01 (carrier PR #879; workflow
32529860337); branch head unchanged.

## remediation (attempt 2)

Independent review of `d35ee3d9fc9f781a8169b68aa3051f3af0660c2c` failed on a
High: hosted
`verify-remediate-operator-ownership-runner.py` extracted base SHA via
`str(run.get("pull_requests") or [{}])[0].get(...)`, which stringifies the list
to `"[{...}]"`, indexes the character `"["`, and raises `AttributeError` before
the PR-view fallback. Fixed by adding pure helper
`expected_base_sha_from_run` and calling it from the runner (fixture + infra).
TEST-11 now asserts the helper and forbids the `str(run.get("pull_requests"`
pattern. Medium: `PINNED_SHA.txt` updated to infra merge
`54573e94e62e671f023f521a07770b1d30889591`. Low: fixture README restored to
pinned-contract framing with an explicit VOC-106 section.

## commands

```bash
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_remediation_ownership.py' -v
python3 -m unittest discover -s tests -p 'test_remediate*.py' -v

cd ..
python3 -m unittest tooling.governance.tests.test_remediate_ownership -v
node --test scripts/foundation/voc106-remediate-ownership.test.mjs

bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## results

- Shared-infrastructure PR `KARSIFT/karsift-ai-infra#94` merged at
  `54573e94e62e671f023f521a07770b1d30889591`. Hosted `actionlint`,
  `shellcheck`, YAML parse, and policy-test jobs all passed on that merge.
- Attempt-2 verifier base-SHA fix applied in calling-repo fixtures
  (`tooling/governance/fixtures/karsift-ai-infra/config/`) and prepared in the
  local untracked `karsift-ai-infra/` tree (same three files). This implementer
  session had no GitHub credentials to open the follow-up infra PR; `@main`
  still needs that one-line adapter helper merged before T01's hosted
  `verify-remediate-operator-ownership` dispatch can succeed against live
  Actions metadata.
- `python3 -m unittest tooling.governance.tests.test_remediate_ownership -v`:
  **9 passed** (TEST-00–07, TEST-11 including base-SHA helper regression).
- `node --test scripts/foundation/voc106-remediate-ownership.test.mjs`:
  **3 passed** (TEST-00–07 matrix via unittest, TEST-10 docs, TEST-11 wiring).
- Local infra `test_remediation_ownership.py`: **6 passed** (includes base-SHA
  extraction regression).
- `validate-governance.sh`: **passed**.
- `classify-change-risk.sh`: **passed** (path floor R4 on fixture/governance
  paths; no PR risk declaration in this local run).
- `git diff --check`: **passed**.

## implementation notes

- `remediate.yml` checks out the exact PR head (`pr-head/`) and runs
  `remediate-ownership-classifier.py` before any `should_retry=true` path.
- Operator / live-actions ownership routes to sanitized PR escalation via
  `remediate-escalate-operator.py`; malformed metadata routes to
  `remediate-fail-closed.py`. Neither path dispatches `implement.yml`.
- Calling-repo `pipeline.yml` exposes read-only
  `verify-remediate-operator-ownership` workflow dispatch (consumed at
  `KARSIFT/karsift-ai-infra/.github/workflows/verify-remediate-operator-ownership.yml@main`).
  It reuses existing generic run/PR/change/task/package inputs so the complete
  `workflow_dispatch` schema remains within GitHub's 25-input hard limit.
- Infra README and `docs/operations/live-evidence.md` document ownership-gated
  FAIL/CI remediation and retained ordinary bounded retry.
- Fixture pin: `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` →
  `54573e94e62e671f023f521a07770b1d30889591`.
- VOC-106-TEST-08 / VOC-106-TEST-09 live proof remains operator-owned T01 work.

No secrets, logs, artifacts, or unrelated package live evidence recorded here.
