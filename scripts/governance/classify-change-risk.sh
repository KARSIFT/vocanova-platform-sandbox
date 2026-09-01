#!/usr/bin/env bash
set -euo pipefail

base=""
head=""
files_from=""
pr_body_file=""
require_declaration=false
base_supplied=false
head_supplied=false

usage() {
  echo "Usage: $0 [--base SHA --head SHA | --files-from FILE] [--pr-body-file FILE] [--require-declaration]" >&2
}

while (($#)); do
  case "$1" in
    --base) base="${2:-}"; base_supplied=true; shift 2 ;;
    --head) head="${2:-}"; head_supplied=true; shift 2 ;;
    --files-from) files_from="${2:-}"; shift 2 ;;
    --pr-body-file) pr_body_file="${2:-}"; shift 2 ;;
    --require-declaration) require_declaration=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) usage; echo "Unknown argument: $1" >&2; exit 2 ;;
  esac
done

if [[ "$base_supplied" == true && "$head_supplied" != true ]] || \
   [[ "$head_supplied" == true && "$base_supplied" != true ]]; then
  echo "A changed-file range requires both --base and --head." >&2
  exit 1
fi

# shellcheck source=load-changed-files.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/load-changed-files.sh"
load_changed_files "$files_from" "$base" "$head"
declare -a files=("${GOVERNANCE_CHANGED_FILES[@]}")

if ((${#files[@]} == 0)); then
  echo "No changed files to classify."
  exit 0
fi

risk_rank() {
  case "$1" in
    R0) echo 0 ;; R1) echo 1 ;; R2) echo 2 ;; R3) echo 3 ;; R4) echo 4 ;;
    *) return 1 ;;
  esac
}

path_risk() {
  local path="$1"
  case "$path" in
    CODEOWNERS|.github/CODEOWNERS|.github/approved-policy/*|.github/workflows/governance-policy.yml|.github/workflows/repository-governance.yml|scripts/governance/*|tooling/governance/*|docs/governance/amendments/*|docs/governance/a003-transition-state.yaml|docs/governance/16-autonomous-development-operating-model.md|docs/governance/approval-matrix.md|docs/governance/change-risk-classification.md|docs/governance/protected-areas.md|docs/governance/post-merge-activation-checklist.md|docs/operations/15-ai-native-product-and-engineering-operating-model.md|docs/architecture/17-autonomous-development-architecture.md|docs/planning/18-autonomous-development-implementation-roadmap.md|specs/changes/VOC-001-repository-foundation/*|specs/changes/VOC-002-a003-governance-transition/*|specs/changes/VOC-003-a003-lifecycle-sync/*|specs/changes/VOC-004-canonical-adoption-doc-17-doc-18/*)
      echo R4 ;;
    .github/workflows/*|.github/pull_request_template.md|CONTRIBUTING.md|SECURITY.md|AGENTS.md|CLAUDE.md|docs/governance/*|specs/README.md|specs/templates/*|*/auth/*|*/authorization/*|*/payments/*|*/billing/*|*/migrations/*|packages/auth/*|packages/database/*|packages/billing/*|packages/ai/*|packages/audio/*|packages/voice/*|packages/storage/*|database/*|db/*|migrations/*|infrastructure/*|infra/*|backups/*|scripts/backup/*|scripts/restore/*|scripts/deploy/*|scripts/release/*|scripts/rollback/*|.env|.env.*|*/.env|*/.env.*|wrangler.json|wrangler.jsonc|wrangler.toml|*/wrangler.json|*/wrangler.jsonc|*/wrangler.toml)
      echo R3 ;;
    package.json|pnpm-lock.yaml|pnpm-workspace.yaml|docs/templates/*|docs/decisions/*|docs/product/*)
      echo R2 ;;
    *.md|*.txt|*.png|*.jpg|*.jpeg|*.gif|*.svg)
      echo R0 ;;
    *)
      echo R1 ;;
  esac
}

floor=R0
floor_rank=0
declare -a floor_paths=()
for path in "${files[@]}"; do
  risk="$(path_risk "$path")"
  rank="$(risk_rank "$risk")"
  printf '%s\t%s\n' "$risk" "$path"
  if ((rank > floor_rank)); then
    floor="$risk"
    floor_rank="$rank"
    floor_paths=("$path")
  elif ((rank == floor_rank)); then
    floor_paths+=("$path")
  fi
done

echo "Detected path-based risk floor: $floor"
echo "Paths establishing the floor:"
printf '  - %s\n' "${floor_paths[@]}"

declared=""
if [[ -n "$pr_body_file" && -r "$pr_body_file" ]]; then
  declared="$({
    grep -Eim1 '^[[:space:]]*Risk classification:[[:space:]]*R[0-4]([[:space:]]|$)' "$pr_body_file" || true
  } | sed -E 's/.*Risk classification:[[:space:]]*(R[0-4]).*/\1/I')"
fi

if [[ -z "$declared" ]]; then
  if [[ "$require_declaration" == true ]]; then
    echo "Missing declaration. Add a line such as 'Risk classification: $floor' to the pull-request body." >&2
    exit 1
  fi
  echo "No pull-request risk declaration supplied; reporting the path floor only."
else
  declared_rank="$(risk_rank "$declared")"
  echo "Declared risk: $declared"
  if ((declared_rank < floor_rank)); then
    echo "Declared risk $declared is below the detected floor $floor." >&2
    exit 1
  fi
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "risk_floor=$floor"
    echo "declared_risk=${declared:-unknown}"
  } >> "$GITHUB_OUTPUT"
fi

echo "Path classification passed. Semantic consequences may require escalation by the independent verifier."
