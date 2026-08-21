# VOC-102-T00 — Evidence

## evidence_id

VOC-102-EV-00

## gate_status

complete

## drafting-time diagnosis

Confirmed `karsift-ai-infra/.github/workflows/auto-advance.yml` previously set
`should_dispatch=true` for every open next roster task without reading
`<package_path>/.karsift/live-evidence/<next_task_id>.yaml` or the exact
`Automation ownership` marker in `tasks.md`.

## commands

```bash
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py'

cd ..
node --test scripts/foundation/voc102-auto-advance-ownership.test.mjs
pnpm validate
pnpm build

bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## results

### Infra policy tests

```
cd karsift-ai-infra && python3 -m unittest tests.test_auto_advance_ownership -v
```

Result: the initial 98/98 policy tests passed locally. Exact infra head
`e2822ea134fb12561c74c4532076beafe7a8a246` passed hosted self-CI run
`32478477231` (actionlint, shellcheck, YAML parse, and policy tests), then merged
through `KARSIFT/karsift-ai-infra#73` as
`8f87d0354602db207c1b67c4996a26f1e244c683`. Exact-SHA review then identified
two low-severity test-coverage gaps. The follow-up strengthened the last-task
no-op boundary and exercised full carrier re-entry at mocked external
boundaries. All 99 policy tests passed hosted self-CI run `32479420670`, and
`KARSIFT/karsift-ai-infra#74` merged as the final pinned revision
`e296e6bf8fc7cb300de986872605ee92f1a9cec0`. A second exact-SHA review found
two additional low-severity coverage gaps. Direct ordinary-stanza parsing and
fail-closed marker trust, sanitization, and deduplication tests brought the
suite to 100/100; hosted self-CI run `32481551754` passed and
`KARSIFT/karsift-ai-infra#75` merged as the final pinned revision
`29a76b99e9aa732096a5243433b21252585cfeeb`. The final review pass requested
full missing-file carrier repair coverage and use of the production marker
constant. All 101 policy tests passed hosted self-CI run `32482230478`, and
`KARSIFT/karsift-ai-infra#76` merged as the final pinned revision
`268be51d7781c499e36a4d35634c397ee0d2e4d9`.

### Calling-repo foundation tests

```
node --test scripts/foundation/voc102-auto-advance-ownership.test.mjs
```

Result: 5/5 tests passed.

### Calling-repo full validation

`pnpm validate` passed workspace validation, formatting, lint, type checks, all
226 foundation tests, API-client tests, and web middleware tests. Its local API
phase reached the two Docker-backed controlled-signup callback tests and stopped
because this WSL environment has no Docker command. No product assertion failed.
`pnpm build` then passed package, Next.js, and Go builds. The unmodified full
validation remains mandatory in hosted exact-SHA CI, where Docker is available.

### Governance validation

```
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Result: governance structure validation passed; detected path floor R4;
`git diff --check` clean.

## implementation summary

- `auto-advance.yml` now classifies next-task ownership before any path-specific
  dispatch decision (`implement`, `prepare-live-evidence`, `fail-closed`, or
  `none`).
- Operator/live-actions tasks use a separate clean App-scoped carrier publisher;
  the classifier remains read-only.
- The publisher derives the package-local evidence path from the task ID, repairs
  missing file/PR/marker states, rejects unexpected changed paths or untrusted
  carriers, and preserves evidence already recorded by an operator.
- Fail-closed metadata posts a sanitized issue marker and creates no carrier.
- Ordinary tasks retain the existing-PR guard and `implement.yml` attempt 1.
- Calling-repo `pipeline.yml` exposes read-only
  `verify-auto-advance-live-evidence` for T01 exact-head proof.
- Docs/template guidance define the exact task-stanza automation-ownership marker.

## notes

- Controlled live workflow proof remains `VOC-102-T01` (`VOC-102-EV-01`).
- The first automated attempt (pipeline `32476149809`) correctly refused to
  publish a protected workflow-file mutation. Its committed recovery artifact
  was reviewed and corrected through the supervised split above; the failed run
  was not blindly retried.
- Calling-repo runtime fixtures are byte-identical to the final infra merge
  `268be51d7781c499e36a4d35634c397ee0d2e4d9`; runtime consumption remains
  `KARSIFT/karsift-ai-infra/...@main`.
- No secrets, logs, artifacts, or unrelated package live evidence recorded here.
