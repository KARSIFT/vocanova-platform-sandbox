#!/usr/bin/env bash
set -euo pipefail

# VOC-094-T00: classify concurrency-superseded deploy cancellations using
# bounded GitHub Actions metadata only (run conclusion and job count via API).
#
# Exit 0 — benign pending-run supersession; the observer should skip issue creation.
# Exit 1 — not benign, or fail-closed when metadata is ambiguous or unavailable.

required_names=(
  GH_TOKEN
  GH_REPOSITORY
  FAILURE_WORKFLOW_NAME
  FAILURE_CONCLUSION
  FAILURE_RUN_ID
)
for required_name in "${required_names[@]}"; do
  if [ -z "${!required_name:-}" ]; then
    echo "Missing required input: ${required_name}" >&2
    exit 1
  fi
done

case "$FAILURE_WORKFLOW_NAME" in
  deploy-staging | deploy-production) ;;
  *)
    exit 1
    ;;
esac

if [ "$FAILURE_CONCLUSION" != "cancelled" ]; then
  exit 1
fi

if [[ ! "$GH_REPOSITORY" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo "Invalid repository identifier" >&2
  exit 1
fi

if [[ ! "$FAILURE_RUN_ID" =~ ^[0-9]+$ ]]; then
  echo "Invalid workflow run id" >&2
  exit 1
fi

jobs_json=""
if ! jobs_json="$(
  gh api \
    --method GET \
    "repos/${GH_REPOSITORY}/actions/runs/${FAILURE_RUN_ID}/jobs"
)"; then
  echo "Unable to read bounded job metadata; fail-closed toward issue creation" >&2
  exit 1
fi

total_count="$(jq -r '.total_count // empty' <<<"$jobs_json")"
if [ -z "$total_count" ] || ! [[ "$total_count" =~ ^[0-9]+$ ]]; then
  echo "Ambiguous job count metadata; fail-closed toward issue creation" >&2
  exit 1
fi

if [ "$total_count" -eq 0 ]; then
  echo "Concurrency-superseded deploy cancellation (zero jobs started)"
  exit 0
fi

exit 1
