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
python3 -m unittest tests.test_auto_advance_ownership -v

cd ..
node --test scripts/foundation/voc102-auto-advance-ownership.test.mjs

bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## results

### Infra policy tests

```
cd karsift-ai-infra && python3 -m unittest tests.test_auto_advance_ownership -v
```

Result: `Ran 11 tests in 0.009s` — OK.

### Calling-repo foundation tests

```
node --test scripts/foundation/voc102-auto-advance-ownership.test.mjs
```

Result: 5/5 tests passed.

### Governance validation

```
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Result: governance structure validation passed; detected path floor R3; `git diff --check` clean.

## implementation summary

- `auto-advance.yml` now classifies next-task ownership before any path-specific
  dispatch decision (`implement`, `prepare-live-evidence`, `fail-closed`, or
  `none`).
- Operator/live-actions tasks use a separate clean App-scoped carrier publisher;
  the classifier remains read-only.
- Fail-closed metadata posts a sanitized issue marker and creates no carrier.
- Ordinary tasks retain the existing-PR guard and `implement.yml` attempt 1.
- Calling-repo `pipeline.yml` exposes read-only
  `verify-auto-advance-live-evidence` for T01 exact-head proof.
- Docs/template guidance define the exact task-stanza automation-ownership marker.

## notes

- Controlled live workflow proof remains `VOC-102-T01` (`VOC-102-EV-01`).
- Infra changes land in the local `karsift-ai-infra/` checkout for the infra PR;
  this calling repo consumes reusable workflows at `@main` after merge.
- No secrets, logs, artifacts, or unrelated package live evidence recorded here.
