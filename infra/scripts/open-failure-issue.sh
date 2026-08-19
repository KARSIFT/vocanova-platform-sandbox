#!/usr/bin/env bash
set -euo pipefail

# VOC-088-T02: open a plain, unlabeled operational-failure issue using only
# bounded GitHub workflow_run metadata. Never add logs or step output here.

required_names=(
  GH_TOKEN
  GH_REPOSITORY
  FAILURE_WORKFLOW_NAME
  FAILURE_CONCLUSION
  FAILURE_RUN_URL
)
for required_name in "${required_names[@]}"; do
  if [ -z "${!required_name:-}" ]; then
    echo "Missing required input: ${required_name}" >&2
    exit 1
  fi
done

case "$FAILURE_WORKFLOW_NAME" in
  scheduled-synthetics | deploy-staging | deploy-production) ;;
  *)
    echo "Refusing unobserved workflow name" >&2
    exit 1
    ;;
esac

case "$FAILURE_CONCLUSION" in
  failure | cancelled | timed_out) ;;
  *)
    echo "Conclusion is not an observed terminal failure" >&2
    exit 1
    ;;
esac

if [[ ! "$GH_REPOSITORY" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo "Invalid repository identifier" >&2
  exit 1
fi

expected_url_prefix="https://github.com/${GH_REPOSITORY}/actions/runs/"
if [[ "$FAILURE_RUN_URL" != "${expected_url_prefix}"* ]]; then
  echo "Refusing non-canonical workflow run URL" >&2
  exit 1
fi
run_id="${FAILURE_RUN_URL#"$expected_url_prefix"}"
if [[ ! "$run_id" =~ ^[0-9]+$ ]]; then
  echo "Refusing non-canonical workflow run URL" >&2
  exit 1
fi

fingerprint="${FAILURE_WORKFLOW_NAME}:${FAILURE_CONCLUSION}"
marker="<!-- operational-failure:${fingerprint} -->"
title="Operational failure: ${FAILURE_WORKFLOW_NAME} (${FAILURE_CONCLUSION})"

open_issues_json="$(
  gh api \
    --method GET \
    --paginate \
    --slurp \
    "repos/${GH_REPOSITORY}/issues?state=open&per_page=100"
)"

if jq -e --arg marker "$marker" \
  'flatten | any(.[]; (.pull_request | not) and ((.body // "") | contains($marker)))' \
  >/dev/null <<<"$open_issues_json"; then
  echo "An open operational issue already owns this failure fingerprint; no duplicate created"
  exit 0
fi

body_file="$(mktemp)"
trap 'rm -f "$body_file"' EXIT
{
  printf '%s\n\n' "$marker"
  printf '%s\n\n' "An expected operational workflow ended unsuccessfully. Investigate through the governed issue-to-plan process."
  printf '%s\n' "- Workflow: \`${FAILURE_WORKFLOW_NAME}\`"
  printf '%s\n' "- Conclusion: \`${FAILURE_CONCLUSION}\`"
  printf '%s\n' "- Run: ${FAILURE_RUN_URL}"
  printf '%s\n' "- Sanitization: no job logs, step output, credentials, sessions, OAuth state, or user identifiers were copied."
} >"$body_file"

# Deliberately omit --label: a plain App-created issue is what triggers the
# repository's plan-from-issue route.
gh api \
  --method POST \
  "repos/${GH_REPOSITORY}/issues" \
  -f "title=${title}" \
  -F "body=@${body_file}" \
  --silent

echo "Created one sanitized operational failure issue"
