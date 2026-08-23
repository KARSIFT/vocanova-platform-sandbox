#!/usr/bin/env bash
# VOC-112-T03 — code-only Graphify extract wrapper (opt-in pilot).
# Requires a valid locked runtime from scripts/graphify/setup.sh.
# Never downloads packages, registers hooks, or enables provider auto-detection.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GRAPHIFY_HOME="${GRAPHIFY_HOME:-$SCRIPT_DIR}"
REPO_ROOT="$(cd "$GRAPHIFY_HOME/../.." && pwd)"

bash "$GRAPHIFY_HOME/check" >/dev/null

cli_relative="$(awk -F': ' '$1 == "graphify_cli_relative" {gsub(/^["'\'']|["'\'']$/, "", $2); print $2; exit}' "$GRAPHIFY_HOME/runtime-identity.yaml")"
GRAPHIFY_CLI="$GRAPHIFY_HOME/$cli_relative"

TARGET="${1:-$REPO_ROOT}"
OUTPUT_DIR="${GRAPHIFY_OUTPUT_DIR:-$REPO_ROOT/graphify-out}"

# Strip provider credentials so Graphify cannot auto-detect LLM backends.
unset OPENAI_API_KEY ANTHROPIC_API_KEY GEMINI_API_KEY GOOGLE_API_KEY \
  AZURE_OPENAI_API_KEY DEEPSEEK_API_KEY OPENROUTER_API_KEY MISTRAL_API_KEY \
  COHERE_API_KEY TOGETHER_API_KEY GROQ_API_KEY XAI_API_KEY

export GRAPHIFY_QUERY_LOG_DISABLE=1
export GRAPHIFY_QUERY_LOG_ENABLE=0
export GRAPHIFY_OUT="$OUTPUT_DIR"

exec "$GRAPHIFY_CLI" extract "$TARGET" --code-only
