# VOC-112-T03 evidence — Graphify pilot (code-only, opt-in)

Draft carrier for implementation evidence. No graph dumps, secrets, or personal data.

## gate_status

complete — T03 implementation recorded 2026-08-23

## Pin table

| Component | Pin | Notes |
|-----------|-----|-------|
| Graphify upstream repository | `https://github.com/Graphify-Labs/graphify` | Reviewed Apache-2.0 upstream |
| Graphify upstream commit/tag | `b2cd36267456c166788c95be6e68574064a92a42` / `v0.9.48` | Tag verified via `git ls-remote` |
| Runtime package (`graphifyy`) | `0.9.48` | PyPI package name uses double-y |
| Transitive environment lock/hash identity | `scripts/graphify/requirements.lock` (`sha256:492f681c167b6d8da7b0970a9dec66e477470343424ec4eb34aa7e33a16e0a5a`) | Generated with `pip-compile --generate-hashes` |
| Runtime compatibility | Python `>=3.12`; 30 locked runtime distributions | The reviewed `numpy==2.5.2` pin requires Python 3.12 or newer |
| Per-skill provenance | `.agents/skills/graphify-pilot/PROVENANCE.yaml` | Adapted record covers the local skill and a transitively verified runtime manifest |
| Retained license/NOTICE | `LICENSE`, `LICENSE-MIT`, `NOTICE` under `.agents/skills/graphify-pilot/` | Byte-identical to upstream commit `b2cd362…`; NOTICE's historical MIT reference is complete |

## Pilot configuration checklist

| Check | Result | Notes |
|-------|--------|-------|
| Code-only / `--code-only` enforced | pass | `scripts/graphify/run.sh` invokes `graphify extract … --code-only` |
| Query logging disabled | pass | `GRAPHIFY_QUERY_LOG_DISABLE=1` and `GRAPHIFY_QUERY_LOG_ENABLE=0` in runner |
| `.graphifyignore` excludes secrets/generated/vendor | pass | Includes `.env*`, `node_modules/`, build output, `graphify-out/`, `.venv/` |
| Skill marked opt-in / auto-invocation disabled | pass | `disable-model-invocation: true` in canonical skill and adapter |
| Generated graph output gitignored by default | pass | `.gitignore` lists `graphify-out/` and `.graphify_python` |
| Runner fails safely without global install | pass | `scripts/graphify/check` exits non-zero when `.venv` absent |
| Ordinary use makes no download/upgrade/network setup action | pass | Only `setup.sh` installs; `check`/`run.sh` are offline |
| No provider auto-detection or provider credentials | pass | Runner starts with `env -i`, an isolated repository-local home, and only a minimal non-secret allowlist |
| No hooks, always-on injection, or user-profile mutation | pass | Scripts avoid `graphify install` / `hook install`; skill forbids them |
| Repository target boundary | pass | Realpath containment rejects targets outside the checkout, including symlink escapes |
| Locked repository-local runtime identity verified | pass | File digests, Graphify version, all 30 distribution name/version pins, and dependency consistency are checked |
| No unhashed setup upgrade/build fallback | pass | Setup uses `--require-hashes --only-binary=:all:` and does not upgrade bootstrap tooling |
| Runtime files covered by provenance | pass | `RUNTIME-MANIFEST.yaml` hashes the runner, setup/check, lock/identity, and ignore policy; foundation tests verify every entry |
| Hint-only language in skill | pass | SKILL.md states verify-in-source and navigation-hint posture |

## Validation commands

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc112-agent-skills.test.mjs` | pass |
| `node --test scripts/foundation/voc112-graphify.test.mjs` | pass (15 tests; includes a deterministic valid locked runtime and hermetic invocation fixture) |
| `bash scripts/graphify/check` | expected fail-closed locally when explicit operator setup has not been run; positive behavior is covered by the deterministic fixture |
| `git diff --check` | pass |

## Limitations (pilot scope)

- Automatic skill invocation remains disabled throughout VOC-112; graduation requires a separate governed change.
- Graph output under `graphify-out/` is a navigation hint only; agents must verify consequential claims in current source.
- Operator must run `bash scripts/graphify/setup.sh` explicitly; CI and ordinary agent sessions do not provision the runtime.
- Semantic document/image extraction and LLM-backed passes are out of scope (`--code-only` only).

## Acceptance mapping

- `VOC-112-AC-03` / `VOC-112-EV-03` — checklist passes and pin table filled.
