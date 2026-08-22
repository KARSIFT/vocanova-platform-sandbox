# VOC-107-T00 — Evidence

## evidence_id

VOC-107-EV-00

## gate_status

complete

## drafting-time diagnosis

Confirmed `karsift-ai-infra/.github/workflows/implement.yml` creates
`base_sha..HEAD` bundles and that attempt 2+ sets `base_sha` to `HEAD` after
checkout/rebase of the existing agent branch. The isolated publish job fetches
only `integration_branch` before `git bundle verify`. Issue #891 records run
`32539352323` where a valid remediation commit was rejected before publication
because that thin-bundle prerequisite was absent from the clean bare repository.

## commands

```bash
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py'
git diff --check

cd vocanova-platform-sandbox
pnpm validate
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## results

- Shared-infrastructure PR `KARSIFT/karsift-ai-infra#100` merged at
  `3b968d89b1a958db1eaa7d20748549f5c3d3658e`.
- Hosted shared-infrastructure `actionlint`, `shellcheck`, YAML parsing, and
  policy-test jobs passed in run `32544512478`; the local suite passed all 154
  tests.
- The implementer now records `integration_sha` separately from the pre-model
  `base_sha`. Soft reset continues to use `base_sha`; bundle creation,
  publisher ancestry, and the denied-workflow scan use the integration anchor.
- The clean publisher still imports the declared exact head and retains the
  SHA-valued attempt-2 force-with-lease. It does not fetch or trust the existing
  task branch before validating the artifact.
- The real-Git fixture passed attempt 1, attempt 2 after rebasing prior task
  commits to new SHAs, and clean bare verify/import. The old incomplete
  pre-model-tip bundle fails verification. A workflow path added and later
  removed within the task lineage is still denied because every commit path is
  scanned.
- The calling repository consumes
  `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml@main`; no pin change
  is required.
- The plan-review wording ambiguity in TEST-02 and the AC-04/TEST-07 mapping
  drift were corrected without changing authorized scope.

- Local `pnpm validate` passed workspace validation, formatting, lint, type
  checks, 235 foundation tests, 28 API-client tests, 16 web middleware/lib
  tests, and all non-container Go packages reached by the API suite. It stopped
  only because this WSL environment has no Docker command for the two
  disposable-Postgres controlled-signup tests; hosted exact-SHA CI remains the
  required acceptance evidence for those tests and the remaining builds.
- `validate-governance.sh`: **passed**.
- `classify-change-risk.sh`: **passed**, with an R0 path floor for the three
  package evidence/wording files (the implemented upstream workflow remains
  semantically R3).
- `git diff --check`: **passed**.

## privacy

Evidence must remain allowlisted metadata only (commands, pass/fail, SHAs, run
IDs). No logs, artifacts, secrets, or user identifiers.
