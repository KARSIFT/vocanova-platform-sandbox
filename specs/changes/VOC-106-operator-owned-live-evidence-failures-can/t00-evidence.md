# VOC-106-T00 — Evidence (pending implementation)

## evidence_id

VOC-106-EV-00

## gate_status

pending

## drafting-time diagnosis

Confirmed `karsift-ai-infra/config/decide-remediation.py` returns `RETRY` on
`ci_failed` or review `FAIL` without reading
`<package_path>/.karsift/live-evidence/<task_id>.yaml` ownership.
`WAITING` already suppresses retry. Issue #882 records cancelled remediate→
implement selection for operator-owned VOC-104-T01 (carrier PR #879; workflow
32529860337); branch head unchanged.

## commands

Record exact commands and results at implementation time. Expected shape:

```bash
# karsift-ai-infra policy tests (names may vary)
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py'

# calling-repo foundation tests (if landed)
node --test scripts/foundation/voc106-*.test.mjs

bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## results

Pending T00 implementation. No secrets, logs, artifacts, or unrelated package
live evidence belong here.
