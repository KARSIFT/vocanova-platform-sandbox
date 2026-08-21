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

## commands

```bash
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_remediate*.py' -v

cd ..
node --test scripts/foundation/voc106-remediate-ownership.test.mjs

bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## results

- Shared-infrastructure PR `KARSIFT/karsift-ai-infra#94` merged at exact head
  `d0405f43fa66beb6642835dcbd40d346cba528db` (merge
  `54573e94e62e671f023f521a07770b1d30889591`). Hosted `actionlint`,
  `shellcheck`, YAML parse, and policy-test jobs all passed.
- The original isolated publisher correctly refused the calling-repository
  bundle because it contains governed workflow changes. The already-validated
  bundle from run `32535480267` was recovered without rerunning the model and
  published through the supervised workflow-change path.
- `karsift-ai-infra` remediate policy + ownership tests: **20 passed** (including
  VOC-106-TEST-00 through TEST-07 and TEST-11 matrix cases in
  `tests/test_remediate_ownership.py`).
- Calling-repo foundation tests: **3 passed**
  (`scripts/foundation/voc106-remediate-ownership.test.mjs`), covering
  VOC-106-TEST-00 through TEST-07, TEST-10, and TEST-11.
- `validate-governance.sh`: **passed** (repository foundation, monitoring impact,
  governance structure).
- `classify-change-risk.sh`: **passed** (path floor reported; no PR risk
  declaration in this local run).
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
- VOC-106-TEST-08 / VOC-106-TEST-09 live proof remains operator-owned T01 work.

No secrets, logs, artifacts, or unrelated package live evidence recorded here.
