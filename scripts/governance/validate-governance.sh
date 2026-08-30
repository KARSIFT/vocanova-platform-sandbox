#!/usr/bin/env bash
set -euo pipefail

base=""
head=""
files_from=""

usage() {
  echo "Usage: $0 [--base SHA --head SHA | --files-from FILE]" >&2
}

while (($#)); do
  case "$1" in
    --base) base="${2:-}"; shift 2 ;;
    --head) head="${2:-}"; shift 2 ;;
    --files-from) files_from="${2:-}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) usage; echo "Unknown argument: $1" >&2; exit 2 ;;
  esac
done

required_files=(
  AGENTS.md
  CLAUDE.md
  SECURITY.md
  .github/CODEOWNERS
  .github/pull_request_template.md
  .github/approved-policy/protected-paths.yaml
  .github/workflows/repository-governance.yml
  docs/governance/16-autonomous-development-operating-model.md
  docs/governance/amendments/A-002-governed-autonomous-releases.md
  docs/governance/amendments/A-003-governed-autonomous-engineering-authority.md
  docs/governance/a003-transition-state.yaml
  docs/governance/approval-matrix.md
  docs/governance/change-risk-classification.md
  docs/governance/protected-areas.md
  docs/governance/post-merge-activation-checklist.md
  docs/governance/repository-settings.md
  docs/governance/technical-steward-appointment.md
  docs/architecture/17-autonomous-development-architecture.md
  docs/planning/18-autonomous-development-implementation-roadmap.md
  docs/templates/change-specification.md
  docs/templates/acceptance-criteria.md
  docs/templates/founder-decision-card.md
  docs/templates/technical-approval-request.md
  docs/templates/verification-report.md
  docs/templates/release-record.md
  docs/templates/rollback-report.md
  specs/README.md
  specs/templates/change-package/change.yaml
  specs/changes/VOC-001-repository-foundation/change.yaml
  specs/changes/VOC-002-a003-governance-transition/change.yaml
  specs/changes/VOC-003-a003-lifecycle-sync/change.yaml
  specs/changes/VOC-004-canonical-adoption-doc-17-doc-18/change.yaml
  tooling/governance/validate_repository_foundation.py
  tooling/governance/tests/test_validate_repository_foundation.py
)

for file in "${required_files[@]}"; do
  [[ -f "$file" ]] || { echo "Missing required governance file: $file" >&2; exit 1; }
  [[ -s "$file" ]] || { echo "Empty governance file: $file" >&2; exit 1; }
done

if [[ -e CODEOWNERS ]]; then
  echo "CODEOWNERS must live only at .github/CODEOWNERS to avoid conflicting policies." >&2
  exit 1
fi

required_pr_fields=(
  "Risk classification:"
  "Affected protected areas"
  "Acceptance-criteria evidence"
  "Commands executed and results"
  "Preview deployment"
  "Security and privacy"
  "Migration/data integrity"
  "Rollback trigger"
  "Analytics/telemetry"
  "Documentation"
  "Independent-verifier report/result"
  "Required approval class"
)

for field in "${required_pr_fields[@]}"; do
  grep -Fq "$field" .github/pull_request_template.md || {
    echo "Pull-request template is missing required field: $field" >&2
    exit 1
  }
done

amendment=docs/governance/amendments/A-002-governed-autonomous-releases.md
operating_model=docs/governance/16-autonomous-development-operating-model.md
appointment=docs/governance/technical-steward-appointment.md
activation=docs/governance/post-merge-activation-checklist.md

grep -Fqx "status: approved" "$operating_model" || {
  echo "DOC-16 must have status: approved." >&2
  exit 1
}
grep -Fqx "approved_at: 2026-07-13" "$operating_model" || {
  echo "DOC-16 must record approved_at: 2026-07-13." >&2
  exit 1
}
grep -Fqx "status: approved" "$amendment" || {
  echo "A-002 must have status: approved." >&2
  exit 1
}
grep -Fqx "approved_at: 2026-07-13" "$amendment" || {
  echo "A-002 must record approved_at: 2026-07-13." >&2
  exit 1
}
if grep -Fq "approval_evidence: pending-github-pull-request" "$operating_model" "$amendment"; then
  echo "Canonical governance contains stale pending approval evidence." >&2
  exit 1
fi

grep -Fq 'Founder GitHub identity: `@m-e-h-r-d-a-a-d`' "$appointment" || {
  echo "The appointment must identify the founder GitHub account." >&2
  exit 1
}
grep -Fq 'Appointed qualified human technical steward: `@m-e-h-r-d-a-a-d`' "$appointment" || {
  echo "The appointment must identify the qualified human technical steward." >&2
  exit 1
}
grep -Fqi "dual-role human appointment" "$appointment" || {
  echo "The appointment must record its dual-role human nature." >&2
  exit 1
}
grep -Fqi "both capacities" "$appointment" || {
  echo "The appointment must define combined approval evidence." >&2
  exit 1
}
if grep -Eqi '^[[:space:]-]*(appointed )?(qualified human )?technical[- ]steward:[[:space:]]*`?@?[^[:space:]]*(claude|codex|bot|automation)' "$appointment"; then
  echo "An AI or automation identity must not be appointed technical steward." >&2
  exit 1
fi

if grep -Eqi 'no qualified (human )?technical[- ]steward( identity or team)? exists' .github/CODEOWNERS; then
  echo "CODEOWNERS contains stale commentary saying no steward exists." >&2
  exit 1
fi

grep -Fq "RL1/RL2 technical activation remain disabled" "$activation" || {
  echo "The activation checklist must keep RL1/RL2 technical activation disabled." >&2
  exit 1
}
grep -Fq "A-004 is the active engineering-workflow authority model" "$activation" || {
  echo "The activation checklist must document active A-004 authority." >&2
  exit 1
}
grep -Fq "repository-controlled" "$activation" || {
  echo "The activation checklist must document the repository-controlled release/deploy path." >&2
  exit 1
}
if grep -Eqi '^[[:space:]]*Status:[[:space:]]*Activated|^- \[x\].*(autonomous|production release).*(enable|activat)' "$activation"; then
  echo "The activation checklist must not claim autonomous release activation." >&2
  exit 1
fi

grep -Fq "Low-risk, reversible R0-R1 production releases may merge" "$amendment"
grep -Fq "technical steward" "$amendment"
grep -Fq "require founder approval" "$amendment"
grep -Fq "initial public launch" "$amendment"
grep -Fq "Initial governance bootstrap adoption" "$operating_model"
grep -Fq "Initial adoption exception" "$amendment"
grep -Fq "historical initial DOC-16/A-002 bootstrap" docs/governance/approval-matrix.md
grep -Fq "R3 production changes remain" docs/governance/post-merge-activation-checklist.md

if grep -Eq 'FOUNDER_GITHUB_USERNAME|TECHNICAL_STEWARD_GITHUB_USERNAME' .github/CODEOWNERS; then
  echo "CODEOWNERS contains an unverifiable identity placeholder." >&2
  exit 1
fi

if ! grep -Ev '^[[:space:]]*#' .github/CODEOWNERS | grep -Eq '@[A-Za-z0-9]'; then
  echo "CODEOWNERS contains no configured identity." >&2
  exit 1
fi

if grep -Ev '^[[:space:]]*#' .github/CODEOWNERS | grep -Eqi '@[^[:space:]]*(claude|codex|bot)([[:space:]]|$)'; then
  echo "A bot-looking identity is configured as a CODEOWNER; human role assignment must be verified." >&2
  exit 1
fi

protected_tokens=(auth database billing ai audio voice storage migrations infra backups deploy release rollback)
for token in "${protected_tokens[@]}"; do
  grep -Fiq "$token" .github/CODEOWNERS || {
    echo "CODEOWNERS is missing protected-area token: $token" >&2
    exit 1
  }
  grep -Fiq "$token" scripts/governance/classify-change-risk.sh || {
    echo "Risk classifier is missing protected-area token: $token" >&2
    exit 1
  }
done

r4_ruleset_paths=(
  /.github/CODEOWNERS
  /.github/workflows/governance-policy.yml
  /scripts/governance/
  /docs/operations/15-ai-native-product-and-engineering-operating-model.md
  /docs/governance/approval-matrix.md
  /docs/governance/change-risk-classification.md
  /docs/governance/protected-areas.md
  /docs/governance/post-merge-activation-checklist.md
  /docs/governance/amendments/
  /docs/governance/a003-transition-state.yaml
  /docs/governance/16-autonomous-development-operating-model.md
  /docs/architecture/17-autonomous-development-architecture.md
  /docs/planning/18-autonomous-development-implementation-roadmap.md
  /specs/changes/VOC-002-a003-governance-transition/
  /specs/changes/VOC-003-a003-lifecycle-sync/
  /specs/changes/VOC-004-canonical-adoption-doc-17-doc-18/
)

for path in "${r4_ruleset_paths[@]}"; do
  grep -Fq "$path" docs/governance/repository-settings.md || {
    echo "Repository settings are missing fixed R4 ruleset path: $path" >&2
    exit 1
  }
  grep -Fq "${path#/}" scripts/governance/classify-change-risk.sh || {
    echo "Risk classifier is missing fixed R4 path family: ${path#/}" >&2
    exit 1
  }
done

bash -n scripts/governance/classify-change-risk.sh
bash -n scripts/governance/validate-monitoring-impact.sh
python3 tooling/governance/validate_repository_foundation.py --repository-root .
python3 tooling/governance/validate_python_bytecode_hygiene.py --repository-root .

monitoring_impact_args=()
if [[ -n "$files_from" ]]; then
  monitoring_impact_args+=(--files-from "$files_from")
elif [[ -n "$base" && -n "$head" ]]; then
  monitoring_impact_args+=(--base "$base" --head "$head")
fi

bash scripts/governance/validate-monitoring-impact.sh --declarations-only
bash scripts/governance/validate-monitoring-impact.sh "${monitoring_impact_args[@]}"
echo "Governance structure validation passed."
