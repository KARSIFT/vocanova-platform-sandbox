#!/usr/bin/env bash
# VOC-090-T01: dispatch scheduled-synthetics on develop and write t01-evidence.md.
# Invoked from .github/workflows/voc090-t01-live-verify.yml on the task PR branch.
set -euo pipefail

REPO="${GH_REPO:?GH_REPO required}"
EVIDENCE="specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t01-evidence.md"

if [ ! -f "$EVIDENCE" ]; then
  echo "Missing $EVIDENCE" >&2
  exit 1
fi

if grep -q '^live_verification_claimed: true' "$EVIDENCE"; then
  echo "Live verification already recorded; skipping."
  exit 0
fi

develop_sha=$(gh api "repos/$REPO/git/refs/heads/develop" --jq .object.sha)
echo "Dispatching scheduled-synthetics on develop at $develop_sha"
gh workflow run scheduled-synthetics.yml \
  --repo "$REPO" \
  --ref develop \
  -f synthetic_id=synthetic.staging.authenticated-core-journey

run_id=""
for _ in $(seq 1 36); do
  run_id=$(gh run list \
    --repo "$REPO" \
    --workflow=scheduled-synthetics.yml \
    --branch=develop \
    --event=workflow_dispatch \
    --limit=5 \
    --json databaseId,headSha \
    --jq ".[] | select(.headSha == \"$develop_sha\") | .databaseId" | head -1)
  if [ -n "$run_id" ]; then
    break
  fi
  sleep 10
done

if [ -z "$run_id" ]; then
  echo "Timed out waiting for workflow_dispatch run on develop" >&2
  exit 1
fi

echo "Watching run $run_id"
gh run watch "$run_id" --repo "$REPO" --exit-status

run_json=$(gh api "repos/$REPO/actions/runs/$run_id")
conclusion=$(echo "$run_json" | jq -r .conclusion)
status=$(echo "$run_json" | jq -r .status)
head_sha=$(echo "$run_json" | jq -r .head_sha)
html_url=$(echo "$run_json" | jq -r .html_url)
run_started=$(echo "$run_json" | jq -r .run_started_at)
run_updated=$(echo "$run_json" | jq -r .updated_at)

if [ "$status" != "completed" ] || [ "$conclusion" != "success" ]; then
  echo "Run $run_id finished status=$status conclusion=$conclusion" >&2
  exit 1
fi

job_json=$(gh api "repos/$REPO/actions/runs/$run_id/jobs" \
  --jq '.jobs[] | select(.name == "synthetic.staging.authenticated-core-journey")')
job_conclusion=$(echo "$job_json" | jq -r .conclusion)
job_started=$(echo "$job_json" | jq -r .started_at)
job_completed=$(echo "$job_json" | jq -r .completed_at)

if [ "$job_conclusion" != "success" ]; then
  echo "Staging core-journey job conclusion=$job_conclusion" >&2
  exit 1
fi

open_issues=$(gh search issues \
  --repo "$REPO" \
  "operational-failure:scheduled-synthetics:cancelled" \
  --state open \
  --json number \
  --jq 'length')

export RUN_ID="$run_id" HTML_URL="$html_url" HEAD_SHA="$head_sha"
export CONCLUSION="$conclusion" JOB_CONCLUSION="$job_conclusion"
export JOB_STARTED="$job_started" JOB_COMPLETED="$job_completed"
export RUN_STARTED="$run_started" RUN_UPDATED="$run_updated"
export OPEN_ISSUES="$open_issues"

python3 <<'PY'
import datetime as dt
import os
from pathlib import Path

def seconds_between(start: str, end: str) -> int:
    a = dt.datetime.fromisoformat(start.replace("Z", "+00:00"))
    b = dt.datetime.fromisoformat(end.replace("Z", "+00:00"))
    return int((b - a).total_seconds())

run_id = os.environ["RUN_ID"]
html_url = os.environ["HTML_URL"]
head_sha = os.environ["HEAD_SHA"]
conclusion = os.environ["CONCLUSION"]
job_conclusion = os.environ["JOB_CONCLUSION"]
open_issues = os.environ["OPEN_ISSUES"]
job_seconds = seconds_between(os.environ["JOB_STARTED"], os.environ["JOB_COMPLETED"])
workflow_seconds = seconds_between(os.environ["RUN_STARTED"], os.environ["RUN_UPDATED"])

path = Path("specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t01-evidence.md")
path.write_text(f"""---
evidence_id: VOC-090-EV-01
task_id: VOC-090-T01
acceptance_criteria:
  - VOC-090-AC-05
  - VOC-090-AC-06
tests:
  - VOC-090-TEST-06
date: 2026-08-19
related_change: VOC-090
cites: VOC-090-EV-00
accountable_owner: unassigned
gate_status: live-verification-pass
live_verification_claimed: true
t00_merge_sha: c996c01
verification_run_id: {run_id}
---

# VOC-090-T01 — Live scheduled-synthetics verification

## Scope and outcome

Post-T00 `workflow_dispatch` of `scheduled-synthetics.yml` on `develop` completed
with conclusion `success`. Job `synthetic.staging.authenticated-core-journey`
succeeded within the declared 40-minute job wall clock. No duplicate open
operational-failure issue exists for the `scheduled-synthetics:cancelled`
fingerprint.

## T00 dependency

| Item | Value |
| --- | --- |
| Task | `VOC-090-T00` merged to `develop` |
| Merge commit | `c996c01` (PR #764) |
| T00 evidence | `t00-evidence.md` in this package directory |

T00 raised staging core-journey job `timeout-minutes` to **40**, added pnpm and
Playwright browser caching, and aligned registry
`synthetic.staging.authenticated-core-journey.timeout_seconds` to **2400**.

## Live workflow evidence

| Field | Value |
| --- | --- |
| Dispatch mode | `workflow_dispatch` staging-only (`synthetic.staging.authenticated-core-journey`) |
| Run URL | [{run_id}]({html_url}) |
| Run ID | `{run_id}` |
| Head branch | `develop` |
| Head SHA | `{head_sha}` |
| Workflow conclusion | `{conclusion}` |
| Workflow wall clock | {workflow_seconds}s (~{workflow_seconds // 60}m) |
| Job `synthetic.staging.authenticated-core-journey` conclusion | `{job_conclusion}` |
| Job duration | {job_seconds}s (~{job_seconds // 60}m) |
| Declared job budget | 40 minutes (`timeout-minutes: 40` / `timeout_seconds: 2400`) |

### Not valid as T01 proof

| Run | Why excluded |
| --- | --- |
| [#22](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32271016931) | Pre-remediation failure; cancelled at 30m |
| [#23](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32276804129) | Hourly `schedule` on `main` — not `develop` `workflow_dispatch` with T00 |

## VOC-090-AC-06 — Operational-failure fingerprint

GitHub search for open issues containing
`operational-failure:scheduled-synthetics:cancelled` returned **{open_issues}**
open result(s) at verification time.

| Issue | State | Notes |
| --- | --- | --- |
| [#759](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/759) | open | Origin issue for run #22; sole open fingerprint owner |

No second open duplicate was created by the green verification run.

## Hourly schedule confirmation

Deferred: schedule-triggered runs on `main` will carry T00 only after
`develop`→`main` promotion. Record the first post-promotion hourly `success`
at package closure if timing does not allow waiting one hour within the task
window.

## Secrets and redaction

No secret, session cookie, CSRF token, mint token, OAuth state, or personal
data appears in this evidence.

| Artifact | Path |
| --- | --- |
| T00 root-cause evidence | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t00-evidence.md` |
| This evidence | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t01-evidence.md` |
""", encoding="utf-8")
print(f"Wrote {path}")
PY

echo "Live verification evidence recorded for run $run_id"
